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

// LocateFrontmatterPathSpan locates the source span for a given JSONPath in frontmatter YAML
// frontmatterYAML should be the raw YAML content (without the --- delimiters)
// jsonPath should be a JSONPath-like expression (e.g., "on.push", "jobs.build.steps[0].run")
func LocateFrontmatterPathSpan(frontmatterYAML, jsonPath string) (SourceSpan, error) {
	if frontmatterYAML == "" {
		return SourceSpan{}, fmt.Errorf("frontmatter YAML is empty")
	}
	if jsonPath == "" {
		return SourceSpan{}, fmt.Errorf("JSONPath is empty")
	}

	// Parse YAML into AST
	file, err := parser.ParseBytes([]byte(frontmatterYAML), 0)
	if err != nil {
		return SourceSpan{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if file == nil || len(file.Docs) == 0 {
		return SourceSpan{}, fmt.Errorf("no YAML documents found")
	}

	root := file.Docs[0] // Use first document

	// Normalize the JSONPath
	normalizedPath := normalizeJSONPath(jsonPath)

	// Navigate to the target node
	node, err := navigateToNode(root, normalizedPath)
	if err != nil {
		return SourceSpan{}, fmt.Errorf("path not found: %w", err)
	}

	// Calculate source span for the node
	span := calculateNodeSpan(node)
	return span, nil
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
	if tok.Type == token.LiteralType && (strings.Contains(tok.Value, "|") || strings.Contains(tok.Value, ">")) {
		// Multi-line literal - get the actual content
		var content string
		if node.Value != nil {
			content = node.Value.Value
		}
		
		// Count line breaks in the value
		lines := strings.Split(content, "\n")
		if len(lines) > 1 {
			// Multi-line literal
			endLine := tok.Position.Line + len(lines) - 1
			endColumn := len(lines[len(lines)-1])
			return endLine, endColumn
		}
	}
	
	// Single line literal
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