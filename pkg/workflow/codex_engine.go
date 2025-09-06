package workflow

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CodexEngine represents the Codex agentic engine (experimental)
type CodexEngine struct {
	BaseEngine
}

func NewCodexEngine() *CodexEngine {
	return &CodexEngine{
		BaseEngine: BaseEngine{
			id:                     "codex",
			displayName:            "Codex",
			description:            "Uses OpenAI Codex CLI with MCP server support",
			experimental:           true,
			supportsToolsWhitelist: true,
			supportsHTTPTransport:  false, // Codex only supports stdio transport
			supportsMaxTurns:       false, // Codex does not support max-turns feature
		},
	}
}

func (e *CodexEngine) GetInstallationSteps(workflowData *WorkflowData) []GitHubActionStep {
	// Build the npm install command, optionally with version
	installCmd := "npm install -g @openai/codex"
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Version != "" {
		installCmd = fmt.Sprintf("npm install -g @openai/codex@%s", workflowData.EngineConfig.Version)
	}

	return []GitHubActionStep{
		{
			"      - name: Setup Node.js",
			"        uses: actions/setup-node@v4",
			"        with:",
			"          node-version: '24'",
		},
		{
			"      - name: Install Codex",
			fmt.Sprintf("        run: %s", installCmd),
		},
	}
}

// GetExecutionSteps returns the GitHub Actions steps for executing Codex
func (e *CodexEngine) GetExecutionSteps(workflowData *WorkflowData, logFile string) []GitHubActionStep {
	var steps []GitHubActionStep

	// Handle custom steps if they exist in engine config
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Steps) > 0 {
		for _, step := range workflowData.EngineConfig.Steps {
			stepYAML, err := e.convertStepToYAML(step)
			if err != nil {
				// Log error but continue with other steps
				continue
			}
			steps = append(steps, GitHubActionStep{stepYAML})
		}
	}

	// Use model from engineConfig if available, otherwise default to o4-mini
	model := "o4-mini"
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Model != "" {
		model = workflowData.EngineConfig.Model
	}

	command := fmt.Sprintf(`set -o pipefail
INSTRUCTION=$(cat /tmp/aw-prompts/prompt.txt)
export CODEX_HOME=/tmp/mcp-config

# Create log directory outside git repo
mkdir -p /tmp/aw-logs

# Run codex with log capture - pipefail ensures codex exit code is preserved
codex exec \
  -c model=%s \
  --full-auto "$INSTRUCTION" 2>&1 | tee %s`, model, logFile)

	env := map[string]string{
		"OPENAI_API_KEY":      "${{ secrets.OPENAI_API_KEY }}",
		"GITHUB_STEP_SUMMARY": "${{ env.GITHUB_STEP_SUMMARY }}",
		"GITHUB_AW_PROMPT":    "/tmp/aw-prompts/prompt.txt",
	}

	// Add GITHUB_AW_SAFE_OUTPUTS if output is needed
	hasOutput := workflowData.SafeOutputs != nil
	if hasOutput {
		env["GITHUB_AW_SAFE_OUTPUTS"] = "${{ env.GITHUB_AW_SAFE_OUTPUTS }}"
	}

	// Add custom environment variables from engine config
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Env) > 0 {
		for key, value := range workflowData.EngineConfig.Env {
			env[key] = value
		}
	}

	// Generate the step for Codex execution
	stepName := "Run Codex"
	var stepLines []string

	stepLines = append(stepLines, fmt.Sprintf("      - name: %s", stepName))
	stepLines = append(stepLines, "        run: |")

	// Split command into lines and indent them properly
	commandLines := strings.Split(command, "\n")
	for _, line := range commandLines {
		stepLines = append(stepLines, "          "+line)
	}

	// Add environment variables
	if len(env) > 0 {
		stepLines = append(stepLines, "        env:")
		// Sort environment keys for consistent output
		envKeys := make([]string, 0, len(env))
		for key := range env {
			envKeys = append(envKeys, key)
		}
		sort.Strings(envKeys)

		for _, key := range envKeys {
			value := env[key]
			stepLines = append(stepLines, fmt.Sprintf("          %s: %s", key, value))
		}
	}

	steps = append(steps, GitHubActionStep(stepLines))

	return steps
}

