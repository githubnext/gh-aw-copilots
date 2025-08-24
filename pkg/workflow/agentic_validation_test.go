package workflow

import (
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

// TestAgenticWorkflowValidationSimplified provides basic validation testing
// for agentic workflow files with various validation errors.
func TestAgenticWorkflowValidationSimplified(t *testing.T) {
	testCases := []struct {
		name           string
		workflowYAML   string
		shouldHaveError bool
	}{
		{
			name:        "valid_agentic_workflow",
			workflowYAML: `---
engine: claude
on: push
tools:
  github:
    allowed: [create_issue]
---`,
			shouldHaveError: false,
		},
		{
			name:        "invalid_engine_type",
			workflowYAML: `---
engine: gpt-4
on: push
---`,
			shouldHaveError: true, // JSON schema should catch this
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse frontmatter for validation
			result, err := parser.ExtractFrontmatterFromContent(tc.workflowYAML)
			if err != nil {
				t.Fatalf("Failed to parse frontmatter: %v", err)
			}

			validator := NewFrontmatterValidator(tc.workflowYAML)
			errors := validator.ValidateFrontmatter(result.Frontmatter)

			if tc.shouldHaveError && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			if !tc.shouldHaveError && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got %d: %v", len(errors), errors)
			}

			if len(errors) > 0 {
				t.Logf("Validation errors: %v", errors)
			}
		})
	}
}