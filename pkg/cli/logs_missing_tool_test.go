package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseMissingToolsFromLog tests parsing missing tool reports from log content
func TestParseMissingToolsFromLog(t *testing.T) {
	testRun := WorkflowRun{
		DatabaseID:   12345,
		WorkflowName: "Test Workflow",
	}

	tests := []struct {
		name         string
		logContent   string
		expected     int
		expectTool   string
		expectReason string
	}{
		{
			name: "single_missing_tool",
			logContent: `
2024-01-01T10:00:00Z Step output: tools_reported=[{"tool":"docker","reason":"Need containerization","alternatives":"VM setup","timestamp":"2024-01-01T10:00:00Z"}]
`,
			expected:     1,
			expectTool:   "docker",
			expectReason: "Need containerization",
		},
		{
			name: "multiple_missing_tools",
			logContent: `
2024-01-01T10:00:00Z Step output: tools_reported=[{"tool":"docker","reason":"Need containerization","timestamp":"2024-01-01T10:00:00Z"},{"tool":"kubectl","reason":"K8s management","timestamp":"2024-01-01T10:00:00Z"}]
`,
			expected:   2,
			expectTool: "docker",
		},
		{
			name:       "no_missing_tools",
			logContent: "This is a regular log line with no missing tool reports",
			expected:   0,
		},
		{
			name:       "empty_log",
			logContent: "",
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := parseMissingToolsFromLog(tt.logContent, testRun, false)

			if len(tools) != tt.expected {
				t.Errorf("Expected %d tools, got %d", tt.expected, len(tools))
				return
			}

			if tt.expected > 0 && len(tools) > 0 {
				if tools[0].Tool != tt.expectTool {
					t.Errorf("Expected tool %s, got %s", tt.expectTool, tools[0].Tool)
				}

				if tt.expectReason != "" && tools[0].Reason != tt.expectReason {
					t.Errorf("Expected reason %s, got %s", tt.expectReason, tools[0].Reason)
				}

				// Check that run information was populated
				if tools[0].WorkflowName != testRun.WorkflowName {
					t.Errorf("Expected workflow name %s, got %s", testRun.WorkflowName, tools[0].WorkflowName)
				}

				if tools[0].RunID != testRun.DatabaseID {
					t.Errorf("Expected run ID %d, got %d", testRun.DatabaseID, tools[0].RunID)
				}
			}
		})
	}
}

// TestExtractMissingToolsFromRun tests extracting missing tools from a run directory
func TestExtractMissingToolsFromRun(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	testRun := WorkflowRun{
		DatabaseID:   67890,
		WorkflowName: "Integration Test",
	}

	// Create a log file with missing tool reports
	logContent := `Step completed successfully
tools_reported=[{"tool":"terraform","reason":"Infrastructure automation needed","alternatives":"Manual setup","timestamp":"2024-01-01T12:00:00Z"}]
Process completed with exit code 0`

	logFile := filepath.Join(tmpDir, "workflow.log")
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Extract missing tools
	tools, err := extractMissingToolsFromRun(tmpDir, testRun, false)
	if err != nil {
		t.Fatalf("Error extracting missing tools: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0]
	if tool.Tool != "terraform" {
		t.Errorf("Expected tool 'terraform', got '%s'", tool.Tool)
	}

	if tool.Reason != "Infrastructure automation needed" {
		t.Errorf("Expected reason 'Infrastructure automation needed', got '%s'", tool.Reason)
	}

	if tool.WorkflowName != testRun.WorkflowName {
		t.Errorf("Expected workflow name '%s', got '%s'", testRun.WorkflowName, tool.WorkflowName)
	}

	if tool.RunID != testRun.DatabaseID {
		t.Errorf("Expected run ID %d, got %d", testRun.DatabaseID, tool.RunID)
	}
}

// TestDisplayMissingToolsAnalysis tests the display functionality
func TestDisplayMissingToolsAnalysis(t *testing.T) {
	// This is a smoke test to ensure the function doesn't panic
	processedRuns := []ProcessedRun{
		{
			Run: WorkflowRun{
				DatabaseID:   1001,
				WorkflowName: "Workflow A",
			},
			MissingTools: []MissingToolReport{
				{
					Tool:         "docker",
					Reason:       "Containerization needed",
					Alternatives: "VM setup",
					WorkflowName: "Workflow A",
					RunID:        1001,
				},
				{
					Tool:         "kubectl",
					Reason:       "K8s management",
					WorkflowName: "Workflow A",
					RunID:        1001,
				},
			},
		},
		{
			Run: WorkflowRun{
				DatabaseID:   1002,
				WorkflowName: "Workflow B",
			},
			MissingTools: []MissingToolReport{
				{
					Tool:         "docker",
					Reason:       "Need containers for deployment",
					WorkflowName: "Workflow B",
					RunID:        1002,
				},
			},
		},
	}

	// Test non-verbose mode (should not panic)
	displayMissingToolsAnalysis(processedRuns, false)

	// Test verbose mode (should not panic)
	displayMissingToolsAnalysis(processedRuns, true)
}
