package workflow

import (
	"fmt"

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

// ValidateFrontmatter performs validation on frontmatter data using JSON schema validation only
func (v *FrontmatterValidator) ValidateFrontmatter(frontmatter map[string]any) []FrontmatterValidationError {
	// Run JSON schema validation using the parser package
	err := parser.ValidateMainWorkflowFrontmatterWithSchema(frontmatter)
	if err != nil {
		// Create a simple validation error from the JSON schema error
		return []FrontmatterValidationError{{
			Path:    "",
			Message: err.Error(),
			Span:    nil,
		}}
	}
	
	return nil
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
			Hint:     "",  // Simplified - no hints for simpler implementation
		}

		compilerErrors = append(compilerErrors, compilerError)
	}

	return compilerErrors
}


