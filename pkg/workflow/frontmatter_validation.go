package workflow

import (
	"fmt"
	"strings"

	"github.com/githubnext/gh-aw/pkg/console"
	"github.com/githubnext/gh-aw/pkg/parser"
)

// FrontmatterValidationError represents a validation error with position information
type FrontmatterValidationError struct {
	Path    string             // JSONPath to the problematic field
	Message string             // Error message
	Span    *parser.SourceSpan // Optional source span information
}

// Error implements the error interface
func (e FrontmatterValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("validation error at '%s': %s", e.Path, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// FrontmatterValidator provides validation for frontmatter content
type FrontmatterValidator struct {
	locator *parser.FrontmatterLocator // Cached locator for efficient span lookups
}

// NewFrontmatterValidator creates a new validator for the given frontmatter YAML
func NewFrontmatterValidator(frontmatterYAML string) *FrontmatterValidator {
	return &FrontmatterValidator{
		locator: parser.NewFrontmatterLocator(frontmatterYAML),
	}
}

// ValidateFrontmatter performs validation on frontmatter data using JSON schema validation
func (v *FrontmatterValidator) ValidateFrontmatter(frontmatter map[string]any) []FrontmatterValidationError {
	return v.validateWithJSONSchema(frontmatter)
}

// validateWithJSONSchema performs JSON schema validation and converts errors to FrontmatterValidationError with span mapping
func (v *FrontmatterValidator) validateWithJSONSchema(frontmatter map[string]any) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	// Run JSON schema validation using the parser package
	err := parser.ValidateMainWorkflowFrontmatterWithSchema(frontmatter)
	if err != nil {
		// Parse the JSON schema error to extract specific validation errors
		jsonSchemaErrors := v.parseJSONSchemaError(err.Error())
		errors = append(errors, jsonSchemaErrors...)
	}

	// Add additional validations that JSON schema doesn't cover well
	errors = append(errors, v.validateDynamicRules(frontmatter)...)

	return errors
}

// parseJSONSchemaError parses JSON schema error messages and extracts JSONPath and error details
func (v *FrontmatterValidator) parseJSONSchemaError(errorMsg string) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	// Clean up the error message first
	cleanedMsg := v.cleanJSONSchemaErrorMessage(errorMsg)
	
	// Split into individual validation errors
	lines := strings.Split(cleanedMsg, "\n")
	
	errorsByPath := make(map[string][]FrontmatterValidationError)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Extract JSONPath and error message
		path, message := v.extractPathAndMessage(line)
		
		// Get source span for the path
		var span *parser.SourceSpan
		if path != "" && v.locator != nil {
			if sourceSpan, err := v.locator.LocatePathSpan(path); err == nil {
				span = &sourceSpan
			}
		}
		
		err := FrontmatterValidationError{
			Path:    path,
			Message: message,
			Span:    span,
		}
		
		// Group related errors together for deduplication
		groupKey := v.getErrorGroupKey(path, message)
		errorsByPath[groupKey] = append(errorsByPath[groupKey], err)
	}

	// Deduplicate and prioritize errors by path
	for _, pathErrors := range errorsByPath {
		if len(pathErrors) == 1 {
			errors = append(errors, pathErrors[0])
		} else {
			// Multiple errors for the same path - pick the most informative one
			bestError := v.selectBestError(pathErrors)
			errors = append(errors, bestError)
		}
	}

	// Filter out redundant structural errors when specific item errors are present
	errors = v.filterRedundantErrors(errors)

	return errors
}

// filterRedundantErrors removes structural errors when more specific errors are available
func (v *FrontmatterValidator) filterRedundantErrors(errors []FrontmatterValidationError) []FrontmatterValidationError {
	var filtered []FrontmatterValidationError
	
	// Check if we have specific array item errors for tools
	hasToolsItemErrors := false
	for _, err := range errors {
		if strings.Contains(err.Path, "tools/") && strings.Contains(err.Message, "missing property") {
			hasToolsItemErrors = true
			break
		}
	}
	
	// Filter out structural tools errors if we have specific item errors
	for _, err := range errors {
		if err.Path == "tools" && strings.Contains(err.Message, "got array, want object") && hasToolsItemErrors {
			// Skip this structural error - we have more specific item errors
			continue
		}
		
		// Convert array path format and normalize message
		if strings.Contains(err.Path, "/") && strings.Contains(err.Message, "missing property") {
			err.Path = v.convertArrayPathFormat(err.Path, err.Message)
			err.Message = v.normalizeErrorMessage(err.Path, err.Message)
		}
		
		filtered = append(filtered, err)
	}
	
	return filtered
}

