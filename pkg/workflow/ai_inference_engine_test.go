package workflow

import (
	"strings"
	"testing"
)

func TestAIInferenceEngine(t *testing.T) {
	engine := NewAIInferenceEngine()

	// Test basic engine properties
	if engine.GetID() != "ai-inference" {
		t.Errorf("Expected ID 'ai-inference', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "AI Inference" {
		t.Errorf("Expected display name 'AI Inference', got '%s'", engine.GetDisplayName())
	}

	if engine.IsExperimental() {
		t.Error("Expected AI Inference engine to be stable (not experimental)")
	}

	if !engine.SupportsToolsWhitelist() {
		t.Error("Expected AI Inference engine to support tools whitelist")
	}

	// Test installation steps (should be empty)
	steps := engine.GetInstallationSteps(nil)
	if len(steps) != 0 {
		t.Errorf("Expected no installation steps, got %d", len(steps))
	}
}

func TestAIInferenceExecutionConfig(t *testing.T) {
	engine := NewAIInferenceEngine()

	// Test default execution config
	config := engine.GetExecutionConfig("test-workflow", "test.log", nil)

	if config.StepName != "Execute AI Inference Action" {
		t.Errorf("Expected step name 'Execute AI Inference Action', got '%s'", config.StepName)
	}

	if config.Action != "actions/ai-inference@v1" {
		t.Errorf("Expected action 'actions/ai-inference@v1', got '%s'", config.Action)
	}

	// Check required inputs
	expectedInputs := map[string]string{
		"prompt-file": "/tmp/aw-prompts/prompt.txt",
		"token":       "${{ secrets.GITHUB_TOKEN }}",
		"mcp-config":  "/tmp/mcp-config/mcp-servers.json",
		"model":       "openai/gpt-4o",
		"max-tokens":  "2000",
	}

	for key, expectedValue := range expectedInputs {
		if actualValue, exists := config.Inputs[key]; !exists {
			t.Errorf("Expected input '%s' to be present", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected input '%s' to be '%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	// Test with custom model configuration
	engineConfig := &EngineConfig{
		ID:    "ai-inference",
		Model: "anthropic/claude-3.5-sonnet",
	}

	configWithModel := engine.GetExecutionConfig("test-workflow", "test.log", engineConfig)
	if configWithModel.Inputs["model"] != "anthropic/claude-3.5-sonnet" {
		t.Errorf("Expected custom model 'anthropic/claude-3.5-sonnet', got '%s'", configWithModel.Inputs["model"])
	}
}

func TestAIInferenceMCPConfig(t *testing.T) {
	engine := NewAIInferenceEngine()
	yaml := &strings.Builder{}

	// Test with GitHub tools only
	tools := map[string]any{
		"github": map[string]any{
			"allowed": []string{"get_issue", "add_issue_comment"},
		},
	}
	mcpTools := []string{"github"}

	engine.RenderMCPConfig(yaml, tools, mcpTools)

	output := yaml.String()
	if !strings.Contains(output, "mcp-servers.json") {
		t.Error("Expected MCP servers configuration file generation")
	}
	if !strings.Contains(output, "\"github\": {") {
		t.Error("Expected GitHub MCP server configuration")
	}
	if !strings.Contains(output, "ghcr.io/github/github-mcp-server:") {
		t.Error("Expected dockerized GitHub MCP server")
	}

	// Test with custom MCP tools
	yaml.Reset()
	toolsWithCustom := map[string]any{
		"github": map[string]any{
			"allowed": []string{"get_issue"},
		},
		"custom": map[string]any{
			"mcp": map[string]any{
				"type":    "stdio",
				"command": "custom-mcp-server",
			},
		},
	}
	mcpToolsCustom := []string{"github", "custom"}

	engine.RenderMCPConfig(yaml, toolsWithCustom, mcpToolsCustom)

	outputWithCustom := yaml.String()
	if !strings.Contains(outputWithCustom, "\"custom\": {") {
		t.Error("Expected custom MCP tool configuration")
	}
	if !strings.Contains(outputWithCustom, "custom-mcp-server") {
		t.Error("Expected custom MCP server command")
	}
}

func TestAIInferenceEngineRegistry(t *testing.T) {
	registry := NewEngineRegistry()

	// Test that AI Inference engine is registered
	engine, err := registry.GetEngine("ai-inference")
	if err != nil {
		t.Errorf("Expected to find ai-inference engine, got error: %v", err)
	}

	if engine.GetID() != "ai-inference" {
		t.Errorf("Expected ai-inference engine ID, got '%s'", engine.GetID())
	}

	// Test that it's included in supported engines
	supportedEngines := registry.GetSupportedEngines()
	found := false
	for _, id := range supportedEngines {
		if id == "ai-inference" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected ai-inference to be in supported engines list")
	}

	// Test engine validation
	if !registry.IsValidEngine("ai-inference") {
		t.Error("Expected ai-inference to be a valid engine")
	}
}
