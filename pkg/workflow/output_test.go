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

	// Test case with output.issue configuration
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  issue:
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
	if workflowData.Output == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.Output.Issue == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}

	// Verify title prefix
	expectedPrefix := "[genai] "
	if workflowData.Output.Issue.TitlePrefix != expectedPrefix {
		t.Errorf("Expected title prefix '%s', got '%s'", expectedPrefix, workflowData.Output.Issue.TitlePrefix)
	}

	// Verify labels
	expectedLabels := []string{"copilot", "automation"}
	if len(workflowData.Output.Issue.Labels) != len(expectedLabels) {
		t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(workflowData.Output.Issue.Labels))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.Output.Issue.Labels) || workflowData.Output.Issue.Labels[i] != expectedLabel {
			t.Errorf("Expected label '%s' at index %d, got '%s'", expectedLabel, i, workflowData.Output.Issue.Labels[i])
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

	// Verify output configuration is nil
	if workflowData.Output != nil {
		t.Error("Expected output configuration to be nil when not specified")
	}
}

func TestOutputIssueJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-issue-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with output.issue configuration
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
tools:
  github:
    allowed: [list_issues]
engine: claude
output:
  issue:
    title-prefix: "[genai] "
    labels: [copilot]
---

# Test Output Issue Job Generation

This workflow tests the create_issue job generation.
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

	// Test case with output.comment configuration
	testContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
output:
  comment: {}
---

# Test Output Comment Configuration

This workflow tests the output.comment configuration parsing.
`

	testFile := filepath.Join(tmpDir, "test-output-comment.md")
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
	if workflowData.Output == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.Output.Comment == nil {
		t.Fatal("Expected comment configuration to be parsed")
	}
}

func TestOutputCommentJobGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-job-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with output.comment configuration
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
output:
  comment: {}
---

# Test Output Comment Job Generation

This workflow tests the create_issue_comment job generation.
`

	testFile := filepath.Join(tmpDir, "test-output-comment.md")
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
	lockFile := filepath.Join(tmpDir, "test-output-comment.lock.yml")
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
	if !strings.Contains(lockContent, "needs: test-output-comment") {
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

	// Test case with output.comment configuration but push trigger (not issue/PR)
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
output:
  comment: {}
---

# Test Output Comment Job Skipping

This workflow tests that comment job is skipped for non-issue/PR events.
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

	// Test case with output.pull-request configuration
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
engine: claude
output:
  pull-request:
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
	if workflowData.Output == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData.Output.PullRequest == nil {
		t.Fatal("Expected pull-request configuration to be parsed")
	}

	// Verify title prefix
	expectedPrefix := "[agent] "
	if workflowData.Output.PullRequest.TitlePrefix != expectedPrefix {
		t.Errorf("Expected title prefix '%s', got '%s'", expectedPrefix, workflowData.Output.PullRequest.TitlePrefix)
	}

	// Verify labels
	expectedLabels := []string{"automation", "bot"}
	if len(workflowData.Output.PullRequest.Labels) != len(expectedLabels) {
		t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(workflowData.Output.PullRequest.Labels))
	}

	for i, expectedLabel := range expectedLabels {
		if i >= len(workflowData.Output.PullRequest.Labels) || workflowData.Output.PullRequest.Labels[i] != expectedLabel {
			t.Errorf("Expected label[%d] to be '%s', got '%s'", i, expectedLabel, workflowData.Output.PullRequest.Labels[i])
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

	// Test case with output.pull-request configuration
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
output:
  pull-request:
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
		t.Fatalf("Unexpected error compiling workflow with output pull-request: %v", err)
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

	// Test case with output.pull-request configuration with draft: false
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
output:
  pull-request:
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

	// Test case with output.pull-request configuration with draft: true
	testContent := `---
on: push
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    allowed: [list_issues]
engine: claude
output:
  pull-request:
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

func TestOutputIssueAllowHTMLConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-issue-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case 1: allow-html set to false
	testContent1 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  issue:
    title-prefix: "[test] "
    labels: [automation]
    allow-html: false
---

# Test Allow HTML Configuration

This workflow tests the allow-html configuration for issues.
`

	testFile1 := filepath.Join(tmpDir, "test-allow-html-false.md")
	if err := os.WriteFile(testFile1, []byte(testContent1), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Parse the workflow data
	workflowData1, err := compiler.parseWorkflowFile(testFile1)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with allow-html false: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData1.Output == nil || workflowData1.Output.Issue == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}

	if workflowData1.Output.Issue.AllowHTML == nil {
		t.Fatal("Expected allow-html to be parsed")
	}

	if *workflowData1.Output.Issue.AllowHTML != false {
		t.Errorf("Expected allow-html to be false, got %v", *workflowData1.Output.Issue.AllowHTML)
	}

	// Test case 2: allow-html set to true
	testContent2 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  issue:
    title-prefix: "[test] "
    labels: [automation]
    allow-html: true
---

# Test Allow HTML Configuration

This workflow tests the allow-html configuration for issues.
`

	testFile2 := filepath.Join(tmpDir, "test-allow-html-true.md")
	if err := os.WriteFile(testFile2, []byte(testContent2), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData2, err := compiler.parseWorkflowFile(testFile2)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with allow-html true: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData2.Output == nil || workflowData2.Output.Issue == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}

	if workflowData2.Output.Issue.AllowHTML == nil {
		t.Fatal("Expected allow-html to be parsed")
	}

	if *workflowData2.Output.Issue.AllowHTML != true {
		t.Errorf("Expected allow-html to be true, got %v", *workflowData2.Output.Issue.AllowHTML)
	}

	// Test case 3: allow-html not specified (should be nil)
	testContent3 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  issue:
    title-prefix: "[test] "
    labels: [automation]
---

# Test Allow HTML Configuration

This workflow tests the allow-html configuration for issues.
`

	testFile3 := filepath.Join(tmpDir, "test-allow-html-unspecified.md")
	if err := os.WriteFile(testFile3, []byte(testContent3), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData3, err := compiler.parseWorkflowFile(testFile3)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow without allow-html: %v", err)
	}

	// Verify allow-html is nil when not specified
	if workflowData3.Output == nil || workflowData3.Output.Issue == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}

	if workflowData3.Output.Issue.AllowHTML != nil {
		t.Errorf("Expected allow-html to be nil when not specified, got %v", *workflowData3.Output.Issue.AllowHTML)
	}
}

