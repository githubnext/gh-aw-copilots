package parser

import (
	"fmt"
	"strings"
)

// ExtractYAMLError extracts line and column information from YAML parsing errors
func ExtractYAMLError(err error, frontmatterStartLine int) (line int, column int, message string) {
	errStr := err.Error()

	// Parse "yaml: line X: column Y: message" format (enhanced parsers that provide column info)
	if strings.Contains(errStr, "yaml: line ") && strings.Contains(errStr, "column ") {
		parts := strings.SplitN(errStr, "yaml: line ", 2)
		if len(parts) > 1 {
			lineInfo := parts[1]

			// Look for column information
			colonIndex := strings.Index(lineInfo, ":")
			if colonIndex > 0 {
				lineStr := lineInfo[:colonIndex]

				// Parse line number
				if _, parseErr := fmt.Sscanf(lineStr, "%d", &line); parseErr == nil {
					// Look for column part
					remaining := lineInfo[colonIndex+1:]
					if strings.Contains(remaining, "column ") {
						columnParts := strings.SplitN(remaining, "column ", 2)
						if len(columnParts) > 1 {
							columnInfo := columnParts[1]
							colonIndex2 := strings.Index(columnInfo, ":")
							if colonIndex2 > 0 {
								columnStr := columnInfo[:colonIndex2]
								message = strings.TrimSpace(columnInfo[colonIndex2+1:])

								// Parse column number
								if _, parseErr := fmt.Sscanf(columnStr, "%d", &column); parseErr == nil {
									// Adjust line number to account for frontmatter position in file
									line += frontmatterStartLine
									return
								}
							}
						}
					}
				}
			}
		}
	}

	// Parse "yaml: line X: message" format (standard format without column info)
	if strings.Contains(errStr, "yaml: line ") {
		parts := strings.SplitN(errStr, "yaml: line ", 2)
		if len(parts) > 1 {
			lineInfo := parts[1]
			colonIndex := strings.Index(lineInfo, ":")
			if colonIndex > 0 {
				lineStr := lineInfo[:colonIndex]
				message = strings.TrimSpace(lineInfo[colonIndex+1:])

				// Parse line number
				if _, parseErr := fmt.Sscanf(lineStr, "%d", &line); parseErr == nil {
					// Adjust line number to account for frontmatter position in file
					line += frontmatterStartLine
					column = 1 // Default to column 1 when not provided
					return
				}
			}
		}
	}

	// Parse "yaml: unmarshal errors: line X: message" format (multiline errors)
	if strings.Contains(errStr, "yaml: unmarshal errors:") && strings.Contains(errStr, "line ") {
		lines := strings.Split(errStr, "\n")
		for _, errorLine := range lines {
			errorLine = strings.TrimSpace(errorLine)
			if strings.Contains(errorLine, "line ") && strings.Contains(errorLine, ":") {
				// Extract the first line number found in the error
				parts := strings.SplitN(errorLine, "line ", 2)
				if len(parts) > 1 {
					colonIndex := strings.Index(parts[1], ":")
					if colonIndex > 0 {
						lineStr := parts[1][:colonIndex]
						restOfMessage := strings.TrimSpace(parts[1][colonIndex+1:])

						// Parse line number
						if _, parseErr := fmt.Sscanf(lineStr, "%d", &line); parseErr == nil {
							// Adjust line number to account for frontmatter position in file
							line += frontmatterStartLine
							column = 1
							message = restOfMessage
							return
						}
					}
				}
			}
		}
	}

	// Fallback: return original error message
	return 0, 0, errStr
}
