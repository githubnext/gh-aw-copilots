package workflow

import (
	"encoding/json"
	"fmt"
)

// EngineConfig represents the parsed engine configuration
type EngineConfig struct {
	ID       string
	Version  string
	Model    string
	MaxTurns string
	Env      map[string]string
	Steps    []map[string]any
}

// NetworkPermissions represents network access permissions
type NetworkPermissions struct {
	Mode    string   `yaml:"mode,omitempty"`    // "defaults" for default access
	Allowed []string `yaml:"allowed,omitempty"` // List of allowed domains
}

// EngineNetworkConfig combines engine configuration with top-level network permissions
type EngineNetworkConfig struct {
	Engine  *EngineConfig
	Network *NetworkPermissions
}

// extractEngineConfig extracts engine configuration from frontmatter, supporting both string and object formats
func (c *Compiler) extractEngineConfig(frontmatter map[string]any) (string, *EngineConfig) {
	if engine, exists := frontmatter["engine"]; exists {
		// Handle string format (backwards compatibility)
		if engineStr, ok := engine.(string); ok {
			return engineStr, &EngineConfig{ID: engineStr}
		}

		// Handle object format
		if engineObj, ok := engine.(map[string]any); ok {
			config := &EngineConfig{}

			// Extract required 'id' field
			if id, hasID := engineObj["id"]; hasID {
				if idStr, ok := id.(string); ok {
					config.ID = idStr
				}
			}

			// Extract optional 'version' field
			if version, hasVersion := engineObj["version"]; hasVersion {
				if versionStr, ok := version.(string); ok {
					config.Version = versionStr
				}
			}

			// Extract optional 'model' field
			if model, hasModel := engineObj["model"]; hasModel {
				if modelStr, ok := model.(string); ok {
					config.Model = modelStr
				}
			}

			// Extract optional 'max-turns' field
			if maxTurns, hasMaxTurns := engineObj["max-turns"]; hasMaxTurns {
				if maxTurnsInt, ok := maxTurns.(int); ok {
					config.MaxTurns = fmt.Sprintf("%d", maxTurnsInt)
				} else if maxTurnsUint64, ok := maxTurns.(uint64); ok {
					config.MaxTurns = fmt.Sprintf("%d", maxTurnsUint64)
				} else if maxTurnsStr, ok := maxTurns.(string); ok {
					config.MaxTurns = maxTurnsStr
				}
			}

			// Extract optional 'env' field (object/map of strings)
			if env, hasEnv := engineObj["env"]; hasEnv {
				if envMap, ok := env.(map[string]any); ok {
					config.Env = make(map[string]string)
					for key, value := range envMap {
						if valueStr, ok := value.(string); ok {
							config.Env[key] = valueStr
						}
					}
				}
			}

			// Extract optional 'steps' field (array of step objects)
			if steps, hasSteps := engineObj["steps"]; hasSteps {
				if stepsArray, ok := steps.([]any); ok {
					config.Steps = make([]map[string]any, 0, len(stepsArray))
					for _, step := range stepsArray {
						if stepMap, ok := step.(map[string]any); ok {
							config.Steps = append(config.Steps, stepMap)
						}
					}
				}
			}

			// Return the ID as the engineSetting for backwards compatibility
			return config.ID, config
		}
	}

	// No engine specified
	return "", nil
}

// validateEngine validates that the given engine ID is supported
func (c *Compiler) validateEngine(engineID string) error {
	if engineID == "" {
		return nil // Empty engine is valid (will use default)
	}

	// First try exact match
	if c.engineRegistry.IsValidEngine(engineID) {
		return nil
	}

	// Try prefix match for backward compatibility (e.g., "codex-experimental")
	_, err := c.engineRegistry.GetEngineByPrefix(engineID)
	return err
}

// getAgenticEngine returns the agentic engine for the given engine setting
func (c *Compiler) getAgenticEngine(engineSetting string) (CodingAgentEngine, error) {
	if engineSetting == "" {
		return c.engineRegistry.GetDefaultEngine(), nil
	}

	// First try exact match
	if c.engineRegistry.IsValidEngine(engineSetting) {
		return c.engineRegistry.GetEngine(engineSetting)
	}

	// Try prefix match for backward compatibility
	return c.engineRegistry.GetEngineByPrefix(engineSetting)
}

// validateEngineConflicts validates that included engines don't conflict with the main engine
func (c *Compiler) validateEngineConflicts(mainEngineSetting string, includedEnginesJSON []string) error {
	if len(includedEnginesJSON) == 0 {
		return nil // No included engines to check
	}

	// Parse main engine configuration
	mainConfig := &EngineConfig{}
	if mainEngineSetting != "" {
		mainConfig.ID = mainEngineSetting
	}

	// Check each included engine
	for _, engineJSON := range includedEnginesJSON {
		var includedEngine interface{}
		if err := json.Unmarshal([]byte(engineJSON), &includedEngine); err != nil {
			return fmt.Errorf("failed to parse included engine configuration: %w", err)
		}

		includedConfig := &EngineConfig{}

		// Handle string format
		if engineStr, ok := includedEngine.(string); ok {
			includedConfig.ID = engineStr
		} else if engineObj, ok := includedEngine.(map[string]interface{}); ok {
			// Handle object format
			if id, hasID := engineObj["id"]; hasID {
				if idStr, ok := id.(string); ok {
					includedConfig.ID = idStr
				}
			}
		}

		// Check for conflicts
		if mainEngineSetting != "" && includedConfig.ID != "" && mainConfig.ID != includedConfig.ID {
			return fmt.Errorf("engine conflict: main workflow specifies engine '%s' but included workflow specifies engine '%s'. Remove the engine specification from either the main workflow or the included workflow", mainConfig.ID, includedConfig.ID)
		}
	}

	return nil
}

// mergeEngineConfigs merges engine configurations, preferring the main workflow's engine
func (c *Compiler) mergeEngineConfigs(mainEngineSetting string, includedEnginesJSON []string) (string, error) {
	// If main workflow has engine setting, use it
	if mainEngineSetting != "" {
		return mainEngineSetting, nil
	}

	// If no main engine but we have included engines, use the first one
	if len(includedEnginesJSON) > 0 {
		var firstEngine interface{}
		if err := json.Unmarshal([]byte(includedEnginesJSON[0]), &firstEngine); err != nil {
			return "", fmt.Errorf("failed to parse included engine configuration: %w", err)
		}

		// Handle string format
		if engineStr, ok := firstEngine.(string); ok {
			return engineStr, nil
		} else if engineObj, ok := firstEngine.(map[string]interface{}); ok {
			// Handle object format - return the ID
			if id, hasID := engineObj["id"]; hasID {
				if idStr, ok := id.(string); ok {
					return idStr, nil
				}
			}
		}
	}

	return "", nil // No engine specified anywhere
}
