package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// DefaultAIInferenceActionVersion is the default version of the AI Inference action
	DefaultAIInferenceActionVersion = "v1"
)

// AIInferenceEngine represents the actions/ai-inference agentic engine
type AIInferenceEngine struct {
	BaseEngine
}

func NewAIInferenceEngine() *AIInferenceEngine {
	return &AIInferenceEngine{
		BaseEngine: BaseEngine{
			id:                     "ai-inference",
			displayName:            "AI Inference",
			description:            "Uses GitHub's actions/ai-inference action with MCP tool support",
			experimental:           false,
			supportsToolsWhitelist: false, // AI Inference doesn't support MCP tool whitelisting yet
			supportsHTTPTransport:  false, // AI Inference doesn't support HTTP transport yet
			supportsMaxTurns:       false, // AI Inference doesn't support max-turns feature yet
		},
	}
}

func (e *AIInferenceEngine) GetInstallationSteps(engineConfig *EngineConfig, networkPermissions *NetworkPermissions) []GitHubActionStep {
	// AI Inference action doesn't require any special installation steps
	// The action itself will be used directly in the execution step
	return []GitHubActionStep{}
}

// GetDeclaredOutputFiles returns the output files that AI Inference may produce
func (e *AIInferenceEngine) GetDeclaredOutputFiles() []string {
	// AI Inference action may produce output files but we'll keep it simple for now
	return []string{}
}

func (e *AIInferenceEngine) GetExecutionConfig(workflowName string, logFile string, engineConfig *EngineConfig, networkPermissions *NetworkPermissions, hasOutput bool) ExecutionConfig {
	// Determine the action version to use
	actionVersion := DefaultAIInferenceActionVersion
	if engineConfig != nil && engineConfig.Version != "" {
		actionVersion = engineConfig.Version
	}

	// Determine the model to use
	model := "gpt-4o-mini"
	if engineConfig != nil && engineConfig.Model != "" {
		model = engineConfig.Model
	}

	// Build environment variables for the AI Inference action
	aiInferenceEnv := ""
	
	// Add the prompt file path as an environment variable so ai-inference can use it
	aiInferenceEnv += "            GITHUB_AW_PROMPT_FILE: /tmp/aw-prompts/prompt.txt"

	// Add safe outputs environment variable if needed
	if hasOutput {
		aiInferenceEnv += "\n            GITHUB_AW_SAFE_OUTPUTS: ${{ env.GITHUB_AW_SAFE_OUTPUTS }}"
	}

	// Add custom environment variables from engine config
	if engineConfig != nil && len(engineConfig.Env) > 0 {
		for key, value := range engineConfig.Env {
			aiInferenceEnv += "\n            " + key + ": " + value
		}
	}

	inputs := map[string]string{
		"model":  model,
		"prompt": "Please read the instructions from the file at $GITHUB_AW_PROMPT_FILE and follow them.",
	}

	// Add environment variables to inputs
	if aiInferenceEnv != "" {
		inputs["env"] = "|\n" + aiInferenceEnv
	}

	config := ExecutionConfig{
		StepName: "Execute AI Inference Action",
		Action:   fmt.Sprintf("actions/ai-inference@%s", actionVersion),
		Inputs:   inputs,
	}

	return config
}

func (e *AIInferenceEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string) {
	// AI Inference doesn't support MCP configuration yet, so we'll create a minimal config
	yaml.WriteString("          mkdir -p /tmp/mcp-config\n")
	yaml.WriteString("          cat > /tmp/mcp-config/mcp-servers.json << 'EOF'\n")
	yaml.WriteString("          {\n")
	yaml.WriteString("            \"mcpServers\": {}\n")
	yaml.WriteString("          }\n")
	yaml.WriteString("          EOF\n")
}

func (e *AIInferenceEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	// Basic log parsing for AI Inference - this could be expanded later
	metrics := LogMetrics{}
	
	// Look for common patterns in AI Inference logs
	lines := strings.Split(logContent, "\n")
	for _, line := range lines {
		// Try to extract JSON metrics from streaming logs
		lineMetrics := ExtractJSONMetrics(line, verbose)
		if lineMetrics.TokenUsage > 0 {
			metrics.TokenUsage += lineMetrics.TokenUsage
		}
		if lineMetrics.EstimatedCost > 0 {
			metrics.EstimatedCost += lineMetrics.EstimatedCost
		}
		
		// Count error and warning patterns
		if strings.Contains(strings.ToLower(line), "error") {
			metrics.ErrorCount++
		}
		if strings.Contains(strings.ToLower(line), "warning") || strings.Contains(strings.ToLower(line), "warn") {
			metrics.WarningCount++
		}
		
		// Additional parsing for AI Inference specific patterns
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			// Try to parse JSON for token information
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(trimmed), &jsonData); err == nil {
				// Check for usage object with prompt_tokens/completion_tokens
				if usage, exists := jsonData["usage"]; exists {
					if usageMap, ok := usage.(map[string]interface{}); ok {
						promptTokens := ConvertToInt(usageMap["prompt_tokens"])
						completionTokens := ConvertToInt(usageMap["completion_tokens"])
						if promptTokens > 0 || completionTokens > 0 {
							tokens := promptTokens + completionTokens
							metrics.TokenUsage += tokens
						}
					}
				}
			}
		}
		
		if verbose && (strings.Contains(line, "model:") || strings.Contains(line, "Model:")) {
			fmt.Printf("Found model reference in log: %s\n", line)
		}
	}
	
	return metrics
}

func (e *AIInferenceEngine) GetLogParserScript() string {
	return "parse_ai_inference_logs.cjs"
}