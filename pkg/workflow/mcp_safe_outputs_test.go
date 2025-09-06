package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeOutputsMCPServerIntegration(t *testing.T) {
	tests := []struct {
		name            string
		engine          string
		workflowContent string
		expectMCPSetup  bool
		expectNodeSetup bool
		expectConfig    []string // expected safe output types in config
	}{
		{
			name:   "Claude with safe outputs",
			engine: "claude",
			workflowContent: `---
engine: claude
safe-outputs:
  create-issue:
  add-issue-comment:
  create-pull-request:
---

# Test Workflow
Test content`,
			expectMCPSetup:  true,
			expectNodeSetup: true,
			expectConfig:    []string{"create-issue", "add-issue-comment", "create-pull-request"},
		},
		{
			name:   "Codex with safe outputs",
			engine: "codex",
			workflowContent: `---
engine: codex
safe-outputs:
  create-issue:
  update-issue:
  add-issue-label:
  create-discussion:
---

# Test Workflow
Test content`,
			expectMCPSetup:  true,
			expectNodeSetup: true,
			expectConfig:    []string{"create-issue", "update-issue", "add-issue-label", "create-discussion"},
		},
		{
			name:   "Claude without safe outputs",
			engine: "claude",
			workflowContent: `---
engine: claude
---

# Test Workflow
Test content`,
			expectMCPSetup:  false,
			expectNodeSetup: false,
			expectConfig:    []string{},
		},
		{
			name:   "Codex without safe outputs",
			engine: "codex",
			workflowContent: `---
engine: codex
---

# Test Workflow
Test content`,
			expectMCPSetup:  false,
			expectNodeSetup: false,
			expectConfig:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files
			tmpDir, err := os.MkdirTemp("", "test-safe-outputs-mcp-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Write test workflow file
			testFile := filepath.Join(tmpDir, "test-workflow.md")
			if err := os.WriteFile(testFile, []byte(tt.workflowContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			compiler := NewCompiler(false, "", "test")

			// Compile the workflow
			err = compiler.CompileWorkflow(testFile)
			if err != nil {
				t.Fatalf("Failed to compile workflow: %v", err)
			}

			// Read the generated lock file
			lockFile := filepath.Join(tmpDir, "test-workflow.lock.yml")
			content, err := os.ReadFile(lockFile)
			if err != nil {
				t.Fatalf("Failed to read generated lock file: %v", err)
			}

			result := string(content)

			// Check if Node.js/TypeScript setup is present
			hasNodeSetup := strings.Contains(result, "Setup Node.js and TypeScript dependencies")
			if hasNodeSetup != tt.expectNodeSetup {
				t.Errorf("Expected Node.js setup: %v, got: %v", tt.expectNodeSetup, hasNodeSetup)
			}

			// Check if MCP server setup is present
			hasMCPServerSetup := strings.Contains(result, "Setup Safe Outputs MCP Server")
			if hasMCPServerSetup != tt.expectMCPSetup {
				t.Errorf("Expected MCP server setup: %v, got: %v", tt.expectMCPSetup, hasMCPServerSetup)
			}

			// Check if safe-outputs MCP server is in configuration
			hasSafeOutputsConfig := strings.Contains(result, `"safe-outputs"`) || strings.Contains(result, "[mcp_servers.safe-outputs]")
			if hasSafeOutputsConfig != tt.expectMCPSetup {
				t.Errorf("Expected safe-outputs MCP config: %v, got: %v", tt.expectMCPSetup, hasSafeOutputsConfig)
			}

			// Check for expected configuration content
			for _, expectedConfig := range tt.expectConfig {
				configPattern := `"` + expectedConfig + `": true`
				if !strings.Contains(result, configPattern) && tt.expectMCPSetup {
					t.Errorf("Expected configuration for %s not found in output", expectedConfig)
				}
			}

			// Check that MCP server script is embedded if expected
			if tt.expectMCPSetup {
				if !strings.Contains(result, "MCP Server for Safe Outputs") {
					t.Error("Expected MCP server script content not found")
				}
				if !strings.Contains(result, "/tmp/mcp-safe-outputs/server.ts") {
					t.Error("Expected MCP server path not found")
				}
			}

			// Engine-specific checks
			if tt.expectMCPSetup {
				switch tt.engine {
				case "claude":
					// Should have JSON format MCP config with npx tsx
					if !strings.Contains(result, `"command": "npx"`) {
						t.Error("Expected Claude JSON MCP config with npx not found")
					}
					if !strings.Contains(result, `"tsx"`) {
						t.Error("Expected tsx argument in Claude MCP config not found")
					}
				case "codex":
					// Should have TOML format MCP config with npx tsx
					if !strings.Contains(result, `command = "npx"`) {
						t.Error("Expected Codex TOML MCP config with npx not found")
					}
					if !strings.Contains(result, `"tsx"`) {
						t.Error("Expected tsx argument in Codex MCP config not found")
					}
				}
			}
		})
	}
}

func TestSafeOutputsMCPServerConfigGeneration(t *testing.T) {
	tests := []struct {
		name        string
		safeOutputs *SafeOutputsConfig
	}{
		{
			name: "Multiple safe outputs",
			safeOutputs: &SafeOutputsConfig{
				CreateIssues:     &CreateIssuesConfig{},
				AddIssueComments: &AddIssueCommentsConfig{},
				UpdateIssues:     &UpdateIssuesConfig{},
			},
		},
		{
			name: "Single safe output",
			safeOutputs: &SafeOutputsConfig{
				CreateIssues: &CreateIssuesConfig{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(false, "", "test")
			var yaml strings.Builder

			compiler.generateSafeOutputsConfigJSON(&yaml, tt.safeOutputs)

			result := yaml.String()

			// Check that all expected keys are present
			if tt.safeOutputs.CreateIssues != nil && !strings.Contains(result, `"create-issue": true`) {
				t.Error("Expected create-issue config not found")
			}
			if tt.safeOutputs.AddIssueComments != nil && !strings.Contains(result, `"add-issue-comment": true`) {
				t.Error("Expected add-issue-comment config not found")
			}
			if tt.safeOutputs.UpdateIssues != nil && !strings.Contains(result, `"update-issue": true`) {
				t.Error("Expected update-issue config not found")
			}
		})
	}
}
