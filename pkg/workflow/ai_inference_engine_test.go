package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIInferenceEngineProperties(t *testing.T) {
	engine := NewAIInferenceEngine()

	// Test basic properties
	if engine.GetID() != "ai-inference" {
		t.Errorf("Expected ID 'ai-inference', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "AI Inference" {
		t.Errorf("Expected display name 'AI Inference', got '%s'", engine.GetDisplayName())
	}

	if engine.IsExperimental() != false {
		t.Errorf("Expected AI Inference engine to not be experimental")
	}

	if engine.SupportsToolsWhitelist() != false {
		t.Errorf("Expected AI Inference engine to not support tools whitelist")
	}

	if engine.SupportsHTTPTransport() != false {
		t.Errorf("Expected AI Inference engine to not support HTTP transport")
	}

	if engine.SupportsMaxTurns() != false {
		t.Errorf("Expected AI Inference engine to not support max turns")
	}
}

func TestAIInferenceEngineInstallation(t *testing.T) {
	engine := NewAIInferenceEngine()
	steps := engine.GetInstallationSteps(nil, nil)

	// AI Inference should not require any installation steps
	if len(steps) != 0 {
		t.Errorf("Expected 0 installation steps, got %d", len(steps))
	}
}

func TestAIInferenceEngineExecutionConfig(t *testing.T) {
	engine := NewAIInferenceEngine()

	tests := []struct {
		name         string
		engineConfig *EngineConfig
		hasOutput    bool
		expectedAction string
		expectedModel  string
	}{
		{
			name:          "default configuration",
			engineConfig:  nil,
			hasOutput:     false,
			expectedAction: "actions/ai-inference@v1",
			expectedModel:  "gpt-4o-mini",
		},
		{
			name: "custom version and model",
			engineConfig: &EngineConfig{
				ID:      "ai-inference",
				Version: "v2",
				Model:   "gpt-4o",
			},
			hasOutput:     true,
			expectedAction: "actions/ai-inference@v2",
			expectedModel:  "gpt-4o",
		},
		{
			name: "with environment variables",
			engineConfig: &EngineConfig{
				ID: "ai-inference",
				Env: map[string]string{
					"CUSTOM_VAR": "custom_value",
				},
			},
			hasOutput:     true,
			expectedAction: "actions/ai-inference@v1",
			expectedModel:  "gpt-4o-mini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := engine.GetExecutionConfig("test-workflow", "/tmp/test.log", tt.engineConfig, nil, tt.hasOutput)

			// Check action
			if config.Action != tt.expectedAction {
				t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, config.Action)
			}

			// Check model input
			if config.Inputs["model"] != tt.expectedModel {
				t.Errorf("Expected model '%s', got '%s'", tt.expectedModel, config.Inputs["model"])
			}

			// Check prompt file environment variable is set
			if !strings.Contains(config.Inputs["env"], "GITHUB_AW_PROMPT_FILE: /tmp/aw-prompts/prompt.txt") {
				t.Errorf("Expected GITHUB_AW_PROMPT_FILE environment variable to be set")
			}

			// Check safe outputs environment variable when hasOutput is true
			if tt.hasOutput {
				if !strings.Contains(config.Inputs["env"], "GITHUB_AW_SAFE_OUTPUTS: ${{ env.GITHUB_AW_SAFE_OUTPUTS }}") {
					t.Errorf("Expected GITHUB_AW_SAFE_OUTPUTS environment variable to be set when hasOutput=true")
				}
			}

			// Check custom environment variables
			if tt.engineConfig != nil && len(tt.engineConfig.Env) > 0 {
				for key, value := range tt.engineConfig.Env {
					expectedEnvVar := key + ": " + value
					if !strings.Contains(config.Inputs["env"], expectedEnvVar) {
						t.Errorf("Expected custom environment variable '%s' to be set", expectedEnvVar)
					}
				}
			}
		})
	}
}