func TestOutputIssueJobGenerationWithAllowHTML(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-issue-job-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with allow-html: false
	testContent := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  issue:
    title-prefix: "[test] "
    labels: [automation]
    allow-html: false
---

# Test Allow HTML in Job Generation

This workflow tests the allow-html environment variable generation.
`

	testFile := filepath.Join(tmpDir, "test-allow-html-job.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow to generate the lock file
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Unexpected error compiling workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	// Convert to string for easier testing
	lockContentStr := string(lockContent)

	// Verify GITHUB_AW_ISSUE_ALLOW_HTML environment variable is present with correct value
	if !strings.Contains(lockContentStr, "GITHUB_AW_ISSUE_ALLOW_HTML: \"false\"") {
		t.Error("Expected GITHUB_AW_ISSUE_ALLOW_HTML to be set to false in generated workflow")
	}

	// Verify other expected environment variables are still present
	if !strings.Contains(lockContentStr, "GITHUB_AW_ISSUE_TITLE_PREFIX: \"[test] \"") {
		t.Error("Expected title prefix to be set as environment variable")
	}

	if !strings.Contains(lockContentStr, "GITHUB_AW_ISSUE_LABELS: \"automation\"") {
		t.Error("Expected automation label to be set as environment variable")
	}

	t.Logf("Generated workflow content:\n%s", lockContentStr)
}

func TestOutputCommentAllowHTMLConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	// Test case 1: allow-html: false
	testContent1 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  comment:
    allow-html: false
---

# Test Allow HTML Configuration for Comments

This workflow tests the allow-html configuration for comments.
`

	testFile1 := filepath.Join(tmpDir, "test-comment-allow-html-false.md")
	if err := os.WriteFile(testFile1, []byte(testContent1), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData1, err := compiler.parseWorkflowFile(testFile1)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with comment allow-html false: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData1.Output == nil || workflowData1.Output.Comment == nil {
		t.Fatal("Expected comment configuration to be parsed")
	}

	if workflowData1.Output.Comment.AllowHTML == nil {
		t.Fatal("Expected comment allow-html to be parsed")
	}

	if *workflowData1.Output.Comment.AllowHTML != false {
		t.Errorf("Expected comment allow-html to be false, got %v", *workflowData1.Output.Comment.AllowHTML)
	}

	// Test case 2: allow-html: true
	testContent2 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  comment:
    allow-html: true
---

# Test Allow HTML Configuration for Comments

This workflow tests the allow-html configuration for comments.
`

	testFile2 := filepath.Join(tmpDir, "test-comment-allow-html-true.md")
	if err := os.WriteFile(testFile2, []byte(testContent2), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData2, err := compiler.parseWorkflowFile(testFile2)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with comment allow-html true: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData2.Output == nil || workflowData2.Output.Comment == nil {
		t.Fatal("Expected comment configuration to be parsed")
	}

	if workflowData2.Output.Comment.AllowHTML == nil {
		t.Fatal("Expected comment allow-html to be parsed")
	}

	if *workflowData2.Output.Comment.AllowHTML != true {
		t.Errorf("Expected comment allow-html to be true, got %v", *workflowData2.Output.Comment.AllowHTML)
	}

	// Test case 3: allow-html not specified (should be nil)
	testContent3 := `---
on: push
permissions:
  contents: read
  issues: write
engine: claude
output:
  comment: {}
---

# Test Allow HTML Configuration for Comments

This workflow tests the allow-html configuration for comments.
`

	testFile3 := filepath.Join(tmpDir, "test-comment-allow-html-unspecified.md")
	if err := os.WriteFile(testFile3, []byte(testContent3), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData3, err := compiler.parseWorkflowFile(testFile3)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow without comment allow-html: %v", err)
	}

	// Verify allow-html is nil when not specified
	if workflowData3.Output == nil || workflowData3.Output.Comment == nil {
		t.Fatal("Expected comment configuration to be parsed")
	}

	if workflowData3.Output.Comment.AllowHTML != nil {
		t.Errorf("Expected comment allow-html to be nil when not specified, got %v", *workflowData3.Output.Comment.AllowHTML)
	}
}

func TestOutputPullRequestAllowHTMLConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-pr-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	// Test case 1: allow-html: false
	testContent1 := `---
on: push
permissions:
  contents: read
  pull-requests: write
engine: claude
output:
  pull-request:
    title-prefix: "[agent] "
    labels: [automation]
    allow-html: false
---

# Test Allow HTML Configuration for Pull Requests

This workflow tests the allow-html configuration for pull requests.
`

	testFile1 := filepath.Join(tmpDir, "test-pr-allow-html-false.md")
	if err := os.WriteFile(testFile1, []byte(testContent1), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData1, err := compiler.parseWorkflowFile(testFile1)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with pull-request allow-html false: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData1.Output == nil || workflowData1.Output.PullRequest == nil {
		t.Fatal("Expected pull-request configuration to be parsed")
	}

	if workflowData1.Output.PullRequest.AllowHTML == nil {
		t.Fatal("Expected pull-request allow-html to be parsed")
	}

	if *workflowData1.Output.PullRequest.AllowHTML != false {
		t.Errorf("Expected pull-request allow-html to be false, got %v", *workflowData1.Output.PullRequest.AllowHTML)
	}

	// Test case 2: allow-html: true
	testContent2 := `---
on: push
permissions:
  contents: read
  pull-requests: write
engine: claude
output:
  pull-request:
    title-prefix: "[agent] "
    labels: [automation]
    allow-html: true
---

# Test Allow HTML Configuration for Pull Requests

This workflow tests the allow-html configuration for pull requests.
`

	testFile2 := filepath.Join(tmpDir, "test-pr-allow-html-true.md")
	if err := os.WriteFile(testFile2, []byte(testContent2), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData2, err := compiler.parseWorkflowFile(testFile2)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with pull-request allow-html true: %v", err)
	}

	// Verify allow-html is parsed correctly
	if workflowData2.Output == nil || workflowData2.Output.PullRequest == nil {
		t.Fatal("Expected pull-request configuration to be parsed")
	}

	if workflowData2.Output.PullRequest.AllowHTML == nil {
		t.Fatal("Expected pull-request allow-html to be parsed")
	}

	if *workflowData2.Output.PullRequest.AllowHTML != true {
		t.Errorf("Expected pull-request allow-html to be true, got %v", *workflowData2.Output.PullRequest.AllowHTML)
	}

	// Verify other fields are still parsed
	expectedPrefix := "[agent] "
	if workflowData2.Output.PullRequest.TitlePrefix != expectedPrefix {
		t.Errorf("Expected title prefix '%s', got '%s'", expectedPrefix, workflowData2.Output.PullRequest.TitlePrefix)
	}

	expectedLabels := []string{"automation"}
	if len(workflowData2.Output.PullRequest.Labels) != len(expectedLabels) {
		t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(workflowData2.Output.PullRequest.Labels))
	}
}

func TestOutputCommentJobGenerationWithAllowHTML(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-comment-job-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case with allow-html: false
	testContent := `---
on: issues
permissions:
  contents: read
  issues: write
engine: claude
output:
  comment:
    allow-html: false
---

# Test Comment Allow HTML in Job Generation

This workflow tests the allow-html environment variable generation for comments.
`

	testFile := filepath.Join(tmpDir, "test-comment-allow-html-job.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler(false, "", "test")

	// Compile the workflow to generate the lock file
	err = compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Failed to compile workflow: %v", err)
	}

	// Read the generated lock file
	lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read generated lock file: %v", err)
	}

	// Check that the environment variable is set correctly
	if !strings.Contains(string(lockContent), "GITHUB_AW_COMMENT_ALLOW_HTML: \"false\"") {
		t.Error("Expected GITHUB_AW_COMMENT_ALLOW_HTML environment variable to be set to 'false' in generated workflow")
	}

	t.Logf("Generated workflow content:\n%s", lockContent)
}

func TestOutputSharedAllowHTMLConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-shared-allow-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	// Test case 1: shared allow-html: false with individual overrides
	testContent1 := `---
on: push
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
output:
  allow-html: false
  issue:
    title-prefix: "[test] "
    labels: [automation]
    allow-html: true  # Override shared setting
  comment: {}  # Use shared setting (false)
  pull-request:
    title-prefix: "[agent] "
    labels: [bot]
    # No allow-html specified, should use shared setting (false)
---

# Test Shared Allow HTML Configuration

This workflow tests the shared allow-html configuration with individual overrides.
`

	testFile1 := filepath.Join(tmpDir, "test-shared-allow-html.md")
	if err := os.WriteFile(testFile1, []byte(testContent1), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData1, err := compiler.parseWorkflowFile(testFile1)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with shared allow-html: %v", err)
	}

	// Verify shared allow-html is parsed correctly
	if workflowData1.Output == nil {
		t.Fatal("Expected output configuration to be parsed")
	}

	if workflowData1.Output.AllowHTML == nil {
		t.Fatal("Expected shared allow-html to be parsed")
	}

	if *workflowData1.Output.AllowHTML != false {
		t.Errorf("Expected shared allow-html to be false, got %v", *workflowData1.Output.AllowHTML)
	}

	// Verify issue config overrides shared setting
	if workflowData1.Output.Issue == nil {
		t.Fatal("Expected issue configuration to be parsed")
	}
	if workflowData1.Output.Issue.AllowHTML == nil || *workflowData1.Output.Issue.AllowHTML != true {
		t.Errorf("Expected issue allow-html to override shared setting to true, got %v", workflowData1.Output.Issue.AllowHTML)
	}

	// Verify comment config uses shared setting
	if workflowData1.Output.Comment == nil {
		t.Fatal("Expected comment configuration to be parsed")
	}
	if workflowData1.Output.Comment.AllowHTML == nil || *workflowData1.Output.Comment.AllowHTML != false {
		t.Errorf("Expected comment allow-html to use shared setting (false), got %v", workflowData1.Output.Comment.AllowHTML)
	}

	// Verify pull-request config uses shared setting
	if workflowData1.Output.PullRequest == nil {
		t.Fatal("Expected pull-request configuration to be parsed")
	}
	if workflowData1.Output.PullRequest.AllowHTML == nil || *workflowData1.Output.PullRequest.AllowHTML != false {
		t.Errorf("Expected pull-request allow-html to use shared setting (false), got %v", workflowData1.Output.PullRequest.AllowHTML)
	}

	// Test case 2: shared allow-html: true with no individual overrides
	testContent2 := `---
on: push
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: claude
output:
  allow-html: true
  issue:
    title-prefix: "[test] "
    labels: [automation]
  comment: {}
  pull-request:
    title-prefix: "[agent] "
    labels: [bot]
---

# Test Shared Allow HTML True

This workflow tests that all output types inherit shared allow-html: true.
`

	testFile2 := filepath.Join(tmpDir, "test-shared-allow-html-true.md")
	if err := os.WriteFile(testFile2, []byte(testContent2), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the workflow data
	workflowData2, err := compiler.parseWorkflowFile(testFile2)
	if err != nil {
		t.Fatalf("Unexpected error parsing workflow with shared allow-html true: %v", err)
	}

	// Verify all configs inherit shared setting
	if workflowData2.Output.Issue.AllowHTML == nil || *workflowData2.Output.Issue.AllowHTML != true {
		t.Errorf("Expected issue allow-html to inherit shared setting (true), got %v", workflowData2.Output.Issue.AllowHTML)
	}
	if workflowData2.Output.Comment.AllowHTML == nil || *workflowData2.Output.Comment.AllowHTML != true {
		t.Errorf("Expected comment allow-html to inherit shared setting (true), got %v", workflowData2.Output.Comment.AllowHTML)
	}
	if workflowData2.Output.PullRequest.AllowHTML == nil || *workflowData2.Output.PullRequest.AllowHTML != true {
		t.Errorf("Expected pull-request allow-html to inherit shared setting (true), got %v", workflowData2.Output.PullRequest.AllowHTML)
	}
}
