package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestLocateFrontmatterPathSpan(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		path         string
		expectedSpan SourceSpan
		wantErr      bool
		errContains  string
	}{
		{
			name: "simple scalar value",
			yaml: `title: Test Workflow
on: push`,
			path: "title",
			expectedSpan: SourceSpan{
				StartLine:   1,
				StartColumn: 8,
				EndLine:     1,
				EndColumn:   20,
			},
		},
		{
			name: "nested object key",
			yaml: `on:
  push:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest`,
			path: "jobs.build.runs-on",
			expectedSpan: SourceSpan{
				StartLine:   6,
				StartColumn: 14,
				EndLine:     6,
				EndColumn:   26,
			},
		},
		{
			name: "sequence element",
			yaml: `on:
  push:
    branches:
      - main
      - develop
      - feature/*`,
			path: "on.push.branches[1]",
			expectedSpan: SourceSpan{
				StartLine:   5,
				StartColumn: 9,
				EndLine:     5,
				EndColumn:   15,
			},
		},
		{
			name: "sequence element (flow style)",
			yaml: `on:
  push:
    branches: [main, develop, "feature/*"]`,
			path: "on.push.branches[2]",
			expectedSpan: SourceSpan{
				StartLine:   3,
				StartColumn: 31,
				EndLine:     3,
				EndColumn:   39,
			},
		},
		{
			name: "whole mapping span",
			yaml: `jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2`,
			path: "jobs.build",
			expectedSpan: SourceSpan{
				StartLine:   3,
				StartColumn: 12,
				EndLine:     5,
				EndColumn:   33,
			},
		},
		{
			name: "multi-line literal scalar (pipe style)",
			yaml: `description: |
  This is a multi-line
  description that spans
  several lines
name: Test`,
			path: "description",
			expectedSpan: SourceSpan{
				StartLine:   1,
				StartColumn: 14,
				EndLine:     3,
				EndColumn:   13,
			},
		},
		{
			name: "multi-line literal scalar (fold style)",
			yaml: `description: >
  This is a folded
  multi-line description
  that will be folded
name: Test`,
			path: "description",
			expectedSpan: SourceSpan{
				StartLine:   1,
				StartColumn: 14,
				EndLine:     1,
				EndColumn:   14,
			},
		},
		{
			name: "path not found",
			yaml: `title: Test
on: push`,
			path: "nonexistent.key",
			wantErr:     true,
			errContains: "path not found",
		},
		{
			name: "array index out of range",
			yaml: `branches:
  - main
  - develop`,
			path: "branches[5]",
			wantErr:     true,
			errContains: "array index 5 out of range",
		},
		{
			name: "invalid array index",
			yaml: `branches:
  - main`,
			path: "branches[invalid]",
			wantErr:     true,
			errContains: "invalid array index",
		},
		{
			name: "empty yaml",
			yaml: "",
			path: "any.path",
			wantErr:     true,
			errContains: "frontmatter YAML is empty",
		},
		{
			name: "empty path",
			yaml: `title: test`,
			path: "",
			wantErr:     true,
			errContains: "JSONPath is empty",
		},
		{
			name: "nested sequence with objects",
			yaml: `jobs:
  build:
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Setup
        run: echo "setup"`,
			path: "jobs.build.steps[0].uses",
			expectedSpan: SourceSpan{
				StartLine:   5,
				StartColumn: 15,
				EndLine:     5,
				EndColumn:   33,
			},
		},
		{
			name: "anchor and alias (anchor definition)",
			yaml: `default: &default
  runs-on: ubuntu-latest
  timeout-minutes: 30

jobs:
  build:
    <<: *default
    steps:
      - run: echo "test"`,
			path: "default.runs-on",
			expectedSpan: SourceSpan{
				StartLine:   2,
				StartColumn: 12,
				EndLine:     2,
				EndColumn:   24,
			},
		},
		{
			name: "complex nested structure",
			yaml: `on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'staging'
        type: choice
        options:
          - staging
          - production`,
			path: "on.workflow_dispatch.inputs.environment.options[1]",
			expectedSpan: SourceSpan{
				StartLine:   11,
				StartColumn: 13,
				EndLine:     11,
				EndColumn:   22,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span, err := LocateFrontmatterPathSpan(tt.yaml, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LocateFrontmatterPathSpan() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LocateFrontmatterPathSpan() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("LocateFrontmatterPathSpan() error = %v", err)
				return
			}

			if span != tt.expectedSpan {
				t.Errorf("LocateFrontmatterPathSpan() = %+v, want %+v", span, tt.expectedSpan)
			}
		})
	}
}

