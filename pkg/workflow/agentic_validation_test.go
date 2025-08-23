package workflow

import (
	"os"
	"strings"
	"testing"

	"github.com/githubnext/gh-aw/pkg/parser"
)

// TestAgenticWorkflowValidationSuite provides comprehensive validation testing
// for realistic agentic workflow files with various schema validation errors
// and verifies correct source location mapping using JSONPath to span mapping.
func TestAgenticWorkflowValidationSuite(t *testing.T) {
	testCases := []struct {
		name               string
		workflowYAML       string
		description        string
		expectedErrors     int
		expectedValidation []ValidationExpectation
	}{
		{
			name:        "missing_required_on_field",
			description: "Agentic workflow missing the required 'on' trigger field",
			workflowYAML: `---
engine: claude
timeout_minutes: 15
permissions:
  issues: write
  contents: read
tools:
  github:
    allowed: [create_issue]
---`,
			expectedErrors: 1,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "on",
					Message:      "missing required field 'on'",
					ExpectSpan:   false, // Missing field has no span
					ExpectedHint: "Add an 'on' field to specify when the workflow should run (e.g., 'on: push')",
				},
			},
		},
		{
			name:        "invalid_engine_type",
			description: "Agentic workflow with unsupported AI engine",
			workflowYAML: `---
engine: gpt-4
on:
  schedule:
    - cron: "0 9 * * 1"
timeout_minutes: 15
permissions:
  issues: write
tools:
  github:
    allowed: [create_issue]
---`,
			expectedErrors: 1,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "engine",
					Message:      "unsupported engine 'gpt-4'",
					ExpectSpan:   true,
					ExpectedLine: 1, // Line numbers are 1-based
					ExpectedHint: "Supported engines: claude, codex",
				},
			},
		},
		{
			name:        "max_turns_out_of_range",
			description: "Agentic workflow with max-turns value exceeding limits",
			workflowYAML: `---
engine: claude
on: push
max-turns: 250
timeout_minutes: 30
permissions:
  contents: read
tools:
  claude:
    allowed: [Write, WebSearch]
---`,
			expectedErrors: 1,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "max-turns",
					Message:      "max-turns must be between 1 and 100",
					ExpectSpan:   true,
					ExpectedLine: 3, // Line numbers are 1-based
					ExpectedHint: "max-turns should be a number between 1 and 100",
				},
			},
		},
		{
			name:        "tools_missing_name_field",
			description: "Agentic workflow with tools array containing entries without required name",
			workflowYAML: `---
engine: claude
on:
  workflow_dispatch:
tools:
  - type: shell
    description: "Git operations"
  - name: github
    type: api
    allowed: [create_issue]
  - type: file
    description: "File operations"
---`,
			expectedErrors: 2,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "tools[0].name",
					Message:      "tool must have a 'name' field",
					ExpectSpan:   false, // Missing field has no span
					ExpectedHint: "Each tool must have a 'name' field specifying the tool identifier",
				},
				{
					Path:         "tools[2].name",
					Message:      "tool must have a 'name' field",
					ExpectSpan:   false, // Missing field has no span
					ExpectedHint: "Each tool must have a 'name' field specifying the tool identifier",
				},
			},
		},
		{
			name:        "complex_nested_validation_errors",
			description: "Complex agentic workflow with multiple validation errors at different nesting levels",
			workflowYAML: `---
engine: invalid-ai
on:
  schedule:
    - cron: "0 9 * * 1"
  workflow_dispatch:
max-turns: 0
timeout_minutes: 15
permissions:
  issues: write
  contents: read
  pull-requests: write
tools:
  - type: shell
    allowed: ["git", "ls"]
  - name: github
    type: api
    allowed: [create_issue, create_comment]
  - description: "File tool without name"
    type: file
output:
  issue:
    title-prefix: "[Weekly] "
    labels: [research, ai]
---`,
			expectedErrors: 4, // invalid engine, invalid max-turns, tools[0].name, tools[2].name
			expectedValidation: []ValidationExpectation{
				{
					Path:         "engine",
					Message:      "unsupported engine 'invalid-ai'",
					ExpectSpan:   true,
					ExpectedLine: 1,
					ExpectedHint: "Supported engines: claude, codex",
				},
				{
					Path:         "max-turns",
					Message:      "max-turns must be between 1 and 100",
					ExpectSpan:   true,
					ExpectedLine: 6,
					ExpectedHint: "max-turns should be a number between 1 and 100",
				},
				{
					Path:         "tools[0].name",
					Message:      "tool must have a 'name' field",
					ExpectSpan:   false, // Missing field has no span
					ExpectedHint: "Each tool must have a 'name' field specifying the tool identifier",
				},
				{
					Path:         "tools[2].name",
					Message:      "tool must have a 'name' field",
					ExpectSpan:   false, // Missing field has no span
					ExpectedHint: "Each tool must have a 'name' field specifying the tool identifier",
				},
			},
		},
		{
			name:        "research_workflow_with_invalid_config",
			description: "Realistic research workflow with various configuration errors",
			workflowYAML: `---
engine: chatgpt
on:
  schedule:
    - cron: "0 9 * * 1"
  workflow_dispatch:
max-turns: 150
timeout_minutes: 45
permissions:
  issues: write
  contents: read
  models: read
  pull-requests: read
tools:
  github:
    allowed: [create_issue]
  claude:
    allowed:
      WebFetch:
      WebSearch:
  shell_tool:
    type: shell
    allowed: ["curl", "grep"]
---`,
			expectedErrors: 2,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "engine",
					Message:      "unsupported engine 'chatgpt'",
					ExpectSpan:   true,
					ExpectedLine: 1,
					ExpectedHint: "Supported engines: claude, codex",
				},
				{
					Path:         "max-turns",
					Message:      "max-turns must be between 1 and 100",
					ExpectSpan:   true,
					ExpectedLine: 6,
					ExpectedHint: "max-turns should be a number between 1 and 100",
				},
			},
		},
		{
			name:        "issue_triage_workflow_errors",
			description: "Issue triage workflow with missing fields and invalid values",
			workflowYAML: `---
on:
  issues:
    types: [opened, labeled]
max-turns: -5
permissions:
  issues: write
  contents: read
tools:
  github:
    allowed: [create_comment, add_labels]
output:
  issue_comment: {}
  labels:
    allowed: ["bug", "feature", "question"]
---`,
			expectedErrors: 1,
			expectedValidation: []ValidationExpectation{
				{
					Path:         "max-turns",
					Message:      "max-turns must be between 1 and 100",
					ExpectSpan:   true,
					ExpectedLine: 4,
					ExpectedHint: "max-turns should be a number between 1 and 100",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse frontmatter from the YAML
			result, err := parser.ExtractFrontmatterFromContent(strings.TrimSpace(tc.workflowYAML))
			if err != nil {
				t.Fatalf("Failed to parse frontmatter: %v", err)
			}
			frontmatter := result.Frontmatter

			// Create validator with the raw YAML
			validator := NewFrontmatterValidator(extractYAMLFromWorkflow(tc.workflowYAML))

			// Validate and get errors
			validationErrors := validator.ValidateFrontmatter(frontmatter)

			// Check error count
			if len(validationErrors) != tc.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tc.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Logf("Error %d: Path=%s, Message=%s", i, err.Path, err.Message)
				}
			}

			// Validate each expected error
			for _, expected := range tc.expectedValidation {
				found := false
				for _, actual := range validationErrors {
					if actual.Path == expected.Path {
						found = true

						// Check message content
						if !strings.Contains(actual.Message, expected.Message) {
							t.Errorf("Expected error message to contain '%s', got '%s'", expected.Message, actual.Message)
						}

						// Check span information
						if expected.ExpectSpan {
							if actual.Span == nil {
								t.Errorf("Expected span information for error at path '%s', but got nil", expected.Path)
							} else {
								// Check if line matches (if specified)
								if expected.ExpectedLine > 0 && actual.Span.StartLine != expected.ExpectedLine {
									t.Errorf("Expected error at line %d, got line %d for path '%s'",
										expected.ExpectedLine, actual.Span.StartLine, expected.Path)
								}

								// Check basic span validity
								if actual.Span.StartLine <= 0 || actual.Span.StartColumn <= 0 {
									t.Errorf("Invalid span coordinates for path '%s': line=%d, col=%d",
										expected.Path, actual.Span.StartLine, actual.Span.StartColumn)
								}

								if actual.Span.EndLine < actual.Span.StartLine ||
									(actual.Span.EndLine == actual.Span.StartLine && actual.Span.EndColumn < actual.Span.StartColumn) {
									t.Errorf("Invalid span range for path '%s': start=(%d,%d), end=(%d,%d)",
										expected.Path, actual.Span.StartLine, actual.Span.StartColumn,
										actual.Span.EndLine, actual.Span.EndColumn)
								}
							}
						} else {
							if actual.Span != nil {
								t.Errorf("Did not expect span information for error at path '%s', but got: %+v",
									expected.Path, actual.Span)
							}
						}

						// Test hint generation
						hint := generateHintForValidationError(actual)
						if expected.ExpectedHint != "" && hint != expected.ExpectedHint {
							t.Errorf("Expected hint '%s', got '%s' for path '%s'",
								expected.ExpectedHint, hint, expected.Path)
						}

						break
					}
				}

				if !found {
					t.Errorf("Expected validation error at path '%s' not found", expected.Path)
				}
			}
		})
	}
}

