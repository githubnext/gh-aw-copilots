package workflow

import "testing"

func TestGeminiEngine(t *testing.T) {
	engine := NewGeminiEngine()

	// Test basic properties
	if engine.GetID() != "gemini" {
		t.Errorf("Expected ID 'gemini', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "Gemini CLI" {
		t.Errorf("Expected display name 'Gemini CLI', got '%s'", engine.GetDisplayName())
	}

	if engine.GetDescription() != "Uses Google Gemini CLI with GitHub integration and tool support" {
		t.Errorf("Expected description 'Uses Google Gemini CLI with GitHub integration and tool support', got '%s'", engine.GetDescription())
	}

	if engine.IsExperimental() {
		t.Error("Gemini engine should not be experimental")
	}

	if !engine.SupportsToolsWhitelist() {
		t.Error("Gemini engine should support tools whitelist")
	}

	// Test installation steps (should be empty for Gemini)
	steps := engine.GetInstallationSteps(nil)
	if len(steps) != 0 {
		t.Errorf("Expected no installation steps for Gemini, got %v", steps)
	}

	// Test execution config
	config := engine.GetExecutionConfig("test-workflow", "test-log", nil)
	if config.StepName != "Execute Gemini CLI Action" {
		t.Errorf("Expected step name 'Execute Gemini CLI Action', got '%s'", config.StepName)
	}

	if config.Action != "google-github-actions/run-gemini-cli@v1" {
		t.Errorf("Expected action 'google-github-actions/run-gemini-cli@v1', got '%s'", config.Action)
	}

	if config.Command != "" {
		t.Errorf("Expected empty command for Gemini (uses action), got '%s'", config.Command)
	}

	// Check that required inputs are present
	if _, hasPrompt := config.Inputs["prompt"]; !hasPrompt {
		t.Error("Expected prompt input to be present")
	}

	if config.Inputs["gemini_api_key"] != "${{ secrets.GEMINI_API_KEY }}" {
		t.Errorf("Expected gemini_api_key input, got '%s'", config.Inputs["gemini_api_key"])
	}

	// Check environment variables
	if config.Environment["GITHUB_TOKEN"] != "${{ secrets.GITHUB_TOKEN }}" {
		t.Errorf("Expected GITHUB_TOKEN environment variable, got '%s'", config.Environment["GITHUB_TOKEN"])
	}
}

func TestGeminiEngineWithModel(t *testing.T) {
	engine := NewGeminiEngine()

	// Test with model configuration
	engineConfig := &EngineConfig{
		ID:    "gemini",
		Model: "gemini-1.5-pro",
	}

	config := engine.GetExecutionConfig("test-workflow", "test-log", engineConfig)

	// Check that model is configured via settings
	expectedSettings := `{"model": "gemini-1.5-pro"}`
	if config.Inputs["settings"] != expectedSettings {
		t.Errorf("Expected settings input '%s', got '%s'", expectedSettings, config.Inputs["settings"])
	}
}

func TestGeminiEngineConfiguration(t *testing.T) {
	engine := NewGeminiEngine()

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
			if config.StepName != "Execute Gemini CLI Action" {
				t.Errorf("Expected step name 'Execute Gemini CLI Action', got '%s'", config.StepName)
			}

			if config.Action != "google-github-actions/run-gemini-cli@v1" {
				t.Errorf("Expected action 'google-github-actions/run-gemini-cli@v1', got '%s'", config.Action)
			}

			// Verify all required inputs are present
			requiredInputs := []string{"prompt", "gemini_api_key"}
			for _, input := range requiredInputs {
				if _, exists := config.Inputs[input]; !exists {
					t.Errorf("Expected input '%s' to be present", input)
				}
			}
		})
	}
}
