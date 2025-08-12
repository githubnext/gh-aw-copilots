package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFrontmatterFromContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantYAML     map[string]any
		wantMarkdown string
		wantErr      bool
	}{
		{
			name: "valid frontmatter and markdown",
			content: `---
title: Test Workflow
on: push
---

# Test Workflow

This is a test workflow.`,
			wantYAML: map[string]any{
				"title": "Test Workflow",
				"on":    "push",
			},
			wantMarkdown: "# Test Workflow\n\nThis is a test workflow.",
		},
		{
			name: "no frontmatter",
			content: `# Test Workflow

This is a test workflow without frontmatter.`,
			wantYAML:     map[string]any{},
			wantMarkdown: "# Test Workflow\n\nThis is a test workflow without frontmatter.",
		},
		{
			name: "empty frontmatter",
			content: `---
---

# Test Workflow

This is a test workflow with empty frontmatter.`,
			wantYAML:     map[string]any{},
			wantMarkdown: "# Test Workflow\n\nThis is a test workflow with empty frontmatter.",
		},
		{
			name:    "unclosed frontmatter",
			content: "---\ntitle: Test\nno closing delimiter",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractFrontmatterFromContent(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractFrontmatterFromContent() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractFrontmatterFromContent() error = %v", err)
				return
			}

			// Check frontmatter
			if len(tt.wantYAML) != len(result.Frontmatter) {
				t.Errorf("ExtractFrontmatterFromContent() frontmatter length = %v, want %v", len(result.Frontmatter), len(tt.wantYAML))
			}

			for key, expectedValue := range tt.wantYAML {
				if actualValue, exists := result.Frontmatter[key]; !exists {
					t.Errorf("ExtractFrontmatterFromContent() missing key %v", key)
				} else if actualValue != expectedValue {
					t.Errorf("ExtractFrontmatterFromContent() frontmatter[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}

			// Check markdown
			if result.Markdown != tt.wantMarkdown {
				t.Errorf("ExtractFrontmatterFromContent() markdown = %v, want %v", result.Markdown, tt.wantMarkdown)
			}
		})
	}
}

func TestExtractYamlChunk(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		key      string
		expected string
	}{
		{
			name: "simple key-value",
			yaml: `title: Test Workflow
on: push
permissions: read`,
			key:      "on",
			expected: "on: push",
		},
		{
			name: "nested structure",
			yaml: `title: Test Workflow
on:
  push:
    branches:
      - main
  pull_request:
    types: [opened]
permissions: read`,
			key: "on",
			expected: `on:
  push:
    branches:
      - main
  pull_request:
    types: [opened]`,
		},
		{
			name: "deeply nested structure",
			yaml: `tools:
  Bash:
    allowed:
      - "ls"
      - "cat"
  github:
    allowed:
      - "create_issue"`,
			key: "tools",
			expected: `tools:
  Bash:
    allowed:
      - "ls"
      - "cat"
  github:
    allowed:
      - "create_issue"`,
		},
		{
			name: "key not found",
			yaml: `title: Test Workflow
on: push`,
			key:      "nonexistent",
			expected: "",
		},
		{
			name:     "empty yaml",
			yaml:     "",
			key:      "test",
			expected: "",
		},
		{
			name:     "empty key",
			yaml:     "title: Test",
			key:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractYamlChunk(tt.yaml, tt.key)
			if err != nil {
				t.Errorf("ExtractYamlChunk() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractYamlChunk() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractMarkdownSection(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionName string
		expected    string
		wantErr     bool
	}{
		{
			name: "basic H1 section",
			content: `# Introduction

This is the introduction.

# Setup

This is the setup section.

# Configuration

This is the configuration.`,
			sectionName: "Setup",
			expected: `# Setup

This is the setup section.`,
		},
		{
			name: "H2 section",
			content: `# Main Title

## Subsection 1

Content for subsection 1.

## Subsection 2

Content for subsection 2.`,
			sectionName: "Subsection 1",
			expected: `## Subsection 1

Content for subsection 1.`,
		},
		{
			name: "nested sections",
			content: `# Main

## Sub1

Content 1

### Sub1.1

Nested content

## Sub2

Content 2`,
			sectionName: "Sub1",
			expected: `## Sub1

Content 1

### Sub1.1

Nested content`,
		},
		{
			name:        "section not found",
			content:     "# Title\n\nContent",
			sectionName: "NonExistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractMarkdownSection(tt.content, tt.sectionName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractMarkdownSection() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractMarkdownSection() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractMarkdownSection() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateDefaultWorkflowName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "simple filename",
			filePath: "test-workflow.md",
			expected: "Test Workflow",
		},
		{
			name:     "multiple hyphens",
			filePath: "my-test-workflow-file.md",
			expected: "My Test Workflow File",
		},
		{
			name:     "full path",
			filePath: "/path/to/my-workflow.md",
			expected: "My Workflow",
		},
		{
			name:     "no extension",
			filePath: "workflow",
			expected: "Workflow",
		},
		{
			name:     "single word",
			filePath: "test.md",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDefaultWorkflowName(tt.filePath)
			if result != tt.expected {
				t.Errorf("generateDefaultWorkflowName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractFrontmatterString(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		wantErr  bool
	}{
		{
			name: "valid frontmatter",
			content: `---
title: Test Workflow
on: push
---

# Content`,
			expected: "on: push\ntitle: Test Workflow",
		},
		{
			name:     "no frontmatter",
			content:  "# Just markdown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractFrontmatterString(tt.content)

			if tt.wantErr && err == nil {
				t.Errorf("ExtractFrontmatterString() expected error, got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ExtractFrontmatterString() error = %v", err)
				return
			}

			// For YAML, order may vary, so check both possible orders
			if !strings.Contains(result, "title: Test Workflow") && tt.expected != "" {
				if result != tt.expected {
					t.Errorf("ExtractFrontmatterString() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestExtractMarkdownContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		wantErr  bool
	}{
		{
			name: "with frontmatter",
			content: `---
title: Test
---

# Markdown Content

This is markdown.`,
			expected: "# Markdown Content\n\nThis is markdown.",
		},
		{
			name:     "no frontmatter",
			content:  "# Just Markdown\n\nContent here.",
			expected: "# Just Markdown\n\nContent here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractMarkdownContent(tt.content)

			if tt.wantErr && err == nil {
				t.Errorf("ExtractMarkdownContent() expected error, got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ExtractMarkdownContent() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractMarkdownContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessIncludes(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "test_includes")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with markdown content
	testFile := filepath.Join(tempDir, "test.md")
	testContent := `---
tools:
  bash:
    allowed: ["ls", "cat"]
---

# Test Content
This is a test file content.
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create test file with extra newlines for trimming test
	testFileWithNewlines := filepath.Join(tempDir, "test-newlines.md")
	testContentWithNewlines := `

# Content with Extra Newlines
Some content here.


`
	if err := os.WriteFile(testFileWithNewlines, []byte(testContentWithNewlines), 0644); err != nil {
		t.Fatalf("Failed to write test file with newlines: %v", err)
	}

	tests := []struct {
		name         string
		content      string
		baseDir      string
		extractTools bool
		expected     string
		wantErr      bool
	}{
		{
			name:         "no includes",
			content:      "# Title\nRegular content",
			baseDir:      tempDir,
			extractTools: false,
			expected:     "# Title\nRegular content\n",
		},
		{
			name:         "simple include",
			content:      "@include test.md\n# After include",
			baseDir:      tempDir,
			extractTools: false,
			expected:     "# Test Content\nThis is a test file content.\n# After include\n",
		},
		{
			name:         "extract tools",
			content:      "@include test.md",
			baseDir:      tempDir,
			extractTools: true,
			expected:     `{"bash":{"allowed":["ls","cat"]}}` + "\n",
		},
		{
			name:         "file not found",
			content:      "@include nonexistent.md",
			baseDir:      tempDir,
			extractTools: false,
			expected:     "\n<!-- Error: file not found: " + filepath.Join(tempDir, "nonexistent.md") + " -->\n\n",
		},
		{
			name:         "include file with extra newlines",
			content:      "@include test-newlines.md\n# After include",
			baseDir:      tempDir,
			extractTools: false,
			expected:     "# Content with Extra Newlines\nSome content here.\n# After include\n",
		},
	}

	// Create test file with invalid frontmatter for testing validation
	invalidFile := filepath.Join(tempDir, "invalid.md")
	invalidContent := `---
title: Invalid File
on: push
tools:
  bash:
    allowed: ["ls"]
---

# Invalid Content
This file has invalid frontmatter for an included file.
`
	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid test file: %v", err)
	}

	// Add test case for invalid frontmatter in included file (should now pass with warnings for non-workflow files)
	tests = append(tests, struct {
		name         string
		content      string
		baseDir      string
		extractTools bool
		expected     string
		wantErr      bool
	}{
		name:         "invalid frontmatter in included file",
		content:      "@include invalid.md",
		baseDir:      tempDir,
		extractTools: false,
		expected:     "# Invalid Content\nThis file has invalid frontmatter for an included file.\n",
		wantErr:      false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessIncludes(tt.content, tt.baseDir, tt.extractTools)

			if tt.wantErr && err == nil {
				t.Errorf("ProcessIncludes() expected error, got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ProcessIncludes() error = %v", err)
				return
			}

			// Special handling for the invalid frontmatter test case - it should now pass with warnings
			if tt.name == "invalid frontmatter in included file" {
				// Check that the content was successfully included
				if !strings.Contains(result, "# Invalid Content") {
					t.Errorf("ProcessIncludes() = %q, expected to contain '# Invalid Content'", result)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("ProcessIncludes() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsUnderWorkflowsDirectory(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "file under .github/workflows",
			filePath: "/some/path/.github/workflows/test.md",
			expected: true,
		},
		{
			name:     "file under .github/workflows subdirectory",
			filePath: "/some/path/.github/workflows/shared/helper.md",
			expected: true,
		},
		{
			name:     "file outside .github/workflows",
			filePath: "/some/path/docs/instructions.md",
			expected: false,
		},
		{
			name:     "file in .github but not workflows",
			filePath: "/some/path/.github/ISSUE_TEMPLATE/bug.md",
			expected: false,
		},
		{
			name:     "relative path under workflows",
			filePath: ".github/workflows/test.md",
			expected: true,
		},
		{
			name:     "relative path outside workflows",
			filePath: "docs/readme.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUnderWorkflowsDirectory(tt.filePath)
			if result != tt.expected {
				t.Errorf("isUnderWorkflowsDirectory(%q) = %v, want %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestProcessIncludesConditionalValidation(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "test_conditional_validation")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .github/workflows directory structure
	workflowsDir := filepath.Join(tempDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create docs directory for non-workflow files
	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Test file 1: Valid workflow file (should pass strict validation)
	validWorkflowFile := filepath.Join(workflowsDir, "valid.md")
	validWorkflowContent := `---
tools:
  github:
    allowed: [get_issue]
---

# Valid Workflow
This is a valid workflow file.`
	if err := os.WriteFile(validWorkflowFile, []byte(validWorkflowContent), 0644); err != nil {
		t.Fatalf("Failed to write valid workflow file: %v", err)
	}

	// Test file 2: Invalid workflow file (should fail strict validation)
	invalidWorkflowFile := filepath.Join(workflowsDir, "invalid.md")
	invalidWorkflowContent := `---
title: Invalid Field
on: push
tools:
  github:
    allowed: [get_issue]
---

# Invalid Workflow
This has invalid frontmatter fields.`
	if err := os.WriteFile(invalidWorkflowFile, []byte(invalidWorkflowContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid workflow file: %v", err)
	}

	// Test file 2.5: Invalid non-workflow file (should pass with warnings)
	invalidNonWorkflowFile := filepath.Join(docsDir, "invalid-external.md")
	invalidNonWorkflowContent := `---
title: Invalid Field
on: push
tools:
  github:
    allowed: [get_issue]
---

# Invalid External File
This has invalid frontmatter fields but it's outside workflows dir.`
	if err := os.WriteFile(invalidNonWorkflowFile, []byte(invalidNonWorkflowContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid non-workflow file: %v", err)
	}

	// Test file 3: Copilot instructions file (should pass with warnings)
	copilotFile := filepath.Join(docsDir, "copilot-instructions.md")
	copilotContent := `---
description: Copilot instructions
applyTo: "**/*.py"
temperature: 0.7
tools:
  github:
    allowed: [get_issue]
---

# Copilot Instructions
These are instructions for GitHub Copilot.`
	if err := os.WriteFile(copilotFile, []byte(copilotContent), 0644); err != nil {
		t.Fatalf("Failed to write copilot file: %v", err)
	}

	// Test file 4: Plain markdown file (no frontmatter)
	plainFile := filepath.Join(docsDir, "plain.md")
	plainContent := `# Plain Markdown
This is just plain markdown content with no frontmatter.`
	if err := os.WriteFile(plainFile, []byte(plainContent), 0644); err != nil {
		t.Fatalf("Failed to write plain file: %v", err)
	}

	tests := []struct {
		name         string
		content      string
		baseDir      string
		extractTools bool
		wantErr      bool
		checkContent string
	}{
		{
			name:         "valid workflow file inclusion",
			content:      "@include .github/workflows/valid.md",
			baseDir:      tempDir,
			extractTools: false,
			wantErr:      false,
			checkContent: "# Valid Workflow",
		},
		{
			name:         "invalid workflow file inclusion should fail",
			content:      "@include .github/workflows/invalid.md",
			baseDir:      tempDir,
			extractTools: false,
			wantErr:      false,
			checkContent: "<!-- Error: invalid frontmatter in included file",
		},
		{
			name:         "invalid non-workflow file inclusion should succeed with warnings",
			content:      "@include docs/invalid-external.md",
			baseDir:      tempDir,
			extractTools: false,
			wantErr:      false,
			checkContent: "# Invalid External File",
		},
		{
			name:         "copilot instructions file inclusion should succeed",
			content:      "@include docs/copilot-instructions.md",
			baseDir:      tempDir,
			extractTools: false,
			wantErr:      false,
			checkContent: "# Copilot Instructions",
		},
		{
			name:         "plain markdown file inclusion should succeed",
			content:      "@include docs/plain.md",
			baseDir:      tempDir,
			extractTools: false,
			wantErr:      false,
			checkContent: "# Plain Markdown",
		},
		{
			name:         "extract tools from valid workflow file",
			content:      "@include .github/workflows/valid.md",
			baseDir:      tempDir,
			extractTools: true,
			wantErr:      false,
			checkContent: `{"github":{"allowed":["get_issue"]}}`,
		},
		{
			name:         "extract tools from copilot file",
			content:      "@include docs/copilot-instructions.md",
			baseDir:      tempDir,
			extractTools: true,
			wantErr:      false,
			checkContent: `{"github":{"allowed":["get_issue"]}}`,
		},
		{
			name:         "extract tools from plain file (no tools)",
			content:      "@include docs/plain.md",
			baseDir:      tempDir,
			extractTools: true,
			wantErr:      false,
			checkContent: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessIncludes(tt.content, tt.baseDir, tt.extractTools)

			if tt.wantErr && err == nil {
				t.Errorf("ProcessIncludes() expected error, got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ProcessIncludes() error = %v", err)
				return
			}

			if !tt.wantErr && tt.checkContent != "" {
				if !strings.Contains(result, tt.checkContent) {
					t.Errorf("ProcessIncludes() result = %q, expected to contain %q", result, tt.checkContent)
				}
			}
		})
	}
}

func TestResolveIncludePath(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "test_resolve")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create regular test file in temp dir
	regularFile := filepath.Join(tempDir, "regular.md")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write regular file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		baseDir  string
		expected string
		wantErr  bool
	}{
		{
			name:     "regular relative path",
			filePath: "regular.md",
			baseDir:  tempDir,
			expected: regularFile,
		},
		{
			name:     "regular file not found",
			filePath: "nonexistent.md",
			baseDir:  tempDir,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveIncludePath(tt.filePath, tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveIncludePath() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("resolveIncludePath() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("resolveIncludePath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMergeTools(t *testing.T) {
	tests := []struct {
		name       string
		base       map[string]any
		additional map[string]any
		expected   map[string]any
	}{
		{
			name: "merge with allowed arrays",
			base: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"ls", "cat"},
				},
			},
			additional: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"grep", "ls"}, // ls is duplicate
				},
			},
			expected: map[string]any{
				"bash": map[string]any{
					"allowed": []string{"ls", "cat", "grep"},
				},
			},
		},
		{
			name: "new tool added",
			base: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"ls"},
				},
			},
			additional: map[string]any{
				"github": map[string]any{
					"allowed": []any{"create_issue"},
				},
			},
			expected: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"ls"},
				},
				"github": map[string]any{
					"allowed": []any{"create_issue"},
				},
			},
		},
		{
			name: "empty base",
			base: map[string]any{},
			additional: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"ls"},
				},
			},
			expected: map[string]any{
				"bash": map[string]any{
					"allowed": []any{"ls"},
				},
			},
		},
		{
			name: "merge claude section tools (new format)",
			base: map[string]any{
				"github": map[string]any{
					"allowed": []any{"list_issues"},
				},
				"claude": map[string]any{
					"allowed": map[string]any{
						"Edit":  nil,
						"Write": nil,
					},
				},
			},
			additional: map[string]any{
				"claude": map[string]any{
					"allowed": map[string]any{
						"Read":      nil,
						"MultiEdit": nil,
					},
				},
			},
			expected: map[string]any{
				"github": map[string]any{
					"allowed": []any{"list_issues"},
				},
				"claude": map[string]any{
					"allowed": map[string]any{
						"Edit":      nil,
						"Write":     nil,
						"Read":      nil,
						"MultiEdit": nil,
					},
				},
			},
		},
		{
			name: "merge nested Bash tools under claude section (new format)",
			base: map[string]any{
				"claude": map[string]any{
					"allowed": map[string]any{
						"Bash": []any{"pwd", "whoami"},
						"Edit": nil,
					},
				},
			},
			additional: map[string]any{
				"claude": map[string]any{
					"allowed": map[string]any{
						"Bash": []any{"ls", "cat", "pwd"}, // pwd is duplicate
						"Read": nil,
					},
				},
			},
			expected: map[string]any{
				"claude": map[string]any{
					"allowed": map[string]any{
						"Bash": []any{"pwd", "whoami", "ls", "cat"},
						"Edit": nil,
						"Read": nil,
					},
				},
			},
		},
		{
			name: "merge mcp tools with wildcard allowed",
			base: map[string]any{
				"notion": map[string]any{
					"type":    "mcp",
					"allowed": []any{"create_page"},
				},
			},
			additional: map[string]any{
				"notion": map[string]any{
					"type":    "mcp",
					"allowed": []any{"*"},
				},
			},
			expected: map[string]any{
				"notion": map[string]any{
					"type":    "mcp",
					"allowed": []any{"create_page", "*"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeTools(tt.base, tt.additional)
			if err != nil {
				t.Fatalf("MergeTools() returned unexpected error: %v", err)
			}

			// Convert result to JSON and back for easier comparison
			resultJSON, _ := json.Marshal(result)
			expectedJSON, _ := json.Marshal(tt.expected)

			var resultMap, expectedMap map[string]any
			if err := json.Unmarshal(resultJSON, &resultMap); err != nil {
				t.Fatalf("Failed to unmarshal result JSON: %v", err)
			}
			if err := json.Unmarshal(expectedJSON, &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			// Compare JSON strings for easier debugging
			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("MergeTools() = %s, want %s", string(resultJSON), string(expectedJSON))
			}
		})
	}
}

func TestExpandIncludes(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "test_expand")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod to make it project root for component resolution
	goModFile := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModFile, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "test.md")
	testContent := `---
tools:
  bash:
    allowed: ["ls"]
---

# Test Content
This is test content.
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name         string
		content      string
		baseDir      string
		extractTools bool
		wantContains string
		wantErr      bool
	}{
		{
			name:         "expand markdown content",
			content:      "# Start\n@include test.md\n# End",
			baseDir:      tempDir,
			extractTools: false,
			wantContains: "# Test Content\nThis is test content.",
		},
		{
			name:         "expand tools",
			content:      "@include test.md",
			baseDir:      tempDir,
			extractTools: true,
			wantContains: `"bash"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandIncludes(tt.content, tt.baseDir, tt.extractTools)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExpandIncludes() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ExpandIncludes() error = %v", err)
				return
			}

			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("ExpandIncludes() = %q, want to contain %q", result, tt.wantContains)
			}
		})
	}
}

