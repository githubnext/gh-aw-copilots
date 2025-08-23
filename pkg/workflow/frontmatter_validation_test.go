package workflow

import (
	"strings"
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

func TestFrontmatterValidator(t *testing.T) {
	tests := []struct {
		name               string
		frontmatterYAML    string
		frontmatter        map[string]any
		expectedErrorCount int
		expectedErrors     []string // Substring matches for error messages
		expectedPaths      []string // Paths that should have errors
	}{
		{
			name: "valid frontmatter",
			frontmatterYAML: `engine: claude
on: push
max-turns: 5`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": 5,
			},
			expectedErrorCount: 0,
		},
		{
			name: "missing required 'on' field",
			frontmatterYAML: `engine: claude
max-turns: 5`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"max-turns": 5,
			},
			expectedErrorCount: 1,
			expectedErrors:     []string{"missing required field 'on'"},
			expectedPaths:      []string{"on"},
		},
		{
			name: "invalid engine",
			frontmatterYAML: `engine: invalid-engine
on: push`,
			frontmatter: map[string]any{
				"engine": "invalid-engine",
				"on":     "push",
			},
			expectedErrorCount: 1,
			expectedErrors:     []string{"got string, want object"}, // JSON schema message
			expectedPaths:      []string{"engine"},
		},
		{
			name: "invalid max-turns value",
			frontmatterYAML: `engine: claude
on: push
max-turns: 150`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": 150,
			},
			expectedErrorCount: 1,
			expectedErrors:     []string{"max-turns must be between 1 and 100"},
			expectedPaths:      []string{"max-turns"},
		},
		{
			name: "tools with missing name",
			frontmatterYAML: `engine: claude
on: push
tools:
  - type: shell
  - name: git
    type: shell`,
			frontmatter: map[string]any{
				"engine": "claude",
				"on":     "push",
				"tools": []any{
					map[string]any{"type": "shell"},
					map[string]any{"name": "git", "type": "shell"},
				},
			},
			expectedErrorCount: 1,
			expectedErrors:     []string{"tool must have a 'name' field"},
			expectedPaths:      []string{"tools[0].name"},
		},
		{
			name: "multiple validation errors",
			frontmatterYAML: `engine: invalid
max-turns: 0`,
			frontmatter: map[string]any{
				"engine":    "invalid",
				"max-turns": 0,
			},
			expectedErrorCount: 3, // missing 'on', invalid engine, invalid max-turns
			expectedErrors:     []string{"missing required field 'on'", "got string, want object", "max-turns must be between 1 and 100"},
			expectedPaths:      []string{"on", "engine", "max-turns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFrontmatterValidator(tt.frontmatterYAML)
			errors := validator.ValidateFrontmatter(tt.frontmatter)

			if len(errors) != tt.expectedErrorCount {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrorCount, len(errors))
			}

			// Check that expected error messages are present
			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Message, expectedError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s' not found", expectedError)
				}
			}

			// Check that expected paths have errors
			for _, expectedPath := range tt.expectedPaths {
				found := false
				for _, err := range errors {
					if err.Path == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error at path '%s' not found", expectedPath)
				}
			}
		})
	}
}

func TestValidationErrorSpans(t *testing.T) {
	frontmatterYAML := `engine: invalid-engine
on: push
max-turns: 150
tools:
  - name: git
    type: shell
  - type: shell`

	frontmatter := map[string]any{
		"engine":    "invalid-engine",
		"on":        "push",
		"max-turns": 150,
		"tools": []any{
			map[string]any{"name": "git", "type": "shell"},
			map[string]any{"type": "shell"},
		},
	}

	validator := NewFrontmatterValidator(frontmatterYAML)
	errors := validator.ValidateFrontmatter(frontmatter)

	// Should have errors for engine, max-turns, and tools[1].name
	if len(errors) != 3 {
		t.Fatalf("Expected 3 errors, got %d", len(errors))
	}

	// Check that span information is available for fields that exist
	spanCount := 0
	for _, err := range errors {
		if err.Span != nil {
			spanCount++
			// Verify that spans have reasonable values
			if err.Span.StartLine <= 0 || err.Span.StartColumn <= 0 {
				t.Errorf("Invalid span for error at path '%s': %+v", err.Path, err.Span)
			}
		}
	}

	// We expect spans for engine and max-turns (existing fields)
	// tools[1].name might not have a span since the field is missing
	if spanCount < 2 {
		t.Errorf("Expected at least 2 errors with spans, got %d", spanCount)
	}
}

