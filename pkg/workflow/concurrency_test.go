package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConcurrencyRules(t *testing.T) {
	// Test the new concurrency rules for pull_request and alias workflows
	tmpDir, err := os.MkdirTemp("", "concurrency-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	tests := []struct {
		name                string
		frontmatter         string
		filename            string
		expectedConcurrency string
		shouldHaveCancel    bool
		description         string
	}{
		{
			name: "PR workflow should have dynamic concurrency with cancel",
			frontmatter: `---
on:
  pull_request:
    types: [opened, edited]
tools:
  github:
    allowed: [list_issues]
---`,
			filename: "pr-workflow.md",
			expectedConcurrency: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}"
  cancel-in-progress: true`,
			shouldHaveCancel: true,
			description:      "PR workflows should use dynamic concurrency with PR number and cancellation",
		},
		{
			name: "command workflow should have dynamic concurrency without cancel",
			frontmatter: `---
on:
  command:
    name: test-bot
tools:
  github:
    allowed: [list_issues]
---`,
			filename: "command-workflow.md",
			expectedConcurrency: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number || github.event.pull_request.number }}"`,
			shouldHaveCancel: false,
			description:      "Alias workflows should use dynamic concurrency with ref but without cancellation",
		},
		{
			name: "regular workflow should use static concurrency without cancel",
			frontmatter: `---
on:
  schedule:
    - cron: "0 9 * * 1"
tools:
  github:
    allowed: [list_issues]
---`,
			filename: "regular-workflow.md",
			expectedConcurrency: `concurrency:
  group: "gh-aw-${{ github.workflow }}"`,
			shouldHaveCancel: false,
			description:      "Regular workflows should use static concurrency without cancellation",
		},
		{
			name: "push workflow should use static concurrency without cancel",
			frontmatter: `---
on:
  push:
    branches: [main]
tools:
  github:
    allowed: [list_issues]
---`,
			filename: "push-workflow.md",
			expectedConcurrency: `concurrency:
  group: "gh-aw-${{ github.workflow }}"`,
			shouldHaveCancel: false,
			description:      "Push workflows should use static concurrency without cancellation",
		},
		{
			name: "issue workflow should have dynamic concurrency with issue number",
			frontmatter: `---
on:
  issues:
    types: [opened, edited]
tools:
  github:
    allowed: [list_issues]
---`,
			filename: "issue-workflow.md",
			expectedConcurrency: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number }}"`,
			shouldHaveCancel: false,
			description:      "Issue workflows should use dynamic concurrency with issue number but no cancellation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testContent := tt.frontmatter + `

# Test Concurrency Workflow

