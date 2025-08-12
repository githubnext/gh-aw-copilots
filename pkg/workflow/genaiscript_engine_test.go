package workflow

import (
	"os"
	"strings"
	"testing"
)

func TestGenAIScriptEngine(t *testing.T) {
	engine := NewGenAIScriptEngine()

	// Test basic properties
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

	// Test installation steps
	steps := engine.GetInstallationSteps(nil)
	if len(steps) != 2 {
		t.Errorf("Expected 2 installation steps, got %d", len(steps))
	}

	// Verify Node.js setup step
	nodeSetupFound := false
	genaiscriptInstallFound := false
	for _, step := range steps {
		stepContent := strings.Join(step, "\n")
		if strings.Contains(stepContent, "Setup Node.js") && strings.Contains(stepContent, "actions/setup-node@v4") {
			nodeSetupFound = true
		}
		if strings.Contains(stepContent, "Install GenAIScript") && strings.Contains(stepContent, "npm install -g genaiscript") {
			genaiscriptInstallFound = true
		}
	}

	if !nodeSetupFound {
		t.Error("Expected Node.js setup step")
	}
	if !genaiscriptInstallFound {
		t.Error("Expected GenAIScript installation step")
	}

	// Test execution config
	config := engine.GetExecutionConfig("test-workflow", "test.log", nil)
	if config.StepName != "Run GenAIScript" {
		t.Errorf("Expected step name 'Run GenAIScript', got '%s'", config.StepName)
	}

	if config.Action != "" {
		t.Error("Expected empty action for CLI-based engine")
	}

	if !strings.Contains(config.Command, "genaiscript run") {
		t.Error("Expected command to contain 'genaiscript run'")
	}

	if !strings.Contains(config.Command, "/tmp/aw-prompts/prompt.txt") {
		t.Error("Expected command to use prompt file")
	}

	if !strings.Contains(config.Command, "--mcps /tmp/mcp-config/mcp-servers.json") {
		t.Error("Expected command to use MCP config")
	}

	if !strings.Contains(config.Command, "--out-output $GITHUB_STEP_SUMMARY") {
		t.Error("Expected command to output to GITHUB_STEP_SUMMARY")
	}

	// Test with model configuration
	engineConfig := &EngineConfig{Model: "gpt-4"}
	configWithModel := engine.GetExecutionConfig("test-workflow", "test.log", engineConfig)
	if configWithModel.Environment["GENAISCRIPT_MODEL"] != "gpt-4" {
		t.Error("Expected GENAISCRIPT_MODEL environment variable to be set")
	}
}

func TestGenAIScriptEngineWithVersion(t *testing.T) {
	engine := NewGenAIScriptEngine()

	// Test installation steps without version
	stepsNoVersion := engine.GetInstallationSteps(nil)
	foundNoVersionInstall := false
	for _, step := range stepsNoVersion {
		for _, line := range step {
			if strings.Contains(line, "npm install -g genaiscript") && !strings.Contains(line, "@") {
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
		ID:      "genaiscript",
		Version: "1.2.3",
	}
	stepsWithVersion := engine.GetInstallationSteps(engineConfig)
	foundVersionInstall := false
	for _, step := range stepsWithVersion {
		for _, line := range step {
			if strings.Contains(line, "npm install -g genaiscript@1.2.3") {
				foundVersionInstall = true
				break
			}
		}
	}
	if !foundVersionInstall {
		t.Error("Expected versioned npm install command with genaiscript@1.2.3")
	}
}

func TestGenAIScriptMCPConfigGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "genaiscript-mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	tests := []struct {
		name                 string
		frontmatter          string
		expectedAI           string
		expectMcpServersJson bool
	}{
		{
			name: "genaiscript with github tools generates mcp-servers.json",
			frontmatter: `---
engine: genaiscript
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "genaiscript",
			expectMcpServersJson: true,
		},
		{
			name: "genaiscript with custom MCP tool",
			frontmatter: `---
engine: genaiscript
tools:
  github:
    allowed: [get_issue]
  custom-tool:
    allowed: [custom_function]
    mcp:
      type: stdio
      command: "node"
      args: ["custom-server.js"]
---`,
			expectedAI:           "genaiscript",
			expectMcpServersJson: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test markdown file
			mdPath := tmpDir + "/test-workflow.md"
			mdContent := tt.frontmatter + "\n\n# Test Workflow\n\nTest content"

			if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile workflow
			if err := compiler.CompileWorkflow(mdPath); err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Read generated lock file
			lockPath := strings.TrimSuffix(mdPath, ".md") + ".lock.yml"
			lockContent, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("Failed to read lock file: %v", err)
			}
			lockContentStr := string(lockContent)

			// Test MCP servers configuration
			if tt.expectMcpServersJson {
				if !strings.Contains(lockContentStr, "mcp-servers.json") {
					t.Errorf("Expected mcp-servers.json generation but didn't find it in:\n%s", lockContentStr)
				}
				if !strings.Contains(lockContentStr, "mcpServers") {
					t.Errorf("Expected mcpServers section but didn't find it in:\n%s", lockContentStr)
				}
			}

			// Verify AI type
			if tt.expectedAI == "genaiscript" {
				if !strings.Contains(lockContentStr, "genaiscript run") {
					t.Errorf("Expected genaiscript run command but didn't find it in:\n%s", lockContentStr)
				}
				if !strings.Contains(lockContentStr, "npm install -g genaiscript") {
					t.Errorf("Expected GenAIScript installation but didn't find it in:\n%s", lockContentStr)
				}
			}
		})
	}
}

func TestGenAIScriptCustomMCPConfig(t *testing.T) {
	engine := NewGenAIScriptEngine()
	var yaml strings.Builder

	// Test with custom MCP tool
	tools := map[string]any{
		"github": map[string]any{
			"allowed": []string{"get_issue"},
		},
		"custom-tool": map[string]any{
			"allowed": []string{"custom_function"},
			"mcp": map[string]any{
				"type":    "stdio",
				"command": "node",
				"args":    []string{"custom-server.js"},
			},
		},
	}
	mcpTools := []string{"github", "custom-tool"}

	engine.RenderMCPConfig(&yaml, tools, mcpTools)
	result := yaml.String()

	// Check that MCP configuration is generated
	if !strings.Contains(result, "mcp-servers.json") {
		t.Error("Expected mcp-servers.json configuration")
	}

	if !strings.Contains(result, "mcpServers") {
		t.Error("Expected mcpServers section")
	}

	if !strings.Contains(result, "github") {
		t.Error("Expected github MCP server configuration")
	}

	if !strings.Contains(result, "custom-tool") {
		t.Error("Expected custom-tool MCP server configuration")
	}
}

func TestGenAIScriptHTTPMCPConfig(t *testing.T) {
	engine := NewGenAIScriptEngine()
	var yaml strings.Builder

	// Test with HTTP-based MCP tool
	tools := map[string]any{
		"http-tool": map[string]any{
			"allowed": []string{"http_function"},
			"mcp": map[string]any{
				"type": "http",
				"url":  "http://localhost:3000/mcp",
			},
		},
	}
	mcpTools := []string{"http-tool"}

	engine.RenderMCPConfig(&yaml, tools, mcpTools)
	result := yaml.String()

	// Check that HTTP MCP configuration is generated
	if !strings.Contains(result, "http-tool") {
		t.Error("Expected http-tool MCP server configuration")
	}

	if !strings.Contains(result, "http://localhost:3000/mcp") {
		t.Error("Expected HTTP URL in configuration")
	}
}
