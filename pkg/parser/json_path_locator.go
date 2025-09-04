package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// JSONPathLocation represents a location in YAML source corresponding to a JSON path
type JSONPathLocation struct {
	Line   int
	Column int
	Found  bool
}

// ExtractJSONPathFromValidationError extracts JSON path information from jsonschema validation errors
func ExtractJSONPathFromValidationError(err error) []JSONPathInfo {
	var paths []JSONPathInfo

	if validationError, ok := err.(*jsonschema.ValidationError); ok {
		// Process each cause (individual validation error)
		for _, cause := range validationError.Causes {
			path := JSONPathInfo{
				Path:     convertInstanceLocationToJSONPath(cause.InstanceLocation),
				Message:  cause.Error(),
				Location: cause.InstanceLocation,
			}
			paths = append(paths, path)
		}
	}

	return paths
}

// JSONPathInfo holds information about a validation error and its path
type JSONPathInfo struct {
	Path     string   // JSON path like "/tools/1" or "/age"
	Message  string   // Error message
	Location []string // Instance location from jsonschema (e.g., ["tools", "1"])
}

// convertInstanceLocationToJSONPath converts jsonschema InstanceLocation to JSON path string
func convertInstanceLocationToJSONPath(location []string) string {
	if len(location) == 0 {
		return ""
	}

	var parts []string
	for _, part := range location {
		parts = append(parts, "/"+part)
	}
	return strings.Join(parts, "")
}

// LocateJSONPathInYAML finds the line/column position of a JSON path in YAML source
func LocateJSONPathInYAML(yamlContent string, jsonPath string) JSONPathLocation {
	if jsonPath == "" {
		// Root level error - return start of content
		return JSONPathLocation{Line: 1, Column: 1, Found: true}
	}

	// Parse the path segments
	pathSegments := parseJSONPath(jsonPath)
	if len(pathSegments) == 0 {
		return JSONPathLocation{Line: 1, Column: 1, Found: true}
	}

	// Use a more sophisticated line-by-line approach to find the path
	location := findPathInYAMLLines(yamlContent, pathSegments)
	return location
}

// LocateJSONPathInYAMLWithAdditionalProperties finds the line/column position of a JSON path in YAML source
// with special handling for additional properties errors
func LocateJSONPathInYAMLWithAdditionalProperties(yamlContent string, jsonPath string, errorMessage string) JSONPathLocation {
	if jsonPath == "" {
		// This might be an additional properties error - try to extract property names
		propertyNames := extractAdditionalPropertyNames(errorMessage)
		if len(propertyNames) > 0 {
			// Find the first additional property in the YAML
			return findFirstAdditionalProperty(yamlContent, propertyNames)
		}
		// Fallback to root level error
		return JSONPathLocation{Line: 1, Column: 1, Found: true}
	}

	// For non-empty paths, use the regular logic
	return LocateJSONPathInYAML(yamlContent, jsonPath)
}

// findPathInYAMLLines finds a JSON path in YAML content using line-by-line analysis
func findPathInYAMLLines(yamlContent string, pathSegments []PathSegment) JSONPathLocation {
	lines := strings.Split(yamlContent, "\n")

	// Start from the beginning
	currentLevel := 0
	arrayContexts := make(map[int]int) // level -> current array index

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-based line numbers
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Calculate indentation level
		lineLevel := (len(line) - len(strings.TrimLeft(line, " \t"))) / 2

		// Check if this line matches our path
		matches, column := matchesPathAtLevel(line, pathSegments, lineLevel, arrayContexts)
		if matches {
			return JSONPathLocation{Line: lineNumber, Column: column, Found: true}
		}

		// Update array contexts for list items
		if strings.HasPrefix(trimmedLine, "-") {
			arrayContexts[lineLevel]++
		} else if lineLevel <= currentLevel {
			// Reset array contexts for deeper levels when we move to a shallower level
			for level := lineLevel + 1; level <= currentLevel; level++ {
				delete(arrayContexts, level)
			}
		}

		currentLevel = lineLevel
	}

	return JSONPathLocation{Line: 1, Column: 1, Found: false}
}

