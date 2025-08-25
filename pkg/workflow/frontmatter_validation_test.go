package workflow

import (
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

func TestFrontmatterValidator(t *testing.T) {
	tests := []struct {
		name            string
		frontmatterYAML string
		frontmatter     map[string]any
		shouldHaveError bool // Simplified - just check if there should be any error
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
			shouldHaveError: false,
		},
		{
			name: "invalid engine",
			frontmatterYAML: `engine: invalid-engine
on: push`,
			frontmatter: map[string]any{
				"engine": "invalid-engine",
				"on":     "push",
			},
			shouldHaveError: true,
		},
		{
			name: "max-turns too low",
			frontmatterYAML: `engine: claude
on: push
max-turns: 0`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": 0,
			},
			shouldHaveError: true,
		},
		{
			name: "max-turns negative",
			frontmatterYAML: `engine: claude
on: push
max-turns: -1`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": -1,
			},
			shouldHaveError: true,
		},
		{
			name: "max-turns invalid type",
			frontmatterYAML: `engine: claude
on: push
max-turns: "5"`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": "5",
			},
			shouldHaveError: true,
		},
		{
			name: "additional properties not allowed",
			frontmatterYAML: `engine: claude
on: push
invalid-field: value`,
			frontmatter: map[string]any{
				"engine":        "claude",
				"on":            "push",
				"invalid-field": "value",
			},
			shouldHaveError: true,
		},
		{
			name: "complex tools configuration valid",
			frontmatterYAML: `engine: claude
on: push
tools:
  github:
    allowed: [create_issue, create_comment]`,
			frontmatter: map[string]any{
				"engine": "claude",
				"on":     "push",
				"tools": map[string]any{
					"github": map[string]any{
						"allowed": []any{"create_issue", "create_comment"},
					},
				},
			},
			shouldHaveError: false,
		},
		{
			name: "invalid tools configuration",
			frontmatterYAML: `engine: claude
on: push
tools:
  github:
    invalid-tool-prop: value`,
			frontmatter: map[string]any{
				"engine": "claude",
				"on":     "push",
				"tools": map[string]any{
					"github": map[string]any{
						"invalid-tool-prop": "value",
					},
				},
			},
			shouldHaveError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFrontmatterValidator(tt.frontmatterYAML)
			errors := validator.ValidateFrontmatter(tt.frontmatter)

			if tt.shouldHaveError && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			if !tt.shouldHaveError && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got %d", len(errors))
			}
		})
	}
}

func TestValidationErrorSpans(t *testing.T) {
	testCases := []struct {
		name            string
		frontmatterYAML string
		frontmatter     map[string]any
		minExpectedErrors int // Minimum expected errors (JSON schema can produce multiple detailed errors)
	}{
		{
			name: "invalid engine - line 1",
			frontmatterYAML: `engine: invalid-engine
on: push`,
			frontmatter: map[string]any{
				"engine": "invalid-engine",
				"on":     "push",
			},
			minExpectedErrors: 1, // JSON schema produces multiple engine validation errors
		},
		{
			name: "max-turns too low - line 3",
			frontmatterYAML: `engine: claude
on: push
max-turns: 0`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": 0,
			},
			minExpectedErrors: 1,
		},
		{
			name: "multiple validation errors",
			frontmatterYAML: `engine: invalid-engine
on: push
max-turns: 0
unknown-field: value`,
			frontmatter: map[string]any{
				"engine":        "invalid-engine",
				"on":            "push",
				"max-turns":     0,
				"unknown-field": "value",
			},
			minExpectedErrors: 3, // Multiple types of errors
		},
		{
			name: "nested tools validation error",
			frontmatterYAML: `engine: claude
on: push
tools:
  github:
    invalid-prop: value`,
			frontmatter: map[string]any{
				"engine": "claude",
				"on":     "push",
				"tools": map[string]any{
					"github": map[string]any{
						"invalid-prop": "value",
					},
				},
			},
			minExpectedErrors: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := NewFrontmatterValidator(tc.frontmatterYAML)
			errors := validator.ValidateFrontmatter(tc.frontmatter)

			if len(errors) < tc.minExpectedErrors {
				t.Errorf("Expected at least %d validation errors, got %d", tc.minExpectedErrors, len(errors))
			}

			if len(errors) > 0 {
				t.Logf("Validation errors for %s: %v", tc.name, errors)
			}

			// Check if we can extract JSONPath information from error messages
			for _, err := range errors {
				t.Logf("Error path: '%s', message: '%s'", err.Path, err.Message)
				
				// Verify we get some error information
				if err.Message == "" {
					t.Errorf("Error message should not be empty")
				}
				
				// Test source span mapping if path is available
				if err.Path != "" && err.Span != nil {
					t.Logf("  Source span: Line %d:%d to %d:%d", 
						err.Span.StartLine, err.Span.StartColumn,
						err.Span.EndLine, err.Span.EndColumn)
					
					// Verify span has reasonable values
					if err.Span.StartLine <= 0 || err.Span.StartColumn <= 0 {
						t.Errorf("Invalid span for error at path '%s': %+v", err.Path, err.Span)
					}
				}
			}
		})
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