func TestLocateFrontmatterPath(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		path         string
		expectedLine int
		expectedCol  int
		wantErr      bool
	}{
		{
			name: "simple scalar - legacy compatibility",
			yaml: `title: Test Workflow
on: push`,
			path:         "title",
			expectedLine: 1,
			expectedCol:  8,
		},
		{
			name: "nested object - legacy compatibility",
			yaml: `jobs:
  build:
    runs-on: ubuntu-latest`,
			path:         "jobs.build.runs-on",
			expectedLine: 3,
			expectedCol:  14,
		},
		{
			name: "error case - legacy compatibility",
			yaml: `title: Test`,
			path: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col, err := LocateFrontmatterPath(tt.yaml, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LocateFrontmatterPath() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("LocateFrontmatterPath() error = %v", err)
				return
			}

			if line != tt.expectedLine || col != tt.expectedCol {
				t.Errorf("LocateFrontmatterPath() = (%d, %d), want (%d, %d)", line, col, tt.expectedLine, tt.expectedCol)
			}
		})
	}
}

func TestNormalizeJSONPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple dot notation",
			path:     "on.push",
			expected: []string{"on", "push"},
		},
		{
			name:     "with leading dollar",
			path:     "$.on.push",
			expected: []string{"on", "push"},
		},
		{
			name:     "with array index",
			path:     "jobs.build.steps[0].run",
			expected: []string{"jobs", "build", "steps", "[0]", "run"},
		},
		{
			name:     "multiple array indices",
			path:     "matrix[0].include[1].os",
			expected: []string{"matrix", "[0]", "include", "[1]", "os"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "just dollar",
			path:     "$",
			expected: []string{},
		},
		{
			name:     "single key",
			path:     "title",
			expected: []string{"title"},
		},
		{
			name:     "array only",
			path:     "[0]",
			expected: []string{"[0]"},
		},
		{
			name:     "complex path",
			path:     "on.workflow_dispatch.inputs.environment.options[1]",
			expected: []string{"on", "workflow_dispatch", "inputs", "environment", "options", "[1]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeJSONPath(tt.path)
			
			if len(result) != len(tt.expected) {
				t.Errorf("normalizeJSONPath(%q) = %v, want %v", tt.path, result, tt.expected)
				return
			}
			
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("normalizeJSONPath(%q) = %v, want %v", tt.path, result, tt.expected)
					break
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("malformed yaml", func(t *testing.T) {
		yaml := `title: Test
  invalid: yaml: syntax`
		_, err := LocateFrontmatterPathSpan(yaml, "title")
		if err == nil {
			t.Error("Expected error for malformed YAML, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse YAML") {
			t.Errorf("Expected 'failed to parse YAML' error, got: %v", err)
		}
	})

	t.Run("accessing non-mapping as mapping", func(t *testing.T) {
		yaml := `title: "Simple String"
on: push`
		_, err := LocateFrontmatterPathSpan(yaml, "title.nested")
		if err == nil {
			t.Error("Expected error when accessing string as mapping, got nil")
		}
	})

	t.Run("accessing non-sequence as sequence", func(t *testing.T) {
		yaml := `title: "Simple String"
on: push`
		_, err := LocateFrontmatterPathSpan(yaml, "title[0]")
		if err == nil {
			t.Error("Expected error when accessing string as sequence, got nil")
		}
	})

	t.Run("empty mapping", func(t *testing.T) {
		yaml := `empty: {}`
		span, err := LocateFrontmatterPathSpan(yaml, "empty")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if span.StartLine != 1 {
			t.Errorf("Expected start line 1, got %d", span.StartLine)
		}
	})

	t.Run("empty sequence", func(t *testing.T) {
		yaml := `empty: []`
		span, err := LocateFrontmatterPathSpan(yaml, "empty")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if span.StartLine != 1 {
			t.Errorf("Expected start line 1, got %d", span.StartLine)
		}
	})
}

// TestMultiLineScalarPositions specifically tests multi-line scalar positioning
func TestMultiLineScalarPositions(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		path         string
		expectedSpan SourceSpan
	}{
		{
			name: "pipe literal multi-line",
			yaml: `description: |
  Line 1 of description
  Line 2 of description
  Line 3 of description
name: test`,
			path: "description",
			expectedSpan: SourceSpan{
				StartLine:   1,
				StartColumn: 14,
				EndLine:     3,
				EndColumn:   21,
			},
		},
		{
			name: "fold literal multi-line",
			yaml: `description: >
  This is a very long description
  that will be folded into a
  single line when processed
name: test`,
			path: "description",
			expectedSpan: SourceSpan{
				StartLine:   1,
				StartColumn: 14,
				EndLine:     1,
				EndColumn:   14,
			},
		},
		{
			name: "single line after multi-line",
			yaml: `description: |
  Multi-line content
  with multiple lines
simple: value`,
			path: "simple",
			expectedSpan: SourceSpan{
				StartLine:   4,
				StartColumn: 9,
				EndLine:     4,
				EndColumn:   13,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span, err := LocateFrontmatterPathSpan(tt.yaml, tt.path)
			if err != nil {
				t.Errorf("LocateFrontmatterPathSpan() error = %v", err)
				return
			}

			if span != tt.expectedSpan {
				t.Errorf("LocateFrontmatterPathSpan() = %+v, want %+v", span, tt.expectedSpan)
				
				// Print the YAML with line numbers for debugging
				lines := strings.Split(tt.yaml, "\n")
				t.Logf("YAML content with line numbers:")
				for i, line := range lines {
					t.Logf("%d: %s", i+1, line)
				}
			}
		})
	}
}

// TestFrontmatterLocatorCaching tests the cached locator performance and functionality
func TestFrontmatterLocatorCaching(t *testing.T) {
	frontmatterYAML := `engine: claude
on:
  push:
    branches: [main, develop]
  schedule:
    - cron: "0 9 * * 1"
max-turns: 5
tools:
  - name: git
    type: shell
  - name: curl
    type: shell`

	// Create a locator (parses YAML once)
	locator := NewFrontmatterLocator(frontmatterYAML)

	// Test multiple path lookups using the same parsed AST
	testPaths := []struct {
		path         string
		expectSpan   bool
		expectError  bool
	}{
		{"engine", true, false},
		{"on.push.branches[0]", true, false},
		{"on.schedule[0].cron", true, false},
		{"max-turns", true, false},
		{"tools[0].name", true, false},
		{"tools[1].type", true, false},
		{"nonexistent", false, true},
		{"tools[999]", false, true},
	}

	for _, testPath := range testPaths {
		t.Run(fmt.Sprintf("path_%s", testPath.path), func(t *testing.T) {
			span, err := locator.LocatePathSpan(testPath.path)
			
			if testPath.expectError {
				if err == nil {
					t.Errorf("Expected error for path '%s', got nil", testPath.path)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for path '%s': %v", testPath.path, err)
				return
			}
			
			if testPath.expectSpan {
				if span.StartLine <= 0 || span.StartColumn <= 0 {
					t.Errorf("Expected valid span for path '%s', got %+v", testPath.path, span)
				}
			}
		})
	}

	// Test legacy compatibility
	line, column, err := locator.LocatePath("engine")
	if err != nil {
		t.Errorf("Legacy LocatePath failed: %v", err)
	}
	if line <= 0 || column <= 0 {
		t.Errorf("Legacy LocatePath returned invalid position: line=%d, column=%d", line, column)
	}
}

// BenchmarkLocatorPerformance compares cached vs non-cached lookups
func BenchmarkLocatorPerformance(b *testing.B) {
	frontmatterYAML := `engine: claude
on:
  push:
    branches: [main, develop, feature/*]
  pull_request:
    types: [opened, synchronize, reopened]
  schedule:
    - cron: "0 9 * * 1"
    - cron: "0 18 * * 5"
max-turns: 10
tools:
  - name: git
    type: shell
  - name: curl
    type: shell
  - name: jq
    type: shell
variables:
  NODE_VERSION: "18"
  PYTHON_VERSION: "3.9"`

	paths := []string{
		"engine",
		"on.push.branches[0]",
		"on.pull_request.types[2]",
		"on.schedule[1].cron",
		"max-turns",
		"tools[0].name",
		"tools[2].type",
		"variables.NODE_VERSION",
		"variables.PYTHON_VERSION",
	}

	b.Run("without_caching", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, path := range paths {
				_, err := LocateFrontmatterPathSpan(frontmatterYAML, path)
				if err != nil {
					b.Fatalf("Error locating path %s: %v", path, err)
				}
			}
		}
	})

	b.Run("with_caching", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			locator := NewFrontmatterLocator(frontmatterYAML)
			for _, path := range paths {
				_, err := locator.LocatePathSpan(path)
				if err != nil {
					b.Fatalf("Error locating path %s: %v", path, err)
				}
			}
		}
	})

	b.Run("with_caching_reused_locator", func(b *testing.B) {
		locator := NewFrontmatterLocator(frontmatterYAML)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, path := range paths {
				_, err := locator.LocatePathSpan(path)
				if err != nil {
					b.Fatalf("Error locating path %s: %v", path, err)
				}
			}
		}
	})
}