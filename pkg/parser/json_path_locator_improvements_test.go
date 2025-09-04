package parser

import (
	"strings"
	"testing"
)

func TestExtractAdditionalPropertyNames(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		expected     []string
	}{
		{
			name:         "single additional property",
			errorMessage: "at '': additional properties 'invalid_key' not allowed",
			expected:     []string{"invalid_key"},
		},
		{
			name:         "multiple additional properties",
			errorMessage: "at '': additional properties 'invalid_prop', 'another_invalid' not allowed",
			expected:     []string{"invalid_prop", "another_invalid"},
		},
		{
			name:         "single property with different format",
			errorMessage: "additional property 'bad_field' not allowed",
			expected:     []string{"bad_field"},
		},
		{
			name:         "no additional properties in message",
			errorMessage: "at '/age': got string, want number",
			expected:     []string{},
		},
		{
			name:         "empty message",
			errorMessage: "",
			expected:     []string{},
		},
		{
			name:         "complex property names",
			errorMessage: "additional properties 'invalid-prop', 'another_bad_one', 'third.prop' not allowed",
			expected:     []string{"invalid-prop", "another_bad_one", "third.prop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAdditionalPropertyNames(tt.errorMessage)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d properties, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected property %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestFindFirstAdditionalProperty(t *testing.T) {
	yamlContent := `name: John Doe
age: 30
invalid_prop: value
tools:
  - name: tool1
another_bad: value2
permissions:
  read: true
  invalid_perm: write`

	tests := []struct {
		name          string
		propertyNames []string
		expectedLine  int
		expectedCol   int
		shouldFind    bool
	}{
		{
			name:          "find first property",
			propertyNames: []string{"invalid_prop", "another_bad"},
			expectedLine:  3,
			expectedCol:   1,
			shouldFind:    true,
		},
		{
			name:          "find second property when first not found",
			propertyNames: []string{"not_exist", "another_bad"},
			expectedLine:  6,
			expectedCol:   1,
			shouldFind:    true,
		},
		{
			name:          "property not found",
			propertyNames: []string{"nonexistent", "also_missing"},
			expectedLine:  1,
			expectedCol:   1,
			shouldFind:    false,
		},
		{
			name:          "nested property found",
			propertyNames: []string{"invalid_perm"},
			expectedLine:  9,
			expectedCol:   3, // Indented
			shouldFind:    true,
		},
		{
			name:          "empty property list",
			propertyNames: []string{},
			expectedLine:  1,
			expectedCol:   1,
			shouldFind:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := findFirstAdditionalProperty(yamlContent, tt.propertyNames)

			if location.Found != tt.shouldFind {
				t.Errorf("Expected Found=%v, got Found=%v", tt.shouldFind, location.Found)
			}

			if location.Line != tt.expectedLine {
				t.Errorf("Expected Line=%d, got Line=%d", tt.expectedLine, location.Line)
			}

			if location.Column != tt.expectedCol {
				t.Errorf("Expected Column=%d, got Column=%d", tt.expectedCol, location.Column)
			}
		})
	}
}

func TestLocateJSONPathInYAMLWithAdditionalProperties(t *testing.T) {
	yamlContent := `name: John
age: 25
invalid_prop: value
another_invalid: value2`

	tests := []struct {
		name         string
		jsonPath     string
		errorMessage string
		expectedLine int
		expectedCol  int
		shouldFind   bool
	}{
		{
			name:         "empty path with additional properties",
			jsonPath:     "",
			errorMessage: "at '': additional properties 'invalid_prop', 'another_invalid' not allowed",
			expectedLine: 3,
			expectedCol:  1,
			shouldFind:   true,
		},
		{
			name:         "empty path with single additional property",
			jsonPath:     "",
			errorMessage: "at '': additional properties 'another_invalid' not allowed",
			expectedLine: 4,
			expectedCol:  1,
			shouldFind:   true,
		},
		{
			name:         "empty path without additional properties message",
			jsonPath:     "",
			errorMessage: "some other error",
			expectedLine: 1,
			expectedCol:  1,
			shouldFind:   true,
		},
		{
			name:         "non-empty path should use regular logic",
			jsonPath:     "/name",
			errorMessage: "any message",
			expectedLine: 1,
			expectedCol:  6, // After "name:"
			shouldFind:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := LocateJSONPathInYAMLWithAdditionalProperties(yamlContent, tt.jsonPath, tt.errorMessage)

			if location.Found != tt.shouldFind {
				t.Errorf("Expected Found=%v, got Found=%v", tt.shouldFind, location.Found)
			}

			if location.Line != tt.expectedLine {
				t.Errorf("Expected Line=%d, got Line=%d", tt.expectedLine, location.Line)
			}

			if location.Column != tt.expectedCol {
				t.Errorf("Expected Column=%d, got Column=%d", tt.expectedCol, location.Column)
			}
		})
	}
}

