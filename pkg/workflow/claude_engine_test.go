package workflow

import "testing"

func TestClaudeEngine(t *testing.T) {
	engine := NewClaudeEngine()

	// Test basic properties
	if engine.GetID() != "claude" {
		t.Errorf("Expected ID 'claude', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "Claude Code" {
		t.Errorf("Expected display name 'Claude Code', got '%s'", engine.GetDisplayName())
	}

	if engine.GetDescription() != "Uses Claude Code with full MCP tool support and allow-listing" {
		t.Errorf("Expected description 'Uses Claude Code with full MCP tool support and allow-listing', got '%s'", engine.GetDescription())
	}

	if engine.IsExperimental() {
		t.Error("Claude engine should not be experimental")
	}

	if !engine.SupportsToolsWhitelist() {
		t.Error("Claude engine should support MCP tools")
	}

	// Test installation steps (should be empty for Claude)
	steps := engine.GetInstallationSteps(nil)
	if len(steps) != 0 {
		t.Errorf("Expected no installation steps for Claude, got %v", steps)
	}

	// Test execution config
	config := engine.GetExecutionConfig("test-workflow", "test-log", nil)
	if config.StepName != "Execute Claude Code Action" {
		t.Errorf("Expected step name 'Execute Claude Code Action', got '%s'", config.StepName)
	}

	if config.Action != "anthropics/claude-code-base-action@beta" {
		t.Errorf("Expected action 'anthropics/claude-code-base-action@beta', got '%s'", config.Action)
	}

	if config.Command != "" {
		t.Errorf("Expected empty command for Claude (uses action), got '%s'", config.Command)
	}

	// Check that required inputs are present
	if config.Inputs["prompt_file"] != "/tmp/aw-prompts/prompt.txt" {
		t.Errorf("Expected prompt_file input, got '%s'", config.Inputs["prompt_file"])
	}

	if config.Inputs["anthropic_api_key"] != "${{ secrets.ANTHROPIC_API_KEY }}" {
		t.Errorf("Expected anthropic_api_key input, got '%s'", config.Inputs["anthropic_api_key"])
	}

	if config.Inputs["mcp_config"] != "/tmp/mcp-config/mcp-servers.json" {
		t.Errorf("Expected mcp_config input, got '%s'", config.Inputs["mcp_config"])
	}

	expectedClaudeEnv := "|\n            GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}"
	if config.Inputs["claude_env"] != expectedClaudeEnv {
		t.Errorf("Expected claude_env input '%s', got '%s'", expectedClaudeEnv, config.Inputs["claude_env"])
	}

	// Check that special fields are present but empty (will be filled during generation)
	if _, hasAllowedTools := config.Inputs["allowed_tools"]; !hasAllowedTools {
		t.Error("Expected allowed_tools input to be present")
	}

	if _, hasTimeoutMinutes := config.Inputs["timeout_minutes"]; !hasTimeoutMinutes {
		t.Error("Expected timeout_minutes input to be present")
	}

	// Check environment variables
	if config.Environment["GH_TOKEN"] != "${{ secrets.GITHUB_TOKEN }}" {
		t.Errorf("Expected GH_TOKEN environment variable, got '%s'", config.Environment["GH_TOKEN"])
	}
}

func TestClaudeEngineConfiguration(t *testing.T) {
	engine := NewClaudeEngine()

	// Test different workflow names and log files
	testCases := []struct {
		workflowName string
		logFile      string
	}{
		{"simple-workflow", "simple-log"},
		{"complex workflow with spaces", "complex-log"},
		{"workflow-with-hyphens", "workflow-log"},
	}

	for _, tc := range testCases {
		t.Run(tc.workflowName, func(t *testing.T) {
			config := engine.GetExecutionConfig(tc.workflowName, tc.logFile, nil)

			// Verify the configuration is consistent regardless of input
			if config.StepName != "Execute Claude Code Action" {
				t.Errorf("Expected step name 'Execute Claude Code Action', got '%s'", config.StepName)
			}

			if config.Action != "anthropics/claude-code-base-action@beta" {
				t.Errorf("Expected action 'anthropics/claude-code-base-action@beta', got '%s'", config.Action)
			}

			// Verify all required inputs are present
			requiredInputs := []string{"prompt_file", "anthropic_api_key", "mcp_config", "claude_env", "allowed_tools", "timeout_minutes"}
			for _, input := range requiredInputs {
				if _, exists := config.Inputs[input]; !exists {
					t.Errorf("Expected input '%s' to be present", input)
				}
			}
		})
	}
}
