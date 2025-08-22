// URL filtering function
function filterURLs(content, allowDomains) {
  if (!content || typeof content !== 'string') {
    return { filteredContent: '', removedURLs: [] };
  }
  
  let removedURLs = [];
  let filteredContent = content;
  
  // Default GitHub-owned domains when no domains are configured
  const defaultGitHubDomains = ['github.com', 'github.io', 'githubusercontent.com', 'githubassets.com', 'githubapp.com', 'github.dev'];
  
  // Helper function to determine if URL should be filtered
  function shouldFilterURL(rawURL) {
    try {
      const url = new URL(rawURL);
      
      // Always filter non-HTTPS URLs
      if (url.protocol !== 'https:') {
        return true;
      }
      
      // Use default GitHub domains if no domains are configured
      const domainsToCheck = (allowDomains && allowDomains.length > 0) ? allowDomains : defaultGitHubDomains;
      
      // Check if hostname matches any allowed domain pattern
      const hostname = url.hostname.toLowerCase();
      if (!hostname) {
        return true;
      }
      
      for (const allowedDomain of domainsToCheck) {
        const domain = allowedDomain.toLowerCase().trim();
        if (!domain) continue;
        
        // Exact match
        if (hostname === domain) {
          return false;
        }
        
        // Subdomain match
        if (hostname.endsWith('.' + domain)) {
          return false;
        }
      }
      
      return true;
    } catch (error) {
      // If we can't parse it, filter it out for safety
      return true;
    }
  }
  
  // Handle markdown links: [text](url)
  const markdownLinkRegex = /\[([^\]]*)\]\(([^)]+)\)/g;
  filteredContent = filteredContent.replace(markdownLinkRegex, (match, linkText, linkURL) => {
    if (shouldFilterURL(linkURL)) {
      removedURLs.push(linkURL);
      return linkText ? linkText + ' [filtered]' : '[filtered]';
    }
    return match;
  });
  
  // Handle plain URLs (including all protocols)
  const urlRegex = /[a-zA-Z][a-zA-Z0-9+.-]*:\/\/[^\s<>"'\[\]{}()]+/g;
  filteredContent = filteredContent.replace(urlRegex, (match) => {
    if (shouldFilterURL(match)) {
      removedURLs.push(match);
      return '[filtered]';
    }
    return match;
  });
  
  return { filteredContent, removedURLs };
}

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

  // Get allowed domains from environment variable
  const allowDomains = process.env.GH_AW_ALLOW_DOMAINS ? process.env.GH_AW_ALLOW_DOMAINS.split(',') : null;
  
  // Filter URLs in the content
  const urlFilterResult = filterURLs(outputContent, allowDomains);
  const filteredContent = urlFilterResult.filteredContent;
  if (urlFilterResult.removedURLs.length > 0) {
    console.log('Filtered URLs:', urlFilterResult.removedURLs);
  }
  // Check if we're in an issue context (triggered by an issue event)
  const parentIssueNumber = context.payload?.issue?.number;
  // Parse labels from environment variable (comma-separated string)
  const labelsEnv = process.env.GITHUB_AW_ISSUE_LABELS;
  const labels = labelsEnv ? labelsEnv.split(',').map(/** @param {string} label */ label => label.trim()).filter(/** @param {string} label */ label => label) : [];

  // Parse the output to extract title and body
  const lines = filteredContent.split('\n');
  let title = '';
  let bodyLines = [];
  let foundTitle = false;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();

    // Skip empty lines until we find the title
    if (!foundTitle && line === '') {
      continue;
    }

    // First non-empty line becomes the title
    if (!foundTitle && line !== '') {
      // Remove markdown heading syntax if present
      title = line.replace(/^#+\s*/, '').trim();
      foundTitle = true;
      continue;
    }

    // Everything else goes into the body
    if (foundTitle) {
      bodyLines.push(lines[i]); // Keep original formatting
    }
  }

  // If no title was found, use a default
  if (!title) {
    title = 'Agent Output';
  }

  // Apply title prefix if provided via environment variable
  const titlePrefix = process.env.GITHUB_AW_ISSUE_TITLE_PREFIX;
  if (titlePrefix && !title.startsWith(titlePrefix)) {
    title = titlePrefix + title;
  }

  if (parentIssueNumber) {
    console.log('Detected issue context, parent issue #' + parentIssueNumber);

    // Add reference to parent issue in the child issue body
    bodyLines.push(`Related to #${parentIssueNumber}`);
  }

  // Add AI disclaimer with run id, run htmlurl
  // Add AI disclaimer with workflow run information
  const runId = context.runId;
  const runUrl = `${context.payload.repository.html_url}/actions/runs/${runId}`;  
  bodyLines.push(``, ``, `> Generated by Agentic Workflow Run [${runId}](${runUrl})`, '');

  // Prepare the body content
  const body = bodyLines.join('\n').trim();

  console.log('Creating issue with title:', title);
  console.log('Labels:', labels);
  console.log('Body length:', body.length);


  // Create the issue using GitHub API
  const { data: issue } = await github.rest.issues.create({
    owner: context.repo.owner,
    repo: context.repo.repo,
    title: title,
    body: body,
    labels: labels
  });

  console.log('Created issue #' + issue.number + ': ' + issue.html_url);

  // If we have a parent issue, add a comment to it referencing the new child issue
  if (parentIssueNumber) {
    try {
      await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: parentIssueNumber,
        body: `Created related issue: #${issue.number}`
      });
      console.log('Added comment to parent issue #' + parentIssueNumber);
    } catch (error) {
      console.log('Warning: Could not add comment to parent issue:', error instanceof Error ? error.message : String(error));
    }
  }

  // Set output for other jobs to use
  core.setOutput('issue_number', issue.number);
  core.setOutput('issue_url', issue.html_url);
  // write issue to summary
  await core.summary.addRaw(`

## GitHub Issue
- Issue ID: ${issue.number}
- Issue URL: ${issue.html_url}
`).write();
}
await main();