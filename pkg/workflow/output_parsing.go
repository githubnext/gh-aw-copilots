package workflow

import "fmt"

// parseAllowHTMLField parses the allow-html field from a configuration map
// Returns a pointer to bool to distinguish between unset (nil) and explicitly false
func parseAllowHTMLField(configMap map[string]any) (*bool, error) {
	if allowHTML, exists := configMap["allow-html"]; exists {
		if allowHTMLBool, ok := allowHTML.(bool); ok {
			return &allowHTMLBool, nil
		}
		return nil, fmt.Errorf("allow-html must be a boolean")
	}
	return nil, nil // Not specified
}

// parseTitlePrefixField parses the title-prefix field from a configuration map
func parseTitlePrefixField(configMap map[string]any) (string, error) {
	if titlePrefix, exists := configMap["title-prefix"]; exists {
		if titlePrefixStr, ok := titlePrefix.(string); ok {
			return titlePrefixStr, nil
		}
		return "", fmt.Errorf("title-prefix must be a string")
	}
	return "", nil // Not specified
}

// parseLabelsField parses the labels field from a configuration map
func parseLabelsField(configMap map[string]any) ([]string, error) {
	if labels, exists := configMap["labels"]; exists {
		if labelsArray, ok := labels.([]any); ok {
			var labelStrings []string
			for _, label := range labelsArray {
				if labelStr, ok := label.(string); ok {
					labelStrings = append(labelStrings, labelStr)
				} else {
					return nil, fmt.Errorf("all labels must be strings")
				}
			}
			return labelStrings, nil
		}
		return nil, fmt.Errorf("labels must be an array")
	}
	return nil, nil // Not specified
}

// parseDraftField parses the draft field from a configuration map
func parseDraftField(configMap map[string]any) (*bool, error) {
	if draft, exists := configMap["draft"]; exists {
		if draftBool, ok := draft.(bool); ok {
			return &draftBool, nil
		}
		return nil, fmt.Errorf("draft must be a boolean")
	}
	return nil, nil // Not specified
}
