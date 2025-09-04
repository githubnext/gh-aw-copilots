package workflow

import (
	"fmt"
	"strconv"
	"strings"
)

// GenAIScriptEngine represents the GenAIScript agentic engine
type GenAIScriptEngine struct {
	BaseEngine
}

func NewGenAIScriptEngine() *GenAIScriptEngine {
	return &GenAIScriptEngine{
		BaseEngine: BaseEngine{
			id:                     "genaiscript",
			displayName:            "GenAIScript",
			description:            "Uses GenAIScript to run markdown-based AI scripts with JavaScript/TypeScript",
			experimental:           true, // New engine, mark as experimental
			supportsToolsWhitelist: true, // GenAIScript has its own tool system
			supportsHTTPTransport:  true, // GenAIScript supports HTTP transports
			supportsMaxTurns:       true, // GenAIScript supports max-turns in the front matter
		},
	}
}

func (e *GenAIScriptEngine) GetInstallationSteps(engineConfig *EngineConfig, networkPermissions *NetworkPermissions) []GitHubActionStep {
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

// GetDeclaredOutputFiles returns the output files that GenAIScript may produce
func (e *GenAIScriptEngine) GetDeclaredOutputFiles() []string {
	return []string{"output.txt", "*.genai.md"}
}

func (e *GenAIScriptEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig, networkPermissions *NetworkPermissions, hasOutput bool) ExecutionConfig {
	// Leave model empty if not specified - GenAIScript will use its default
	var model string
	if engineConfig != nil && engineConfig.Model != "" {
		model = engineConfig.Model
	}

	// Build model argument for genaiscript command
	modelArg := ""
	if model != "" {
		modelArg = fmt.Sprintf(" --model %s", model)
	}

	command := fmt.Sprintf(`set -o pipefail

# Create GenAIScript workspace
mkdir -p /tmp/genaiscript-workspace

# Read the workflow content 
WORKFLOW_CONTENT=$(cat /tmp/aw-prompts/prompt.txt)

# Create the GenAIScript markdown file directly
cat > /tmp/genaiscript-workspace/workflow.genai.md << EOF
---
title: GitHub Agentic Workflow
description: Execute workflow instructions using GenAIScript
model: %s
---

$WORKFLOW_CONTENT
EOF

# Change to workspace directory
cd /tmp/genaiscript-workspace

# Run genaiscript with the markdown file - pipefail ensures genaiscript exit code is preserved
genaiscript run workflow.genai.md%s --out /tmp/genaiscript-output 2>&1 | tee %s`, model, modelArg, logFile)

	env := map[string]string{
		"GITHUB_STEP_SUMMARY": "${{ env.GITHUB_STEP_SUMMARY }}",
	}

	// Add GITHUB_AW_SAFE_OUTPUTS if output is needed
	if hasOutput {
		env["GITHUB_AW_SAFE_OUTPUTS"] = "${{ env.GITHUB_AW_SAFE_OUTPUTS }}"
	}

	return ExecutionConfig{
		StepName:    "Run GenAIScript",
		Command:     command,
		Environment: env,
	}
}

func (e *GenAIScriptEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	// GenAIScript supports MCPs - configure MCP servers
	if len(mcpTools) > 0 {
		yaml.WriteString("          # GenAIScript MCP configuration\n")
		yaml.WriteString("          GENAISCRIPT_MCP_SERVERS: |\n")
		for _, tool := range mcpTools {
			yaml.WriteString(fmt.Sprintf("            %s\n", tool))
		}
	} else {
		yaml.WriteString("          # No MCP servers configured for GenAIScript\n")
	}
}

// ParseLogMetrics implements engine-specific log parsing for GenAIScript
func (e *GenAIScriptEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	var metrics LogMetrics
	var totalTokenUsage int

	lines := strings.Split(logContent, "\n")

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Extract GenAIScript-specific token usage
		if tokenUsage := e.extractGenAIScriptTokenUsage(line); tokenUsage > 0 {
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

// extractGenAIScriptTokenUsage extracts token usage from GenAIScript-specific log lines
func (e *GenAIScriptEngine) extractGenAIScriptTokenUsage(line string) int {
	// GenAIScript typically logs token usage in formats like:
	// "tokens: 1234" or "completion_tokens: 567" or "total_tokens: 1801"
	patterns := []string{
		`total[_\s]tokens[:\s]+(\d+)`,
		`tokens[:\s]+(\d+)`,
		`completion[_\s]tokens[:\s]+(\d+)`,
	}

	for _, pattern := range patterns {
		if match := ExtractFirstMatch(line, pattern); match != "" {
			if count, err := strconv.Atoi(match); err == nil {
				return count
			}
		}
	}

	return 0
}

// GetLogParserScript returns the JavaScript script name for parsing GenAIScript logs
func (e *GenAIScriptEngine) GetLogParserScript() string {
	return "parse_genaiscript_log"
}