// matchesPathAtLevel checks if a line matches the target path at the current level
func matchesPathAtLevel(line string, pathSegments []PathSegment, level int, arrayContexts map[int]int) (bool, int) {
	if len(pathSegments) == 0 {
		return false, 0
	}

	trimmedLine := strings.TrimSpace(line)

	// For now, implement a simple key matching approach
	// This is a simplified version - in a full implementation we'd need to track
	// the complete path context as we traverse the YAML

	if level < len(pathSegments) {
		segment := pathSegments[level]

		if segment.Type == "key" {
			// Look for "key:" pattern
			keyPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(segment.Value) + `\s*:`)
			if keyPattern.MatchString(trimmedLine) {
				// Found the key - return position after the colon
				colonIndex := strings.Index(line, ":")
				if colonIndex != -1 {
					return level == len(pathSegments)-1, colonIndex + 2
				}
			}
		} else if segment.Type == "index" {
			// For array elements, check if this is a list item at the right index
			if strings.HasPrefix(trimmedLine, "-") {
				currentIndex := arrayContexts[level]
				if currentIndex == segment.Index {
					return level == len(pathSegments)-1, strings.Index(line, "-") + 2
				}
			}
		}
	}

	return false, 0
}

// parseJSONPath parses a JSON path string into segments
func parseJSONPath(path string) []PathSegment {
	if path == "" || path == "/" {
		return []PathSegment{}
	}

	// Remove leading slash and split by slash
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	var segments []PathSegment
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Check if this is an array index
		if index, err := strconv.Atoi(part); err == nil {
			segments = append(segments, PathSegment{Type: "index", Value: part, Index: index})
		} else {
			segments = append(segments, PathSegment{Type: "key", Value: part})
		}
	}

	return segments
}

// PathSegment represents a segment in a JSON path
type PathSegment struct {
	Type  string // "key" or "index"
	Value string // The raw value
	Index int    // Parsed index for array elements
}

// extractAdditionalPropertyNames extracts property names from additional properties error messages
// Example: "additional properties 'invalid_prop', 'another_invalid' not allowed" -> ["invalid_prop", "another_invalid"]
func extractAdditionalPropertyNames(errorMessage string) []string {
	// Look for the pattern: additional properties ... not allowed
	// Use regex to match the full property list section
	re := regexp.MustCompile(`additional propert(?:y|ies) (.+?) not allowed`)
	match := re.FindStringSubmatch(errorMessage)

	if len(match) < 2 {
		return []string{}
	}

	// Extract all quoted property names from the matched string
	propPattern := regexp.MustCompile(`'([^']+)'`)
	propMatches := propPattern.FindAllStringSubmatch(match[1], -1)

	var properties []string
	for _, propMatch := range propMatches {
		if len(propMatch) > 1 {
			prop := strings.TrimSpace(propMatch[1])
			if prop != "" {
				properties = append(properties, prop)
			}
		}
	}

	return properties
}

// findFirstAdditionalProperty finds the first occurrence of any of the given property names in YAML
func findFirstAdditionalProperty(yamlContent string, propertyNames []string) JSONPathLocation {
	lines := strings.Split(yamlContent, "\n")

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Check if this line contains any of the additional properties
		for _, propName := range propertyNames {
			// Look for "propName:" pattern at the start of the trimmed line
			keyPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(propName) + `\s*:`)
			if keyPattern.MatchString(trimmedLine) {
				// Found the property - return position of the property name
				propIndex := strings.Index(line, propName)
				if propIndex != -1 {
					return JSONPathLocation{
						Line:   lineNum + 1,   // 1-based line numbers
						Column: propIndex + 1, // 1-based column numbers
						Found:  true,
					}
				}
			}
		}
	}

	// If we can't find any of the properties, return the default location
	return JSONPathLocation{Line: 1, Column: 1, Found: false}
}