// convertStepToYAML converts a step map to YAML string - temporary helper
func (e *CodexEngine) convertStepToYAML(stepMap map[string]any) (string, error) {
	// Simple YAML generation for steps - this mirrors the compiler logic
	var stepYAML []string

	// Add step name
	if name, hasName := stepMap["name"]; hasName {
		if nameStr, ok := name.(string); ok {
			stepYAML = append(stepYAML, fmt.Sprintf("      - name: %s", nameStr))
		}
	}

	// Add id field if present
	if id, hasID := stepMap["id"]; hasID {
		if idStr, ok := id.(string); ok {
			stepYAML = append(stepYAML, fmt.Sprintf("        id: %s", idStr))
		}
	}

	// Add continue-on-error field if present
	if continueOnError, hasContinueOnError := stepMap["continue-on-error"]; hasContinueOnError {
		// Handle both string and boolean values for continue-on-error
		switch v := continueOnError.(type) {
		case bool:
			stepYAML = append(stepYAML, fmt.Sprintf("        continue-on-error: %t", v))
		case string:
			stepYAML = append(stepYAML, fmt.Sprintf("        continue-on-error: %s", v))
		}
	}

	// Add uses action
	if uses, hasUses := stepMap["uses"]; hasUses {
		if usesStr, ok := uses.(string); ok {
			stepYAML = append(stepYAML, fmt.Sprintf("        uses: %s", usesStr))
		}
	}

	// Add run command
	if run, hasRun := stepMap["run"]; hasRun {
		if runStr, ok := run.(string); ok {
			stepYAML = append(stepYAML, "        run: |")
			// Split command into lines and indent them properly
			runLines := strings.Split(runStr, "\n")
			for _, line := range runLines {
				stepYAML = append(stepYAML, "          "+line)
			}
		}
	}

	// Add with parameters
	if with, hasWith := stepMap["with"]; hasWith {
		if withMap, ok := with.(map[string]any); ok {
			stepYAML = append(stepYAML, "        with:")
			// Sort keys for stable output
			keys := make([]string, 0, len(withMap))
			for key := range withMap {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				stepYAML = append(stepYAML, fmt.Sprintf("          %s: %v", key, withMap[key]))
			}
		}
	}

	return strings.Join(stepYAML, "\n"), nil
}

func (e *CodexEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	yaml.WriteString("          cat > /tmp/mcp-config/config.toml << EOF\n")

	// Add history configuration to disable persistence
	yaml.WriteString("          [history]\n")
	yaml.WriteString("          persistence = \"none\"\n")

	// Generate [mcp_servers] section
	for _, toolName := range mcpTools {
		switch toolName {
		case "github":
			githubTool := tools["github"]
			e.renderGitHubCodexMCPConfig(yaml, githubTool)
		case "safe-outputs":
			e.renderSafeOutputsCodexMCPConfig(yaml)
		default:
			// Handle custom MCP tools (those with MCP-compatible type)
			if toolConfig, ok := tools[toolName].(map[string]any); ok {
				if hasMcp, _ := hasMCPConfig(toolConfig); hasMcp {
					if err := e.renderCodexMCPConfig(yaml, toolName, toolConfig); err != nil {
						fmt.Printf("Error generating custom MCP configuration for %s: %v\n", toolName, err)
					}
				}
			}
		}
	}

	yaml.WriteString("          EOF\n")
}

