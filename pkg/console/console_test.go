package console

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      CompilerError
		expected []string // Substrings that should be present in output
	}{
		{
			name: "basic error with position",
			err: CompilerError{
				Position: ErrorPosition{
					File:   "test.md",
					Line:   5,
					Column: 10,
				},
				Type:    "error",
				Message: "invalid syntax",
			},
			expected: []string{
				"test.md:5:10:",
				"error:",
				"invalid syntax",
			},
		},
		{
			name: "warning with hint",
			err: CompilerError{
				Position: ErrorPosition{
					File:   "workflow.md",
					Line:   2,
					Column: 1,
				},
				Type:    "warning",
				Message: "deprecated field",
				Hint:    "use 'new_field' instead",
			},
			expected: []string{
				"workflow.md:2:1:",
				"warning:",
				"deprecated field",
				"hint:",
				"use 'new_field' instead",
			},
		},
		{
			name: "error with context",
			err: CompilerError{
				Position: ErrorPosition{
					File:   "test.md",
					Line:   3,
					Column: 5,
				},
				Type:    "error",
				Message: "missing colon",
				Context: []string{
					"tools:",
					"  github",
					"    allowed: [list_issues]",
				},
			},
			expected: []string{
				"test.md:3:5:",
				"error:",
				"missing colon",
				"2 |",
				"3 |",
				"4 |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatError(tt.err)

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestFormatSuccessMessage(t *testing.T) {
	output := FormatSuccessMessage("compilation completed")
	if !strings.Contains(output, "compilation completed") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Expected output to contain checkmark, got: %s", output)
	}
}

func TestFormatInfoMessage(t *testing.T) {
	output := FormatInfoMessage("processing file")
	if !strings.Contains(output, "processing file") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "ℹ") {
		t.Errorf("Expected output to contain info icon, got: %s", output)
	}
}

func TestFormatWarningMessage(t *testing.T) {
	output := FormatWarningMessage("deprecated syntax")
	if !strings.Contains(output, "deprecated syntax") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "⚠") {
		t.Errorf("Expected output to contain warning icon, got: %s", output)
	}
}

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name     string
		config   TableConfig
		expected []string // Substrings that should be present in output
	}{
		{
			name: "simple table",
			config: TableConfig{
				Headers: []string{"ID", "Name", "Status"},
				Rows: [][]string{
					{"1", "Test", "Active"},
					{"2", "Demo", "Inactive"},
				},
			},
			expected: []string{
				"ID",
				"Name",
				"Status",
				"Test",
				"Demo",
				"Active",
				"Inactive",
			},
		},
		{
			name: "table with title and total",
			config: TableConfig{
				Title:   "Workflow Results",
				Headers: []string{"Run", "Duration", "Cost"},
				Rows: [][]string{
					{"123", "5m", "$0.50"},
					{"456", "3m", "$0.30"},
				},
				ShowTotal: true,
				TotalRow:  []string{"TOTAL", "8m", "$0.80"},
			},
			expected: []string{
				"Workflow Results",
				"Run",
				"Duration",
				"Cost",
				"123",
				"456",
				"TOTAL",
				"8m",
				"$0.80",
			},
		},
		{
			name: "empty table",
			config: TableConfig{
				Headers: []string{},
				Rows:    [][]string{},
			},
			expected: []string{}, // Should return empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderTable(tt.config)

			if len(tt.expected) == 0 {
				if output != "" {
					t.Errorf("Expected empty output for empty table config, got: %s", output)
				}
				return
			}

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestFormatLocationMessage(t *testing.T) {
	output := FormatLocationMessage("Downloaded to: /path/to/logs")
	if !strings.Contains(output, "Downloaded to: /path/to/logs") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "📁") {
		t.Errorf("Expected output to contain folder icon, got: %s", output)
	}
}

func TestToRelativePath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedFunc func(string, string) bool // Compare function that takes result and expected pattern
	}{
		{
			name: "relative path unchanged",
			path: "test.md",
			expectedFunc: func(result, expected string) bool {
				return result == "test.md"
			},
		},
		{
			name: "nested relative path unchanged",
			path: "pkg/console/test.md",
			expectedFunc: func(result, expected string) bool {
				return result == "pkg/console/test.md"
			},
		},
		{
			name: "absolute path converted to relative",
			path: "/tmp/test.md",
			expectedFunc: func(result, expected string) bool {
				// Should be a relative path that doesn't start with /
				return !strings.HasPrefix(result, "/") && strings.HasSuffix(result, "test.md")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToRelativePath(tt.path)
			if !tt.expectedFunc(result, tt.path) {
				t.Errorf("ToRelativePath(%s) = %s, but validation failed", tt.path, result)
			}
		})
	}
}