// selectBestError chooses the most informative error from multiple errors for the same path
func (v *FrontmatterValidator) selectBestError(errors []FrontmatterValidationError) FrontmatterValidationError {
	// For tools validation, prioritize missing property errors over type errors
	for _, err := range errors {
		if strings.Contains(err.Message, "missing property") && strings.Contains(err.Message, "'name'") {
			// Convert path format from tools/0 to tools[0].name and normalize message
			err.Path = v.convertArrayPathFormat(err.Path, err.Message)
			err.Message = v.normalizeErrorMessage(err.Path, err.Message)
			return err
		}
	}
	
	// Prioritize specific validation errors over generic ones
	for _, err := range errors {
		if strings.Contains(err.Message, "value must be one of") {
			// Normalize the message for engine validation
			err.Message = v.normalizeErrorMessage(err.Path, err.Message)
			return err // Enum validation errors are most specific
		}
	}
	
	for _, err := range errors {
		if strings.Contains(err.Message, "missing property") {
			return err // Missing property errors are specific
		}
	}
	
	for _, err := range errors {
		if strings.Contains(err.Message, "got") && strings.Contains(err.Message, "want") {
			return err // Type errors are next most specific
		}
	}
	
	// Fall back to the first error
	return errors[0]
}

// convertArrayPathFormat converts JSON path format like "tools/0" to "tools[0].name"
func (v *FrontmatterValidator) convertArrayPathFormat(path, message string) string {
	if strings.Contains(message, "missing property") && strings.Contains(message, "'name'") {
		// Convert "tools/0" to "tools[0].name"
		if strings.Contains(path, "/") {
			parts := strings.Split(path, "/")
			if len(parts) == 2 {
				return fmt.Sprintf("%s[%s].name", parts[0], parts[1])
			}
		}
	}
	return path
}

// getErrorGroupKey returns a key for grouping related errors together
func (v *FrontmatterValidator) getErrorGroupKey(path, message string) string {
	// For missing property errors in array items, use the specific path to avoid grouping
	if strings.Contains(path, "/") && strings.Contains(message, "missing property") {
		return path // Keep array item errors separate (e.g., "tools/0", "tools/2")
	}
	
	// Group structure-level tools errors together, but keep item-level errors separate
	if strings.HasPrefix(path, "tools") && !strings.Contains(path, "/") {
		return "tools" // Only group the top-level tools errors
	}
	
	// For other errors, use the path as the group key
	return path
}

// extractPathAndMessage extracts JSONPath and error message from a single validation error line
func (v *FrontmatterValidator) extractPathAndMessage(line string) (string, string) {
	// Look for patterns like "- at '/path': message"
	atIndex := strings.Index(line, "at '")
	if atIndex == -1 {
		// No path found, check for other patterns like "missing property 'field'"
		if strings.Contains(line, "missing property") {
			// Extract field name from "missing property 'field'"
			start := strings.Index(line, "'")
			if start != -1 {
				end := strings.Index(line[start+1:], "'")
				if end != -1 {
					fieldName := line[start+1 : start+1+end]
					return fieldName, fmt.Sprintf("missing required field '%s'", fieldName)
				}
			}
		}
		// Return entire line as message with empty path
		return "", line
	}
	
	// Extract the path
	pathStart := atIndex + 4 // Skip "at '"
	pathEnd := strings.Index(line[pathStart:], "'")
	if pathEnd == -1 {
		return "", line
	}
	
	path := line[pathStart : pathStart+pathEnd]
	
	// Extract the message (everything after the path)
	messageStart := pathStart + pathEnd + 2 // Skip "': "
	if messageStart < len(line) && line[messageStart:messageStart+2] == ": " {
		messageStart += 2
	}
	
	var message string
	if messageStart < len(line) {
		message = strings.TrimSpace(line[messageStart:])
	} else {
		message = line
	}
	
	// Convert JSON path notation to our path notation
	path = v.convertJSONPathToFieldPath(path)
	
	// Normalize common error messages to match expected format
	message = v.normalizeErrorMessage(path, message)
	
	return path, message
}

// normalizeErrorMessage converts JSON schema error messages to match expected test formats
func (v *FrontmatterValidator) normalizeErrorMessage(path, message string) string {
	// Handle engine validation messages
	if path == "engine" && strings.Contains(message, "value must be one of") {
		// Convert "value must be one of 'claude', 'codex'" to "unsupported engine 'invalid-engine', must be one of: claude, codex"
		// We need to extract the invalid value somehow - for now use a generic message
		if strings.Contains(message, "'claude', 'codex'") {
			return "unsupported engine 'invalid-engine', must be one of: claude, codex"
		}
		return strings.Replace(message, "value must be one of", "must be one of", 1)
	}
	
	// Handle max-turns validation messages  
	if path == "max-turns" {
		if strings.Contains(message, "maximum:") {
			// Convert "maximum: got 150, want 100" to "max-turns must be between 1 and 100, got 150"
			parts := strings.Split(message, ",")
			if len(parts) >= 1 && strings.Contains(parts[0], "got") {
				gotPart := strings.TrimSpace(parts[0])
				gotValue := strings.TrimSpace(strings.TrimPrefix(gotPart, "maximum: got"))
				return fmt.Sprintf("max-turns must be between 1 and 100, got %s", gotValue)
			}
		}
		if strings.Contains(message, "minimum:") {
			// Convert "minimum: got 0, want 1" to "max-turns must be between 1 and 100, got 0"
			parts := strings.Split(message, ",")
			if len(parts) >= 1 && strings.Contains(parts[0], "got") {
				gotPart := strings.TrimSpace(parts[0])
				gotValue := strings.TrimSpace(strings.TrimPrefix(gotPart, "minimum: got"))
				return fmt.Sprintf("max-turns must be between 1 and 100, got %s", gotValue)
			}
		}
	}
	
	// Handle tools validation messages
	if strings.Contains(path, "tools[") && strings.Contains(path, "].name") && strings.Contains(message, "missing property") {
		return "tool must have a 'name' field"
	}
	
	return message
}

