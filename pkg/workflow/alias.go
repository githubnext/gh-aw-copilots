package workflow

import "fmt"

// buildAliasOnlyCondition creates a condition that only applies to alias mentions in comment-related events
// Unlike buildEventAwareAliasCondition, this does NOT allow non-comment events to pass through
func buildAliasOnlyCondition(aliasName string) ConditionNode {
	// Define the alias condition using proper expression nodes
	aliasText := fmt.Sprintf("@%s", aliasName)

	// Build alias checks for different content sources using expression nodes
	issueBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.issue.body"),
		BuildStringLiteral(aliasText),
	)
	commentBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.comment.body"),
		BuildStringLiteral(aliasText),
	)
	prBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.pull_request.body"),
		BuildStringLiteral(aliasText),
	)

	// Combine all alias checks with OR - only true when alias is mentioned
	return &DisjunctionNode{
		Terms: []ConditionNode{
			issueBodyCheck,
			commentBodyCheck,
			prBodyCheck,
		},
	}
}

// buildEventAwareAliasCondition creates a condition that only applies alias checks to comment-related events
func buildEventAwareAliasCondition(aliasName string, hasOtherEvents bool) ConditionNode {
	// Define the alias condition using proper expression nodes
	aliasText := fmt.Sprintf("@%s", aliasName)

	// Build alias checks for different content sources using expression nodes
	issueBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.issue.body"),
		BuildStringLiteral(aliasText),
	)
	commentBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.comment.body"),
		BuildStringLiteral(aliasText),
	)
	prBodyCheck := BuildContains(
		BuildPropertyAccess("github.event.pull_request.body"),
		BuildStringLiteral(aliasText),
	)

	// Combine all alias checks with OR
	aliasCondition := &OrNode{
		Left: &OrNode{
			Left:  issueBodyCheck,
			Right: commentBodyCheck,
		},
		Right: prBodyCheck,
	}

	if !hasOtherEvents {
		// If there are no other events, just use the simple alias condition
		return aliasCondition
	}

	// Define which events should be checked for alias using expression nodes
	commentEventChecks := &DisjunctionNode{
		Terms: []ConditionNode{
			BuildEventTypeEquals("issues"),
			BuildEventTypeEquals("issue_comment"),
			BuildEventTypeEquals("pull_request"),
			BuildEventTypeEquals("pull_request_review_comment"),
		},
	}

	// For comment events: check alias; for other events: allow unconditionally
	commentEventCheck := &AndNode{
		Left:  commentEventChecks,
		Right: aliasCondition,
	}

	// Allow all non-comment events to run
	nonCommentEvents := &NotNode{Child: commentEventChecks}

	// Combine: (comment events && alias check) || (non-comment events)
	return &OrNode{
		Left:  commentEventCheck,
		Right: nonCommentEvents,
	}
}