// TestLocateJSONPathInYAMLWithAdditionalPropertiesNested tests the new functionality for nested additional properties
func TestLocateJSONPathInYAMLWithAdditionalPropertiesNested(t *testing.T) {
	yamlContent := `name: Test Workflow
on:
  push: 
    branches: [main]
  foobar: invalid
permissions:
  contents: read
  invalid_perm: write
nested:
  deeply:
    more_nested: true
    bad_prop: invalid`

	tests := []struct {
		name         string
		jsonPath     string
		errorMessage string
		expectedLine int
		expectedCol  int
		shouldFind   bool
	}{
		{
			name:         "nested additional property under 'on'",
			jsonPath:     "/on",
			errorMessage: "at '/on': additional properties 'foobar' not allowed",
			expectedLine: 5,
			expectedCol:  3, // Position of 'foobar'
			shouldFind:   true,
		},
		{
			name:         "nested additional property under 'permissions'",
			jsonPath:     "/permissions",
			errorMessage: "at '/permissions': additional properties 'invalid_perm' not allowed",
			expectedLine: 8,
			expectedCol:  3, // Position of 'invalid_perm'
			shouldFind:   true,
		},
		{
			name:         "deeply nested additional property",
			jsonPath:     "/nested/deeply",
			errorMessage: "at '/nested/deeply': additional properties 'bad_prop' not allowed",
			expectedLine: 12,
			expectedCol:  5, // Position of 'bad_prop' (indented further)
			shouldFind:   true,
		},
		{
			name:         "multiple additional properties - should find first",
			jsonPath:     "/on",
			errorMessage: "at '/on': additional properties 'foobar', 'another_prop' not allowed",
			expectedLine: 5,
			expectedCol:  3, // Position of 'foobar' (first one found)
			shouldFind:   true,
		},
		{
			name:         "non-existent path with additional properties",
			jsonPath:     "/nonexistent",
			errorMessage: "at '/nonexistent': additional properties 'some_prop' not allowed",
			expectedLine: 1, // Falls back to global search, which won't find 'some_prop'
			expectedCol:  1,
			shouldFind:   false,
		},
		{
			name:         "nested path without additional properties error",
			jsonPath:     "/on/push",
			errorMessage: "at '/on/push': some other validation error",
			expectedLine: 3, // Should find the 'push' key location using regular logic
			expectedCol:  8, // After "push:"
			shouldFind:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := LocateJSONPathInYAMLWithAdditionalProperties(yamlContent, tt.jsonPath, tt.errorMessage)

			if location.Found != tt.shouldFind {
				t.Errorf("Expected Found=%v, got Found=%v", tt.shouldFind, location.Found)
			}

			if location.Line != tt.expectedLine {
				t.Errorf("Expected Line=%d, got Line=%d", tt.expectedLine, location.Line)
			}

			if location.Column != tt.expectedCol {
				t.Errorf("Expected Column=%d, got Column=%d", tt.expectedCol, location.Column)
			}
		})
	}
}

func TestFindFrontmatterBounds(t *testing.T) {
	tests := []struct {
		name                     string
		lines                    []string
		expectedStartIdx         int
		expectedEndIdx           int
		expectedFrontmatterLines int
	}{
		{
			name: "normal frontmatter",
			lines: []string{
				"---",
				"name: test",
				"age: 30",
				"---",
				"# Markdown content",
			},
			expectedStartIdx:         0,
			expectedEndIdx:           3,
			expectedFrontmatterLines: 2,
		},
		{
			name: "frontmatter with comments before",
			lines: []string{
				"# Comment at top",
				"",
				"---",
				"name: test",
				"---",
				"Content",
			},
			expectedStartIdx:         2,
			expectedEndIdx:           4,
			expectedFrontmatterLines: 1,
		},
		{
			name: "no frontmatter",
			lines: []string{
				"# Just a markdown file",
				"Some content",
			},
			expectedStartIdx:         -1,
			expectedEndIdx:           -1,
			expectedFrontmatterLines: 0,
		},
		{
			name: "incomplete frontmatter (no closing)",
			lines: []string{
				"---",
				"name: test",
				"Some content without closing",
			},
			expectedStartIdx:         -1,
			expectedEndIdx:           -1,
			expectedFrontmatterLines: 0,
		},
		{
			name: "empty frontmatter",
			lines: []string{
				"---",
				"---",
				"Content",
			},
			expectedStartIdx:         0,
			expectedEndIdx:           1,
			expectedFrontmatterLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startIdx, endIdx, frontmatterContent := findFrontmatterBounds(tt.lines)

			if startIdx != tt.expectedStartIdx {
				t.Errorf("Expected startIdx=%d, got startIdx=%d", tt.expectedStartIdx, startIdx)
			}

			if endIdx != tt.expectedEndIdx {
				t.Errorf("Expected endIdx=%d, got endIdx=%d", tt.expectedEndIdx, endIdx)
			}

			// Count the lines in frontmatterContent
			actualLines := 0
			if frontmatterContent != "" {
				actualLines = len(strings.Split(frontmatterContent, "\n"))
			}

			if actualLines != tt.expectedFrontmatterLines {
				t.Errorf("Expected %d frontmatter lines, got %d", tt.expectedFrontmatterLines, actualLines)
			}
		})
	}
}
