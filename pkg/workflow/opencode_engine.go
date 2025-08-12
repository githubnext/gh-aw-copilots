package workflow

import (
	"fmt"
	"strings"
)

// OpenCodeEngine represents the OpenCode agentic engine (experimental)
type OpenCodeEngine struct {
	BaseEngine
}

func NewOpenCodeEngine() *OpenCodeEngine {
	return &OpenCodeEngine{
		BaseEngine: BaseEngine{
			id:                     "opencode",
			displayName:            "OpenCode",
			description:            "Uses OpenCode AI coding assistant (experimental)",
			experimental:           true,
			supportsToolsWhitelist: true,
		},
	}
}

func (e *OpenCodeEngine) GetInstallationSteps(engineConfig *EngineConfig) []GitHubActionStep {
	// Build the npm install command, optionally with version
	installCmd := "npm install -g opencode"
	if engineConfig != nil && engineConfig.Version != "" {
		installCmd = fmt.Sprintf("npm install -g opencode@%s", engineConfig.Version)
	}

	return []GitHubActionStep{
		{
			"      - name: Setup Node.js",
			"        uses: actions/setup-node@v4",
			"        with:",
			"          node-version: '24'",
			"          cache: 'npm'",
		},
		{
			"      - name: Install OpenCode",
			fmt.Sprintf("        run: %s", installCmd),
		},
	}
}

func (e *OpenCodeEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig) ExecutionConfig {
	// Configure model and API settings based on engineConfig
	modelConfig := ""
	apiKeyEnv := "OPENCODE_API_KEY"
	apiKeySecret := "${{ secrets.OPENCODE_API_KEY }}"

	if engineConfig != nil && engineConfig.Model != "" {
		// If a specific model is configured, use it
		modelConfig = fmt.Sprintf("--model %s", engineConfig.Model)

		// For Claude models, use Anthropic API key
		if strings.HasPrefix(engineConfig.Model, "claude") || strings.HasPrefix(engineConfig.Model, "anthropic/") {
			apiKeyEnv = "ANTHROPIC_API_KEY"
			apiKeySecret = "${{ secrets.ANTHROPIC_API_KEY }}"
		}
	}

	command := fmt.Sprintf(`INSTRUCTION=$(cat /tmp/aw-prompts/prompt.txt)
export OPENCODE_CONFIG=/tmp/mcp-config

# Create log directory outside git repo
mkdir -p /tmp/aw-logs

# Run opencode with log capture
opencode exec \
  --config /tmp/mcp-config/opencode.json \
  %s \
  --auto "$INSTRUCTION" 2>&1 | tee /tmp/aw-logs/%s.log`, modelConfig, logFile)

	return ExecutionConfig{
		StepName: "Run OpenCode",
		Command:  command,
		Environment: map[string]string{
			apiKeyEnv: apiKeySecret,
		},
	}
}

func (e *OpenCodeEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	yaml.WriteString("          cat > /tmp/mcp-config/opencode.json << 'EOF'\n")
	yaml.WriteString("          {\n")
	yaml.WriteString("            \"mcpServers\": {\n")

	// Generate configuration for each MCP tool
	for i, toolName := range mcpTools {
		isLast := i == len(mcpTools)-1

		switch toolName {
		case "github":
			githubTool := tools["github"]
			e.renderGitHubOpenCodeMCPConfig(yaml, githubTool, isLast)
		default:
			// Handle custom MCP tools (those with MCP-compatible type)
			if toolConfig, ok := tools[toolName].(map[string]any); ok {
				if hasMcp, _ := hasMCPConfig(toolConfig); hasMcp {
					if err := e.renderOpenCodeMCPConfig(yaml, toolName, toolConfig, isLast); err != nil {
						fmt.Printf("Error generating custom MCP configuration for %s: %v\n", toolName, err)
					}
				}
			}
		}
	}

	yaml.WriteString("            }\n")
	yaml.WriteString("          }\n")
	yaml.WriteString("          EOF\n")
}

// renderGitHubOpenCodeMCPConfig generates the GitHub MCP server configuration
// Uses Docker MCP as the default for OpenCode
func (e *OpenCodeEngine) renderGitHubOpenCodeMCPConfig(yaml *strings.Builder, githubTool any, isLast bool) {
	yaml.WriteString("              \"github\": {\n")
	yaml.WriteString("                \"command\": \"docker\",\n")
	yaml.WriteString("                \"args\": [\n")
	yaml.WriteString("                  \"run\",\n")
	yaml.WriteString("                  \"--rm\",\n")
	yaml.WriteString("                  \"-e\", \"GITHUB_TOKEN\",\n")
	yaml.WriteString("                  \"ghcr.io/githubnext/github-mcp-server:latest\"\n")
	yaml.WriteString("                ],\n")
	yaml.WriteString("                \"env\": {\n")
	yaml.WriteString("                  \"GITHUB_TOKEN\": \"${{ secrets.GITHUB_TOKEN }}\"\n")
	yaml.WriteString("                }\n")

	if isLast {
		yaml.WriteString("              }\n")
	} else {
		yaml.WriteString("              },\n")
	}
}

// renderOpenCodeMCPConfig generates custom MCP server configuration for a single tool in OpenCode workflow opencode.json
func (e *OpenCodeEngine) renderOpenCodeMCPConfig(yaml *strings.Builder, toolName string, toolConfig map[string]any, isLast bool) error {
	yaml.WriteString(fmt.Sprintf("              \"%s\": {\n", toolName))

	// Use the shared MCP config renderer with JSON format
	renderer := MCPConfigRenderer{
		IndentLevel: "                ",
		Format:      "json",
	}

	err := renderSharedMCPConfig(yaml, toolName, toolConfig, isLast, renderer)
	if err != nil {
		return err
	}

	if isLast {
		yaml.WriteString("              }\n")
	} else {
		yaml.WriteString("              },\n")
	}

	return nil
}