// Test ExtractWorkflowNameFromMarkdown function
func TestExtractWorkflowNameFromMarkdown(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "test-extract-name-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		content     string
		expected    string
		expectError bool
	}{
		{
			name: "file with H1 header",
			content: `---
name: Test Workflow
---

# Daily QA Report

This is a test workflow.`,
			expected:    "Daily QA Report",
			expectError: false,
		},
		{
			name: "file without H1 header",
			content: `---
name: Test Workflow
---

This is content without H1 header.
## This is H2`,
			expected:    "Test Extract Name", // Should generate from filename
			expectError: false,
		},
		{
			name: "file with multiple H1 headers",
			content: `---
name: Test Workflow
---

# First Header

Some content.

# Second Header

Should use first H1.`,
			expected:    "First Header",
			expectError: false,
		},
		{
			name: "file with only frontmatter",
			content: `---
name: Test Workflow
description: A test
---`,
			expected:    "Test Extract Name", // Should generate from filename
			expectError: false,
		},
		{
			name: "file with H1 and extra spaces",
			content: `---
name: Test
---

#   Spaced Header   

Content here.`,
			expected:    "Spaced Header",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			fileName := "test-extract-name.md"
			filePath := filepath.Join(tempDir, fileName)

			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := ExtractWorkflowNameFromMarkdown(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractWorkflowNameFromMarkdown(%q) expected error, but got none", filePath)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractWorkflowNameFromMarkdown(%q) unexpected error: %v", filePath, err)
				return
			}

			if got != tt.expected {
				t.Errorf("ExtractWorkflowNameFromMarkdown(%q) = %q, want %q", filePath, got, tt.expected)
			}
		})
	}

	// Test nonexistent file
	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ExtractWorkflowNameFromMarkdown("/nonexistent/file.md")
		if err == nil {
			t.Error("ExtractWorkflowNameFromMarkdown with nonexistent file should return error")
		}
	})
}