// ParseLogMetrics implements engine-specific log parsing for Codex
func (e *CodexEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	var metrics LogMetrics
	var totalTokenUsage int

	lines := strings.Split(logContent, "\n")

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Extract Codex-specific token usage (always sum for Codex)
		if tokenUsage := e.extractCodexTokenUsage(line); tokenUsage > 0 {
			totalTokenUsage += tokenUsage
		}

		// Count errors and warnings
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") {
			metrics.ErrorCount++
		}
		if strings.Contains(lowerLine, "warning") {
			metrics.WarningCount++
		}
	}

	metrics.TokenUsage = totalTokenUsage

	return metrics
}

// extractCodexTokenUsage extracts token usage from Codex-specific log lines
func (e *CodexEngine) extractCodexTokenUsage(line string) int {
	// Codex format: "tokens used: 13934"
	codexPattern := `tokens\s+used[:\s]+(\d+)`
	if match := ExtractFirstMatch(line, codexPattern); match != "" {
		if count, err := strconv.Atoi(match); err == nil {
			return count
		}
	}
	return 0
}

// renderGitHubCodexMCPConfig generates GitHub MCP server configuration for codex config.toml
// Always uses Docker MCP as the default
func (e *CodexEngine) renderGitHubCodexMCPConfig(yaml *strings.Builder, githubTool any) {
	githubDockerImageVersion := getGitHubDockerImageVersion(githubTool)
	yaml.WriteString("          \n")
	yaml.WriteString("          [mcp_servers.github]\n")

	// Always use Docker-based GitHub MCP server (services mode has been removed)
	yaml.WriteString("          command = \"docker\"\n")
	yaml.WriteString("          args = [\n")
	yaml.WriteString("            \"run\",\n")
	yaml.WriteString("            \"-i\",\n")
	yaml.WriteString("            \"--rm\",\n")
	yaml.WriteString("            \"-e\",\n")
	yaml.WriteString("            \"GITHUB_PERSONAL_ACCESS_TOKEN\",\n")
	yaml.WriteString("            \"ghcr.io/github/github-mcp-server:" + githubDockerImageVersion + "\"\n")
	yaml.WriteString("          ]\n")
	yaml.WriteString("          env = { \"GITHUB_PERSONAL_ACCESS_TOKEN\" = \"${{ secrets.GITHUB_TOKEN }}\" }\n")
}

// renderSafeOutputsCodexMCPConfig generates the Safe Outputs MCP server configuration for Codex
func (e *CodexEngine) renderSafeOutputsCodexMCPConfig(yaml *strings.Builder) {
	yaml.WriteString("          \n")
	yaml.WriteString("          [mcp_servers.safe-outputs]\n")
	yaml.WriteString("          command = \"npx\"\n")
	yaml.WriteString("          args = [\"tsx\", \"/tmp/mcp-safe-outputs/server.ts\"]\n")
	yaml.WriteString("          env = {\n")
	yaml.WriteString("            \"MCP_SAFE_OUTPUTS_CONFIG\" = \"$(cat /tmp/mcp-safe-outputs/config.json)\",\n")
	yaml.WriteString("            \"GITHUB_AW_SAFE_OUTPUTS\" = \"${{ env.GITHUB_AW_SAFE_OUTPUTS }}\"\n")
	yaml.WriteString("          }\n")
}

// renderCodexMCPConfig generates custom MCP server configuration for a single tool in codex workflow config.toml
func (e *CodexEngine) renderCodexMCPConfig(yaml *strings.Builder, toolName string, toolConfig map[string]any) error {
	yaml.WriteString("          \n")
	fmt.Fprintf(yaml, "          [mcp_servers.%s]\n", toolName)

	// Use the shared MCP config renderer with TOML format
	renderer := MCPConfigRenderer{
		IndentLevel: "          ",
		Format:      "toml",
	}

	err := renderSharedMCPConfig(yaml, toolName, toolConfig, renderer)
	if err != nil {
		return err
	}

	return nil
}

// GetLogParserScript returns the JavaScript script name for parsing Codex logs
func (e *CodexEngine) GetLogParserScript() string {
	return "parse_codex_log"
}
