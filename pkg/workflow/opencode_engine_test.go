package workflow

import (
	"strings"
	"testing"
)

func TestOpenCodeEngine(t *testing.T) {
	engine := NewOpenCodeEngine()

	// Test basic properties
	if engine.GetID() != "opencode" {
		t.Errorf("Expected ID 'opencode', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "OpenCode" {
		t.Errorf("Expected display name 'OpenCode', got '%s'", engine.GetDisplayName())
	}

	if !engine.IsExperimental() {
		t.Error("OpenCode engine should be experimental")
	}

	if !engine.SupportsToolsWhitelist() {
		t.Error("OpenCode engine should support MCP tools")
	}

	// Test installation steps
	steps := engine.GetInstallationSteps(nil)
	expectedStepCount := 2 // Setup Node.js and Install OpenCode
	if len(steps) != expectedStepCount {
		t.Errorf("Expected %d installation steps, got %d", expectedStepCount, len(steps))
	}

	// Verify first step is Setup Node.js
	if len(steps) > 0 && len(steps[0]) > 0 {
		if !strings.Contains(steps[0][0], "Setup Node.js") {
			t.Errorf("Expected first step to contain 'Setup Node.js', got '%s'", steps[0][0])
		}
	}

	// Verify second step is Install OpenCode
	if len(steps) > 1 && len(steps[1]) > 0 {
		if !strings.Contains(steps[1][0], "Install OpenCode") {
			t.Errorf("Expected second step to contain 'Install OpenCode', got '%s'", steps[1][0])
		}
	}

	// Test execution config
	config := engine.GetExecutionConfig("test-workflow", "test-log", nil)
	if config.StepName != "Run OpenCode" {
		t.Errorf("Expected step name 'Run OpenCode', got '%s'", config.StepName)
	}

	if config.Action != "" {
		t.Errorf("Expected empty action for OpenCode (uses command), got '%s'", config.Action)
	}

	if !strings.Contains(config.Command, "opencode exec") {
		t.Errorf("Expected command to contain 'opencode exec', got '%s'", config.Command)
	}

	if !strings.Contains(config.Command, "test-log.log") {
		t.Errorf("Expected command to contain log file name, got '%s'", config.Command)
	}

	// Check environment variables
	if config.Environment["OPENCODE_API_KEY"] != "${{ secrets.OPENCODE_API_KEY }}" {
		t.Errorf("Expected OPENCODE_API_KEY environment variable, got '%s'", config.Environment["OPENCODE_API_KEY"])
	}
}

func TestOpenCodeMCPConfigGeneration(t *testing.T) {
	engine := NewOpenCodeEngine()

	// Test MCP config generation with GitHub tool
	tools := map[string]any{
		"github": map[string]any{
			"allowed": []string{"get_issue", "add_issue_comment"},
		},
	}
	mcpTools := []string{"github"}

	var yaml strings.Builder
	engine.RenderMCPConfig(&yaml, tools, mcpTools)

	config := yaml.String()

	// Check that opencode.json is generated (not mcp-servers.json or config.toml)
	if !strings.Contains(config, "cat > /tmp/mcp-config/opencode.json") {
		t.Errorf("Expected config to contain opencode.json generation for OpenCode but it didn't.\nContent:\n%s", config)
	}

	// Check for GitHub MCP configuration
	if !strings.Contains(config, "\"github\": {") {
		t.Errorf("Expected config to contain GitHub MCP server configuration but it didn't.\nContent:\n%s", config)
	}

	// Check for Docker-based GitHub MCP server (following pattern)
	if !strings.Contains(config, "\"command\": \"docker\"") {
		t.Errorf("Expected config to contain Docker command for GitHub MCP server but it didn't.\nContent:\n%s", config)
	}

	// Check JSON structure
	if !strings.Contains(config, "\"mcpServers\": {") {
		t.Errorf("Expected config to contain 'mcpServers' section but it didn't.\nContent:\n%s", config)
	}
}

func TestOpenCodeCustomMCPConfig(t *testing.T) {
	engine := NewOpenCodeEngine()

	// Test custom MCP tool configuration
	tools := map[string]any{
		"custom-tool": map[string]any{
			"mcp": map[string]any{
				"type":    "stdio",
				"command": "node",
				"args":    []any{"custom-server.js"},
				"env": map[string]any{
					"API_KEY": "test-key",
				},
			},
			"allowed": []string{"custom_function"},
		},
	}
	mcpTools := []string{"custom-tool"}

	var yaml strings.Builder
	engine.RenderMCPConfig(&yaml, tools, mcpTools)

	config := yaml.String()

	// Check that custom tool is configured
	if !strings.Contains(config, "\"custom-tool\": {") {
		t.Errorf("Expected config to contain custom-tool configuration but it didn't.\nContent:\n%s", config)
	}

	// Check command configuration
	if !strings.Contains(config, "\"command\": \"node\"") {
		t.Errorf("Expected config to contain node command but it didn't.\nContent:\n%s", config)
	}

	// Check args configuration
	if !strings.Contains(config, "\"custom-server.js\"") {
		t.Errorf("Expected config to contain custom-server.js in args but it didn't.\nContent:\n%s", config)
	}

	// Check env configuration
	if !strings.Contains(config, "\"API_KEY\": \"test-key\"") {
		t.Errorf("Expected config to contain API_KEY environment variable but it didn't.\nContent:\n%s", config)
	}
}

func TestOpenCodeHTTPMCPConfig(t *testing.T) {
	engine := NewOpenCodeEngine()

	// Test HTTP MCP tool configuration
	tools := map[string]any{
		"http-tool": map[string]any{
			"mcp": map[string]any{
				"type": "http",
				"url":  "http://localhost:3000",
				"headers": map[string]any{
					"Authorization": "Bearer token",
				},
			},
			"allowed": []string{"http_function"},
		},
	}
	mcpTools := []string{"http-tool"}

	var yaml strings.Builder
	engine.RenderMCPConfig(&yaml, tools, mcpTools)

	config := yaml.String()

	// Check that HTTP tool is configured
	if !strings.Contains(config, "\"http-tool\": {") {
		t.Errorf("Expected config to contain http-tool configuration but it didn't.\nContent:\n%s", config)
	}

	// Check URL configuration
	if !strings.Contains(config, "\"url\": \"http://localhost:3000\"") {
		t.Errorf("Expected config to contain URL but it didn't.\nContent:\n%s", config)
	}

	// Check headers configuration
	if !strings.Contains(config, "\"Authorization\": \"Bearer token\"") {
		t.Errorf("Expected config to contain Authorization header but it didn't.\nContent:\n%s", config)
	}
}

func TestOpenCodeEngineWithVersion(t *testing.T) {
	engine := NewOpenCodeEngine()

	// Test installation steps without version
	stepsNoVersion := engine.GetInstallationSteps(nil)
	foundNoVersionInstall := false
	for _, step := range stepsNoVersion {
		for _, line := range step {
			if strings.Contains(line, "npm install -g opencode") && !strings.Contains(line, "@") {
				foundNoVersionInstall = true
				break
			}
		}
	}
	if !foundNoVersionInstall {
		t.Error("Expected default npm install command without version")
	}

	// Test installation steps with version
	engineConfig := &EngineConfig{
		ID:      "opencode",
		Version: "2.1.0",
	}
	stepsWithVersion := engine.GetInstallationSteps(engineConfig)
	foundVersionInstall := false
	for _, step := range stepsWithVersion {
		for _, line := range step {
			if strings.Contains(line, "npm install -g opencode@2.1.0") {
				foundVersionInstall = true
				break
			}
		}
	}
	if !foundVersionInstall {
		t.Error("Expected versioned npm install command with opencode@2.1.0")
	}
}
