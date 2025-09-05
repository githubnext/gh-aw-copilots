package workflow

import (
	"strings"
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

func TestGitCommandsIntegrationWithCreatePullRequest(t *testing.T) {
	// Create a simple workflow with create-pull-request enabled
	workflowContent := `---
name: Test Git Commands Integration
tools:
  claude:
    allowed:
      Read: null
      Write: null
safe-outputs:
  create-pull-request:
    max: 1
---

This is a test workflow that should automatically get Git commands when create-pull-request is enabled.
`

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow content and get both result and allowed tools string
	_, allowedToolsStr, err := compiler.parseWorkflowMarkdownContentWithToolsString(workflowContent)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Verify that Git commands are present in the allowed tools string
	expectedGitCommands := []string{"Bash(git checkout:*)", "Bash(git add:*)", "Bash(git commit:*)", "Bash(git branch:*)", "Bash(git switch:*)", "Bash(git rm:*)", "Bash(git merge:*)"}

	for _, expectedCmd := range expectedGitCommands {
		if !strings.Contains(allowedToolsStr, expectedCmd) {
			t.Errorf("Expected allowed tools to contain %s, got: %s", expectedCmd, allowedToolsStr)
		}
	}

	// Verify that the basic tools are also present
	if !strings.Contains(allowedToolsStr, "Read") {
		t.Errorf("Expected allowed tools to contain Read tool, got: %s", allowedToolsStr)
	}
	if !strings.Contains(allowedToolsStr, "Write") {
		t.Errorf("Expected allowed tools to contain Write tool, got: %s", allowedToolsStr)
	}
}

func TestGitCommandsNotAddedWithoutPullRequestOutput(t *testing.T) {
	// Create a workflow with only create-issue (no PR-related outputs)
	workflowContent := `---
name: Test No Git Commands
tools:
  claude:
    allowed:
      Read: null
      Write: null
safe-outputs:
  create-issue:
    max: 1
---

This workflow should NOT get Git commands since it doesn't use create-pull-request or push-to-branch.
`

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow content and get allowed tools string
	_, allowedToolsStr, err := compiler.parseWorkflowMarkdownContentWithToolsString(workflowContent)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Verify allowed tools do not include Git commands
	gitCommands := []string{"Bash(git checkout:*)", "Bash(git add:*)", "Bash(git commit:*)", "Bash(git branch:*)", "Bash(git switch:*)", "Bash(git rm:*)", "Bash(git merge:*)"}
	for _, gitCmd := range gitCommands {
		if strings.Contains(allowedToolsStr, gitCmd) {
			t.Errorf("Did not expect allowed tools to contain Git command %s, got: %s", gitCmd, allowedToolsStr)
		}
	}

	// Verify basic tools are still present
	if !strings.Contains(allowedToolsStr, "Read") {
		t.Errorf("Expected allowed tools to contain Read tool, got: %s", allowedToolsStr)
	}
	if !strings.Contains(allowedToolsStr, "Write") {
		t.Errorf("Expected allowed tools to contain Write tool, got: %s", allowedToolsStr)
	}
}

func TestAdditionalClaudeToolsIntegrationWithCreatePullRequest(t *testing.T) {
	// Create a simple workflow with create-pull-request enabled
	workflowContent := `---
name: Test Additional Claude Tools Integration
tools:
  claude:
    allowed:
      Read: null
      Task: null
safe-outputs:
  create-pull-request:
    max: 1
---

This is a test workflow that should automatically get additional Claude tools when create-pull-request is enabled.
`

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow content and get allowed tools string
	_, allowedToolsStr, err := compiler.parseWorkflowMarkdownContentWithToolsString(workflowContent)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Verify that additional Claude tools are present in the allowed tools string
	expectedAdditionalTools := []string{"Edit", "MultiEdit", "Write", "NotebookEdit"}
	for _, expectedTool := range expectedAdditionalTools {
		if !strings.Contains(allowedToolsStr, expectedTool) {
			t.Errorf("Expected allowed tools to contain %s, got: %s", expectedTool, allowedToolsStr)
		}
	}

	// Verify that pre-existing tools are still there
	if !strings.Contains(allowedToolsStr, "Read") {
		t.Error("Expected pre-existing Read tool to be preserved")
	}
	if !strings.Contains(allowedToolsStr, "Task") {
		t.Error("Expected pre-existing Task tool to be preserved")
	}

	// Verify Git commands are also present (since create-pull-request is enabled)
	expectedGitCommands := []string{"Bash(git checkout:*)", "Bash(git add:*)", "Bash(git commit:*)"}
	for _, expectedCmd := range expectedGitCommands {
		if !strings.Contains(allowedToolsStr, expectedCmd) {
			t.Errorf("Expected allowed tools to contain %s, got: %s", expectedCmd, allowedToolsStr)
		}
	}
}

func TestAdditionalClaudeToolsIntegrationWithPushToBranch(t *testing.T) {
	// Create a simple workflow with push-to-branch enabled
	workflowContent := `---
name: Test Additional Claude Tools Integration with Push to Branch
tools:
  claude:
    allowed:
      Read: null
safe-outputs:
  push-to-branch:
    branch: "feature-branch"
---

This is a test workflow that should automatically get additional Claude tools when push-to-branch is enabled.
`

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow content and get allowed tools string
	_, allowedToolsStr, err := compiler.parseWorkflowMarkdownContentWithToolsString(workflowContent)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Verify that additional Claude tools are present in the allowed tools string
	expectedAdditionalTools := []string{"Edit", "MultiEdit", "Write", "NotebookEdit"}
	for _, expectedTool := range expectedAdditionalTools {
		if !strings.Contains(allowedToolsStr, expectedTool) {
			t.Errorf("Expected additional Claude tool %s to be present, got: %s", expectedTool, allowedToolsStr)
		}
	}

	// Verify that pre-existing tools are still there
	if !strings.Contains(allowedToolsStr, "Read") {
		t.Error("Expected pre-existing Read tool to be preserved")
	}

	// Verify Git commands are also present (since push-to-branch is enabled)
	expectedGitCommands := []string{"Bash(git checkout:*)", "Bash(git add:*)", "Bash(git commit:*)"}
	for _, expectedCmd := range expectedGitCommands {
		if !strings.Contains(allowedToolsStr, expectedCmd) {
			t.Errorf("Expected allowed tools to contain %s when push-to-branch is enabled, got: %s", expectedCmd, allowedToolsStr)
		}
	}
}

// Helper function to parse workflow content and return both WorkflowData and allowed tools string
func (c *Compiler) parseWorkflowMarkdownContentWithToolsString(content string) (*WorkflowData, string, error) {
	// This would normally be in parseWorkflowFile, but we'll extract the core logic for testing
	result, err := parser.ExtractFrontmatterFromContent(content)
	if err != nil {
		return nil, "", err
	}
	engine := NewClaudeEngine()

	// Extract SafeOutputs early
	safeOutputs := c.extractSafeOutputsConfig(result.Frontmatter)

	// Extract and process tools
	topTools := extractToolsFromFrontmatter(result.Frontmatter)
	topTools = c.applyDefaultGitHubMCPTools(topTools)

	// Build basic workflow data for testing
	workflowData := &WorkflowData{
		Name:        "Test Workflow",
		Tools:       topTools,
		SafeOutputs: safeOutputs,
		AI:          "claude",
	}
	allowedToolsStr := engine.computeAllowedClaudeToolsString(topTools, safeOutputs)

	return workflowData, allowedToolsStr, nil
}