// ValidationExpectation defines expected validation error details
type ValidationExpectation struct {
	Path         string // JSONPath where error should occur
	Message      string // Expected error message substring
	ExpectSpan   bool   // Whether span information should be available
	ExpectedLine int    // Expected line number (0 if not checking)
	ExpectedHint string // Expected hint message
}

// extractYAMLFromWorkflow extracts the YAML frontmatter from a workflow file
func extractYAMLFromWorkflow(workflow string) string {
	lines := strings.Split(strings.TrimSpace(workflow), "\n")
	if len(lines) < 2 || lines[0] != "---" {
		return ""
	}

	var yamlLines []string
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}

	return strings.Join(yamlLines, "\n")
}

// TestAgenticWorkflowCompilerErrorIntegration tests the integration between
// frontmatter validation and compiler error formatting
func TestAgenticWorkflowCompilerErrorIntegration(t *testing.T) {
	workflowContent := `---
engine: invalid-engine
on: push
max-turns: 200
tools:
  shell_tool:
    type: shell
    allowed: ["git", "ls"]
  github:
    type: api
    allowed: [create_issue]
---

# Test Workflow

This is a test workflow with multiple validation errors.

## Job Description

Test the error reporting system.
`

	// Parse frontmatter from full workflow content
	result, err := parser.ExtractFrontmatterFromContent(workflowContent)
	if err != nil {
		t.Fatalf("Failed to parse frontmatter: %v", err)
	}
	frontmatter := result.Frontmatter

	// Extract just the YAML for the validator
	yamlContent := extractYAMLFromWorkflow(workflowContent)

	// Validate
	validator := NewFrontmatterValidator(yamlContent)
	validationErrors := validator.ValidateFrontmatter(frontmatter)

	// Should have 2 errors: invalid engine, invalid max-turns
	if len(validationErrors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(validationErrors))
		for i, err := range validationErrors {
			t.Logf("Error %d: Path=%s, Message=%s", i, err.Path, err.Message)
		}
		// Continue the test even if error count doesn't match to see what we got
	}

	// Convert to compiler errors
	filePath := "test-workflow.md"
	frontmatterStart := 2 // Frontmatter starts at line 2 in the file
	compilerErrors := ConvertValidationErrorsToCompilerErrors(filePath, frontmatterStart, validationErrors)

	if len(compilerErrors) != len(validationErrors) {
		t.Fatalf("Expected %d compiler errors to match validation errors, got %d", len(validationErrors), len(compilerErrors))
	}

	// Test error position adjustment
	for _, compilerErr := range compilerErrors {
		if compilerErr.Position.File != filePath {
			t.Errorf("Expected file '%s', got '%s'", filePath, compilerErr.Position.File)
		}

		// Check that line numbers are adjusted for frontmatter position
		if compilerErr.Position.Line < frontmatterStart {
			t.Errorf("Expected adjusted line >= %d, got %d", frontmatterStart, compilerErr.Position.Line)
		}

		// Check that hints are generated
		if compilerErr.Hint == "" {
			t.Errorf("Expected hint to be generated for error: %s", compilerErr.Message)
		}
	}
}

