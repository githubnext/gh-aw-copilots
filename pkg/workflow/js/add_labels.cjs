async function main() {
  // Read the agent output content from environment variable
  const outputContent = process.env.GITHUB_AW_AGENT_OUTPUT;
  if (!outputContent) {
    console.log('No GITHUB_AW_AGENT_OUTPUT environment variable found');
    return;
  }

  if (outputContent.trim() === '') {
    console.log('Agent output content is empty');
    return;
  }

  console.log('Agent output content length:', outputContent.length);

  // Read the allowed labels from environment variable (mandatory)
  const allowedLabelsEnv = process.env.GITHUB_AW_LABELS_ALLOWED;
  if (!allowedLabelsEnv) {
    core.setFailed('GITHUB_AW_LABELS_ALLOWED environment variable is required but missing');
    return;
  }

  const allowedLabels = allowedLabelsEnv.split(',').map(label => label.trim()).filter(label => label);
  if (allowedLabels.length === 0) {
    core.setFailed('Allowed labels list is empty. At least one allowed label must be specified');
    return;
  }

  console.log('Allowed labels:', allowedLabels);

  // Check if we're in an issue or pull request context
  const isIssueContext = context.eventName === 'issues' || context.eventName === 'issue_comment';
  const isPRContext = context.eventName === 'pull_request' || context.eventName === 'pull_request_review' || context.eventName === 'pull_request_review_comment';

  if (!isIssueContext && !isPRContext) {
    console.log('Not running in issue or pull request context, skipping label addition');
    return;
  }

  // Determine the issue/PR number
  let issueNumber;
  let contextType;

  if (isIssueContext) {
    if (context.payload.issue) {
      issueNumber = context.payload.issue.number;
      contextType = 'issue';
    } else {
      console.log('Issue context detected but no issue found in payload');
      return;
    }
  } else if (isPRContext) {
    if (context.payload.pull_request) {
      issueNumber = context.payload.pull_request.number;
      contextType = 'pull request';
    } else {
      console.log('Pull request context detected but no pull request found in payload');
      return;
    }
  }

  if (!issueNumber) {
    console.log('Could not determine issue or pull request number');
    return;
  }

  // Parse labels from agent output (one per line, ignore empty lines)
  const lines = outputContent.split('\n');
  const requestedLabels = [];

  for (const line of lines) {
    const trimmedLine = line.trim();
    
    // Skip empty lines
    if (trimmedLine === '') {
      continue;
    }

    // Reject lines that start with '-' (removal indication)
    if (trimmedLine.startsWith('-')) {
      core.setFailed(`Label removal is not permitted. Found line starting with '-': ${trimmedLine}`);
      return;
    }

    requestedLabels.push(trimmedLine);
  }

  console.log('Requested labels:', requestedLabels);

  // Validate that all requested labels are in the allowed list
  const invalidLabels = requestedLabels.filter(label => !allowedLabels.includes(label));
  if (invalidLabels.length > 0) {
    core.setFailed(`The following labels are not in the allowed list: ${invalidLabels.join(', ')}. Allowed labels: ${allowedLabels.join(', ')}`);
    return;
  }

  // Remove duplicates from requested labels
  const uniqueLabels = [...new Set(requestedLabels)];

  if (uniqueLabels.length === 0) {
    console.log('No labels to add');
    core.setOutput('labels_added', '');
    await core.summary.addRaw(`
## Label Addition

No labels were added (no valid labels found in agent output).
`).write();
    return;
  }

  console.log(`Adding labels to ${contextType} #${issueNumber}:`, uniqueLabels);

  try {
    // Add labels using GitHub API
    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      labels: uniqueLabels
    });

    console.log(`Successfully added ${uniqueLabels.length} labels to ${contextType} #${issueNumber}`);

    // Set output for other jobs to use
    core.setOutput('labels_added', uniqueLabels.join('\n'));

    // Write summary
    const labelsListMarkdown = uniqueLabels.map(label => `- \`${label}\``).join('\n');
    await core.summary.addRaw(`
## Label Addition

Successfully added ${uniqueLabels.length} label(s) to ${contextType} #${issueNumber}:

${labelsListMarkdown}
`).write();

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to add labels:', errorMessage);
    core.setFailed(`Failed to add labels: ${errorMessage}`);
  }
}
await main();