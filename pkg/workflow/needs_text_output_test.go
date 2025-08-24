package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNeedsTextOutputDetection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "text-output-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name         string
		content      string
		expectNeeded bool
	}{
		{
			name: "workflow with text output usage",
			content: `---
name: Test Workflow With Text Output
on: push
---

This workflow uses the task output text.

Here is the current issue text: ${{ needs.task.outputs.text }}

## Job: test

Process the text above.
`,
			expectNeeded: true,
		},
		{
			name: "workflow without text output usage",
			content: `---
name: Test Workflow Without Text Output  
on: push
---

This is a basic test workflow.

## Job: test

This does not use any text outputs.
`,
			expectNeeded: false,
		},
		{
			name: "workflow with multiple text output usages",
			content: `---
name: Test Workflow With Multiple Text Outputs
on: push
---

Using text output: ${{ needs.task.outputs.text }}

More content here.

Using it again: ${{ needs.task.outputs.text }}

## Job: test

Process both occurrences.
`,
			expectNeeded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create workflow file
			workflowFile := filepath.Join(tempDir, strings.ReplaceAll(tt.name, " ", "-")+".md")
			if err := os.WriteFile(workflowFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create workflow file: %v", err)
			}

			// Compile the workflow
			compiler := NewCompiler(true, "", "test-version")
			workflowData, err := compiler.parseWorkflowFile(workflowFile)
			if err != nil {
				t.Fatalf("Failed to parse workflow: %v", err)
			}

			// Check if NeedsTextOutput is set correctly
			if workflowData.NeedsTextOutput != tt.expectNeeded {
				t.Errorf("Expected NeedsTextOutput to be %v, got %v", tt.expectNeeded, workflowData.NeedsTextOutput)
			}

			// Check if task job is needed correctly
			isTaskJobNeeded := compiler.isTaskJobNeeded(workflowData)
			expectedTaskJobNeeded := tt.expectNeeded // Task job should be needed if text output is needed
			if isTaskJobNeeded != expectedTaskJobNeeded {
				t.Errorf("Expected task job needed to be %v, got %v", expectedTaskJobNeeded, isTaskJobNeeded)
			}

			// If text output is needed, compile full workflow and check for compute-text step
			if tt.expectNeeded {
				yamlContent, err := compiler.generateYAML(workflowData)
				if err != nil {
					t.Fatalf("Failed to generate YAML: %v", err)
				}

				// Check that the generated YAML contains the compute-text step
				if !strings.Contains(yamlContent, "Compute current body text") {
					t.Errorf("Expected generated YAML to contain compute-text step")
				}

				// Check that the task job outputs include text
				if !strings.Contains(yamlContent, "text: ${{ steps.compute-text.outputs.text }}") {
					t.Errorf("Expected task job to have text output")
				}

				// Check that the compute-text step uses github-script action
				if !strings.Contains(yamlContent, "uses: actions/github-script@v7") {
					t.Errorf("Expected compute-text step to use github-script action")
				}
			}
		})
	}
}

func TestDetectTextOutputUsage(t *testing.T) {
	compiler := NewCompiler(false, "", "test-version") // Use non-verbose mode for cleaner test output

	tests := []struct {
		name          string
		content       string
		expectedUsage bool
	}{
		{
			name:          "content with text output usage",
			content:       "This uses ${{ needs.task.outputs.text }} in the middle.",
			expectedUsage: true,
		},
		{
			name:          "content without text output usage",
			content:       "This is just regular content with no special expressions.",
			expectedUsage: false,
		},
		{
			name:          "content with other GitHub expressions but not text output",
			content:       "This uses ${{ github.event.issue.body }} but not the text output.",
			expectedUsage: false,
		},
		{
			name:          "content with multiple text output usages",
			content:       "First usage: ${{ needs.task.outputs.text }}, second usage: ${{ needs.task.outputs.text }}",
			expectedUsage: true,
		},
		{
			name:          "empty content",
			content:       "",
			expectedUsage: false,
		},
		{
			name:          "content with similar but not exact expression",
			content:       "This has ${{ needs.task.outputs.other }} which is similar but different.",
			expectedUsage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compiler.detectTextOutputUsage(tt.content)
			if result != tt.expectedUsage {
				t.Errorf("detectTextOutputUsage() = %v, expected %v", result, tt.expectedUsage)
			}
		})
	}
}