// TestAgenticWorkflowSpanAccuracy tests the accuracy of span information
// for complex nested structures in agentic workflows
func TestAgenticWorkflowSpanAccuracy(t *testing.T) {
	workflowYAML := `engine: claude
on:
  schedule:
    - cron: "0 9 * * 1"
  workflow_dispatch:
max-turns: 50
permissions:
  issues: write
  contents: read
tools:
  github:
    allowed: [create_issue]
  claude:
    allowed:
      WebFetch:
      WebSearch:
  time:
    mcp:
      type: stdio
      container: "mcp/time"
    allowed: ["get_current_time"]`

	locator := parser.NewFrontmatterLocator(workflowYAML)

	testCases := []struct {
		path           string
		expectedExists bool
		expectedLine   int
	}{
		{"engine", true, 1},
		{"on.schedule[0].cron", true, 4},
		{"max-turns", true, 6},
		{"permissions.issues", true, 8},
		{"tools.github.allowed[0]", true, 12},
		{"tools.claude.allowed.WebFetch", true, 15},
		{"tools.time.mcp.container", true, 20},
		{"tools.time.allowed[0]", true, 21},
		{"nonexistent", false, 0},
		{"tools.nonexistent", false, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			span, err := locator.LocatePathSpan(tc.path)

			if tc.expectedExists {
				if err != nil {
					t.Errorf("Expected to find span for path '%s', got error: %v", tc.path, err)
				} else {
					if span.StartLine != tc.expectedLine {
						t.Errorf("Expected line %d for path '%s', got %d", tc.expectedLine, tc.path, span.StartLine)
					}

					// Verify span validity
					if span.StartLine <= 0 || span.StartColumn <= 0 {
						t.Errorf("Invalid span coordinates for path '%s': %+v", tc.path, span)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for nonexistent path '%s', but got span: %+v", tc.path, span)
				}
			}
		})
	}
}

