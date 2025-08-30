package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskJobWithIfConditionHasDummyStep(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "task-job-if-test*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowContent := `---
on:
  workflow_run:
    workflows: ["Daily Perf Improver", "Daily Test Coverage Improver"]
    types:
      - completed
  stop-after: +48h

if: ${{ github.event.workflow_run.conclusion == 'failure' }}
---

# CI Doctor

This workflow runs when CI workflows fail to help diagnose issues.

Check the failed workflow and provide analysis.`

	// Write the test workflow file
	workflowFile := filepath.Join(tmpDir, "test-workflow.md")
	err = os.WriteFile(workflowFile, []byte(workflowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	// Compile the workflow
	compiler := NewCompiler(false, "", "test")

	err = compiler.CompileWorkflow(workflowFile)
	if err != nil {
		t.Fatalf("Failed to compile workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(workflowFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockContentStr := string(lockContent)

	// Test 1: Verify task job exists
	if !strings.Contains(lockContentStr, "task:") {
		t.Error("Expected task job to be present in generated workflow")
	}

	// Test 2: Verify task job has the if condition
	if !strings.Contains(lockContentStr, "if: ${{ github.event.workflow_run.conclusion == 'failure' }}") {
		t.Error("Expected task job to have the if condition")
	}

	// Test 3: Verify task job has steps (specifically the dummy step)
	if !strings.Contains(lockContentStr, "steps:") {
		t.Error("Task job should contain steps section")
	}

	// Test 4: Verify the dummy step is present
	if !strings.Contains(lockContentStr, "Task job condition barrier") {
		t.Error("Task job should contain the dummy step 'Task job condition barrier'")
	}

	// Test 5: Verify the dummy step has a run command
	if !strings.Contains(lockContentStr, "run: echo \"Task job executed - conditions satisfied\"") {
		t.Error("Task job should contain the dummy step run command")
	}

	// Test 6: Verify main job depends on task job
	if !strings.Contains(lockContentStr, "needs: task") {
		t.Error("Main job should depend on task job")
	}

	// Test 7: Verify the generated YAML is valid (no empty task job)
	// Check that there's no task job section that has only "if:" and "runs-on:" without steps
	lines := strings.Split(lockContentStr, "\n")
	inTaskJob := false
	hasSteps := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "task:" {
			inTaskJob = true
			hasSteps = false
			continue
		}

		if inTaskJob {
			// If we hit another job at the same level, stop checking task job
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") && strings.Contains(line, ":") {
				break
			}

			// Check if we found steps
			if strings.TrimSpace(line) == "steps:" {
				hasSteps = true
			}
		}
	}

	if inTaskJob && !hasSteps {
		t.Error("Task job must have steps to be valid GitHub Actions YAML")
	}
}
