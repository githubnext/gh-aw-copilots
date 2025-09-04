package workflow

import (
	"testing"
)

func TestGenAIScriptEngine(t *testing.T) {
	engine := NewGenAIScriptEngine()

	if engine.GetID() != "genaiscript" {
		t.Errorf("Expected engine ID 'genaiscript', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "GenAIScript" {
		t.Errorf("Expected display name 'GenAIScript', got '%s'", engine.GetDisplayName())
	}

	if !engine.IsExperimental() {
		t.Error("Expected GenAIScript engine to be experimental")
	}

	if !engine.SupportsToolsWhitelist() {
		t.Error("Expected GenAIScript engine to support tools whitelist")
	}

	if !engine.SupportsHTTPTransport() {
		t.Error("Expected GenAIScript engine to support HTTP transport")
	}

	if !engine.SupportsMaxTurns() {
		t.Error("Expected GenAIScript engine to support max-turns")
	}
}

func TestGenAIScriptEngineInstallationSteps(t *testing.T) {
	engine := NewGenAIScriptEngine()

	// Test without version
	steps := engine.GetInstallationSteps(nil, nil)
	if len(steps) != 2 {
		t.Errorf("Expected 2 installation steps, got %d", len(steps))
	}

	// Check Node.js setup step
	nodeStep := steps[0]
	if len(nodeStep) == 0 || nodeStep[0] != "      - name: Setup Node.js" {
		t.Error("Expected first step to be Node.js setup")
	}

	// Check GenAIScript installation step
	installStep := steps[1]
	if len(installStep) == 0 || installStep[0] != "      - name: Install GenAIScript" {
		t.Error("Expected second step to be GenAIScript installation")
	}
}

func TestGenAIScriptEngineWithVersion(t *testing.T) {
	engine := NewGenAIScriptEngine()

	engineConfig := &EngineConfig{
		Version: "2.4.0",
	}

	steps := engine.GetInstallationSteps(engineConfig, nil)
	if len(steps) != 2 {
		t.Errorf("Expected 2 installation steps, got %d", len(steps))
	}

	// Check that version is included in install command
	installStep := steps[1]
	found := false
	for _, line := range installStep {
		if line == "        run: npm install -g genaiscript@2.4.0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected install command to include version 2.4.0")
	}
}

func TestGenAIScriptEngineExecutionConfig(t *testing.T) {
	engine := NewGenAIScriptEngine()

	config := engine.GetExecutionConfig("test-workflow", "/tmp/test.log", nil, nil, false)

	if config.StepName != "Run GenAIScript" {
		t.Errorf("Expected step name 'Run GenAIScript', got '%s'", config.StepName)
	}

	if config.Command == "" {
		t.Error("Expected execution command to be set")
	}

	// Check that required environment variables are configured
	if config.Environment["GITHUB_STEP_SUMMARY"] != "${{ env.GITHUB_STEP_SUMMARY }}" {
		t.Error("Expected GITHUB_STEP_SUMMARY to be configured")
	}

	// API keys are no longer configured by default in GenAIScript engine
	if _, exists := config.Environment["OPENAI_API_KEY"]; exists {
		t.Error("Expected OPENAI_API_KEY to not be configured by default")
	}
}

func TestGenAIScriptEngineWithModel(t *testing.T) {
	engine := NewGenAIScriptEngine()

	engineConfig := &EngineConfig{
		Model: "gpt-4",
	}

	config := engine.GetExecutionConfig("test-workflow", "/tmp/test.log", engineConfig, nil, false)

	// Check that the command includes the specified model
	if !containsStringInCommand(config.Command, "gpt-4") {
		t.Error("Expected execution command to include model 'gpt-4'")
	}
}

func TestGenAIScriptEngineWithOutput(t *testing.T) {
	engine := NewGenAIScriptEngine()

	config := engine.GetExecutionConfig("test-workflow", "/tmp/test.log", nil, nil, true)

	// Check that GITHUB_AW_SAFE_OUTPUTS is configured when hasOutput is true
	if config.Environment["GITHUB_AW_SAFE_OUTPUTS"] != "${{ env.GITHUB_AW_SAFE_OUTPUTS }}" {
		t.Error("Expected GITHUB_AW_SAFE_OUTPUTS to be configured when output is needed")
	}
}

func TestGenAIScriptEngineLogParsing(t *testing.T) {
	engine := NewGenAIScriptEngine()

	// Test empty log
	metrics := engine.ParseLogMetrics("", false)
	if metrics.TokenUsage != 0 || metrics.ErrorCount != 0 || metrics.WarningCount != 0 {
		t.Error("Expected empty metrics for empty log")
	}

	// Test log with token usage
	logWithTokens := `Running GenAIScript...
total_tokens: 1500
Completed successfully`

	metrics = engine.ParseLogMetrics(logWithTokens, false)
	if metrics.TokenUsage != 1500 {
		t.Errorf("Expected token usage 1500, got %d", metrics.TokenUsage)
	}

	// Test log with errors and warnings
	logWithErrorsWarnings := `Warning: deprecated feature used
Error: API rate limit exceeded
tokens: 800`

	metrics = engine.ParseLogMetrics(logWithErrorsWarnings, false)
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.ErrorCount)
	}
	if metrics.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", metrics.WarningCount)
	}
	if metrics.TokenUsage != 800 {
		t.Errorf("Expected token usage 800, got %d", metrics.TokenUsage)
	}
}

func TestGenAIScriptEngineLogParserScript(t *testing.T) {
	engine := NewGenAIScriptEngine()

	scriptName := engine.GetLogParserScript()
	if scriptName != "parse_genaiscript_log" {
		t.Errorf("Expected log parser script 'parse_genaiscript_log', got '%s'", scriptName)
	}
}

func TestGenAIScriptEngineDeclaredOutputFiles(t *testing.T) {
	engine := NewGenAIScriptEngine()

	outputFiles := engine.GetDeclaredOutputFiles()
	if len(outputFiles) != 2 {
		t.Errorf("Expected 2 declared output files, got %d", len(outputFiles))
	}

	if outputFiles[0] != "output.txt" {
		t.Errorf("Expected first output file 'output.txt', got '%s'", outputFiles[0])
	}

	if outputFiles[1] != "*.genai.md" {
		t.Errorf("Expected second output file '*.genai.md', got '%s'", outputFiles[1])
	}
}

// Helper function to check if a string contains a substring
func containsStringInCommand(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			indexOfStringInCommand(s, substr) >= 0)
}

func indexOfStringInCommand(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
