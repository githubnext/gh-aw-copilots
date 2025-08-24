package workflow

import (
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