This is a test workflow for concurrency behavior.
`

			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Parse the workflow to get its data
			workflowData, err := compiler.parseWorkflowFile(testFile)
			if err != nil {
				t.Errorf("Failed to parse workflow: %v", err)
				return
			}

			t.Logf("Workflow: %s", tt.description)
			t.Logf("  On: %s", workflowData.On)
			t.Logf("  Concurrency: %s", workflowData.Concurrency)

			// Check that the concurrency field matches expected pattern
			if !strings.Contains(workflowData.Concurrency, "gh-aw-${{ github.workflow }}") {
				t.Errorf("Expected concurrency to use gh-aw-${{ github.workflow }}, got: %s", workflowData.Concurrency)
			}

			// Check for cancel-in-progress based on workflow type
			hasCancel := strings.Contains(workflowData.Concurrency, "cancel-in-progress: true")
			if tt.shouldHaveCancel && !hasCancel {
				t.Errorf("Expected cancel-in-progress: true for %s workflow, but not found in: %s", tt.name, workflowData.Concurrency)
			} else if !tt.shouldHaveCancel && hasCancel {
				t.Errorf("Did not expect cancel-in-progress: true for %s workflow, but found in: %s", tt.name, workflowData.Concurrency)
			}

			// For PR workflows, check for PR number inclusion; for alias workflows, check for issue/PR numbers; for issue workflows, check for issue number
			isPRWorkflow := strings.Contains(tt.name, "PR workflow")
			isAliasWorkflow := strings.Contains(tt.name, "alias workflow")
			isIssueWorkflow := strings.Contains(tt.name, "issue workflow")

			if isPRWorkflow {
				if !strings.Contains(workflowData.Concurrency, "github.event.pull_request.number") {
					t.Errorf("Expected concurrency to include github.event.pull_request.number for %s workflow, got: %s", tt.name, workflowData.Concurrency)
				}
			} else if isAliasWorkflow {
				if !strings.Contains(workflowData.Concurrency, "github.event.issue.number || github.event.pull_request.number") {
					t.Errorf("Expected concurrency to include issue/PR numbers for %s workflow, got: %s", tt.name, workflowData.Concurrency)
				}
			} else if isIssueWorkflow {
				if !strings.Contains(workflowData.Concurrency, "github.event.issue.number") {
					t.Errorf("Expected concurrency to include github.event.issue.number for %s workflow, got: %s", tt.name, workflowData.Concurrency)
				}
			} else {
				if strings.Contains(workflowData.Concurrency, "github.ref") {
					t.Errorf("Did not expect concurrency to include github.ref for %s workflow, got: %s", tt.name, workflowData.Concurrency)
				}
			}
		})
	}
}

func TestGenerateConcurrencyConfig(t *testing.T) {
	tests := []struct {
		name           string
		workflowData   *WorkflowData
		isAliasTrigger bool
		expected       string
		description    string
	}{
		{
			name: "PR workflow should have dynamic concurrency with cancel and PR number",
			workflowData: &WorkflowData{
				On: `on:
  pull_request:
    types: [opened, synchronize]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}"
  cancel-in-progress: true`,
			description: "PR workflows should use PR number or ref with cancellation",
		},
		{
			name: "Alias workflow should have dynamic concurrency without cancel",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited, reopened]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: true,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number || github.event.pull_request.number }}"`,
			description: "Alias workflows should use dynamic concurrency with ref but without cancellation",
		},
		{
			name: "Regular workflow should use static concurrency without cancel",
			workflowData: &WorkflowData{
				On: `on:
  schedule:
    - cron: "0 9 * * 1"`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}"`,
			description: "Regular workflows should use static concurrency without cancellation",
		},
		{
			name: "Issue workflow should have dynamic concurrency with issue number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number }}"`,
			description: "Issue workflows should use issue number without cancellation",
		},
		{
			name: "Issue comment workflow should have dynamic concurrency with issue number",
			workflowData: &WorkflowData{
				On: `on:
  issue_comment:
    types: [created, edited]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number }}"`,
			description: "Issue comment workflows should use issue number without cancellation",
		},
		{
			name: "Mixed issue and PR workflow should have dynamic concurrency with issue/PR number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]
  pull_request:
    types: [opened, synchronize]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number || github.event.pull_request.number }}"
  cancel-in-progress: true`,
			description: "Mixed workflows should use issue/PR number with cancellation enabled",
		},
		{
			name: "Discussion workflow should have dynamic concurrency with discussion number",
			workflowData: &WorkflowData{
				On: `on:
  discussion:
    types: [created, edited]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.discussion.number }}"`,
			description: "Discussion workflows should use discussion number without cancellation",
		},
		{
			name: "Mixed issue and discussion workflow should have dynamic concurrency with issue/discussion number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]
  discussion:
    types: [created, edited]`,
				Concurrency: "", // Empty, should be generated
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "gh-aw-${{ github.workflow }}-${{ github.event.issue.number || github.event.discussion.number }}"`,
			description: "Mixed issue and discussion workflows should use issue/discussion number without cancellation",
		},
		{
			name: "Existing concurrency should not be overridden",
			workflowData: &WorkflowData{
				On: `on:
  pull_request:
    types: [opened, synchronize]`,
				Concurrency: `concurrency:
  group: "custom-group"`,
			},
			isAliasTrigger: false,
			expected: `concurrency:
  group: "custom-group"`,
			description: "Existing concurrency configuration should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateConcurrencyConfig(tt.workflowData, tt.isAliasTrigger)

			if result != tt.expected {
				t.Errorf("GenerateConcurrencyConfig() failed for %s\nExpected:\n%s\nGot:\n%s", tt.description, tt.expected, result)
			}
		})
	}
}

func TestIsPullRequestWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		on       string
		expected bool
	}{
		{
			name: "Pull request workflow should be identified",
			on: `on:
  pull_request:
    types: [opened, synchronize]`,
			expected: true,
		},
		{
			name: "Pull request review comment workflow should be identified",
			on: `on:
  pull_request_review_comment:
    types: [created]`,
			expected: true,
		},
		{
			name: "Schedule workflow should not be identified as PR workflow",
			on: `on:
  schedule:
    - cron: "0 9 * * 1"`,
			expected: false,
		},
		{
			name: "Issues workflow should not be identified as PR workflow",
			on: `on:
  issues:
    types: [opened, edited]`,
			expected: false,
		},
		{
			name: "Mixed workflow with PR should be identified",
			on: `on:
  issues:
    types: [opened, edited]
  pull_request:
    types: [opened, synchronize]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPullRequestWorkflow(tt.on)
			if result != tt.expected {
				t.Errorf("isPullRequestWorkflow() for %s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsIssueWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		on       string
		expected bool
	}{
		{
			name: "Issues workflow should be identified",
			on: `on:
  issues:
    types: [opened, edited]`,
			expected: true,
		},
		{
			name: "Issue comment workflow should be identified",
			on: `on:
  issue_comment:
    types: [created]`,
			expected: true,
		},
		{
			name: "Pull request workflow should not be identified as issue workflow",
			on: `on:
  pull_request:
    types: [opened, synchronize]`,
			expected: false,
		},
		{
			name: "Schedule workflow should not be identified as issue workflow",
			on: `on:
  schedule:
    - cron: "0 9 * * 1"`,
			expected: false,
		},
		{
			name: "Mixed workflow with issues should be identified",
			on: `on:
  issues:
    types: [opened, edited]
  push:
    branches: [main]`,
			expected: true,
		},
		{
			name: "Mixed workflow with issue_comment should be identified",
			on: `on:
  issue_comment:
    types: [created]
  schedule:
    - cron: "0 9 * * 1"`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIssueWorkflow(tt.on)
			if result != tt.expected {
				t.Errorf("isIssueWorkflow() for %s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsDiscussionWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		on       string
		expected bool
	}{
		{
			name: "Discussion workflow should be identified",
			on: `on:
  discussion:
    types: [created, edited]`,
			expected: true,
		},
		{
			name: "Discussion comment workflow should be identified",
			on: `on:
  discussion_comment:
    types: [created]`,
			expected: true,
		},
		{
			name: "Issues workflow should not be identified as discussion workflow",
			on: `on:
  issues:
    types: [opened, edited]`,
			expected: false,
		},
		{
			name: "Mixed workflow with discussion should be identified",
			on: `on:
  discussion:
    types: [created, edited]
  push:
    branches: [main]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDiscussionWorkflow(tt.on)
			if result != tt.expected {
				t.Errorf("isDiscussionWorkflow() for %s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestBuildConcurrencyGroupKeys(t *testing.T) {
	tests := []struct {
		name           string
		workflowData   *WorkflowData
		isAliasTrigger bool
		expected       []string
		description    string
	}{
		{
			name: "Alias workflow should include issue/PR number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]`,
			},
			isAliasTrigger: true,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.issue.number || github.event.pull_request.number }}"},
			description:    "Alias workflows should use issue/PR number",
		},
		{
			name: "Pure PR workflow should include PR number",
			workflowData: &WorkflowData{
				On: `on:
  pull_request:
    types: [opened, synchronize]`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.pull_request.number || github.ref }}"},
			description:    "Pure PR workflows should use PR number",
		},
		{
			name: "Pure issue workflow should include issue number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.issue.number }}"},
			description:    "Pure issue workflows should use issue number",
		},
		{
			name: "Mixed issue and PR workflow should include issue/PR number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]
  pull_request:
    types: [opened, synchronize]`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.issue.number || github.event.pull_request.number }}"},
			description:    "Mixed workflows should use issue/PR number",
		},
		{
			name: "Pure discussion workflow should include discussion number",
			workflowData: &WorkflowData{
				On: `on:
  discussion:
    types: [created, edited]`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.discussion.number }}"},
			description:    "Pure discussion workflows should use discussion number",
		},
		{
			name: "Mixed issue and discussion workflow should include issue/discussion number",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]
  discussion:
    types: [created, edited]`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}", "${{ github.event.issue.number || github.event.discussion.number }}"},
			description:    "Mixed issue and discussion workflows should use issue/discussion number",
		},
		{
			name: "Other workflow should not include additional keys",
			workflowData: &WorkflowData{
				On: `on:
  schedule:
    - cron: "0 9 * * 1"`,
			},
			isAliasTrigger: false,
			expected:       []string{"gh-aw", "${{ github.workflow }}"},
			description:    "Other workflows should use just workflow name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConcurrencyGroupKeys(tt.workflowData, tt.isAliasTrigger)

			if len(result) != len(tt.expected) {
				t.Errorf("buildConcurrencyGroupKeys() for %s returned %d keys, expected %d", tt.description, len(result), len(tt.expected))
				return
			}

			for i, key := range result {
				if key != tt.expected[i] {
					t.Errorf("buildConcurrencyGroupKeys() for %s key[%d] = %s, expected %s", tt.description, i, key, tt.expected[i])
				}
			}
		})
	}
}

