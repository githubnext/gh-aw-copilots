package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/githubnext/gh-aw/pkg/console"
	"github.com/githubnext/gh-aw/pkg/parser"
)

// TestFrontmatterSpanIntegration tests the complete integration from
// frontmatter validation through span calculation to error formatting
func TestFrontmatterSpanIntegration(t *testing.T) {
	tests := []struct {
		name                    string
		workflowContent         string
		expectedErrorSubstrings []string
		expectedSpanFormat      bool // Should the error include span formatting?
	}{
		{
			name: "invalid engine with span",
			workflowContent: `---
engine: invalid-engine
on: push
---

# Test Workflow

This workflow tests span-based error reporting.`,
			expectedErrorSubstrings: []string{
				"test.md:2:9-22:", // Span format for invalid engine (corrected)
				"unsupported engine 'invalid-engine'",
				"Supported engines: claude, codex",
			},
			expectedSpanFormat: true,
		},
		{
			name: "invalid max-turns with span",
			workflowContent: `---
engine: claude
on: push
max-turns: 150
---

# Test Workflow

This workflow has invalid max-turns value.`,
			expectedErrorSubstrings: []string{
				"test.md:4:12-14:", // Span format for max-turns value (corrected)
				"max-turns must be between 1 and 100, got 150",
				"max-turns should be a number between 1 and 100",
			},
			expectedSpanFormat: true,
		},
		{
			name: "missing on field (no span)",
			workflowContent: `---
engine: claude
max-turns: 5
---

# Test Workflow

This workflow is missing the 'on' field.`,
			expectedErrorSubstrings: []string{
				"test.md:2:1:", // No span for missing field
				"missing required field 'on'",
				"Add an 'on' field to specify when the workflow should run",
			},
			expectedSpanFormat: false,
		},
		{
			name: "tools validation with complex path",
			workflowContent: `---
engine: claude
on: push
tools:
  - name: git
    type: shell
  - type: shell
    description: Missing name
---

# Test Workflow

This workflow has tools with missing name field.`,
			expectedErrorSubstrings: []string{
				"tool must have a 'name' field",
				"Each tool must have a 'name' field specifying the tool identifier",
			},
			expectedSpanFormat: false, // Complex path might not resolve to span
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir, err := os.MkdirTemp("", "span-integration-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.workflowContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Extract frontmatter and perform validation
			result, err := parser.ExtractFrontmatterFromContent(tt.workflowContent)
			if err != nil {
				t.Fatal(err)
			}

			// Get the raw frontmatter YAML for span calculation
			frontmatterYAML := strings.Join(result.FrontmatterLines, "\n")

			// Create validator and validate
			validator := NewFrontmatterValidator(frontmatterYAML)
			validationErrors := validator.ValidateFrontmatter(result.Frontmatter)

			if len(validationErrors) == 0 {
				t.Fatal("Expected validation errors, got none")
			}

			// Convert to compiler errors with span information
			compilerErrors := ConvertValidationErrorsToCompilerErrors(
				testFile,
				result.FrontmatterStart,
				validationErrors,
			)

			if len(compilerErrors) == 0 {
				t.Fatal("Expected compiler errors, got none")
			}

			// Format errors and check output
			for _, compilerError := range compilerErrors {
				formattedError := console.FormatError(compilerError)

				// Check that expected substrings are present
				for _, expectedSubstring := range tt.expectedErrorSubstrings {
					if !strings.Contains(formattedError, expectedSubstring) {
						t.Errorf("Expected error output to contain '%s', got:\n%s", expectedSubstring, formattedError)
					}
				}

				// Check span formatting expectation
				hasSpanFormat := compilerError.Position.IsSpan()
				if hasSpanFormat != tt.expectedSpanFormat {
					t.Errorf("Expected span format: %v, got span format: %v in error:\n%s",
						tt.expectedSpanFormat, hasSpanFormat, formattedError)
				}
			}
		})
	}
}

