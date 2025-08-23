package workflow

import (
	"fmt"
	"strings"

	"github.com/githubnext/gh-aw/pkg/console"
	"github.com/githubnext/gh-aw/pkg/parser"
)

// FrontmatterValidationError represents a validation error with position information
type FrontmatterValidationError struct {
	Path    string  // JSONPath to the problematic field
	Message string  // Error message
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
	frontmatterYAML string // Original YAML content for span calculation
}

// NewFrontmatterValidator creates a new validator for the given frontmatter YAML
func NewFrontmatterValidator(frontmatterYAML string) *FrontmatterValidator {
	return &FrontmatterValidator{
		frontmatterYAML: frontmatterYAML,
	}
}

// ValidateFrontmatter performs validation on frontmatter data and returns errors with spans
func (v *FrontmatterValidator) ValidateFrontmatter(frontmatter map[string]any) []FrontmatterValidationError {
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
				span, err := parser.LocateFrontmatterPathSpan(v.frontmatterYAML, "engine")
				var spanPtr *parser.SourceSpan
				if err == nil {
					spanPtr = &span
				}
				
				errors = append(errors, FrontmatterValidationError{
					Path:    "engine",
					Message: fmt.Sprintf("unsupported engine '%s', must be one of: claude, codex", engineStr),
					Span:    spanPtr,
				})
			}
		}
	}

	// Example validation: check max-turns if present
	if maxTurns, exists := frontmatter["max-turns"]; exists {
		if maxTurnsInt, ok := maxTurns.(int); ok {
			if maxTurnsInt < 1 || maxTurnsInt > 100 {
				span, err := parser.LocateFrontmatterPathSpan(v.frontmatterYAML, "max-turns")
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
						span, err := parser.LocateFrontmatterPathSpan(v.frontmatterYAML, path)
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

// isValidEngine checks if the engine name is supported
func isValidEngine(engine string) bool {
	validEngines := []string{"claude", "codex"}
	for _, valid := range validEngines {
		if engine == valid {
			return true
		}
	}
	return false
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
		return "Supported engines: claude, codex"
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