package workflow

import (
	"strings"
)

// CustomEngine represents a custom agentic engine that executes user-defined GitHub Actions steps
type CustomEngine struct {
	BaseEngine
}

// NewCustomEngine creates a new CustomEngine instance
func NewCustomEngine() *CustomEngine {
	return &CustomEngine{
		BaseEngine: BaseEngine{
			id:                     "custom",
			displayName:            "Custom Steps",
			description:            "Executes user-defined GitHub Actions steps",
			experimental:           false,
			supportsToolsWhitelist: false,
			supportsHTTPTransport:  false,
			supportsMaxTurns:       false,
		},
	}
}

// GetInstallationSteps returns empty installation steps since custom engine doesn't need installation
func (e *CustomEngine) GetInstallationSteps(engineConfig *EngineConfig, networkPermissions *NetworkPermissions) []GitHubActionStep {
	return []GitHubActionStep{}
}

// GetExecutionConfig returns the execution configuration for custom steps
func (e *CustomEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig, networkPermissions *NetworkPermissions, hasOutput bool) ExecutionConfig {
	// The custom engine doesn't execute itself - the steps are handled directly by the compiler
	// This method is called but the actual execution logic is handled in the compiler
	config := ExecutionConfig{
		StepName: "Custom Steps Execution",
		Command:  "echo \"Custom steps are handled directly by the compiler\"",
		Environment: map[string]string{
			"WORKFLOW_NAME": workflowName,
		},
	}

	// If the engine configuration has custom steps, include them in the execution config
	if engineConfig != nil && len(engineConfig.Steps) > 0 {
		config.Steps = engineConfig.Steps
	}

	return config
}

// RenderMCPConfig renders empty MCP configuration since custom engine doesn't use MCP
func (e *CustomEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	// Custom engine doesn't use MCP servers
}

// ParseLogMetrics implements basic log parsing for custom engine
func (e *CustomEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	var metrics LogMetrics

	lines := strings.Split(logContent, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
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

	return metrics
}

// GetLogParserScript returns the JavaScript script name for parsing custom engine logs
func (e *CustomEngine) GetLogParserScript() string {
	return "parse_custom_log"
}
