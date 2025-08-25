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

// ValidateFrontmatter performs validation on frontmatter data using JSON schema validation only
func (v *FrontmatterValidator) ValidateFrontmatter(frontmatter map[string]any) []FrontmatterValidationError {
	// Run JSON schema validation using the parser package
	err := parser.ValidateMainWorkflowFrontmatterWithSchema(frontmatter)
	if err != nil {
		// Parse JSON schema error to extract validation errors with paths
		return v.parseJSONSchemaError(err)
	}

	return nil
}

// parseJSONSchemaError parses JSON schema validation errors and extracts path information
func (v *FrontmatterValidator) parseJSONSchemaError(err error) []FrontmatterValidationError {
	errorMsg := err.Error()
	
	// Extract path information from JSON schema error messages
	// Example error format: "at '/engine': value must be one of 'claude', 'codex'"
	paths := extractJSONPathsFromError(errorMsg)
	
	var validationErrors []FrontmatterValidationError
	
	if len(paths) > 0 {
		// Create individual errors for each path found
		for _, pathInfo := range paths {
			span := v.getSourceSpanForPath(pathInfo.path)
			validationErrors = append(validationErrors, FrontmatterValidationError{
				Path:    pathInfo.path,
				Message: pathInfo.message,
				Span:    span,
			})
		}
	} else {
		// Fallback: create a single error without specific path
		validationErrors = append(validationErrors, FrontmatterValidationError{
			Path:    "",
			Message: errorMsg,
			Span:    nil,
		})
	}
	
	return validationErrors
}

// pathErrorInfo represents extracted path and message from JSON schema error
type pathErrorInfo struct {
	path    string
	message string
}

// extractJSONPathsFromError extracts JSONPath and error message pairs from JSON schema errors
func extractJSONPathsFromError(errorMsg string) []pathErrorInfo {
	var pathErrors []pathErrorInfo
	
	// Parse the error message to extract individual path errors
	// Look for patterns like "- at '/path': message"
	lines := strings.Split(errorMsg, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for lines starting with "- at '/path':"
		if strings.HasPrefix(line, "- at '") {
			if endQuote := strings.Index(line[6:], "'"); endQuote != -1 {
				path := line[6 : 6+endQuote]
				
				// Extract message after the path
				messageStart := strings.Index(line, ": ")
				if messageStart != -1 {
					message := line[messageStart+2:]
					
					// Clean up the path (remove leading slash if present)
					if strings.HasPrefix(path, "/") {
						path = path[1:]
					}
					
					pathErrors = append(pathErrors, pathErrorInfo{
						path:    path,
						message: message,
					})
				}
			}
		}
	}
	
	// If no specific paths found, try to extract a general path from the error
	if len(pathErrors) == 0 {
		// Look for single path references like "at '/engine'"
		if strings.Contains(errorMsg, "at '/") {
			startIdx := strings.Index(errorMsg, "at '/") + 4
			if endIdx := strings.Index(errorMsg[startIdx:], "'"); endIdx != -1 {
				path := errorMsg[startIdx : startIdx+endIdx]
				if strings.HasPrefix(path, "/") {
					path = path[1:]
				}
				
				pathErrors = append(pathErrors, pathErrorInfo{
					path:    path,
					message: errorMsg,
				})
			}
		}
	}
	
	return pathErrors
}

// getSourceSpanForPath attempts to get source span for a given JSONPath
func (v *FrontmatterValidator) getSourceSpanForPath(jsonPath string) *parser.SourceSpan {
	if v.locator == nil || jsonPath == "" {
		return nil
	}
	
	// Try the original path first
	span, err := v.locator.LocatePathSpan(jsonPath)
	if err == nil {
		return &span
	}
	
	// If the original path fails, try converting slash notation to dot notation
	// JSON schema uses '/tools/github' but our locator expects 'tools.github'
	if strings.Contains(jsonPath, "/") {
		dotPath := strings.ReplaceAll(jsonPath, "/", ".")
		span, err := v.locator.LocatePathSpan(dotPath)
		if err == nil {
			return &span
		}
	}
	
	// If we still can't find the path, return nil
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
