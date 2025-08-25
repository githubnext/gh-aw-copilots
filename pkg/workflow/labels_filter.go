package workflow

import (
	"fmt"
	"strings"
)

// applyLabelFilter applies label name filter conditions for label triggers
func (c *Compiler) applyLabelFilter(data *WorkflowData, frontmatter map[string]any) {
	// Check if there's an "on" section in the frontmatter
	onValue, hasOn := frontmatter["on"]
	if !hasOn {
		return
	}

	// Check if "on" is an object (not a string)
	onMap, isOnMap := onValue.(map[string]any)
	if !isOnMap {
		return
	}

	// Check if there's a label section
	labelValue, hasLabel := onMap["label"]
	if !hasLabel {
		return
	}

	// Check if label is an object with name settings
	labelMap, isLabelMap := labelValue.(map[string]any)
	if !isLabelMap {
		return
	}

	// Check if name is specified
	nameValue, hasName := labelMap["name"]
	if !hasName {
		return
	}

	// Check if name is an array of strings
	nameArray, isNameArray := nameValue.([]any)
	if !isNameArray {
		return
	}

	// Convert to string array and validate
	var labelNames []string
	for _, name := range nameArray {
		if nameStr, ok := name.(string); ok {
			labelNames = append(labelNames, nameStr)
		}
	}

	// If no valid label names found, don't add filter
	if len(labelNames) == 0 {
		return
	}

	// Build label filter conditions using expression nodes
	var labelConditions []ConditionNode
	for _, labelName := range labelNames {
		labelCondition := BuildLabelContains(labelName)
		labelConditions = append(labelConditions, labelCondition)
	}

	// Combine all label conditions with OR logic
	var combinedLabelCondition ConditionNode
	if len(labelConditions) == 1 {
		combinedLabelCondition = labelConditions[0]
	} else {
		combinedLabelCondition = &DisjunctionNode{Terms: labelConditions}
	}

	// The condition should apply to label events on issues/PRs
	// Only filter when it's a label event, otherwise allow all events
	isLabelEvent := &OrNode{
		Left: BuildEquals(
			BuildPropertyAccess("github.event_name"),
			BuildStringLiteral("issues"),
		),
		Right: BuildEquals(
			BuildPropertyAccess("github.event_name"),
			BuildStringLiteral("pull_request"),
		),
	}

	// Check if it's a labeling action
	isLabelingAction := &OrNode{
		Left: BuildEquals(
			BuildPropertyAccess("github.event.action"),
			BuildStringLiteral("labeled"),
		),
		Right: BuildEquals(
			BuildPropertyAccess("github.event.action"),
			BuildStringLiteral("unlabeled"),
		),
	}

	// For label events with labeling actions, apply the label filter
	// For non-label events or non-labeling actions, allow through
	labelEventWithFilter := &AndNode{
		Left: &AndNode{
			Left:  isLabelEvent,
			Right: isLabelingAction,
		},
		Right: combinedLabelCondition,
	}

	notLabelEvent := &NotNode{Child: isLabelEvent}
	notLabelingAction := &NotNode{Child: isLabelingAction}

	finalCondition := &OrNode{
		Left: &OrNode{
			Left:  notLabelEvent,
			Right: notLabelingAction,
		},
		Right: labelEventWithFilter,
	}

	// Build condition tree and render
	existingCondition := strings.TrimPrefix(data.If, "if: ")
	conditionTree := buildConditionTree(existingCondition, finalCondition.Render())
	data.If = fmt.Sprintf("if: %s", conditionTree.Render())
}

// commentOutLabelNameInOnSection comments out name fields in label sections within the YAML string
// The name field is processed separately by applyLabelFilter and should be commented for documentation
func (c *Compiler) commentOutLabelNameInOnSection(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	var result []string
	inLabel := false

	for _, line := range lines {
		// Check if we're entering a label section
		if strings.Contains(line, "label:") {
			inLabel = true
			result = append(result, line)
			continue
		}

		// Check if we're leaving the label section (new top-level key or end of indent)
		if inLabel {
			// If line is not indented or is a new top-level key, we're out of label
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
				inLabel = false
			}
		}

		// If we're in label section and this line contains name:, comment it out
		if inLabel && strings.Contains(strings.TrimSpace(line), "name:") {
			// Preserve the original indentation and comment out the line
			indentation := ""
			trimmed := strings.TrimLeft(line, " \t")
			if len(line) > len(trimmed) {
				indentation = line[:len(line)-len(trimmed)]
			}

			commentedLine := indentation + "# " + trimmed + " # Label filtering applied via job conditions"
			result = append(result, commentedLine)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
