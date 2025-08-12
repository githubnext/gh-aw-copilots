package workflow

import (
	"fmt"
	"strings"
)

// AIInferenceEngine represents the AI Inference agentic engine using GitHub Models
type AIInferenceEngine struct {
	BaseEngine
}

func NewAIInferenceEngine() *AIInferenceEngine {
	return &AIInferenceEngine{
		BaseEngine: BaseEngine{
			id:                     "ai-inference",
			displayName:            "AI Inference",
			description:            "Uses GitHub Models via actions/ai-inference with GitHub MCP support",
			experimental:           false,
			supportsToolsWhitelist: true,
		},
	}
}

func (e *AIInferenceEngine) GetInstallationSteps(engineConfig *EngineConfig) []GitHubActionStep {
	// ai-inference doesn't require installation as it's a GitHub Action
	return []GitHubActionStep{}
}

func (e *AIInferenceEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig) ExecutionConfig {
	config := ExecutionConfig{
		StepName: "Execute AI Inference Action",
		Action:   "actions/ai-inference@v1",
		Inputs: map[string]string{
			"prompt-file": "/tmp/aw-prompts/prompt.txt",
			"token":       "${{ secrets.GITHUB_TOKEN }}",
			"mcp-config":  "/tmp/mcp-config/mcp-servers.json",
			"max-tokens":  "2000", // Increased default for workflow responses
		},
		Environment: map[string]string{
			"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
		},
	}

	// Add model configuration if specified
	if engineConfig != nil && engineConfig.Model != "" {
		config.Inputs["model"] = engineConfig.Model
	} else {
		// Use default model from ai-inference action
		config.Inputs["model"] = "openai/gpt-4o"
	}

	return config
}

func (e *AIInferenceEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	yaml.WriteString("          cat > /tmp/mcp-config/mcp-servers.json << 'EOF'\n")
	yaml.WriteString("          {\n")
	yaml.WriteString("            \"mcpServers\": {\n")

	// Generate configuration for each MCP tool
	for i, toolName := range mcpTools {
		isLast := i == len(mcpTools)-1

		switch toolName {
		case "github":
			githubTool := tools["github"]
			e.renderGitHubAIInferenceMCPConfig(yaml, githubTool, isLast)
		default:
			// Handle custom MCP tools (those with MCP-compatible type)
			if toolConfig, ok := tools[toolName].(map[string]any); ok {
				if hasMcp, _ := hasMCPConfig(toolConfig); hasMcp {
					if err := e.renderAIInferenceMCPConfig(yaml, toolName, toolConfig, isLast); err != nil {
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

// renderGitHubAIInferenceMCPConfig generates the GitHub MCP server configuration
// Uses Docker MCP as the default for AI Inference
func (e *AIInferenceEngine) renderGitHubAIInferenceMCPConfig(yaml *strings.Builder, githubTool any, isLast bool) {
	githubDockerImageVersion := getGitHubDockerImageVersion(githubTool)

	yaml.WriteString("              \"github\": {\n")

	// Always use Docker-based GitHub MCP server (services mode has been removed)
	yaml.WriteString("                \"command\": \"docker\",\n")
	yaml.WriteString("                \"args\": [\n")
	yaml.WriteString("                  \"run\",\n")
	yaml.WriteString("                  \"-i\",\n")
	yaml.WriteString("                  \"--rm\",\n")
	yaml.WriteString("                  \"-e\",\n")
	yaml.WriteString("                  \"GITHUB_PERSONAL_ACCESS_TOKEN\",\n")
	yaml.WriteString("                  \"ghcr.io/github/github-mcp-server:" + githubDockerImageVersion + "\"\n")
	yaml.WriteString("                ],\n")
	yaml.WriteString("                \"env\": {\n")
	yaml.WriteString("                  \"GITHUB_PERSONAL_ACCESS_TOKEN\": \"${{ secrets.GITHUB_TOKEN }}\"\n")
	yaml.WriteString("                }\n")

	if isLast {
		yaml.WriteString("              }\n")
	} else {
		yaml.WriteString("              },\n")
	}
}

// renderAIInferenceMCPConfig generates custom MCP server configuration for a single tool in AI Inference workflow mcp-servers.json
func (e *AIInferenceEngine) renderAIInferenceMCPConfig(yaml *strings.Builder, toolName string, toolConfig map[string]any, isLast bool) error {
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