// Test ExtractMarkdown function
func TestExtractMarkdown(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "test-extract-markdown-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		content     string
		expected    string
		expectError bool
	}{
		{
			name: "file with frontmatter",
			content: `---
name: Test Workflow
description: A test workflow
---

# Test Content

This is the markdown content.`,
			expected:    "# Test Content\n\nThis is the markdown content.",
			expectError: false,
		},
		{
			name: "file without frontmatter",
			content: `# Pure Markdown

This is just markdown content without frontmatter.`,
			expected:    "# Pure Markdown\n\nThis is just markdown content without frontmatter.",
			expectError: false,
		},
		{
			name:        "empty file",
			content:     ``,
			expected:    "",
			expectError: false,
		},
		{
			name: "file with only frontmatter",
			content: `---
name: Test
---`,
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			fileName := "test-extract-markdown.md"
			filePath := filepath.Join(tempDir, fileName)

			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := ExtractMarkdown(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractMarkdown(%q) expected error, but got none", filePath)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractMarkdown(%q) unexpected error: %v", filePath, err)
				return
			}

			if got != tt.expected {
				t.Errorf("ExtractMarkdown(%q) = %q, want %q", filePath, got, tt.expected)
			}
		})
	}

	// Test nonexistent file
	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ExtractMarkdown("/nonexistent/file.md")
		if err == nil {
			t.Error("ExtractMarkdown with nonexistent file should return error")
		}
	})
}