func TestShouldEnableCancelInProgress(t *testing.T) {
	tests := []struct {
		name           string
		workflowData   *WorkflowData
		isAliasTrigger bool
		expected       bool
		description    string
	}{
		{
			name: "Alias workflow should not enable cancellation",
			workflowData: &WorkflowData{
				On: `on:
  pull_request:
    types: [opened, synchronize]`,
			},
			isAliasTrigger: true,
			expected:       false,
			description:    "Alias workflows should never enable cancellation",
		},
		{
			name: "PR workflow should enable cancellation",
			workflowData: &WorkflowData{
				On: `on:
  pull_request:
    types: [opened, synchronize]`,
			},
			isAliasTrigger: false,
			expected:       true,
			description:    "PR workflows should enable cancellation",
		},
		{
			name: "Issue workflow should not enable cancellation",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]`,
			},
			isAliasTrigger: false,
			expected:       false,
			description:    "Issue workflows should not enable cancellation",
		},
		{
			name: "Mixed issue and PR workflow should enable cancellation",
			workflowData: &WorkflowData{
				On: `on:
  issues:
    types: [opened, edited]
  pull_request:
    types: [opened, synchronize]`,
			},
			isAliasTrigger: false,
			expected:       true,
			description:    "Mixed workflows with PR should enable cancellation",
		},
		{
			name: "Other workflow should not enable cancellation",
			workflowData: &WorkflowData{
				On: `on:
  schedule:
    - cron: "0 9 * * 1"`,
			},
			isAliasTrigger: false,
			expected:       false,
			description:    "Other workflows should not enable cancellation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldEnableCancelInProgress(tt.workflowData, tt.isAliasTrigger)
			if result != tt.expected {
				t.Errorf("shouldEnableCancelInProgress() for %s = %v, expected %v", tt.description, result, tt.expected)
			}
		})
	}
}
