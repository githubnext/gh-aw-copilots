async function main() {
  // Read inputs from environment variables
  const reaction = process.env.GITHUB_AW_REACTION || 'eyes';

  console.log('Reaction type:', reaction);

  // Validate reaction type
  const validReactions = ['+1', '-1', 'laugh', 'confused', 'heart', 'hooray', 'rocket', 'eyes'];
  if (!validReactions.includes(reaction)) {
    core.setFailed(`Invalid reaction type: ${reaction}. Valid reactions are: ${validReactions.join(', ')}`);
    return;
  }

  // Determine the API endpoint based on the event type
  let endpoint;
  const eventName = context.eventName;
  const owner = context.repo.owner;
  const repo = context.repo.repo;

  try {
    switch (eventName) {
      case 'issues':
        const issueNumber = context.payload?.issue?.number;
        if (!issueNumber) {
          core.setFailed('Issue number not found in event payload');
          return;
        }
        endpoint = `/repos/${owner}/${repo}/issues/${issueNumber}/reactions`;
        break;

      case 'issue_comment':
        const commentId = context.payload?.comment?.id;
        if (!commentId) {
          core.setFailed('Comment ID not found in event payload');
          return;
        }
        endpoint = `/repos/${owner}/${repo}/issues/comments/${commentId}/reactions`;
        break;

      case 'pull_request':
      case 'pull_request_target':
        const prNumber = context.payload?.pull_request?.number;
        if (!prNumber) {
          core.setFailed('Pull request number not found in event payload');
          return;
        }
        // PRs are "issues" for the reactions endpoint
        endpoint = `/repos/${owner}/${repo}/issues/${prNumber}/reactions`;
        break;

      case 'pull_request_review_comment':
        const reviewCommentId = context.payload?.comment?.id;
        if (!reviewCommentId) {
          core.setFailed('Review comment ID not found in event payload');
          return;
        }
        endpoint = `/repos/${owner}/${repo}/pulls/comments/${reviewCommentId}/reactions`;
        break;

      default:
        core.setFailed(`Unsupported event type: ${eventName}`);
        return;
    }

    console.log('API endpoint:', endpoint);

    await addReaction(endpoint, reaction);

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to add reaction:', errorMessage);
    core.setFailed(`Failed to add reaction: ${errorMessage}`);
  }
}

/**
 * Add a reaction to a GitHub issue, PR, or comment
 * @param {string} endpoint - The GitHub API endpoint to add the reaction to
 * @param {string} reaction - The reaction type to add
 */
async function addReaction(endpoint, reaction) {
  try {
    // Try to create the reaction
    const response = await github.request('POST ' + endpoint, {
      content: reaction,
      headers: {
        'Accept': 'application/vnd.github+json'
      }
    });

    const reactionId = response.data?.id;
    if (reactionId) {
      console.log(`Successfully added reaction: ${reaction} (id: ${reactionId})`);
      core.setOutput('reaction-id', reactionId.toString());
      return;
    }

    // If we couldn't get the ID from the create response, fall back to listing
    console.log('Could not get reaction ID from create response, falling back to list...');
    await fallbackToListReaction(endpoint, reaction);

  } catch (error) {
    // If creation failed (e.g., reaction already exists), try to find existing reaction
    console.log('Create reaction failed, trying to find existing reaction...');
    await fallbackToListReaction(endpoint, reaction);
  }
}

/**
 * Fallback to list reactions and find the bot's reaction
 * @param {string} endpoint - The GitHub API endpoint to list reactions
 * @param {string} reaction - The reaction type to find
 */
async function fallbackToListReaction(endpoint, reaction) {
  try {
    const response = await github.request('GET ' + endpoint, {
      headers: {
        'Accept': 'application/vnd.github+json'
      }
    });

    const reactions = response.data || [];
    const botReaction = reactions.find(/** @param {any} r */ r => 
      r.content === reaction && 
      r.user && 
      r.user.login === 'github-actions[bot]'
    );

    if (botReaction) {
      console.log(`Found existing reaction: ${reaction} (id: ${botReaction.id})`);
      core.setOutput('reaction-id', botReaction.id.toString());
    } else {
      console.warn('Warning: could not determine reaction id; cleanup will list/filter.');
      core.setOutput('reaction-id', '');
    }

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.warn(`Warning: could not list reactions: ${errorMessage}`);
    core.setOutput('reaction-id', '');
    // Rethrow the error so it can be caught by the main error handler
    throw error;
  }
}

await main();