func TestAIInferenceWorkflowCompilation(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ai-inference-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	tests := []struct {
		name        string
		frontmatter string
		expectError bool
	}{
		{
			name: "simple ai-inference workflow",
			frontmatter: `---
on:
  issues:
    types: [opened]
engine: ai-inference
safe-outputs:
  add-issue-label:
---`,
			expectError: false,
		},
		{
			name: "ai-inference with custom model",
			frontmatter: `---
on:
  issues:
    types: [opened]
engine:
  id: ai-inference
  model: gpt-4o
safe-outputs:
  add-issue-label:
---`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.md")
			content := tt.frontmatter + "\n\nAnalyze the issue and apply appropriate labels."
			if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Compile workflow
			err := compiler.CompileWorkflow(testFile)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but compilation succeeded")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error compiling workflow: %v", err)
			}

			if !tt.expectError {
				// Read the generated lock file
				lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
				content, err := os.ReadFile(lockFile)
				if err != nil {
					t.Fatalf("Failed to read lock file: %v", err)
				}

				lockContent := string(content)

				// Check that AI Inference action is present
				if !strings.Contains(lockContent, "actions/ai-inference@") {
					t.Errorf("Expected lock file to contain 'actions/ai-inference@' but it didn't.\nContent:\n%s", lockContent)
				}

				// Check that prompt file environment variable is set
				if !strings.Contains(lockContent, "GITHUB_AW_PROMPT_FILE: /tmp/aw-prompts/prompt.txt") {
					t.Errorf("Expected lock file to contain 'GITHUB_AW_PROMPT_FILE' environment variable but it didn't.\nContent:\n%s", lockContent)
				}

				// Check that safe outputs environment variable is set
				if !strings.Contains(lockContent, "GITHUB_AW_SAFE_OUTPUTS: ${{ env.GITHUB_AW_SAFE_OUTPUTS }}") {
					t.Errorf("Expected lock file to contain 'GITHUB_AW_SAFE_OUTPUTS' environment variable but it didn't.\nContent:\n%s", lockContent)
				}
			}
		})
	}
}

func TestAIInferenceLogParsing(t *testing.T) {
	engine := NewAIInferenceEngine()

	tests := []struct {
		name           string
		logContent     string
		expectedTokens int
		expectedErrors int
		expectedWarnings int
	}{
		{
			name:           "empty log",
			logContent:     "",
			expectedTokens: 0,
			expectedErrors: 0,
			expectedWarnings: 0,
		},
		{
			name: "log with JSON metrics",
			logContent: `{"usage": {"total_tokens": 100}}
{"cost": 0.002}
Error: Something went wrong
Warning: This is a warning`,
			expectedTokens: 100,
			expectedErrors: 1,
			expectedWarnings: 1,
		},
		{
			name: "log with separate input/output tokens",
			logContent: `{"usage": {"prompt_tokens": 50, "completion_tokens": 25}}
Another error occurred`,
			expectedTokens: 75,
			expectedErrors: 1,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := engine.ParseLogMetrics(tt.logContent, false)

			if metrics.TokenUsage != tt.expectedTokens {
				t.Errorf("Expected token usage %d, got %d", tt.expectedTokens, metrics.TokenUsage)
			}

			if metrics.ErrorCount != tt.expectedErrors {
				t.Errorf("Expected error count %d, got %d", tt.expectedErrors, metrics.ErrorCount)
			}

			if metrics.WarningCount != tt.expectedWarnings {
				t.Errorf("Expected warning count %d, got %d", tt.expectedWarnings, metrics.WarningCount)
			}
		})
	}
}

func TestAIInferenceEngineRegistry(t *testing.T) {
	registry := NewEngineRegistry()
	
	// Test that ai-inference engine is registered
	engine, err := registry.GetEngine("ai-inference")
	if err != nil {
		t.Errorf("Expected ai-inference engine to be registered, got error: %v", err)
	}

	if engine.GetID() != "ai-inference" {
		t.Errorf("Expected engine ID 'ai-inference', got '%s'", engine.GetID())
	}

	// Test that ai-inference is in supported engines list
	supportedEngines := registry.GetSupportedEngines()
	found := false
	for _, engineID := range supportedEngines {
		if engineID == "ai-inference" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'ai-inference' to be in supported engines list: %v", supportedEngines)
	}
}