func TestConvertValidationErrorsToCompilerErrors(t *testing.T) {
	validationErrors := []FrontmatterValidationError{
		{
			Path:    "engine",
			Message: "unsupported engine 'invalid'",
			Span: &parser.SourceSpan{
				StartLine:   1,
				StartColumn: 9,
				EndLine:     1,
				EndColumn:   15,
			},
		},
		{
			Path:    "on",
			Message: "missing required field 'on'",
			Span:    nil, // No span for missing field
		},
	}

	filePath := "test.md"
	frontmatterStart := 2 // Frontmatter starts at line 2 in the file

	compilerErrors := ConvertValidationErrorsToCompilerErrors(filePath, frontmatterStart, validationErrors)

	if len(compilerErrors) != 2 {
		t.Fatalf("Expected 2 compiler errors, got %d", len(compilerErrors))
	}

	// Check first error (with span)
	err1 := compilerErrors[0]
	if err1.Position.File != filePath {
		t.Errorf("Expected file '%s', got '%s'", filePath, err1.Position.File)
	}
	if err1.Position.Line != 2 { // 1 + 2 - 1 = 2 (adjusted for frontmatter position)
		t.Errorf("Expected line 2, got %d", err1.Position.Line)
	}
	if !err1.Position.IsSpan() {
		t.Error("Expected first error to have span information")
	}

	// Check second error (no span)
	err2 := compilerErrors[1]
	if err2.Position.IsSpan() {
		t.Error("Expected second error to not have span information")
	}
}

func TestGenerateHintForValidationError(t *testing.T) {
	tests := []struct {
		name         string
		err          FrontmatterValidationError
		expectedHint string
	}{
		{
			name: "engine error",
			err: FrontmatterValidationError{
				Path:    "engine",
				Message: "unsupported engine",
			},
			expectedHint: "Supported engines: claude, codex",
		},
		{
			name: "max-turns error",
			err: FrontmatterValidationError{
				Path:    "max-turns",
				Message: "invalid value",
			},
			expectedHint: "max-turns should be a number between 1 and 100",
		},
		{
			name: "tools name error",
			err: FrontmatterValidationError{
				Path:    "tools[0].name",
				Message: "tool must have a 'name' field",
			},
			expectedHint: "Each tool must have a 'name' field specifying the tool identifier",
		},
		{
			name: "missing on field",
			err: FrontmatterValidationError{
				Path:    "on",
				Message: "missing required field",
			},
			expectedHint: "Add an 'on' field to specify when the workflow should run (e.g., 'on: push')",
		},
		{
			name: "unknown error",
			err: FrontmatterValidationError{
				Path:    "unknown",
				Message: "unknown error",
			},
			expectedHint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := generateHintForValidationError(tt.err)
			if hint != tt.expectedHint {
				t.Errorf("Expected hint '%s', got '%s'", tt.expectedHint, hint)
			}
		})
	}
}

func TestIsValidEngine(t *testing.T) {
	tests := []struct {
		engine string
		valid  bool
	}{
		{"claude", true},
		{"codex", true},
		{"invalid", false},
		{"", false},
		{"Claude", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			if isValidEngine(tt.engine) != tt.valid {
				t.Errorf("isValidEngine(%q) = %v, want %v", tt.engine, !tt.valid, tt.valid)
			}
		})
	}
}
