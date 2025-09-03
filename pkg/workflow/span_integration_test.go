package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestSimplifiedSpanIntegration tests basic span integration with JSON schema validation
func TestSimplifiedSpanIntegration(t *testing.T) {
	tests := []struct {
		name            string
		workflowContent string
		shouldHaveError bool
	}{
		{
			name: "invalid engine",
			workflowContent: `---
engine: invalid-engine
on: push
---

# Test Workflow

This workflow has an invalid engine.`,
			shouldHaveError: true,
		},
		{
			name: "valid workflow",
			workflowContent: `---
engine: claude
on: push
---

# Test Workflow

This is a valid workflow.`,
			shouldHaveError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir, err := os.MkdirTemp("", "span-integration-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.workflowContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile workflow to test integration
			compiler := NewCompiler(false, "", "")
			err = compiler.CompileWorkflow(testFile)

			if tt.shouldHaveError && err == nil {
				t.Errorf("Expected compilation error but got none")
			}
			if !tt.shouldHaveError && err != nil {
				t.Errorf("Expected no compilation error but got: %v", err)
			}
		})
	}
}

// TestComprehensiveSpanIntegration tests comprehensive span integration with detailed validation errors
func TestComprehensiveSpanIntegration(t *testing.T) {
	testCases := []struct {
		name            string
		workflowContent string
		shouldHaveError bool
		description     string
	}{

		{
			name: "additional_properties_error",
			workflowContent: `---
engine: claude
on: push
invalid-field: value
unknown-prop: another-value
---

# Test Workflow

This workflow has unknown fields.`,
			shouldHaveError: true,
			description:     "Unknown frontmatter fields should be rejected",
		},
		{
			name: "nested_tools_error",
			workflowContent: `---
engine: claude
on: push
tools:
  github:
    invalid-nested: property
    another-bad: field
---

# Test Workflow

This workflow has invalid tools configuration.`,
			shouldHaveError: true,
			description:     "Invalid nested tools properties should be rejected",
		},
		{
			name: "type_mismatch_error",
			workflowContent: `---
engine: claude
on: push
max-turns: "not-a-number"
---

# Test Workflow

This workflow has a type mismatch for max-turns.`,
			shouldHaveError: true,
			description:     "String value for max-turns should be rejected",
		},
		{
			name: "complex_valid_workflow",
			workflowContent: `---
engine: claude
on:
  push:
    branches: [main]
  workflow_dispatch:
max-turns: 15
tools:
  github:
    allowed: [create_issue, create_comment, create_pr]
    use_docker_mcp: true
permissions: write-all
---

# Complex Test Workflow

This is a complex but valid workflow that should compile successfully.`,
			shouldHaveError: false,
			description:     "Complex valid workflow should compile without errors",
		},
		{
			name: "engine_object_format_error",
			workflowContent: `---
engine:
  id: invalid-engine-id
  model: some-model
on: push
---

# Test Workflow

This workflow has an invalid engine object configuration.`,
			shouldHaveError: true,
			description:     "Invalid engine ID in object format should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Create temporary test file
			tmpDir, err := os.MkdirTemp("", "comprehensive-span-integration-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, fmt.Sprintf("%s.md", tc.name))
			if err := os.WriteFile(testFile, []byte(tc.workflowContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile workflow to test integration
			compiler := NewCompiler(false, "", "")
			err = compiler.CompileWorkflow(testFile)

			if tc.shouldHaveError && err == nil {
				t.Errorf("Expected compilation error but got none")
			}
			if !tc.shouldHaveError && err != nil {
				t.Errorf("Expected no compilation error but got: %v", err)
			}

			// Log error details for analysis
			if err != nil {
				t.Logf("Compilation error: %v", err)
			} else {
				t.Logf("Compilation successful")
			}
		})
	}
}

// TestValidationErrorLocationAccuracy tests the accuracy of error location reporting
func TestValidationErrorLocationAccuracy(t *testing.T) {
	testCases := []struct {
		name            string
		workflowContent string
		expectedLine    int // Expected line number where error should be reported
		description     string
	}{
		{
			name: "engine_error_line_2",
			workflowContent: `---
engine: invalid-engine
on: push
---

# Test Workflow`,
			expectedLine: 2,
			description:  "Engine error should be reported on line 2",
		},

		{
			name: "nested_tools_error_line_6",
			workflowContent: `---
engine: claude
on: push
tools:
  github:
    invalid-prop: value
---

# Test Workflow`,
			expectedLine: 6,
			description:  "Nested tools error should be reported around line 6",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Create temporary test file
			tmpDir, err := os.MkdirTemp("", "location-accuracy-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, fmt.Sprintf("%s.md", tc.name))
			if err := os.WriteFile(testFile, []byte(tc.workflowContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile workflow to test location accuracy
			compiler := NewCompiler(false, "", "")
			err = compiler.CompileWorkflow(testFile)

			if err == nil {
				t.Errorf("Expected compilation error but got none")
				return
			}

			// Log the error for manual inspection of location accuracy
			t.Logf("Compilation error (should reference line %d): %v", tc.expectedLine, err)

			// Note: For now, we're primarily testing that errors are caught
			// Future enhancement could parse error messages to verify exact line numbers
		})
	}
}
