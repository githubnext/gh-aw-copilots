package workflow

import (
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

// TestAgenticWorkflowValidationSimplified provides basic validation testing
// for agentic workflow files with various validation errors.
func TestAgenticWorkflowValidationSimplified(t *testing.T) {
	testCases := []struct {
		name            string
		workflowYAML    string
		shouldHaveError bool
	}{
		{
			name: "valid_agentic_workflow",
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
			name: "invalid_engine_type",
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

// TestAgenticWorkflowValidationComprehensive provides comprehensive validation testing
// for agentic workflow files with various validation errors and source location tracking.
func TestAgenticWorkflowValidationComprehensive(t *testing.T) {
	testCases := []struct {
		name              string
		workflowYAML      string
		shouldHaveError   bool
		expectedErrorType string
		description       string
	}{
		{
			name: "valid_complex_workflow",
			workflowYAML: `---
engine: claude
on:
  push:
    branches: [main]
  workflow_dispatch:
max-turns: 10
tools:
  github:
    allowed: [create_issue, create_comment]
    use_docker_mcp: false
permissions: write-all
---`,
			shouldHaveError:   false,
			expectedErrorType: "",
			description:       "Complex valid workflow with multiple configurations",
		},
		{
			name: "invalid_engine_unsupported",
			workflowYAML: `---
engine: gpt-4
on: push
---`,
			shouldHaveError:   true,
			expectedErrorType: "engine_validation",
			description:       "Unsupported engine should be rejected",
		},
		{
			name: "invalid_max_turns_zero",
			workflowYAML: `---
engine: claude
on: push
max-turns: 0
---`,
			shouldHaveError:   true,
			expectedErrorType: "max_turns_validation",
			description:       "Max-turns value of 0 should be rejected (minimum is 1)",
		},
		{
			name: "invalid_max_turns_negative",
			workflowYAML: `---
engine: claude
on: push
max-turns: -5
---`,
			shouldHaveError:   true,
			expectedErrorType: "max_turns_validation",
			description:       "Negative max-turns should be rejected",
		},
		{
			name: "invalid_max_turns_type",
			workflowYAML: `---
engine: claude
on: push
max-turns: "not-a-number"
---`,
			shouldHaveError:   true,
			expectedErrorType: "type_validation",
			description:       "String value for max-turns should be rejected",
		},
		{
			name: "invalid_additional_properties",
			workflowYAML: `---
engine: claude
on: push
unknown-field: value
another-invalid: field
---`,
			shouldHaveError:   true,
			expectedErrorType: "additional_properties",
			description:       "Unknown fields should be rejected by schema",
		},
		{
			name: "invalid_tools_configuration",
			workflowYAML: `---
engine: claude
on: push
tools:
  github:
    invalid-prop: value
    also-invalid: another-value
---`,
			shouldHaveError:   true,
			expectedErrorType: "tools_validation",
			description:       "Invalid tools properties should be rejected",
		},
		{
			name: "valid_engine_object_format",
			workflowYAML: `---
engine:
  id: claude
  model: claude-3-sonnet
on: push
---`,
			shouldHaveError:   false,
			expectedErrorType: "",
			description:       "Engine object format should be valid",
		},
		{
			name: "invalid_engine_object_id",
			workflowYAML: `---
engine:
  id: invalid-engine
  model: some-model
on: push
---`,
			shouldHaveError:   true,
			expectedErrorType: "engine_validation",
			description:       "Invalid engine ID in object format should be rejected",
		},
		{
			name: "complex_valid_tools_object_format",
			workflowYAML: `---
engine: claude
on: push
tools:
  github:
    allowed: ["create_issue", "create_comment"]
---`,
			shouldHaveError:   false,
			expectedErrorType: "",
			description:       "Tools object format with allowed array should be valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

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

			// Log detailed error information for debugging
			if len(errors) > 0 {
				t.Logf("Found %d validation error(s):", len(errors))
				for i, err := range errors {
					t.Logf("  Error %d: Path='%s', Message='%s'", i+1, err.Path, err.Message)
					if err.Span != nil {
						t.Logf("    Span: Line %d:%d to %d:%d",
							err.Span.StartLine, err.Span.StartColumn,
							err.Span.EndLine, err.Span.EndColumn)
					}
				}
			}
		})
	}
}