func TestFormatErrorWithAbsolutePaths(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	err := CompilerError{
		Position: ErrorPosition{
			File:   tmpFile,
			Line:   5,
			Column: 10,
		},
		Type:    "error",
		Message: "invalid syntax",
	}

	output := FormatError(err)

	// The output should contain test.md and line:column information
	if !strings.Contains(output, "test.md:5:10:") {
		t.Errorf("Expected output to contain relative file path with line:column, got: %s", output)
	}

	// The output should not start with an absolute path (no leading /)
	lines := strings.Split(output, "\n")
	if strings.HasPrefix(lines[0], "/") {
		t.Errorf("Expected output to start with relative path, but found absolute path: %s", lines[0])
	}

	// Should contain error message
	if !strings.Contains(output, "invalid syntax") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}
}

func TestErrorPosition(t *testing.T) {
	tests := []struct {
		name   string
		pos    ErrorPosition
		isSpan bool
	}{
		{
			name:   "basic position",
			pos:    NewErrorPosition("test.yaml", 5, 10),
			isSpan: false,
		},
		{
			name:   "single line span",
			pos:    NewErrorPositionWithSpan("test.yaml", 5, 10, 5, 20),
			isSpan: true,
		},
		{
			name:   "multi-line span",
			pos:    NewErrorPositionWithSpan("test.yaml", 5, 10, 8, 15),
			isSpan: true,
		},
		{
			name:   "zero end values not a span",
			pos:    ErrorPosition{File: "test.yaml", Line: 5, Column: 10, EndLine: 0, EndColumn: 0},
			isSpan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pos.IsSpan() != tt.isSpan {
				t.Errorf("IsSpan() = %v, want %v", tt.pos.IsSpan(), tt.isSpan)
			}
		})
	}
}

func TestFormatErrorWithSpan(t *testing.T) {
	tests := []struct {
		name     string
		err      CompilerError
		expected []string // Substrings that should be present
	}{
		{
			name: "error with single line span",
			err: CompilerError{
				Position: NewErrorPositionWithSpan("test.yaml", 5, 10, 5, 20),
				Type:     "error",
				Message:  "invalid syntax",
			},
			expected: []string{
				"test.yaml:5:10-20:",
				"error:",
				"invalid syntax",
			},
		},
		{
			name: "error with multi-line span",
			err: CompilerError{
				Position: NewErrorPositionWithSpan("workflow.md", 3, 8, 6, 15),
				Type:     "warning",
				Message:  "deprecated field",
			},
			expected: []string{
				"workflow.md:3:8-6:15:",
				"warning:",
				"deprecated field",
			},
		},
		{
			name: "backwards compatibility - no span",
			err: CompilerError{
				Position: NewErrorPosition("old.yaml", 2, 5),
				Type:     "error",
				Message:  "missing field",
			},
			expected: []string{
				"old.yaml:2:5:",
				"error:",
				"missing field",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatError(tt.err)
			
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestSourceSpanToErrorPosition(t *testing.T) {
	span := struct {
		StartLine   int
		StartColumn int
		EndLine     int
		EndColumn   int
	}{
		StartLine:   3,
		StartColumn: 8,
		EndLine:     5,
		EndColumn:   12,
	}
	
	pos := SourceSpanToErrorPosition("test.yaml", span)
	
	expected := NewErrorPositionWithSpan("test.yaml", 3, 8, 5, 12)
	if pos != expected {
		t.Errorf("SourceSpanToErrorPosition() = %+v, want %+v", pos, expected)
	}
	
	if !pos.IsSpan() {
		t.Error("Expected converted position to be a span")
	}
}