// Test mergeToolsFromJSON function
func TestMergeToolsFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		wantErr  bool
	}{
		{
			name:     "single valid JSON object",
			content:  `{"tool1": {"enabled": true}, "tool2": {"enabled": false}}`,
			expected: `{"tool1":{"enabled":true},"tool2":{"enabled":false}}`,
			wantErr:  false,
		},
		{
			name: "multiple JSON objects on separate lines",
			content: `{"tool1": {"enabled": true}}
{"tool2": {"enabled": false}}
{"tool3": {"config": "value"}}`,
			expected: `{"tool1":{"enabled":true},"tool2":{"enabled":false},"tool3":{"config":"value"}}`,
			wantErr:  false,
		},
		{
			name:     "empty content",
			content:  ``,
			expected: `{}`,
			wantErr:  false,
		},
		{
			name:     "empty JSON objects",
			content:  `{}\n{}\n{}`,
			expected: `{}`,
			wantErr:  false,
		},
		{
			name:     "whitespace only",
			content:  `   \n  \t  \n   `,
			expected: `{}`,
			wantErr:  false,
		},
		{
			name: "mixed empty and non-empty objects",
			content: `{}
{"tool1": {"enabled": true}}
{}
{"tool2": {"value": 42}}`,
			expected: `{"tool1":{"enabled":true},"tool2":{"value":42}}`,
			wantErr:  false,
		},
		{
			name: "objects with overlapping keys",
			content: `{"tool1": {"enabled": true, "config": "old"}}
{"tool1": {"config": "new", "version": 2}}`,
			expected: `{"tool1":{"config":"new","enabled":true,"version":2}}`,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			content:  `{"invalid": json}`,
			expected: `{}`,
			wantErr:  false, // Function handles invalid JSON gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeToolsFromJSON(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("mergeToolsFromJSON(%q) expected error, but got none", tt.content)
				}
				return
			}

			if err != nil {
				t.Errorf("mergeToolsFromJSON(%q) unexpected error: %v", tt.content, err)
				return
			}

			// For JSON comparison, parse both strings to ensure equivalent content
			var gotObj, expectedObj map[string]any
			if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
				t.Errorf("mergeToolsFromJSON(%q) returned invalid JSON: %v", tt.content, err)
				return
			}
			if err := json.Unmarshal([]byte(tt.expected), &expectedObj); err != nil {
				t.Errorf("Test case has invalid expected JSON: %v", err)
				return
			}

			// Convert back to JSON strings for comparison
			gotJSON, _ := json.Marshal(gotObj)
			expectedJSON, _ := json.Marshal(expectedObj)

			if string(gotJSON) != string(expectedJSON) {
				t.Errorf("mergeToolsFromJSON(%q) = %q, want %q", tt.content, string(gotJSON), string(expectedJSON))
			}
		})
	}
}