// TestSampleAgenticWorkflowFiles tests the sample workflow files with validation errors
// located in .github/workflows/test-*.md to ensure they produce expected validation errors
func TestSampleAgenticWorkflowFiles(t *testing.T) {
	testFiles := []struct {
		filename       string
		expectedErrors int
		description    string
	}{
		{
			filename:       ".github/workflows/test-validation-errors.md",
			expectedErrors: 3, // invalid engine, max-turns too high, missing tool name
			description:    "Multiple validation errors workflow",
		},
		{
			filename:       ".github/workflows/test-missing-on.md",
			expectedErrors: 1, // missing 'on' field
			description:    "Missing required 'on' field workflow",
		},
		{
			filename:       ".github/workflows/test-tools-validation.md",
			expectedErrors: 4, // max-turns 0, and 3 missing tool names
			description:    "Tools array validation errors workflow",
		},
	}

	for _, tf := range testFiles {
		t.Run(tf.filename, func(t *testing.T) {
			// Read the workflow file
			workflowPath := "/home/runner/work/gh-aw-copilots/gh-aw-copilots/" + tf.filename

			// Check if file exists (skip if not found since these are test files)
			if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
				t.Skipf("Test workflow file %s not found, skipping", tf.filename)
			}

			content, err := os.ReadFile(workflowPath)
			if err != nil {
				t.Fatalf("Failed to read workflow file %s: %v", tf.filename, err)
			}

			// Parse frontmatter
			result, err := parser.ExtractFrontmatterFromContent(string(content))
			if err != nil {
				t.Fatalf("Failed to parse frontmatter from %s: %v", tf.filename, err)
			}

			// Extract YAML for validator
			yamlContent := extractYAMLFromWorkflow(string(content))
			if yamlContent == "" {
				t.Fatalf("No frontmatter YAML found in %s", tf.filename)
			}

			// Validate
			validator := NewFrontmatterValidator(yamlContent)
			validationErrors := validator.ValidateFrontmatter(result.Frontmatter)

			// Check error count
			if len(validationErrors) != tf.expectedErrors {
				t.Errorf("File %s: expected %d validation errors, got %d",
					tf.filename, tf.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Logf("  Error %d: Path=%s, Message=%s, HasSpan=%v",
						i, err.Path, err.Message, err.Span != nil)
				}
			}

			// Test error formatting for each error
			for _, validationErr := range validationErrors {
				// Test hint generation
				hint := generateHintForValidationError(validationErr)
				if hint == "" && (strings.Contains(validationErr.Path, "engine") ||
					strings.Contains(validationErr.Path, "max-turns") ||
					strings.Contains(validationErr.Path, "tools") ||
					validationErr.Path == "on") {
					t.Errorf("Expected hint for common validation error at path '%s'", validationErr.Path)
				}

				// Test span information for existing fields
				if validationErr.Span != nil {
					if validationErr.Span.StartLine <= 0 || validationErr.Span.StartColumn <= 0 {
						t.Errorf("Invalid span coordinates for path '%s': %+v",
							validationErr.Path, validationErr.Span)
					}
				}
			}

			// Test compiler error conversion
			compilerErrors := ConvertValidationErrorsToCompilerErrors(
				tf.filename, 1, validationErrors) // Assume frontmatter starts at line 1

			if len(compilerErrors) != len(validationErrors) {
				t.Errorf("Compiler error count mismatch for %s: validation=%d, compiler=%d",
					tf.filename, len(validationErrors), len(compilerErrors))
			}

			// Verify compiler errors have proper structure
			for _, compilerErr := range compilerErrors {
				if compilerErr.Position.File != tf.filename {
					t.Errorf("Compiler error file mismatch: expected %s, got %s",
						tf.filename, compilerErr.Position.File)
				}
				if compilerErr.Type != "error" {
					t.Errorf("Expected error type 'error', got '%s'", compilerErr.Type)
				}
				if compilerErr.Message == "" {
					t.Error("Compiler error message should not be empty")
				}
			}
		})
	}
}
