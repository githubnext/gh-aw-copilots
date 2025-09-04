package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputConfigParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-issue configuration
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
safe-outputs:
  create-issue:
    title-prefix: "[genai] "
    labels: [copilot, automation]
---

# Test Output Configuration

This workflow tests the output configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-config.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with output config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.CreateIssues == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}

	// Verify title prefix
	expectedPrefix := "[genai] "
	if workflowData.SafeOutputs.CreateIssues.TitlePrefix != expectedPrefix {
		t.Errorf("Expected title prefix '%s', got '%s'", expectedPrefix, workflowData.SafeOutputs.CreateIssues.TitlePrefix)
	}

	// Verify labels
	expectedLabels := []string{"copilot", "automation"}
	if len(workflowData.SafeOutputs.CreateIssues.Labels) != len(expectedLabels) {
		t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(workflowData.SafeOutputs.CreateIssues.Labels))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.SafeOutputs.CreateIssues.Labels) || workflowData.SafeOutputs.CreateIssues.Labels[i] != expectedLabel {
			t.Errorf("Expected label '%s' at index %d, got '%s'", expectedLabel, i, workflowData.SafeOutputs.CreateIssues.Labels[i])
		}
	}
}

func TestOutputConfigEmpty(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-config-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case without output configuration
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
---

# Test No Output Configuration

This workflow has no output configuration.
`

	testFile := filepath.Join(tmpDir, "test-no-output.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow without output config: %v", err)
	}

	// Verify output configuration contains only missing-tool (always enabled)
	if workflowData.SafeOutputs == nil {
		t.Error("Expected SafeOutputs to be non-nil since missing-tool is always enabled")
	} else {
		// Check that only missing-tool is enabled
		if workflowData.SafeOutputs.MissingTool == nil {
			t.Error("Expected MissingTool to be enabled by default")
		}
		if workflowData.SafeOutputs.CreateIssues != nil {
			t.Error("Expected CreateIssues to be nil when not specified")
		}
		if workflowData.SafeOutputs.CreatePullRequests != nil {
			t.Error("Expected CreatePullRequests to be nil when not specified")
		}
		if workflowData.SafeOutputs.AddIssueComments != nil {
			t.Error("Expected AddIssueComments to be nil when not specified")
		}
	}
}

func TestOutputConfigNull(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-config-null-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with null values for create-issue and create-pull-request
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  create-issue:
  create-pull-request:
  add-issue-comment:
  add-issue-label:
---

# Test Null Output Configuration

This workflow tests the null output configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-null-output-config.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with null output config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	// Verify create-issue configuration is parsed with empty values
	if workflowData.SafeOutputs.CreateIssues == nil {
		t.Fatal("Expected create-issue configuration to be parsed with null value")
	}
	if workflowData.SafeOutputs.CreateIssues.TitlePrefix != "" {
		t.Errorf("Expected empty title prefix for null create-issue, got '%s'", workflowData.SafeOutputs.CreateIssues.TitlePrefix)
	}
	if len(workflowData.SafeOutputs.CreateIssues.Labels) != 0 {
		t.Errorf("Expected empty labels for null create-issue, got %v", workflowData.SafeOutputs.CreateIssues.Labels)
	}

	// Verify create-pull-request configuration is parsed with empty values
	if workflowData.SafeOutputs.CreatePullRequests == nil {
		t.Fatal("Expected create-pull-request configuration to be parsed with null value")
	}
	if workflowData.SafeOutputs.CreatePullRequests.TitlePrefix != "" {
		t.Errorf("Expected empty title prefix for null create-pull-request, got '%s'", workflowData.SafeOutputs.CreatePullRequests.TitlePrefix)
	}
	if len(workflowData.SafeOutputs.CreatePullRequests.Labels) != 0 {
		t.Errorf("Expected empty labels for null create-pull-request, got %v", workflowData.SafeOutputs.CreatePullRequests.Labels)
	}

	// Verify add-issue-comment configuration is parsed with empty values
	if workflowData.SafeOutputs.AddIssueComments == nil {
		t.Fatal("Expected add-issue-comment configuration to be parsed with null value")
	}

	// Verify add-issue-label configuration is parsed with empty values
	if workflowData.SafeOutputs.AddIssueLabels == nil {
		t.Fatal("Expected add-issue-label configuration to be parsed with null value")
	}
	if len(workflowData.SafeOutputs.AddIssueLabels.Allowed) != 0 {
		t.Errorf("Expected empty allowed labels for null add-issue-label, got %v", workflowData.SafeOutputs.AddIssueLabels.Allowed)
	}
	if workflowData.SafeOutputs.AddIssueLabels.MaxCount != nil {
		t.Errorf("Expected nil MaxCount for null add-issue-label, got %v", *workflowData.SafeOutputs.AddIssueLabels.MaxCount)
	}
}

func TestOutputIssueJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-issue-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-issue configuration
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
tools:
  github:
    allowed: [list_issues]
engine: claude
safe-outputs:
  create-issue:
    title-prefix: "[genai] "
    labels: [copilot]
---

# Test Output Issue Job Generation

This workflow tests the create-issue job generation.
`

	testFile := filepath.Join(tmpDir, "test-output-issue.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output issue: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-output-issue.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify create_issue job exists
	if !strings.Contains(lockContent, "create_issue:") {
		t.Error("Expected 'create_issue' job to be in generated workflow")
	}

	// Verify job properties
	if !strings.Contains(lockContent, "timeout-minutes: 10") {
		t.Error("Expected 10-minute timeout in create_issue job")
	}

	if !strings.Contains(lockContent, "permissions:\n      contents: read\n      issues: write") {
		t.Error("Expected correct permissions in create_issue job")
	}

	// Verify the job uses github-script
	if !strings.Contains(lockContent, "uses: actions/github-script@v7") {
		t.Error("Expected github-script action to be used in create_issue job")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContent, "GITHUB_AW_ISSUE_TITLE_PREFIX: \"[genai] \"") {
		t.Error("Expected title prefix to be set as environment variable")
	}

	if !strings.Contains(lockContent, "GITHUB_AW_ISSUE_LABELS: \"copilot\"") {
		t.Error("Expected copilot label to be set as environment variable")
	}

	// Verify job dependencies
	if !strings.Contains(lockContent, "needs: test-output-issue") {
		t.Error("Expected create_issue job to depend on main job")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputCommentConfigParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with output.add-issue-comment configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-comment:
---

# Test Output Issue Comment Configuration

This workflow tests the output.add-issue-comment configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-issue-comment.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with output comment config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments == nil {
		t.Fatal("Expected issue_comment configuration to be parsed")
	}
}

func TestOutputCommentConfigParsingNull(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-config-null-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with output.add-issue-comment: null (no {} brackets)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-comment:
---

# Test Output Issue Comment Configuration with Null Value

This workflow tests the output.add-issue-comment configuration parsing with null value.
`

	testFile := filepath.Join(tmpDir, "test-output-issue-comment-null.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with null output comment config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments == nil {
		t.Fatal("Expected issue_comment configuration to be parsed even with null value")
	}
}

func TestOutputCommentConfigTargetParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-target-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with target: "*"
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-comment:
    target: "*"
---

# Test Output Issue Comment Target Configuration

This workflow tests the output.add-issue-comment target configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-issue-comment-target.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with target comment config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments == nil {
		t.Fatal("Expected issue_comment configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments.Target != "*" {
		t.Fatalf("Expected target to be '*', got '%s'", workflowData.SafeOutputs.AddIssueComments.Target)
	}
}

func TestOutputCommentMaxTargetParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-max-target-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with max and target configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-comment:
    max: 3
    target: "123"
---

# Test Output Issue Comments Max Target Configuration

This workflow tests the add-issue-comment max and target configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-issue-comment-max-target.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with max target comment config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments == nil {
		t.Fatal("Expected issue_comment configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueComments.Max != 3 {
		t.Fatalf("Expected max to be 3, got %d", workflowData.SafeOutputs.AddIssueComments.Max)
	}

	if workflowData.SafeOutputs.AddIssueComments.Target != "123" {
		t.Fatalf("Expected target to be '123', got '%s'", workflowData.SafeOutputs.AddIssueComments.Target)
	}
}

func TestOutputCommentJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with output.add-issue-comment configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
tools:
  github:
    allowed: [get_issue]
engine: claude
safe-outputs:
  add-issue-comment:
---

# Test Output Issue Comment Job Generation

This workflow tests the create_issue_comment job generation.
`

	testFile := filepath.Join(tmpDir, "test-output-issue-comment.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output comment: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-output-issue-comment.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify create_issue_comment job exists
	if !strings.Contains(lockContent, "create_issue_comment:") {
		t.Error("Expected 'create_issue_comment' job to be in generated workflow")
	}

	// Verify job properties
	if !strings.Contains(lockContent, "timeout-minutes: 10") {
		t.Error("Expected 10-minute timeout in create_issue_comment job")
	}

	if !strings.Contains(lockContent, "permissions:\n      contents: read\n      issues: write\n      pull-requests: write") {
		t.Error("Expected correct permissions in create_issue_comment job")
	}

	// Verify the job uses github-script
	if !strings.Contains(lockContent, "uses: actions/github-script@v7") {
		t.Error("Expected github-script action to be used in create_issue_comment job")
	}

	// Verify job has conditional execution
	if !strings.Contains(lockContent, "if: github.event.issue.number || github.event.pull_request.number") {
		t.Error("Expected create_issue_comment job to have conditional execution")
	}

	// Verify job dependencies
	if !strings.Contains(lockContent, "needs: test-output-issue-comment") {
		t.Error("Expected create_issue_comment job to depend on main job")
	}

	// Verify JavaScript content includes environment variable for agent output
	if !strings.Contains(lockContent, "GITHUB_AW_AGENT_OUTPUT:") {
		t.Error("Expected agent output content to be passed as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputCommentJobSkippedForNonIssueEvents(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-skip-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-comment configuration but push trigger (not issue/PR)
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-comment:
---

# Test Output Issue Comment Job Skipping

This workflow tests that issue comment job is skipped for non-issue/PR events.
`

	testFile := filepath.Join(tmpDir, "test-comment-skip.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output comment: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-comment-skip.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify create_issue_comment job exists (it should be generated regardless of trigger)
	if !strings.Contains(lockContent, "create_issue_comment:") {
		t.Error("Expected 'create_issue_comment' job to be in generated workflow")
	}

	// Verify job has conditional execution to skip when not in issue/PR context
	if !strings.Contains(lockContent, "if: github.event.issue.number || github.event.pull_request.number") {
		t.Error("Expected create_issue_comment job to have conditional execution for skipping")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputPullRequestConfigParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-pr-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-pull-request configuration
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
engine: claude
safe-outputs:
  create-pull-request:
    title-prefix: "[agent] "
    labels: [automation, bot]
---

# Test Output Pull Request Configuration

This workflow tests the output pull request configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-pr-config.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with output pull-request config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.CreatePullRequests == nil {
		t.Fatal("Expected pull-request configuration to be parsed")
	}

	// Verify title prefix
	expectedPrefix := "[agent] "
	if workflowData.SafeOutputs.CreatePullRequests.TitlePrefix != expectedPrefix {
		t.Errorf("Expected title prefix '%s', got '%s'", expectedPrefix, workflowData.SafeOutputs.CreatePullRequests.TitlePrefix)
	}

	// Verify labels
	expectedLabels := []string{"automation", "bot"}
	if len(workflowData.SafeOutputs.CreatePullRequests.Labels) != len(expectedLabels) {
		t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(workflowData.SafeOutputs.CreatePullRequests.Labels))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.SafeOutputs.CreatePullRequests.Labels) || workflowData.SafeOutputs.CreatePullRequests.Labels[i] != expectedLabel {
			t.Errorf("Expected label[%d] to be '%s', got '%s'", i, expectedLabel, workflowData.SafeOutputs.CreatePullRequests.Labels[i])
		}
	}
}

func TestOutputPullRequestJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-pr-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-pull-request configuration
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
safe-outputs:
  create-pull-request:
    title-prefix: "[agent] "
    labels: [automation]
---

# Test Output Pull Request Job Generation

This workflow tests the create_pull_request job generation.
`

	testFile := filepath.Join(tmpDir, "test-output-pr.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output create-pull-request: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	// Convert to string for easier testing
	lockContentStr := string(lockContent)

	// Verify create_pull_request job is present
	if !strings.Contains(lockContentStr, "create_pull_request:") {
		t.Error("Expected 'create_pull_request' job to be in generated workflow")
	}

	// Verify permissions
	if !strings.Contains(lockContentStr, "contents: write") {
		t.Error("Expected contents: write permission in create_pull_request job")
	}

	if !strings.Contains(lockContentStr, "pull-requests: write") {
		t.Error("Expected pull-requests: write permission in create_pull_request job")
	}

	// Verify steps
	if !strings.Contains(lockContentStr, "Download patch artifact") {
		t.Error("Expected 'Download patch artifact' step in create_pull_request job")
	}

	if !strings.Contains(lockContentStr, "actions/download-artifact@v4") {
		t.Error("Expected download-artifact action to be used in create_pull_request job")
	}

	if !strings.Contains(lockContentStr, "Checkout repository") {
		t.Error("Expected 'Checkout repository' step in create_pull_request job")
	}

	if !strings.Contains(lockContentStr, "Create Pull Request") {
		t.Error("Expected 'Create Pull Request' step in create_pull_request job")
	}

	if !strings.Contains(lockContentStr, "uses: actions/github-script@v7") {
		t.Error("Expected github-script action to be used in create_pull_request job")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_TITLE_PREFIX: \"[agent] \"") {
		t.Error("Expected title prefix to be set as environment variable")
	}

	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_LABELS: \"automation\"") {
		t.Error("Expected automation label to be set as environment variable")
	}

	// Verify draft setting defaults to true
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_DRAFT: \"true\"") {
		t.Error("Expected draft to default to true when not specified")
	}

	// Verify job dependencies
	if !strings.Contains(lockContentStr, "needs: test-output-pull-request-job-generation") {
		t.Error("Expected create_pull_request job to depend on main job")
	}

	t.Logf("Generated workflow content:\n%s", lockContentStr)
}

func TestOutputPullRequestDraftFalse(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-pr-draft-false-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-pull-request configuration with draft: false
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
safe-outputs:
  create-pull-request:
    title-prefix: "[agent] "
    labels: [automation]
    draft: false
---

# Test Output Pull Request with Draft False

This workflow tests the create_pull_request job generation with draft: false.
`

	testFile := filepath.Join(tmpDir, "test-output-pr-draft-false.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output pull-request draft: false: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	// Convert to string for easier testing
	lockContentStr := string(lockContent)

	// Verify create_pull_request job is present
	if !strings.Contains(lockContentStr, "create_pull_request:") {
		t.Error("Expected 'create_pull_request' job to be in generated workflow")
	}

	// Verify draft setting is false
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_DRAFT: \"false\"") {
		t.Error("Expected draft to be set to false when explicitly specified")
	}

	// Verify other expected environment variables are still present
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_TITLE_PREFIX: \"[agent] \"") {
		t.Error("Expected title prefix to be set as environment variable")
	}

	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_LABELS: \"automation\"") {
		t.Error("Expected automation label to be set as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContentStr)
}

func TestOutputPullRequestDraftTrue(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-pr-draft-true-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with create-pull-request configuration with draft: true
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
safe-outputs:
  create-pull-request:
    title-prefix: "[agent] "
    labels: [automation]
    draft: true
---

# Test Output Pull Request with Draft True

This workflow tests the create_pull_request job generation with draft: true.
`

	testFile := filepath.Join(tmpDir, "test-output-pr-draft-true.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output pull-request draft: true: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	// Convert to string for easier testing
	lockContentStr := string(lockContent)

	// Verify create_pull_request job is present
	if !strings.Contains(lockContentStr, "create_pull_request:") {
		t.Error("Expected 'create_pull_request' job to be in generated workflow")
	}

	// Verify draft setting is true
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_DRAFT: \"true\"") {
		t.Error("Expected draft to be set to true when explicitly specified")
	}

	// Verify other expected environment variables are still present
	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_TITLE_PREFIX: \"[agent] \"") {
		t.Error("Expected title prefix to be set as environment variable")
	}

	if !strings.Contains(lockContentStr, "GITHUB_AW_PR_LABELS: \"automation\"") {
		t.Error("Expected automation label to be set as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContentStr)
}

func TestOutputLabelConfigParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement, needs-review]
---

# Test Output Label Configuration

This workflow tests the output labels configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-labels.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with output labels config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueLabels == nil {
		t.Fatal("Expected labels configuration to be parsed")
	}

	// Verify allowed labels
	expectedLabels := []string{"triage", "bug", "enhancement", "needs-review"}
	if len(workflowData.SafeOutputs.AddIssueLabels.Allowed) != len(expectedLabels) {
		t.Errorf("Expected %d allowed labels, got %d", len(expectedLabels), len(workflowData.SafeOutputs.AddIssueLabels.Allowed))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.SafeOutputs.AddIssueLabels.Allowed) || workflowData.SafeOutputs.AddIssueLabels.Allowed[i] != expectedLabel {
			t.Errorf("Expected label[%d] to be '%s', got '%s'", i, expectedLabel, workflowData.SafeOutputs.AddIssueLabels.Allowed[i])
		}
	}
}

func TestOutputLabelJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
tools:
  github:
    allowed: [get_issue]
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement]
---

# Test Output Label Job Generation

This workflow tests the add_labels job generation.
`

	testFile := filepath.Join(tmpDir, "test-output-labels.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output labels: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-output-labels.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify add_labels job exists
	if !strings.Contains(lockContent, "add_labels:") {
		t.Error("Expected 'add_labels' job to be in generated workflow")
	}

	// Verify job properties
	if !strings.Contains(lockContent, "timeout-minutes: 10") {
		t.Error("Expected 10-minute timeout in add_labels job")
	}

	if !strings.Contains(lockContent, "permissions:\n      contents: read\n      issues: write\n      pull-requests: write") {
		t.Error("Expected correct permissions in add_labels job")
	}

	// Verify the job uses github-script
	if !strings.Contains(lockContent, "uses: actions/github-script@v7") {
		t.Error("Expected github-script action to be used in add_labels job")
	}

	// Verify job has conditional execution
	if !strings.Contains(lockContent, "if: github.event.issue.number || github.event.pull_request.number") {
		t.Error("Expected add_labels job to have conditional execution")
	}

	// Verify job dependencies
	if !strings.Contains(lockContent, "needs: test-output-label-job-generation") {
		t.Error("Expected add_labels job to depend on main job")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContent, "GITHUB_AW_AGENT_OUTPUT:") {
		t.Error("Expected agent output content to be passed as environment variable")
	}

	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_ALLOWED: \"triage,bug,enhancement\"") {
		t.Error("Expected allowed labels to be set as environment variable")
	}

	// Verify output variables
	if !strings.Contains(lockContent, "labels_added: ${{ steps.add_labels.outputs.labels_added }}") {
		t.Error("Expected labels_added output to be available")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputLabelJobGenerationNoAllowedLabels(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-no-allowed-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test workflow with no allowed labels (any labels permitted)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
engine: claude
safe-outputs:
  add-issue-label:
    max: 5
---

# Test Output Label No Allowed Labels

This workflow tests label addition with no allowed labels restriction.
Write your labels to ${{ env.GITHUB_AW_SAFE_OUTPUTS }}, one per line.
`

	testFile := filepath.Join(tmpDir, "test-label-no-allowed.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Failed to compile workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockBytes, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatal(err)
	}
	lockContent := string(lockBytes)

	// Verify job has conditional execution
	if !strings.Contains(lockContent, "if: github.event.issue.number || github.event.pull_request.number") {
		t.Error("Expected add_labels job to have conditional execution")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContent, "GITHUB_AW_AGENT_OUTPUT:") {
		t.Error("Expected agent output content to be passed as environment variable")
	}

	// Verify empty allowed labels environment variable is set
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_ALLOWED: \"\"") {
		t.Error("Expected empty allowed labels to be set as environment variable")
	}

	// Verify max is set correctly
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_MAX_COUNT: 5") {
		t.Error("Expected max to be set correctly")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputLabelJobGenerationNullConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-null-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test workflow with null add-issue-label configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