// TestEndToEndSpanErrorReporting tests the complete flow from a markdown file
// with frontmatter errors through the compiler to formatted error output
func TestEndToEndSpanErrorReporting(t *testing.T) {
	// Create a temporary markdown file with multiple frontmatter errors
	workflowContent := `---
engine: unsupported-ai
on: push
max-turns: 200
tools:
  - name: git
  - type: shell
---

# Error Reporting Test

This workflow contains multiple frontmatter validation errors to test span-based error reporting.

The errors should be:
1. Invalid engine (should show span)
2. Invalid max-turns value (should show span) 
3. Missing tool name (may not show span due to complex path)
`

	tmpDir, err := os.MkdirTemp("", "end-to-end-span-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "error-test.md")
	if err := os.WriteFile(testFile, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate the compiler flow
	result, err := parser.ExtractFrontmatterFromContent(workflowContent)
	if err != nil {
		t.Fatal(err)
	}

	frontmatterYAML := strings.Join(result.FrontmatterLines, "\n")
	validator := NewFrontmatterValidator(frontmatterYAML)
	validationErrors := validator.ValidateFrontmatter(result.Frontmatter)

	// Should have 3 errors: invalid engine, invalid max-turns, missing tool name
	if len(validationErrors) != 3 {
		t.Fatalf("Expected 3 validation errors, got %d", len(validationErrors))
	}

	// Convert to compiler errors
	compilerErrors := ConvertValidationErrorsToCompilerErrors(
		testFile,
		result.FrontmatterStart,
		validationErrors,
	)

	// Verify error details
	var engineError, maxTurnsError, toolError *console.CompilerError
	for i := range compilerErrors {
		err := &compilerErrors[i]
		formatted := console.FormatError(*err)

		if strings.Contains(formatted, "unsupported engine") {
			engineError = err
		} else if strings.Contains(formatted, "max-turns must be") {
			maxTurnsError = err
		} else if strings.Contains(formatted, "tool must have") {
			toolError = err
		}
	}

	// Verify engine error has span
	if engineError == nil {
		t.Fatal("Engine error not found")
	}
	if !engineError.Position.IsSpan() {
		t.Error("Expected engine error to have span information")
	}
	engineFormatted := console.FormatError(*engineError)
	if !strings.Contains(engineFormatted, "error-test.md:2:9-22:") {
		t.Errorf("Expected engine error to show span format, got: %s", engineFormatted)
	}

	// Verify max-turns error has span
	if maxTurnsError == nil {
		t.Fatal("Max-turns error not found")
	}
	if !maxTurnsError.Position.IsSpan() {
		t.Error("Expected max-turns error to have span information")
	}
	maxTurnsFormatted := console.FormatError(*maxTurnsError)
	if !strings.Contains(maxTurnsFormatted, "error-test.md:4:12-14:") {
		t.Errorf("Expected max-turns error to show span format, got: %s", maxTurnsFormatted)
	}

	// Tool error might not have span (complex path)
	if toolError == nil {
		t.Fatal("Tool error not found")
	}
	// Note: Path is stored separately in the validation error, not in the message

	// All errors should have helpful hints
	allFormatted := make([]string, len(compilerErrors))
	for i, err := range compilerErrors {
		allFormatted[i] = console.FormatError(err)
		if !strings.Contains(allFormatted[i], "hint:") {
			t.Errorf("Expected error %d to have a hint, got: %s", i, allFormatted[i])
		}
	}
}

// TestSpanAccuracyWithRealFrontmatter tests that span calculations are accurate
// by comparing expected positions with actual frontmatter content
func TestSpanAccuracyWithRealFrontmatter(t *testing.T) {
	workflowContent := `---
title: Test Workflow
engine: invalid-engine
on:
  push:
    branches: [main]
max-turns: 999
---

# Test

Content here.`

	lines := strings.Split(workflowContent, "\n")

	result, err := parser.ExtractFrontmatterFromContent(workflowContent)
	if err != nil {
		t.Fatal(err)
	}

	frontmatterYAML := strings.Join(result.FrontmatterLines, "\n")
	validator := NewFrontmatterValidator(frontmatterYAML)
	validationErrors := validator.ValidateFrontmatter(result.Frontmatter)

	// Find the engine error
	var engineError *FrontmatterValidationError
	for _, err := range validationErrors {
		if err.Path == "engine" {
			engineError = &err
			break
		}
	}

	if engineError == nil {
		t.Fatal("Expected engine validation error")
	}

	if engineError.Span == nil {
		t.Fatal("Expected engine error to have span")
	}

	// Debug the actual span values
	t.Logf("Engine error span: StartLine=%d, StartColumn=%d, EndLine=%d, EndColumn=%d",
		engineError.Span.StartLine, engineError.Span.StartColumn,
		engineError.Span.EndLine, engineError.Span.EndColumn)

	// Verify the span points to the correct text
	// The engine value "invalid-engine" should be on line 2 within frontmatter (not line 3)
	expectedLine := 2 // Within frontmatter
	if engineError.Span.StartLine != expectedLine {
		t.Errorf("Expected engine error span to start at frontmatter line %d, got %d",
			expectedLine, engineError.Span.StartLine)
	}

	// The actual line in the file would be line 3 (2 + 1 for frontmatter start adjustment)
	actualFileLineIndex := engineError.Span.StartLine + result.FrontmatterStart - 2 // Convert to 0-based file index
	if actualFileLineIndex >= len(lines) {
		t.Fatalf("Span line %d exceeds file length %d", actualFileLineIndex, len(lines))
	}

	actualLine := lines[actualFileLineIndex]
	t.Logf("Actual line %d: %q (length: %d)", actualFileLineIndex+1, actualLine, len(actualLine))

	expectedValue := "invalid-engine"

	// Extract the substring using the span columns
	if engineError.Span.StartColumn <= 0 || engineError.Span.EndColumn > len(actualLine) {
		t.Fatalf("Span columns (%d-%d) are out of range for line %q (length %d)",
			engineError.Span.StartColumn, engineError.Span.EndColumn, actualLine, len(actualLine))
	}

	// Note: Columns are 1-based, so subtract 1 for 0-based string indexing
	spannedText := actualLine[engineError.Span.StartColumn-1 : engineError.Span.EndColumn]
	t.Logf("Spanned text: %q", spannedText)
	if spannedText != expectedValue {
		t.Errorf("Expected span to cover %q, but got %q from line %q (columns %d-%d)",
			expectedValue, spannedText, actualLine, engineError.Span.StartColumn, engineError.Span.EndColumn)
	}
}
