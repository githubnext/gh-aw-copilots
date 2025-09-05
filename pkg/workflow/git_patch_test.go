package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func TestGitPatchGeneration(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a test markdown file with minimal agentic workflow
	testMarkdown := `---
on:
  workflow_dispatch:
safe-outputs:
  create-pull-request:
---

# Test Git Patch

This is a test workflow to validate git patch generation.

Please do the following tasks:
1. Check current status
2. Make some changes
3. Verify the git patch is generated
`

	// Write the test file
	mdFile := filepath.Join(tmpDir, "test-git-patch.md")
	if err := os.WriteFile(mdFile, []byte(testMarkdown), 0644); err != nil {
		t.Fatalf("Failed to write test markdown file: %v", err)
	}

	// Create compiler with verbose enabled for testing
	compiler := NewCompiler(false, "", "test-version")

	// Compile the workflow
	if err := compiler.CompileWorkflow(mdFile); err != nil {
		t.Fatalf("Failed to compile workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-git-patch.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify git patch generation step exists
	if !strings.Contains(lockContent, "- name: Generate git patch") {
		t.Error("Expected 'Generate git patch' step to be in generated workflow")
	}

	// Verify the git patch step contains the expected commands
	if !strings.Contains(lockContent, "git status") {
		t.Error("Expected 'git status' command in git patch step")
	}

	if !strings.Contains(lockContent, "git add -A || true") {
		t.Error("Expected 'git add -A || true' command in git patch step")
	}

	if !strings.Contains(lockContent, "INITIAL_SHA=\"$GITHUB_SHA\"") {
		t.Error("Expected INITIAL_SHA variable assignment in git patch step")
	}

	if !strings.Contains(lockContent, "git format-patch") {
		t.Error("Expected 'git format-patch' command in git patch step")
	}

	if !strings.Contains(lockContent, "/tmp/aw.patch") {
		t.Error("Expected '/tmp/aw.patch' path in git patch step")
	}

	// Verify it skips patch generation when no changes
	if !strings.Contains(lockContent, "Skipping patch generation - no committed changes to create patch from") {
		t.Error("Expected message about skipping patch generation when no changes")
	}

	// Verify git patch upload step exists
	if !strings.Contains(lockContent, "- name: Upload git patch") {
		t.Error("Expected 'Upload git patch' step to be in generated workflow")
	}

	// Verify the upload step uses actions/upload-artifact@v5
	if !strings.Contains(lockContent, "uses: actions/upload-artifact@v5") {
		t.Error("Expected upload-artifact action to be used for git patch upload step")
	}

	// Verify the artifact upload configuration
	if !strings.Contains(lockContent, "name: aw.patch") {
		t.Error("Expected artifact name 'aw.patch' in upload step")
	}

	if !strings.Contains(lockContent, "path: /tmp/aw.patch") {
		t.Error("Expected artifact path '/tmp/aw.patch' in upload step")
	}

	if !strings.Contains(lockContent, "if-no-files-found: ignore") {
		t.Error("Expected 'if-no-files-found: ignore' in upload step")
	}

	// Verify the git patch step runs with 'if: always()'
	gitPatchIndex := strings.Index(lockContent, "- name: Generate git patch")
	if gitPatchIndex == -1 {
		t.Fatal("Git patch step not found")
	}

	// Find the next step after git patch step
	nextStepStart := gitPatchIndex + len("- name: Generate git patch")
	stepEnd := strings.Index(lockContent[nextStepStart:], "- name:")
	if stepEnd == -1 {
		stepEnd = len(lockContent) - nextStepStart
	}
	gitPatchStep := lockContent[gitPatchIndex : nextStepStart+stepEnd]

	if !strings.Contains(gitPatchStep, "if: always()") {
		t.Error("Expected git patch step to have 'if: always()' condition")
	}

	// Verify the upload step runs with conditional logic for file existence
	uploadPatchIndex := strings.Index(lockContent, "- name: Upload git patch")
	if uploadPatchIndex == -1 {
		t.Fatal("Upload git patch step not found")
	}

	// Find the next step after upload patch step
	nextUploadStart := uploadPatchIndex + len("- name: Upload git patch")
	uploadStepEnd := strings.Index(lockContent[nextUploadStart:], "- name:")
	if uploadStepEnd == -1 {
		uploadStepEnd = len(lockContent) - nextUploadStart
	}
	uploadPatchStep := lockContent[uploadPatchIndex : nextUploadStart+uploadStepEnd]

	if !strings.Contains(uploadPatchStep, "if: always()") {
		t.Error("Expected upload git patch step to have 'if: always()' condition")
	}

	// Verify step ordering: git patch steps should be after agentic execution but before other uploads
	agenticIndex := strings.Index(lockContent, "Execute Claude Code")
	if agenticIndex == -1 {
		// Try alternative agentic step names
		agenticIndex = strings.Index(lockContent, "uses: anthropics/claude-code-base-action")
		if agenticIndex == -1 {
			agenticIndex = strings.Index(lockContent, "uses: githubnext/claude-action")
		}
	}

	uploadEngineLogsIndex := strings.Index(lockContent, "Upload agentic engine logs")

	if agenticIndex != -1 && gitPatchIndex != -1 && uploadEngineLogsIndex != -1 {
		if gitPatchIndex <= agenticIndex {
			t.Error("Git patch step should appear after agentic execution step")
		}

		if gitPatchIndex >= uploadEngineLogsIndex {
			t.Error("Git patch step should appear before engine logs upload step")
		}
	}
}
