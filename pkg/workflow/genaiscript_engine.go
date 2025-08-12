package workflow

import (
	"fmt"
	"strings"
)

// GenAIScriptEngine represents the GenAIScript agentic engine (experimental)
type GenAIScriptEngine struct {
	BaseEngine
}

func NewGenAIScriptEngine() *GenAIScriptEngine {
	return &GenAIScriptEngine{
		BaseEngine: BaseEngine{
			id:                     "genaiscript",
			displayName:            "GenAIScript",
			description:            "Uses GenAIScript with markdown scripts and MCP support (experimental)",
			experimental:           true,
			supportsToolsWhitelist: true,
		},
	}
}

func (e *GenAIScriptEngine) GetInstallationSteps(engineConfig *EngineConfig) []GitHubActionStep {
	// Build the npm install command, optionally with version
	installCmd := "npm install -g genaiscript"
	if engineConfig != nil && engineConfig.Version != "" {
		installCmd = fmt.Sprintf("npm install -g genaiscript@%s", engineConfig.Version)
	}

	return []GitHubActionStep{
		{
			"      - name: Setup Node.js",
			"        uses: actions/setup-node@v4",
			"        with:",
			"          node-version: '24'",
		},
		{
			"      - name: Install GenAIScript",
			fmt.Sprintf("        run: %s", installCmd),
		},
	}
}

func (e *GenAIScriptEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig) ExecutionConfig {
	// Build the genaiscript command
	// Based on comment: genaiscript run prompt.md --mcps ./mcpservers.json --out-output $GITHUB_STEP_SUMMARY
	command := fmt.Sprintf(`# Create log directory outside git repo
mkdir -p /tmp/aw-logs

# Run GenAIScript with MCP config and log capture
genaiscript run /tmp/aw-prompts/prompt.txt \
  --mcps /tmp/mcp-config/mcp-servers.json \
  --out-output $GITHUB_STEP_SUMMARY 2>&1 | tee /tmp/aw-logs/%s.log`, logFile)

	config := ExecutionConfig{
		StepName: "Run GenAIScript",
		Command:  command,
		Environment: map[string]string{
			"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
		},
	}

	// Add model configuration if specified
	if engineConfig != nil && engineConfig.Model != "" {
		// GenAIScript supports model specification via environment or CLI args
		config.Environment["GENAISCRIPT_MODEL"] = engineConfig.Model
	}

	return config
}

func (e *GenAIScriptEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	// GenAIScript uses Claude-compatible MCP configuration format
	// Generate mcp-servers.json in the same format as Claude engine

	yaml.WriteString("          # Create MCP configuration directory\n")
	yaml.WriteString("          mkdir -p /tmp/mcp-config\n")
	yaml.WriteString("          \n")
	yaml.WriteString("          # Generate MCP servers configuration for GenAIScript\n")
	yaml.WriteString("          cat > /tmp/mcp-config/mcp-servers.json << 'EOF'\n")
	yaml.WriteString("          {\n")
	yaml.WriteString("            \"mcpServers\": {\n")

	// Process tools and generate MCP configuration
	mcpServerCount := 0

	for i, toolName := range mcpTools {
		if toolConfig, ok := tools[toolName].(map[string]any); ok {
			if toolName == "github" {
				e.renderGitHubGenAIScriptMCPConfig(yaml, toolConfig, i == len(mcpTools)-1)
				mcpServerCount++
			} else {
				// Handle custom MCP tools
				if hasMcp, _ := hasMCPConfig(toolConfig); hasMcp {
					if err := e.renderGenAIScriptMCPConfig(yaml, toolName, toolConfig); err == nil {
						if i < len(mcpTools)-1 {
							yaml.WriteString(",\n")
						}
						mcpServerCount++
					}
				}
			}
		}
	}

	yaml.WriteString("\n            }\n")
	yaml.WriteString("          }\n")
	yaml.WriteString("          EOF\n")
	yaml.WriteString("          \n")
}

// renderGitHubGenAIScriptMCPConfig generates GitHub MCP server configuration for GenAIScript
// Uses the same format as Claude since GenAIScript supports Claude MCP config format
func (e *GenAIScriptEngine) renderGitHubGenAIScriptMCPConfig(yaml *strings.Builder, githubTool any, isLast bool) {
	yaml.WriteString("              \"github\": {\n")
	yaml.WriteString("                \"command\": \"docker\",\n")
	yaml.WriteString("                \"args\": [\n")
	yaml.WriteString("                  \"run\", \"--rm\",\n")
	yaml.WriteString("                  \"-e\", \"GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }}\",\n")
	yaml.WriteString("                  \"ghcr.io/modelcontextprotocol/servers/github:latest\"\n")
	yaml.WriteString("                ]\n")
	yaml.WriteString("              }")

	if !isLast {
		yaml.WriteString(",")
	}
	yaml.WriteString("\n")
}

// renderGenAIScriptMCPConfig generates custom MCP server configuration for a single tool in GenAIScript workflow
func (e *GenAIScriptEngine) renderGenAIScriptMCPConfig(yaml *strings.Builder, toolName string, toolConfig map[string]any) error {
	yaml.WriteString(fmt.Sprintf("              \"%s\": {\n", toolName))

	// Use the shared MCP config renderer with JSON format
	renderer := MCPConfigRenderer{
		IndentLevel: "                ",
		Format:      "json",
	}

	if err := renderSharedMCPConfig(yaml, toolName, toolConfig, true, renderer); err != nil {
		return err
	}

	yaml.WriteString("              }")
	return nil
}
