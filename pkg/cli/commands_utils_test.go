package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractWorkflowNameFromFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		filename    string
		expected    string
		expectError bool
	}{
		{
			name: "file with H1 header",
			content: `---
title: Test Workflow
---

# Daily Test Coverage Improvement

This is a test workflow.`,
			filename:    "test-workflow.md",
			expected:    "Daily Test Coverage Improvement",
			expectError: false,
		},
		{
			name: "file with H1 header with extra spaces",
			content: `# Weekly Research   

This is a research workflow.`,
			filename:    "weekly-research.md",
			expected:    "Weekly Research",
			expectError: false,
		},
		{
			name: "file without H1 header - generates from filename",
			content: `This is content without H1 header.

## Some H2 header

Content here.`,
			filename:    "daily-dependency-updates.md",
			expected:    "Daily Dependency Updates",
			expectError: false,
		},
		{
			name:        "file with complex filename",
			content:     `No headers here.`,
			filename:    "complex-workflow-name-test.md",
			expected:    "Complex Workflow Name Test",
			expectError: false,
		},
		{
			name:        "file with single word filename",
			content:     `No headers.`,
			filename:    "workflow.md",
			expected:    "Workflow",
			expectError: false,
		},
		{
			name:        "empty file - generates from filename",
			content:     "",
			filename:    "empty-workflow.md",
			expected:    "Empty Workflow",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test the function
			result, err := extractWorkflowNameFromFile(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestExtractWorkflowNameFromFile_NonExistentFile(t *testing.T) {
	_, err := extractWorkflowNameFromFile("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestUpdateWorkflowTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		number   int
		expected string
	}{
		{
			name: "content with H1 header",
			content: `---
title: Test
---

# Daily Test Coverage

This is a workflow.`,
			number: 2,
			expected: `---
title: Test
---

# Daily Test Coverage 2

This is a workflow.`,
		},
		{
			name: "content with H1 header with extra spaces",
			content: `   # Weekly Research   

Content here.`,
			number: 3,
			expected: `# # Weekly Research 3

Content here.`,
		},
		{
			name: "content without H1 header",
			content: `## H2 Header

Content without H1.`,
			number: 1,
			expected: `## H2 Header

Content without H1.`,
		},
		{
			name:     "empty content",
			content:  "",
			number:   1,
			expected: "",
		},
		{
			name: "multiple H1 headers - only first is modified",
			content: `# First Header

Some content.

# Second Header

More content.`,
			number: 5,
			expected: `# First Header 5

Some content.

# Second Header

More content.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateWorkflowTitle(tt.content, tt.number)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	// Test in current directory (should be a git repo based on project setup)
	result := isGitRepo()

	// Since we're running in a git repository, this should return true
	if !result {
		t.Error("Expected isGitRepo() to return true in git repository")
	}
}

// TestFindGitRoot is already tested in gitroot_test.go, skipping duplicate

func TestParseRepoSpec(t *testing.T) {
	tests := []struct {
		name            string
		repoSpec        string
		expectedRepo    string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "simple org/repo",
			repoSpec:        "githubnext/gh-aw",
			expectedRepo:    "githubnext/gh-aw",
			expectedVersion: "",
			expectError:     false,
		},
		{
			name:            "org/repo with version",
			repoSpec:        "githubnext/gh-aw@v1.0.0",
			expectedRepo:    "githubnext/gh-aw",
			expectedVersion: "v1.0.0",
			expectError:     false,
		},
		{
			name:            "org/repo with branch",
			repoSpec:        "githubnext/gh-aw@main",
			expectedRepo:    "githubnext/gh-aw",
			expectedVersion: "main",
			expectError:     false,
		},
		{
			name:            "invalid format - no slash",
			repoSpec:        "invalid-repo",
			expectedRepo:    "",
			expectedVersion: "",
			expectError:     true,
		},
		{
			name:            "invalid format - empty org",
			repoSpec:        "/repo",
			expectedRepo:    "",
			expectedVersion: "",
			expectError:     true,
		},
		{
			name:            "invalid format - empty repo",
			repoSpec:        "org/",
			expectedRepo:    "",
			expectedVersion: "",
			expectError:     true,
		},
		{
			name:            "empty repo spec",
			repoSpec:        "",
			expectedRepo:    "",
			expectedVersion: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, version, err := parseRepoSpec(tt.repoSpec)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if repo != tt.expectedRepo {
					t.Errorf("Expected repo %q, got %q", tt.expectedRepo, repo)
				}
				if version != tt.expectedVersion {
					t.Errorf("Expected version %q, got %q", tt.expectedVersion, version)
				}
			}
		})
	}
}

func TestExtractWorkflowNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple workflow file",
			path:     ".github/workflows/daily-test.lock.yml",
			expected: "daily-test",
		},
		{
			name:     "workflow file without lock suffix",
			path:     ".github/workflows/weekly-research.yml",
			expected: "weekly-research",
		},
		{
			name:     "nested path",
			path:     "/home/user/project/.github/workflows/complex-workflow-name.lock.yml",
			expected: "complex-workflow-name",
		},
		{
			name:     "file without extension",
			path:     ".github/workflows/workflow",
			expected: "workflow",
		},
		{
			name:     "single file name",
			path:     "test.yml",
			expected: "test",
		},
		{
			name:     "file with multiple dots",
			path:     "test.lock.yml",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWorkflowNameFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFindIncludesInContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no includes",
			content:  "This is regular content without includes.",
			expected: []string{},
		},
		{
			name: "single include",
			content: `This is content with include:
@include shared/tools.md
More content here.`,
			expected: []string{"shared/tools.md"},
		},
		{
			name: "multiple includes",
			content: `Content with multiple includes:
@include shared/tools.md
Some content between.
@include shared/config.md
More content.
@include another/file.md`,
			expected: []string{"shared/tools.md", "shared/config.md", "another/file.md"},
		},
		{
			name: "includes with different whitespace",
			content: `Content:
@include shared/tools.md
@include  shared/config.md  
@include	shared/tabs.md`,
			expected: []string{"shared/tools.md", "shared/config.md", "shared/tabs.md"},
		},
		{
			name: "includes with section references",
			content: `Content:
@include shared/tools.md#Tools
@include shared/config.md#Configuration`,
			expected: []string{"shared/tools.md", "shared/config.md"},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findIncludesInContent(tt.content, "", false)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d includes, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected include %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkExtractWorkflowNameFromFile(b *testing.B) {
	// Create temporary test file
	tmpDir := b.TempDir()
	content := `---
title: Test Workflow
---

# Daily Test Coverage Improvement

This is a test workflow with some content.`

	filePath := filepath.Join(tmpDir, "test-workflow.md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractWorkflowNameFromFile(filePath)
	}
}

func BenchmarkUpdateWorkflowTitle(b *testing.B) {
	content := `---
title: Test
---

# Daily Test Coverage

This is a workflow with some content that needs title updating.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = updateWorkflowTitle(content, i+1)
	}
}

func BenchmarkFindIncludesInContent(b *testing.B) {
	content := `This is content with includes:
@include shared/tools.md
Some content between includes.
@include shared/config.md
More content here.
@include another/file.md
Final content.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = findIncludesInContent(content, "", false)
	}
}

func TestCopyMarkdownFiles(t *testing.T) {
	tests := []struct {
		name           string
		sourceFiles    map[string]string // path -> content
		expectedTarget map[string]string // relative path -> content
		verbose        bool
		expectError    bool
	}{
		{
			name: "copy single markdown file",
			sourceFiles: map[string]string{
				"workflow.md": `# Test Workflow
This is a test workflow.`,
			},
			expectedTarget: map[string]string{
				"workflow.md": `# Test Workflow
This is a test workflow.`,
			},
			verbose:     false,
			expectError: false,
		},
		{
			name: "copy multiple markdown files",
			sourceFiles: map[string]string{
				"daily.md": `# Daily Workflow
Daily tasks`,
				"weekly.md": `# Weekly Workflow
Weekly tasks`,
			},
			expectedTarget: map[string]string{
				"daily.md": `# Daily Workflow
Daily tasks`,
				"weekly.md": `# Weekly Workflow
Weekly tasks`,
			},
			verbose:     false,
			expectError: false,
		},
		{
			name: "copy markdown files in subdirectories",
			sourceFiles: map[string]string{
				"workflows/daily.md": `# Daily
Content`,
				"workflows/weekly.md": `# Weekly
Content`,
				"shared/utils.md": `# Utils
Shared content`,
			},
			expectedTarget: map[string]string{
				"workflows/daily.md": `# Daily
Content`,
				"workflows/weekly.md": `# Weekly
Content`,
				"shared/utils.md": `# Utils
Shared content`,
			},
			verbose:     true,
			expectError: false,
		},
		{
			name: "skip non-markdown files",
			sourceFiles: map[string]string{
				"workflow.md": `# Test Workflow`,
				"config.yaml": `name: test`,
				"readme.txt":  `This is a readme`,
				"script.sh":   `#!/bin/bash\necho "hello"`,
			},
			expectedTarget: map[string]string{
				"workflow.md": `# Test Workflow`,
			},
			verbose:     false,
			expectError: false,
		},
		{
			name: "handle empty source directory",
			sourceFiles: map[string]string{
				"not-markdown.txt": `This won't be copied`,
			},
			expectedTarget: map[string]string{},
			verbose:        false,
			expectError:    false,
		},
		{
			name: "copy nested markdown files with complex structure",
			sourceFiles: map[string]string{
				"level1/workflow1.md":               `# Level 1 Workflow 1`,
				"level1/level2/workflow2.md":        `# Level 2 Workflow 2`,
				"level1/level2/level3/workflow3.md": `# Level 3 Workflow 3`,
				"other.txt":                         `Not copied`,
			},
			expectedTarget: map[string]string{
				"level1/workflow1.md":               `# Level 1 Workflow 1`,
				"level1/level2/workflow2.md":        `# Level 2 Workflow 2`,
				"level1/level2/level3/workflow3.md": `# Level 3 Workflow 3`,
			},
			verbose:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary source and target directories
			sourceDir := t.TempDir()
			targetDir := t.TempDir()

			// Create source files
			for path, content := range tt.sourceFiles {
				fullPath := filepath.Join(sourceDir, path)
				// Create directory if needed
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create source directory %s: %v", dir, err)
				}
				// Write file
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create source file %s: %v", fullPath, err)
				}
			}

			// Test the function
			err := copyMarkdownFiles(sourceDir, targetDir, tt.verbose)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			// Verify expected files were copied
			for expectedPath, expectedContent := range tt.expectedTarget {
				fullTargetPath := filepath.Join(targetDir, expectedPath)

				// Check if file exists
				if _, err := os.Stat(fullTargetPath); os.IsNotExist(err) {
					t.Errorf("Expected file %s was not copied", expectedPath)
					continue
				}

				// Check file content
				content, err := os.ReadFile(fullTargetPath)
				if err != nil {
					t.Errorf("Failed to read copied file %s: %v", expectedPath, err)
					continue
				}

				if string(content) != expectedContent {
					t.Errorf("File %s content mismatch:\nExpected: %q\nGot: %q",
						expectedPath, expectedContent, string(content))
				}
			}

			// Verify no unexpected files were copied (check that only .md files exist)
			err = filepath.Walk(targetDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}

				if !info.IsDir() {
					relPath, err := filepath.Rel(targetDir, path)
					if err != nil {
						return err
					}

					// All files in target should be .md files
					if !strings.HasSuffix(relPath, ".md") {
						t.Errorf("Unexpected non-markdown file copied: %s", relPath)
					}
				}
				return nil
			})

			if err != nil {
				t.Errorf("Error walking target directory: %v", err)
			}
		})
	}
}

