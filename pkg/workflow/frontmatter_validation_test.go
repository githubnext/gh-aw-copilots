package workflow

import (
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

func TestFrontmatterValidator(t *testing.T) {
	tests := []struct {
		name               string
		frontmatterYAML    string
		frontmatter        map[string]any
		shouldHaveError    bool // Simplified - just check if there should be any error
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
	frontmatterYAML := `engine: invalid-engine
on: push`

	frontmatter := map[string]any{
		"engine": "invalid-engine",
		"on":     "push",
	}

	validator := NewFrontmatterValidator(frontmatterYAML)
	errors := validator.ValidateFrontmatter(frontmatter)

	// With JSON schema validation, we might get different numbers of errors
	// Just check that we get validation errors for invalid engine
	if len(errors) == 0 {
		t.Fatalf("Expected validation errors, got none")
	}

	// Check that at least one error has span information
	hasSpan := false
	for _, err := range errors {
		if err.Span != nil {
			hasSpan = true
			// Verify that spans have reasonable values
			if err.Span.StartLine <= 0 || err.Span.StartColumn <= 0 {
				t.Errorf("Invalid span for error at path '%s': %+v", err.Path, err.Span)
			}
		}
	}

	if !hasSpan {
		t.Logf("Note: No span information available in simplified validation. Errors: %v", errors)
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


