package workflow

import (
	"strings"
	"testing"
)

func TestCustomEngine(t *testing.T) {
	engine := NewCustomEngine()

	// Test basic engine properties
	if engine.GetID() != "custom" {
		t.Errorf("Expected ID 'custom', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "Custom Steps" {
		t.Errorf("Expected display name 'Custom Steps', got '%s'", engine.GetDisplayName())
	}

	if engine.GetDescription() != "Executes user-defined GitHub Actions steps" {
		t.Errorf("Expected description 'Executes user-defined GitHub Actions steps', got '%s'", engine.GetDescription())
	}

	if engine.IsExperimental() {
		t.Error("Expected custom engine to not be experimental")
	}

	if engine.SupportsToolsWhitelist() {
		t.Error("Expected custom engine to not support tools whitelist")
	}

	if engine.SupportsHTTPTransport() {
		t.Error("Expected custom engine to not support HTTP transport")
	}

	if engine.SupportsMaxTurns() {
		t.Error("Expected custom engine to not support max turns")
	}
}

func TestCustomEngineGetInstallationSteps(t *testing.T) {
	engine := NewCustomEngine()

	steps := engine.GetInstallationSteps(nil, nil)
	if len(steps) != 0 {
		t.Errorf("Expected 0 installation steps for custom engine, got %d", len(steps))
	}
}

func TestCustomEngineGetExecutionConfig(t *testing.T) {
	engine := NewCustomEngine()

	config := engine.GetExecutionConfig("test-workflow", "/tmp/test.log", nil, nil, false)

	if config.StepName != "Custom Steps Execution" {
		t.Errorf("Expected step name 'Custom Steps Execution', got '%s'", config.StepName)
	}

	if !strings.Contains(config.Command, "Custom steps are handled directly by the compiler") {
		t.Errorf("Expected command to mention compiler handling, got '%s'", config.Command)
	}

	if config.Environment["WORKFLOW_NAME"] != "test-workflow" {
		t.Errorf("Expected WORKFLOW_NAME env var to be 'test-workflow', got '%s'", config.Environment["WORKFLOW_NAME"])
	}
}

func TestCustomEngineRenderMCPConfig(t *testing.T) {
	engine := NewCustomEngine()
	var yaml strings.Builder

	// This should not generate any MCP configuration
	engine.RenderMCPConfig(&yaml, map[string]any{}, []string{})

	output := yaml.String()
	if output != "" {
		t.Errorf("Expected empty MCP config for custom engine, got '%s'", output)
	}
}

func TestCustomEngineParseLogMetrics(t *testing.T) {
	engine := NewCustomEngine()

	logContent := `This is a test log
Error: Something went wrong
Warning: This is a warning
Another line
ERROR: Another error`

	metrics := engine.ParseLogMetrics(logContent, false)

	if metrics.ErrorCount != 2 {
		t.Errorf("Expected 2 errors, got %d", metrics.ErrorCount)
	}

	if metrics.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", metrics.WarningCount)
	}

	if metrics.TokenUsage != 0 {
		t.Errorf("Expected 0 token usage, got %d", metrics.TokenUsage)
	}

	if metrics.EstimatedCost != 0 {
		t.Errorf("Expected 0 estimated cost, got %f", metrics.EstimatedCost)
	}
}

func TestCustomEngineGetLogParserScript(t *testing.T) {
	engine := NewCustomEngine()

	script := engine.GetLogParserScript()
	if script != "parse_custom_log" {
		t.Errorf("Expected log parser script 'parse_custom_log', got '%s'", script)
	}
}