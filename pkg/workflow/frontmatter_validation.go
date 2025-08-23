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

// ValidateFrontmatter performs validation on frontmatter data and returns errors with spans
func (v *FrontmatterValidator) ValidateFrontmatter(frontmatter map[string]any) []FrontmatterValidationError {
	return v.ValidateFrontmatterWithOptions(frontmatter, ValidationOptions{UseJSONSchema: false})
}

// ValidationOptions configures how frontmatter validation is performed
type ValidationOptions struct {
	UseJSONSchema bool // If true, use JSON schema validation instead of custom validation
}

// ValidateFrontmatterWithOptions performs validation with configurable options
func (v *FrontmatterValidator) ValidateFrontmatterWithOptions(frontmatter map[string]any, options ValidationOptions) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	if options.UseJSONSchema {
		// Use JSON schema validation
		schemaErrors := v.validateWithJSONSchema(frontmatter)
		errors = append(errors, schemaErrors...)
	} else {
		// Use custom validation (existing behavior)
		errors = append(errors, v.validateCustomRules(frontmatter)...)
	}

	return errors
}

// validateCustomRules performs custom validation rules not covered by JSON schema
func (v *FrontmatterValidator) validateCustomRules(frontmatter map[string]any) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	// Example validation: check for required fields
	if _, exists := frontmatter["on"]; !exists {
		errors = append(errors, FrontmatterValidationError{
			Path:    "on",
			Message: "missing required field 'on'",
			Span:    nil, // No span for missing field
		})
	}

	// Example validation: check engine field if present
	if engine, exists := frontmatter["engine"]; exists {
		if engineStr, ok := engine.(string); ok {
			if !isValidEngine(engineStr) {
				span, err := v.locator.LocatePathSpan("engine")
				var spanPtr *parser.SourceSpan
				if err == nil {
					spanPtr = &span
				}

				registry := GetGlobalEngineRegistry()
				engines := registry.GetSupportedEngines()
				errors = append(errors, FrontmatterValidationError{
					Path:    "engine",
					Message: fmt.Sprintf("unsupported engine '%s', must be one of: %s", engineStr, strings.Join(engines, ", ")),
					Span:    spanPtr,
				})
			}
		}
	}

	// Example validation: check max-turns if present
	if maxTurns, exists := frontmatter["max-turns"]; exists {
		var maxTurnsInt int
		var ok bool

		// Handle both int and uint64 types (YAML can parse numbers as different types)
		switch v := maxTurns.(type) {
		case int:
			maxTurnsInt = v
			ok = true
		case uint64:
			maxTurnsInt = int(v)
			ok = true
		case int64:
			maxTurnsInt = int(v)
			ok = true
		case float64:
			maxTurnsInt = int(v)
			ok = true
		}

		if ok {
			if maxTurnsInt < 1 || maxTurnsInt > 100 {
				span, err := v.locator.LocatePathSpan("max-turns")
				var spanPtr *parser.SourceSpan
				if err == nil {
					spanPtr = &span
				}

				errors = append(errors, FrontmatterValidationError{
					Path:    "max-turns",
					Message: fmt.Sprintf("max-turns must be between 1 and 100, got %d", maxTurnsInt),
					Span:    spanPtr,
				})
			}
		}
	}

	// Example validation: check tools array structure
	if tools, exists := frontmatter["tools"]; exists {
		if toolsArray, ok := tools.([]any); ok {
			for i, tool := range toolsArray {
				if toolMap, ok := tool.(map[string]any); ok {
					if _, hasName := toolMap["name"]; !hasName {
						path := fmt.Sprintf("tools[%d].name", i)
						span, err := v.locator.LocatePathSpan(path)
						var spanPtr *parser.SourceSpan
						if err == nil {
							spanPtr = &span
						}

						errors = append(errors, FrontmatterValidationError{
							Path:    path,
							Message: "tool must have a 'name' field",
							Span:    spanPtr,
						})
					}
				}
			}
		}
	}

	return errors
}

// validateWithJSONSchema performs JSON schema validation and converts errors to FrontmatterValidationError
func (v *FrontmatterValidator) validateWithJSONSchema(frontmatter map[string]any) []FrontmatterValidationError {
	var errors []FrontmatterValidationError

	// Run JSON schema validation using the parser package
	err := parser.ValidateMainWorkflowFrontmatterWithSchema(frontmatter)
	if err != nil {
		// Convert JSON schema error to FrontmatterValidationError
		// For now, we'll create a general error without specific path/span information
		// since JSON schema errors don't map directly to source locations
		errors = append(errors, FrontmatterValidationError{
			Path:    "schema", // Use a general path for schema errors
			Message: err.Error(),
			Span:    nil, // JSON schema errors don't have source spans yet
		})
	}

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
	case strings.Contains(err.Path, "engine"):
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
