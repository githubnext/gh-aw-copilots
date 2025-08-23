package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// SourceSpan represents a range in source code with start and end positions
type SourceSpan struct {
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// FrontmatterLocator provides cached access to frontmatter YAML for efficient
// multiple path lookups. It parses the YAML once and reuses the AST.
type FrontmatterLocator struct {
	frontmatterYAML string
	root            ast.Node
	parseError      error
}

// NewFrontmatterLocator creates a new locator with cached YAML parsing
func NewFrontmatterLocator(frontmatterYAML string) *FrontmatterLocator {
	locator := &FrontmatterLocator{
		frontmatterYAML: frontmatterYAML,
	}

	// Parse YAML once and cache the result
	if frontmatterYAML != "" {
		file, err := parser.ParseBytes([]byte(frontmatterYAML), 0)
		if err != nil {
			locator.parseError = fmt.Errorf("failed to parse YAML: %w", err)
		} else if file == nil || len(file.Docs) == 0 {
			locator.parseError = fmt.Errorf("no YAML documents found")
		} else {
			locator.root = file.Docs[0]
		}
	} else {
		locator.parseError = fmt.Errorf("frontmatter YAML is empty")
	}

	return locator
}

// LocatePathSpan finds the source span for the given JSONPath using cached AST
func (l *FrontmatterLocator) LocatePathSpan(jsonPath string) (SourceSpan, error) {
	if l.parseError != nil {
		return SourceSpan{}, l.parseError
	}

	if jsonPath == "" {
		return SourceSpan{}, fmt.Errorf("JSONPath is empty")
	}

	// Normalize the JSONPath
	normalizedPath := normalizeJSONPath(jsonPath)

	// Navigate to the target node using cached AST
	node, err := navigateToNode(l.root, normalizedPath)
	if err != nil {
		return SourceSpan{}, fmt.Errorf("path not found: %w", err)
	}

	// Calculate source span for the node
	span := calculateNodeSpan(node)
	return span, nil
}

// LocatePath provides backward compatibility by returning only start position
func (l *FrontmatterLocator) LocatePath(jsonPath string) (line int, column int, err error) {
	span, err := l.LocatePathSpan(jsonPath)
	if err != nil {
		return 0, 0, err
	}
	return span.StartLine, span.StartColumn, nil
}

// LocateFrontmatterPathSpan locates the source span for a given JSONPath in frontmatter YAML
// frontmatterYAML should be the raw YAML content (without the --- delimiters)
// jsonPath should be a JSONPath-like expression (e.g., "on.push", "jobs.build.steps[0].run")
func LocateFrontmatterPathSpan(frontmatterYAML, jsonPath string) (SourceSpan, error) {
	// Use the cached locator for single lookups
	locator := NewFrontmatterLocator(frontmatterYAML)
	return locator.LocatePathSpan(jsonPath)
}

// LocateFrontmatterPath provides backward compatibility by returning only start position
// This is a legacy function that delegates to LocateFrontmatterPathSpan
func LocateFrontmatterPath(frontmatterYAML, jsonPath string) (line int, column int, err error) {
	span, err := LocateFrontmatterPathSpan(frontmatterYAML, jsonPath)
	if err != nil {
		return 0, 0, err
	}
	return span.StartLine, span.StartColumn, nil
}

// normalizeJSONPath converts various JSONPath formats to a standard form
// Supports: "on.push", "$.on.push", "on[push]", "jobs.build.steps[0].run"
func normalizeJSONPath(path string) []string {
	// Remove leading $. if present
	path = strings.TrimPrefix(path, "$.")
	// Remove leading $ if present (just dollar)
	path = strings.TrimPrefix(path, "$")

	// Handle empty path
	if path == "" {
		return []string{}
	}

	var parts []string

	// Use regex to split on dots and extract array indices
	// This handles: "jobs.build.steps[0].run" -> ["jobs", "build", "steps", "[0]", "run"]
	re := regexp.MustCompile(`([^.\[\]]+)|\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		if match[1] != "" {
			// Regular property name
			parts = append(parts, match[1])
		} else if match[2] != "" {
			// Array index
			parts = append(parts, "["+match[2]+"]")
		}
	}

	return parts
}

// navigateToNode traverses the AST to find the node at the given path
func navigateToNode(root ast.Node, pathParts []string) (ast.Node, error) {
	current := root

	for i, part := range pathParts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Array index
			indexStr := part[1 : len(part)-1]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s' at part %d", indexStr, i)
			}

			// Navigate to array element
			current, err = navigateToArrayElement(current, index)
			if err != nil {
				return nil, fmt.Errorf("failed to navigate to array index %d at part %d: %w", index, i, err)
			}
		} else {
			// Object key
			var err error
			current, err = navigateToObjectKey(current, part)
			if err != nil {
				return nil, fmt.Errorf("failed to navigate to key '%s' at part %d: %w", part, i, err)
			}
		}
	}

	return current, nil
}

// navigateToObjectKey finds a key in a mapping node
func navigateToObjectKey(node ast.Node, key string) (ast.Node, error) {
	switch n := node.(type) {
	case *ast.DocumentNode:
		return navigateToObjectKey(n.Body, key)
	case *ast.MappingNode:
		for _, value := range n.Values {
			if value.Key != nil {
				keyStr := getNodeStringValue(value.Key)
				if keyStr == key {
					return value.Value, nil
				}
			}
		}
		return nil, fmt.Errorf("key '%s' not found in mapping", key)
	case *ast.AnchorNode:
		// For anchor nodes, navigate to the actual value
		return navigateToObjectKey(n.Value, key)
	default:
		return nil, fmt.Errorf("expected mapping node, got %T", node)
	}
}

// navigateToArrayElement finds an element in a sequence node
func navigateToArrayElement(node ast.Node, index int) (ast.Node, error) {
	switch n := node.(type) {
	case *ast.DocumentNode:
		return navigateToArrayElement(n.Body, index)
	case *ast.SequenceNode:
		if index < 0 || index >= len(n.Values) {
			return nil, fmt.Errorf("array index %d out of range (length: %d)", index, len(n.Values))
		}
		return n.Values[index], nil
	case *ast.AnchorNode:
		// For anchor nodes, navigate to the actual value
		return navigateToArrayElement(n.Value, index)
	default:
		return nil, fmt.Errorf("expected sequence node, got %T", node)
	}
}

// getNodeStringValue extracts string value from various node types
func getNodeStringValue(node ast.Node) string {
	switch n := node.(type) {
	case *ast.StringNode:
		return n.Value
	case *ast.LiteralNode:
		if n.Value != nil {
			return n.Value.Value
		}
		return ""
	case *ast.IntegerNode:
		return fmt.Sprintf("%d", n.Value)
	case *ast.FloatNode:
		return fmt.Sprintf("%g", n.Value)
	case *ast.BoolNode:
		return fmt.Sprintf("%t", n.Value)
	case *ast.NullNode:
		return "null"
	default:
		// Fallback to token value if available
		if node.GetToken() != nil {
			return node.GetToken().Value
		}
		return ""
	}
}

// calculateNodeSpan determines the source span for a given AST node
func calculateNodeSpan(node ast.Node) SourceSpan {
	if node == nil {
		return SourceSpan{}
	}

	// Get the token for position information
	tok := node.GetToken()
	if tok == nil {
		return SourceSpan{}
	}

	startLine := tok.Position.Line
	startColumn := tok.Position.Column
	endLine := startLine
	endColumn := startColumn

	// For different node types, calculate appropriate end positions
	switch n := node.(type) {
	case *ast.StringNode:
		endColumn = calculateStringNodeEnd(tok)
	case *ast.LiteralNode:
		endLine, endColumn = calculateLiteralNodeEnd(n, tok)
	case *ast.MappingNode:
		endLine, endColumn = calculateMappingNodeEnd(n, tok)
	case *ast.SequenceNode:
		endLine, endColumn = calculateSequenceNodeEnd(n, tok)
	case *ast.AnchorNode:
		// For anchor nodes, return the span of the anchor definition itself
		endColumn = calculateStringNodeEnd(tok)
	default:
		// For other node types, use token value length
		if tok.Value != "" {
			endColumn = startColumn + len(tok.Value) - 1
		}
	}

	return SourceSpan{
		StartLine:   startLine,
		StartColumn: startColumn,
		EndLine:     endLine,
		EndColumn:   endColumn,
	}
}

// calculateStringNodeEnd calculates end position for string nodes
func calculateStringNodeEnd(tok *token.Token) int {
	if tok.Value == "" {
		return tok.Position.Column
	}
	return tok.Position.Column + len(tok.Value) - 1
}

// calculateLiteralNodeEnd calculates end position for literal nodes (including multi-line)
func calculateLiteralNodeEnd(node *ast.LiteralNode, tok *token.Token) (int, int) {
	// For multi-line literals, we need to look at the original source structure
	// The token points to the | or > indicator
	if tok.Type == token.LiteralType && (strings.Contains(tok.Value, "|") || strings.Contains(tok.Value, ">")) {
		// This is a multi-line literal indicator
		// For now, let's calculate based on the content lines
		var content string
		if node.Value != nil {
			content = node.Value.Value
		}

		// Count actual newlines in the content to estimate span
		lines := strings.Split(content, "\n")
		// Remove empty trailing line if present (common in YAML literals)
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		if len(lines) > 1 {
			// Multi-line literal spans from the indicator to the last content line
			// Calculate approximate end line based on content lines
			endLine := tok.Position.Line + len(lines) - 1
			// For end column, use the length of the last line of content
			endColumn := len(lines[len(lines)-1])
			if endColumn == 0 {
				endColumn = 1 // Minimum column
			}
			return endLine, endColumn
		}
	}

	// Single line literal or other cases
	endColumn := tok.Position.Column + len(tok.Value) - 1
	return tok.Position.Line, endColumn
}

// calculateMappingNodeEnd calculates end position for mapping nodes
func calculateMappingNodeEnd(node *ast.MappingNode, tok *token.Token) (int, int) {
	if len(node.Values) == 0 {
		return tok.Position.Line, tok.Position.Column
	}

	// Find the last value in the mapping
	lastValue := node.Values[len(node.Values)-1]
	if lastValue.Value != nil {
		lastSpan := calculateNodeSpan(lastValue.Value)
		return lastSpan.EndLine, lastSpan.EndColumn
	}

	return tok.Position.Line, tok.Position.Column
}

// calculateSequenceNodeEnd calculates end position for sequence nodes
func calculateSequenceNodeEnd(node *ast.SequenceNode, tok *token.Token) (int, int) {
	if len(node.Values) == 0 {
		return tok.Position.Line, tok.Position.Column
	}

	// Find the last value in the sequence
	lastValue := node.Values[len(node.Values)-1]
	if lastValue != nil {
		lastSpan := calculateNodeSpan(lastValue)
		return lastSpan.EndLine, lastSpan.EndColumn
	}

	return tok.Position.Line, tok.Position.Column
}