engine: claude
safe-outputs:
  add-issue-label:
---

# Test Output Label Null Config

This workflow tests label addition with null configuration (any labels allowed).
Write your labels to ${{ env.GITHUB_AW_SAFE_OUTPUTS }}, one per line.
`

	testFile := filepath.Join(tmpDir, "test-label-null-config.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Failed to compile workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockBytes, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatal(err)
	}
	lockContent := string(lockBytes)

	// Verify add_labels job exists
	if !strings.Contains(lockContent, "add_labels:") {
		t.Error("Expected 'add_labels' job to be in generated workflow")
	}

	// Verify job has conditional execution
	if !strings.Contains(lockContent, "if: github.event.issue.number || github.event.pull_request.number") {
		t.Error("Expected add_labels job to have conditional execution")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContent, "GITHUB_AW_AGENT_OUTPUT:") {
		t.Error("Expected agent output content to be passed as environment variable")
	}

	// Verify empty allowed labels environment variable is set
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_ALLOWED: \"\"") {
		t.Error("Expected empty allowed labels to be set as environment variable")
	}

	// Verify default max is set correctly
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_MAX_COUNT: 3") {
		t.Error("Expected default max to be set correctly")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputLabelConfigNullParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-null-parsing-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with null add-issue-label configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-label:
---

# Test Output Label Null Configuration Parsing

This workflow tests the output labels null configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-labels-null.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with null labels config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueLabels == nil {
		t.Fatal("Expected labels configuration to be parsed (not nil)")
	}

	// Verify allowed labels is empty (no restrictions)
	if len(workflowData.SafeOutputs.AddIssueLabels.Allowed) != 0 {
		t.Errorf("Expected 0 allowed labels for null config, got %d", len(workflowData.SafeOutputs.AddIssueLabels.Allowed))
	}

	// Verify max is nil (will use default)
	if workflowData.SafeOutputs.AddIssueLabels.MaxCount != nil {
		t.Errorf("Expected max to be nil for null config, got %d", *workflowData.SafeOutputs.AddIssueLabels.MaxCount)
	}
}

func TestOutputLabelConfigMaxCountParsing(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-max-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration including max
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement, needs-review]
    max: 5
---

# Test Output Label Max Count Configuration

This workflow tests the output labels max configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-labels-max.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with output labels max config: %v", err)
	}

	// Verify output configuration is parsed correctly
	if workflowData.SafeOutputs == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.SafeOutputs.AddIssueLabels == nil {
		t.Fatal("Expected labels configuration to be parsed")
	}

	// Verify allowed labels
	expectedLabels := []string{"triage", "bug", "enhancement", "needs-review"}
	if len(workflowData.SafeOutputs.AddIssueLabels.Allowed) != len(expectedLabels) {
		t.Errorf("Expected %d allowed labels, got %d", len(expectedLabels), len(workflowData.SafeOutputs.AddIssueLabels.Allowed))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.SafeOutputs.AddIssueLabels.Allowed) || workflowData.SafeOutputs.AddIssueLabels.Allowed[i] != expectedLabel {
			t.Errorf("Expected label[%d] to be '%s', got '%s'", i, expectedLabel, workflowData.SafeOutputs.AddIssueLabels.Allowed[i])
		}
	}

	// Verify max
	if workflowData.SafeOutputs.AddIssueLabels.MaxCount == nil {
		t.Fatal("Expected max to be parsed")
	}

	expectedMaxCount := 5
	if *workflowData.SafeOutputs.AddIssueLabels.MaxCount != expectedMaxCount {
		t.Errorf("Expected max to be %d, got %d", expectedMaxCount, *workflowData.SafeOutputs.AddIssueLabels.MaxCount)
	}
}

func TestOutputLabelConfigDefaultMaxCount(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-default-max-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration without max (should use default)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement]
---

# Test Output Label Default Max Count

This workflow tests the default max behavior.
`

	testFile := filepath.Join(tmpDir, "test-output-labels-default.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData, err := compiler.parseWorkflowFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow without max: %v", err)
	}

	// Verify max is nil (will use default in job generation)
	if workflowData.SafeOutputs.AddIssueLabels.MaxCount != nil {
		t.Errorf("Expected max to be nil (default), got %d", *workflowData.SafeOutputs.AddIssueLabels.MaxCount)
	}
}