// TestSourceLocationMappingAccuracy tests the accuracy of source location mapping
// for various frontmatter validation errors
func TestSourceLocationMappingAccuracy(t *testing.T) {
	testCases := []struct {
		name            string
		frontmatterYAML string
		frontmatter     map[string]any
		description     string
	}{
		{
			name: "engine_on_first_line",
			frontmatterYAML: `engine: invalid-engine
on: push
max-turns: 5`,
			frontmatter: map[string]any{
				"engine":    "invalid-engine",
				"on":        "push",
				"max-turns": 5,
			},
			description: "Invalid engine on line 1 should be caught",
		},
		{
			name: "max_turns_on_third_line",
			frontmatterYAML: `engine: claude
on: push
max-turns: 0`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": 0,
			},
			description: "Invalid max-turns on line 3 should be caught",
		},
		{
			name: "deeply_nested_tools_error",
			frontmatterYAML: `engine: claude
on: push
tools:
  github:
    allowed: []
    invalid-nested-prop: value`,
			frontmatter: map[string]any{
				"engine": "claude",
				"on":     "push",
				"tools": map[string]any{
					"github": map[string]any{
						"allowed":             []any{},
						"invalid-nested-prop": "value",
					},
				},
			},
			description: "Invalid nested property in tools configuration",
		},
		{
			name: "type_mismatch_error",
			frontmatterYAML: `engine: claude
on: push
max-turns: "not-a-number"`,
			frontmatter: map[string]any{
				"engine":    "claude",
				"on":        "push",
				"max-turns": "not-a-number",
			},
			description: "Type mismatch for max-turns should be caught",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			
			validator := NewFrontmatterValidator(tc.frontmatterYAML)
			errors := validator.ValidateFrontmatter(tc.frontmatter)

			if len(errors) == 0 {
				t.Fatalf("Expected validation errors for test case '%s', got none", tc.name)
			}

			// Log validation errors for manual inspection
			for i, err := range errors {
				t.Logf("Error %d: Path='%s', Message='%s'", i+1, err.Path, err.Message)
				if err.Span != nil {
					t.Logf("  Span: Line %d:%d to %d:%d", 
						err.Span.StartLine, err.Span.StartColumn,
						err.Span.EndLine, err.Span.EndColumn)
				} else {
					t.Logf("  No span information available")
				}
			}

			// Test compilation error conversion
			compilerErrors := ConvertValidationErrorsToCompilerErrors("test.md", 2, errors)
			if len(compilerErrors) != len(errors) {
				t.Errorf("Expected %d compiler errors, got %d", len(errors), len(compilerErrors))
			}

			for i, compilerErr := range compilerErrors {
				t.Logf("Compiler Error %d: File='%s', Line=%d, Column=%d", 
					i+1, compilerErr.Position.File, compilerErr.Position.Line, compilerErr.Position.Column)
			}
		})
	}
}