func TestCopyMarkdownFiles_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (sourceDir, targetDir string, cleanup func())
		expectError bool
		errorText   string
	}{
		{
			name: "nonexistent source directory",
			setup: func() (string, string, func()) {
				targetDir := t.TempDir()
				return "/nonexistent/source", targetDir, func() {}
			},
			expectError: true,
			errorText:   "no such file or directory",
		},
		{
			name: "permission denied on target directory",
			setup: func() (string, string, func()) {
				sourceDir := t.TempDir()
				targetDir := t.TempDir()

				// Create a source file
				sourceFile := filepath.Join(sourceDir, "test.md")
				os.WriteFile(sourceFile, []byte("# Test"), 0644)

				// Make target directory read-only
				os.Chmod(targetDir, 0444)

				cleanup := func() {
					os.Chmod(targetDir, 0755) // Restore permissions for cleanup
				}

				return sourceDir, targetDir, cleanup
			},
			expectError: true,
			errorText:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceDir, targetDir, cleanup := tt.setup()
			defer cleanup()

			err := copyMarkdownFiles(sourceDir, targetDir, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorText != "" && !containsIgnoreCase(err.Error(), tt.errorText) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestIsRunnable(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    bool
		expectError bool
	}{
		{
			name: "workflow with schedule trigger",
			content: `---
on:
  schedule:
    - cron: "0 9 * * *"
---
# Test Workflow
This workflow runs on schedule.`,
			expected:    true,
			expectError: false,
		},
		{
			name: "workflow with workflow_dispatch trigger",
			content: `---
on:
  workflow_dispatch:
---
# Manual Workflow
This workflow can be triggered manually.`,
			expected:    true,
			expectError: false,
		},
		{
			name: "workflow with both schedule and workflow_dispatch",
			content: `---
on:
  schedule:
    - cron: "0 9 * * 1"  
  workflow_dispatch:
  push:
    branches: [main]
---
# Mixed Triggers Workflow`,
			expected:    true,
			expectError: false,
		},
		{
			name: "workflow with only push trigger (not runnable)",
			content: `---
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
---
# CI Workflow
This is not runnable via schedule or manual dispatch.`,
			expected:    false,
			expectError: false,
		},
		{
			name: "workflow with no 'on' section (defaults to runnable)",
			content: `---
name: Default Workflow
---
# Default Workflow
No on section means it defaults to runnable.`,
			expected:    true,
			expectError: false,
		},
		{
			name: "workflow with cron trigger (alternative schedule format)",
			content: `---
on:
  cron: "0 */6 * * *"
---
# Cron Workflow
Uses cron format directly.`,
			expected:    true,
			expectError: false,
		},
		{
			name: "case insensitive schedule detection",
			content: `---
on:
  SCHEDULE:
    - cron: "0 12 * * 0"
---
# Case Test Workflow`,
			expected:    true,
			expectError: false,
		},
		{
			name: "case insensitive workflow_dispatch detection",
			content: `---
on:
  WORKFLOW_DISPATCH:
---
# Case Test Manual Workflow`,
			expected:    true,
			expectError: false,
		},
		{
			name: "complex on section with schedule buried in text",
			content: `---
on:
  push:
    branches: [main]
  schedule:
    - cron: "0 0 * * 0"  # Weekly
  issues:
    types: [opened]
---
# Complex Workflow`,
			expected:    true,
			expectError: false,
		},
		{
			name: "empty on section (not runnable)",
			content: `---
on: {}
---
# Empty On Section`,
			expected:    false,
			expectError: false,
		},
		{
			name: "malformed frontmatter",
			content: `---
invalid yaml structure {
on:
  schedule
---
# Malformed YAML`,
			expected:    false,
			expectError: true,
		},
		{
			name: "no frontmatter at all (defaults to runnable)",
			content: `# Simple Markdown
This file has no frontmatter.
Just plain markdown content.`,
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test-workflow.md")

			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test the function
			result, err := IsRunnable(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestIsRunnable_FileErrors(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		expectErr bool
	}{
		{
			name:      "nonexistent file",
			filePath:  "/nonexistent/path/workflow.md",
			expectErr: true,
		},
		{
			name:      "empty file path",
			filePath:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsRunnable(tt.filePath)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				// Result should be false when there's an error
				if result {
					t.Errorf("Expected false result on error, got true")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFindMatchingLockFile(t *testing.T) {
	// Change to a temporary directory and create .github/workflows structure
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		os.Chdir(originalDir)
	}()

	// Create .github/workflows directory
	workflowsDir := ".github/workflows"
	err = os.MkdirAll(workflowsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workflows directory: %v", err)
	}

	// Set up test lock files
	lockFiles := []string{
		"daily-test-coverage.lock.yml",
		"weekly-research.lock.yml",
		"monthly-report.lock.yml",
		"my_custom_daily.lock.yml",
		"complex-workflow-name.lock.yml",
		"simple.lock.yml",
		"another-test.lock.yml",
		"test-integration.lock.yml",
	}

	for _, fileName := range lockFiles {
		filePath := filepath.Join(workflowsDir, fileName)
		err := os.WriteFile(filePath, []byte("# Mock lock file content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create lock file %s: %v", fileName, err)
		}
	}

	tests := []struct {
		name         string
		workflowName string
		verbose      bool
		expected     string
	}{
		{
			name:         "exact suffix match with underscore",
			workflowName: "daily",
			verbose:      false,
			expected:     "my_custom_daily.lock.yml",
		},
		{
			name:         "contains match when no suffix match",
			workflowName: "test",
			verbose:      false,
			expected:     "another-test.lock.yml", // First match found (alphabetical order)
		},
		{
			name:         "no match found",
			workflowName: "nonexistent",
			verbose:      false,
			expected:     "",
		},
		{
			name:         "exact filename match",
			workflowName: "simple",
			verbose:      false,
			expected:     "simple.lock.yml",
		},
		{
			name:         "complex workflow name match",
			workflowName: "complex-workflow-name",
			verbose:      false,
			expected:     "complex-workflow-name.lock.yml",
		},
		{
			name:         "partial match at beginning",
			workflowName: "daily",
			verbose:      true,                       // Test verbose mode
			expected:     "my_custom_daily.lock.yml", // Suffix match takes priority
		},
		{
			name:         "multiple possible matches - suffix priority",
			workflowName: "test",
			verbose:      false,
			expected:     "another-test.lock.yml", // Contains match (suffix match not found, alphabetical order)
		},
		{
			name:         "case sensitive matching",
			workflowName: "Daily",
			verbose:      false,
			expected:     "", // Should not match "daily"
		},
		{
			name:         "empty workflow name",
			workflowName: "",
			verbose:      false,
			expected:     "another-test.lock.yml", // First file that contains empty string (alphabetical order)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMatchingLockFile(tt.workflowName, tt.verbose)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Helper function to initialize a git repository in test directory
func initTestGitRepo(dir string) error {
	// Create .git directory structure to simulate being in a git repo
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return err
	}

	// Create subdirectories
	subdirs := []string{"objects", "refs", "refs/heads", "refs/tags"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(gitDir, subdir), 0755); err != nil {
			return err
		}
	}

	// Create HEAD file pointing to main branch
	headFile := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		return err
	}

	// Create a minimal git config
	configFile := filepath.Join(gitDir, "config")
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[user]
	name = Test User
	email = test@example.com`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return err
	}

	// Create description file
	descFile := filepath.Join(gitDir, "description")
	if err := os.WriteFile(descFile, []byte("Test repository"), 0644); err != nil {
		return err
	}

	return nil
}
