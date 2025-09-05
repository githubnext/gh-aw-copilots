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

	if !engine.SupportsMaxTurns() {
		t.Error("Expected custom engine to support max turns for consistency with other engines")
	}
}

func TestCustomEngineGetInstallationSteps(t *testing.T) {
	engine := NewCustomEngine()

	steps := engine.GetInstallationSteps(&WorkflowData{})
	if len(steps) != 0 {
		t.Errorf("Expected 0 installation steps for custom engine, got %d", len(steps))
	}
}

func TestCustomEngineGetExecutionSteps(t *testing.T) {
	engine := NewCustomEngine()

	workflowData := &WorkflowData{
		Name: "test-workflow",
	}
	steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")

	// Custom engine without steps should return just the log step
	if len(steps) != 1 {
		t.Errorf("Expected 1 step (log step) when no engine config provided, got %d", len(steps))
	}
}

func TestCustomEngineStepIDHandling(t *testing.T) {
	engine := NewCustomEngine()

	// Create engine config with steps that have IDs
	engineConfig := &EngineConfig{
		ID: "custom",
		Steps: []map[string]any{
			{
				"name": "Step with ID",
				"id":   "step_with_id",
				"run":  "echo 'test'",
			},
			{
				"name": "Step without ID",
				"run":  "echo 'no id'",
			},
		},
	}

	workflowData := &WorkflowData{
		Name:         "test-workflow",
		EngineConfig: engineConfig,
	}

	config := engine.GetExecutionSteps(workflowData, "/tmp/test.log")

	// Should have 2 custom steps + 1 log step
	if len(config) != 3 {
		t.Errorf("Expected 3 steps (2 custom + 1 log), got %d", len(config))
	}

	// Check the first step has the ID
	if len(config) > 0 {
		firstStepContent := strings.Join([]string(config[0]), "\n")
		if !strings.Contains(firstStepContent, "id: step_with_id") {
			t.Errorf("Expected first step to contain 'id: step_with_id', got:\n%s", firstStepContent)
		}
	}

	// Check the second step doesn't have an ID line
	if len(config) > 1 {
		secondStepContent := strings.Join([]string(config[1]), "\n")
		if strings.Contains(secondStepContent, "id:") {
			t.Errorf("Expected second step to NOT contain 'id:' since none was provided, got:\n%s", secondStepContent)
		}
	}
}

func TestCustomEngineGetExecutionStepsWithSteps(t *testing.T) {
	engine := NewCustomEngine()

	// Create engine config with steps
	engineConfig := &EngineConfig{
		ID: "custom",
		Steps: []map[string]any{
			{
				"name": "Setup Node.js",
				"uses": "actions/setup-node@v4",
				"with": map[string]any{
					"node-version": "18",
				},
			},
			{
				"name": "Run tests",
				"run":  "npm test",
			},
		},
	}

	workflowData := &WorkflowData{
		Name:         "test-workflow",
		EngineConfig: engineConfig,
	}

	config := engine.GetExecutionSteps(workflowData, "/tmp/test.log")

	// Test with engine config - steps should be populated (2 custom steps + 1 log step)
	if len(config) != 3 {
		t.Errorf("Expected 3 steps when engine config has 2 steps (2 custom + 1 log), got %d", len(config))
	}

	// Check the first step content
	if len(config) > 0 {
		firstStepContent := strings.Join([]string(config[0]), "\n")
		if !strings.Contains(firstStepContent, "name: Setup Node.js") {
			t.Errorf("Expected first step to contain 'name: Setup Node.js', got:\n%s", firstStepContent)
		}
		if !strings.Contains(firstStepContent, "uses: actions/setup-node@v4") {
			t.Errorf("Expected first step to contain 'uses: actions/setup-node@v4', got:\n%s", firstStepContent)
		}
	}

	// Check the second step content
	if len(config) > 1 {
		secondStepContent := strings.Join([]string(config[1]), "\n")
		if !strings.Contains(secondStepContent, "name: Run tests") {
			t.Errorf("Expected second step to contain 'name: Run tests', got:\n%s", secondStepContent)
		}
		if !strings.Contains(secondStepContent, "run:") && !strings.Contains(secondStepContent, "npm test") {
			t.Errorf("Expected second step to contain run command 'npm test', got:\n%s", secondStepContent)
		}
	}
}

