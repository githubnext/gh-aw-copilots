package workflow

import (
	"fmt"
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

	// Build max-runs argument if specified
	maxRunsArg := ""
	if engineConfig != nil && engineConfig.MaxTurns != "" {
		maxRunsArg = fmt.Sprintf(" --max-runs %s", engineConfig.MaxTurns)
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
genaiscript run workflow.genai.md%s%s --out /tmp/genaiscript-output 2>&1 | tee %s`, model, modelArg, maxRunsArg, logFile)

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

// ParseLogMetrics implements a simple log metrics parser for GenAIScript
// Returns empty metrics since GenAIScript logs are sent directly to summary
func (e *GenAIScriptEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	// Return empty metrics as GenAIScript logs are sent directly to summary without parsing
	return LogMetrics{}
}

// GetLogParserScript returns the JavaScript script name for parsing GenAIScript logs
func (e *GenAIScriptEngine) GetLogParserScript() string {
	return "parse_genaiscript_log"
}