func TestOutputLabelJobGenerationWithMaxCount(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-job-max-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration including max
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
tools:
  github:
    allowed: [get_issue]
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement]
    max: 2
---

# Test Output Label Job Generation with Max Count

This workflow tests the add_labels job generation with max.
`

	testFile := filepath.Join(tmpDir, "test-output-labels-max.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output labels max: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-output-labels-max.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify add_labels job exists
	if !strings.Contains(lockContent, "add_labels:") {
		t.Error("Expected 'add_labels' job to be in generated workflow")
	}

	// Verify JavaScript content includes environment variables for configuration
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_ALLOWED: \"triage,bug,enhancement\"") {
		t.Error("Expected allowed labels to be set as environment variable")
	}

	// Verify max environment variable is set
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_MAX_COUNT: 2") {
		t.Error("Expected max to be set as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputLabelJobGenerationWithDefaultMaxCount(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-job-default-max-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with add-issue-label configuration without max (should use default of 3)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
tools:
  github:
    allowed: [get_issue]
engine: claude
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement]
---

# Test Output Label Job Generation with Default Max Count

This workflow tests the add_labels job generation with default max.
`

	testFile := filepath.Join(tmpDir, "test-output-labels-default-max.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow with output labels default max: %v", err)
	}

	// Read the generated lock file
	lockFile := filepath.Join(tmpDir, "test-output-labels-default-max.lock.yml")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	lockContent := string(content)

	// Verify add_labels job exists
	if !strings.Contains(lockContent, "add_labels:") {
		t.Error("Expected 'add_labels' job to be in generated workflow")
	}

	// Verify max environment variable is set to default value of 3
	if !strings.Contains(lockContent, "GITHUB_AW_LABELS_MAX_COUNT: 3") {
		t.Error("Expected max to be set to default value of 3 as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputLabelConfigValidation(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-validation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with empty allowed labels (should fail)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
engine: claude
safe-outputs:
  add-issue-label:
    allowed: []
---

# Test Output Label Validation

This workflow tests validation of empty allowed labels.
`

	testFile := filepath.Join(tmpDir, "test-label-validation.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow - should fail with empty allowed labels
	err = compiler.CompileWorkflow(testFile)
	if err == nil {
		t.Fatal("Expected error when compiling workflow with empty allowed labels")
	}

	if !strings.Contains(err.Error(), "minItems: got 0, want 1") {
		t.Errorf("Expected schema validation error about minItems, got: %v", err)
	}
}

func TestOutputLabelConfigMissingAllowed(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-label-missing-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with missing allowed field (should now succeed)
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
engine: claude
safe-outputs:
  add-issue-label: {}
---

# Test Output Label Missing Allowed

This workflow tests that missing allowed field is now optional.
`

	testFile := filepath.Join(tmpDir, "test-label-missing.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow - should now succeed with missing allowed labels
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected compilation to succeed with missing allowed labels, got error: %v", err)
	}

	// Verify the workflow was compiled successfully
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Fatal("Expected lock file to be created")
	}
}
