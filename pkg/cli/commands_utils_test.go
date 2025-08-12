package cli

import (
	"os"
	"path/filepath"
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
