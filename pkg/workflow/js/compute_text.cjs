/**
 * Computes the current body text based on the GitHub event context.
 * This extracts the relevant text content from different event types.
 */

const fs = require('fs');

try {
  const eventName = context.eventName;
  const payload = context.payload;
  let bodyText = '';

  console.log(`Computing text for event: ${eventName}`);

  switch (eventName) {
    case 'issues':
      // For issues, get the issue body
      if (payload.issue && payload.issue.body) {
        bodyText = payload.issue.body;
        console.log('Using issue body text');
      }
      break;

    case 'issue_comment':
      // For issue comments, get the comment body
      if (payload.comment && payload.comment.body) {
        bodyText = payload.comment.body;
        console.log('Using issue comment body text');
      } else if (payload.issue && payload.issue.body) {
        // Fallback to issue body if comment body is not available
        bodyText = payload.issue.body;
        console.log('Using issue body text as fallback');
      }
      break;

    case 'pull_request':
      // For pull requests, get the PR body
      if (payload.pull_request && payload.pull_request.body) {
        bodyText = payload.pull_request.body;
        console.log('Using pull request body text');
      }
      break;

    case 'pull_request_review_comment':
      // For PR review comments, get the comment body
      if (payload.comment && payload.comment.body) {
        bodyText = payload.comment.body;
        console.log('Using pull request review comment body text');
      } else if (payload.pull_request && payload.pull_request.body) {
        // Fallback to PR body if comment body is not available
        bodyText = payload.pull_request.body;
        console.log('Using pull request body text as fallback');
      }
      break;

    case 'push':
      // For push events, get the commit messages
      if (payload.commits && payload.commits.length > 0) {
        // @ts-ignore
        const commitMessages = payload.commits.map((commit) => commit.message).join('\n\n');
        bodyText = commitMessages;
        console.log('Using commit messages text');
      }
      break;

    case 'schedule':
    case 'workflow_dispatch':
      // For scheduled or manual triggers, use empty text
      bodyText = '';
      console.log('No specific text for scheduled/manual trigger');
      break;

    default:
      // For other events, try to find any available text content
      if (payload.comment && payload.comment.body) {
        bodyText = payload.comment.body;
        console.log('Using comment body text from unknown event');
      } else if (payload.issue && payload.issue.body) {
        bodyText = payload.issue.body;
        console.log('Using issue body text from unknown event');
      } else if (payload.pull_request && payload.pull_request.body) {
        bodyText = payload.pull_request.body;
        console.log('Using pull request body text from unknown event');
      } else {
        bodyText = '';
        console.log(`No text content found for event: ${eventName}`);
      }
      break;
  }

  // Ensure bodyText is a string and handle null/undefined
  if (typeof bodyText !== 'string') {
    bodyText = bodyText ? String(bodyText) : '';
  }

  console.log(`Computed text length: ${bodyText.length} characters`);

  // Set the output for GitHub Actions
  core.setOutput('text', bodyText);

  // Also log a summary
  const truncatedText = bodyText.length > 200 ? bodyText.substring(0, 200) + '...' : bodyText;
  console.log(`Computed text preview: ${truncatedText}`);

} catch (error) {
  console.error('Error computing text:', error);
  const errorMessage = error instanceof Error ? error.message : String(error);
  core.setFailed(`Failed to compute text: ${errorMessage}`);
}