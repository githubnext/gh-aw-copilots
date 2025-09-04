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
			experimental:           true,  // New engine, mark as experimental
			supportsToolsWhitelist: false, // GenAIScript has its own tool system
			supportsHTTPTransport:  false, // GenAIScript uses its own transport
			supportsMaxTurns:       false, // GenAIScript doesn't support max-turns feature
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
	// Use model from engineConfig if available, otherwise default to gpt-4o-mini
	model := "gpt-4o-mini"
	if engineConfig != nil && engineConfig.Model != "" {
		model = engineConfig.Model
	}

	command := fmt.Sprintf(`set -o pipefail
# Read the instruction/prompt from the aw-prompts file
INSTRUCTION=$(cat /tmp/aw-prompts/prompt.txt)

# Create a temporary GenAIScript script file that contains the markdown prompt
mkdir -p /tmp/genaiscript-workspace/genaisrc
cat > /tmp/genaiscript-workspace/genaisrc/workflow.genai.mts << 'EOF'
script({
  title: "GitHub Agentic Workflow",
  description: "Execute workflow instructions using GenAIScript",
  model: "%s"
})

// Use the instruction from the workflow
$INSTRUCTION = process.env.WORKFLOW_INSTRUCTION || ""
$`+"`"+`${$INSTRUCTION}`+"`"+`
EOF

# Set the working directory for GenAIScript
cd /tmp/genaiscript-workspace

# Configure GenAIScript to work in this directory
export GENAISCRIPT_VAR_WORKSPACE_INSTRUCTION="$INSTRUCTION"

# Run genaiscript with log capture - pipefail ensures genaiscript exit code is preserved
genaiscript run workflow \
  --model %s \
  --out /tmp/genaiscript-output 2>&1 | tee %s`, model, model, logFile)

	env := map[string]string{
		"OPENAI_API_KEY":       "${{ secrets.OPENAI_API_KEY }}",
		"GITHUB_STEP_SUMMARY":  "${{ env.GITHUB_STEP_SUMMARY }}",
		"WORKFLOW_INSTRUCTION": "$(cat /tmp/aw-prompts/prompt.txt)",
	}

	// Add other API keys that GenAIScript might use
	env["ANTHROPIC_API_KEY"] = "${{ secrets.ANTHROPIC_API_KEY }}"
	env["AZURE_OPENAI_API_KEY"] = "${{ secrets.AZURE_OPENAI_API_KEY }}"
	env["GOOGLE_API_KEY"] = "${{ secrets.GOOGLE_API_KEY }}"

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
	// GenAIScript has its own tool/MCP system, so we don't need to configure MCP servers here
	// For now, we'll just create an empty placeholder in case it's needed
	yaml.WriteString("          # GenAIScript uses its own tool system\n")
	yaml.WriteString("          # No MCP configuration needed\n")
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