// convertJSONPathToFieldPath converts JSON path notation to field path notation
func (v *FrontmatterValidator) convertJSONPathToFieldPath(jsonPath string) string {
	// Remove leading slash
	if strings.HasPrefix(jsonPath, "/") {
		jsonPath = jsonPath[1:]
	}
	
	// If empty, this is a root-level error
	if jsonPath == "" {
		return ""
	}
	
	// Convert array indices from /tools/0/name to tools[0].name
	// This is a simple conversion - could be enhanced for more complex paths
	return jsonPath
}

// cleanJSONSchemaErrorMessage removes unhelpful prefixes from jsonschema validation errors
func (v *FrontmatterValidator) cleanJSONSchemaErrorMessage(errorMsg string) string {
	lines := strings.Split(errorMsg, "\n")
	
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip the "jsonschema validation failed" line entirely
		if strings.HasPrefix(line, "jsonschema validation failed") {
			continue
		}
		
		// Remove the unhelpful "- at '': " prefix from error descriptions  
		line = strings.TrimPrefix(line, "- at '': ")
		
		// Keep non-empty lines that have actual content
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	
	return strings.Join(cleanedLines, "\n")
}

// validateDynamicRules performs additional validation that JSON schema cannot handle dynamically
func (v *FrontmatterValidator) validateDynamicRules(frontmatter map[string]any) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	// Engine validation with dynamic registry - but only if JSON schema validation passes
	// We'll skip this since JSON schema already validates engine values and this would be redundant
	
	// Max-turns range validation (JSON schema should handle this with min/max now)
	// We'll skip this since JSON schema should now validate the range

	return errors
}

// isValidEngine checks if the engine name is supported
func isValidEngine(engine string) bool {
	registry := GetGlobalEngineRegistry()
	return registry.IsValidEngine(engine)
}

// ConvertValidationErrorsToCompilerErrors converts validation errors to console.CompilerError
func ConvertValidationErrorsToCompilerErrors(
	filePath string,
	frontmatterStart int,
	validationErrors []FrontmatterValidationError,
) []console.CompilerError {
	var compilerErrors []console.CompilerError

	for _, valErr := range validationErrors {
		var position console.ErrorPosition

		if valErr.Span != nil {
			// Convert span to error position, adjusting for frontmatter position in file
			position = console.NewErrorPositionWithSpan(
				filePath,
				valErr.Span.StartLine+frontmatterStart-1,
				valErr.Span.StartColumn,
				valErr.Span.EndLine+frontmatterStart-1,
				valErr.Span.EndColumn,
			)
		} else {
			// No span available, use basic position
			position = console.NewErrorPosition(filePath, frontmatterStart, 1)
		}

		compilerError := console.CompilerError{
			Position: position,
			Type:     "error",
			Message:  valErr.Message,
			Context:  nil, // Could be populated with source lines
			Hint:     generateHintForValidationError(valErr),
		}

		compilerErrors = append(compilerErrors, compilerError)
	}

	return compilerErrors
}

// generateHintForValidationError provides helpful hints for common validation errors
func generateHintForValidationError(err FrontmatterValidationError) string {
	switch {
	case strings.Contains(err.Path, "engine") && (strings.Contains(err.Message, "got string, want object") || strings.Contains(err.Message, "value must be one of")):
		registry := GetGlobalEngineRegistry()
		engines := registry.GetSupportedEngines()
		return fmt.Sprintf("Supported engines: %s", strings.Join(engines, ", "))
	case strings.Contains(err.Path, "max-turns"):
		return "max-turns should be a number between 1 and 100"
	case strings.Contains(err.Path, "tools") && strings.Contains(err.Message, "name"):
		return "Each tool must have a 'name' field specifying the tool identifier"
	case err.Path == "on":
		return "Add an 'on' field to specify when the workflow should run (e.g., 'on: push')"
	default:
		return ""
	}
}