func TestCustomEngineRenderMCPConfig(t *testing.T) {
	engine := NewCustomEngine()
	var yaml strings.Builder

	// This should generate MCP configuration structure like Claude
	engine.RenderMCPConfig(&yaml, map[string]any{}, []string{})

	output := yaml.String()
	expectedPrefix := "          cat > /tmp/mcp-config/mcp-servers.json << 'EOF'"
	if !strings.Contains(output, expectedPrefix) {
		t.Errorf("Expected MCP config to contain setup prefix, got '%s'", output)
	}

	if !strings.Contains(output, "\"mcpServers\"") {
		t.Errorf("Expected MCP config to contain mcpServers section, got '%s'", output)
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

func TestCustomEnginePromptFileHandling(t *testing.T) {
	engine := NewCustomEngine()

	// Create workflow data with an actions/ai-inference step that has a prompt parameter
	workflowData := &WorkflowData{
		EngineConfig: &EngineConfig{
			ID: "custom",
			Steps: []map[string]any{
				{
					"name": "Analyze Issue with AI Inference",
					"id":   "analyze_issue",
					"uses": "actions/ai-inference@v1",
					"with": map[string]any{
						"model":       "gpt-4",
						"prompt":      "Analyze this GitHub issue and suggest appropriate labels based on the content.\n\nIssue Title: ${{ github.event.issue.title }}\nIssue Body: ${{ github.event.issue.body }}",
						"temperature": 0.1,
					},
				},
				{
					"name": "Process Results",
					"run":  "echo 'Processing AI results'",
				},
			},
		},
	}

	steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")

	// Should have at least 4 steps: create prompt files, analyze issue, process results, ensure log file
	if len(steps) < 4 {
		t.Errorf("Expected at least 4 steps, got %d", len(steps))
	}

	// Check the first step creates prompt files
	firstStepContent := strings.Join([]string(steps[0]), "\n")
	if !strings.Contains(firstStepContent, "Create AI inference prompt files") {
		t.Errorf("Expected first step to create prompt files, got:\n%s", firstStepContent)
	}
	if !strings.Contains(firstStepContent, "mkdir -p /tmp/aw-prompts") {
		t.Errorf("Expected first step to create prompt directory, got:\n%s", firstStepContent)
	}
	if !strings.Contains(firstStepContent, "/tmp/aw-prompts/analyze_issue_with_ai_inference.txt") {
		t.Errorf("Expected first step to create prompt file, got:\n%s", firstStepContent)
	}
	if !strings.Contains(firstStepContent, "Analyze this GitHub issue and suggest appropriate labels") {
		t.Errorf("Expected first step to contain prompt content, got:\n%s", firstStepContent)
	}

	// Check the second step (the AI inference step) uses prompt-file instead of prompt
	secondStepContent := strings.Join([]string(steps[1]), "\n")
	if !strings.Contains(secondStepContent, "uses: actions/ai-inference@v1") {
		t.Errorf("Expected second step to use actions/ai-inference, got:\n%s", secondStepContent)
	}
	if !strings.Contains(secondStepContent, "prompt-file: /tmp/aw-prompts/analyze_issue_with_ai_inference.txt") {
		t.Errorf("Expected second step to use prompt-file parameter, got:\n%s", secondStepContent)
	}
	if strings.Contains(secondStepContent, "prompt: |") {
		t.Errorf("Expected second step to NOT contain inline prompt parameter, got:\n%s", secondStepContent)
	}
	if !strings.Contains(secondStepContent, "GITHUB_AW_PROMPT: /tmp/aw-prompts/analyze_issue_with_ai_inference.txt") {
		t.Errorf("Expected second step to have GITHUB_AW_PROMPT environment variable, got:\n%s", secondStepContent)
	}

	// Check that other parameters are preserved
	if !strings.Contains(secondStepContent, "model: gpt-4") {
		t.Errorf("Expected second step to preserve model parameter, got:\n%s", secondStepContent)
	}
	if !strings.Contains(secondStepContent, "temperature: 0.1") {
		t.Errorf("Expected second step to preserve temperature parameter, got:\n%s", secondStepContent)
	}

	// Check that the step ID is preserved
	if !strings.Contains(secondStepContent, "id: analyze_issue") {
		t.Errorf("Expected second step to preserve step ID, got:\n%s", secondStepContent)
	}
}

func TestCustomEngineExtractPromptForFile(t *testing.T) {
	engine := NewCustomEngine()

	// Test with actions/ai-inference step that has prompt
	step1 := map[string]any{
		"name": "AI Step",
		"id":   "ai_step",
		"uses": "actions/ai-inference@v1",
		"with": map[string]any{
			"prompt": "Test prompt content",
		},
	}
	promptFile := engine.extractPromptForFile(step1)
	expected := "/tmp/aw-prompts/ai_step.txt"
	if promptFile != expected {
		t.Errorf("Expected prompt file '%s', got '%s'", expected, promptFile)
	}

	// Test with step that doesn't use actions/ai-inference
	step2 := map[string]any{
		"name": "Regular Step",
		"run":  "echo 'hello'",
	}
	promptFile2 := engine.extractPromptForFile(step2)
	if promptFile2 != "" {
		t.Errorf("Expected no prompt file for regular step, got '%s'", promptFile2)
	}

	// Test with actions/ai-inference step that doesn't have prompt
	step3 := map[string]any{
		"uses": "actions/ai-inference@v1",
		"with": map[string]any{
			"model": "gpt-4",
		},
	}
	promptFile3 := engine.extractPromptForFile(step3)
	if promptFile3 != "" {
		t.Errorf("Expected no prompt file for step without prompt, got '%s'", promptFile3)
	}
}

func TestCustomEngineModifyStepForPromptFile(t *testing.T) {
	engine := NewCustomEngine()

	originalStep := map[string]any{
		"name": "AI Step",
		"id":   "ai_step",
		"uses": "actions/ai-inference@v1",
		"with": map[string]any{
			"model":       "gpt-4",
			"prompt":      "Original prompt content",
			"temperature": 0.1,
		},
	}

	promptFileName := "/tmp/aw-prompts/test.txt"
	modifiedStep := engine.modifyStepForPromptFile(originalStep, promptFileName)

	// Check that original step is not modified
	originalWith := originalStep["with"].(map[string]any)
	if _, hasPrompt := originalWith["prompt"]; !hasPrompt {
		t.Error("Original step should still have prompt parameter")
	}

	// Check that modified step has prompt-file instead of prompt
	modifiedWith := modifiedStep["with"].(map[string]any)
	if _, hasPrompt := modifiedWith["prompt"]; hasPrompt {
		t.Error("Modified step should not have prompt parameter")
	}
	if promptFile, hasPromptFile := modifiedWith["prompt-file"]; !hasPromptFile || promptFile != promptFileName {
		t.Errorf("Modified step should have prompt-file parameter set to '%s', got '%v'", promptFileName, promptFile)
	}

	// Check that other parameters are preserved
	if model, hasModel := modifiedWith["model"]; !hasModel || model != "gpt-4" {
		t.Error("Modified step should preserve model parameter")
	}
	if temp, hasTemp := modifiedWith["temperature"]; !hasTemp || temp != 0.1 {
		t.Error("Modified step should preserve temperature parameter")
	}
}
