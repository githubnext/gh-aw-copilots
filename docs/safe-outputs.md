# 🔒 Safe Output Processing

One of the primary security features of GitHub Agentic Workflows is "safe output processing", enabling the creation of GitHub issues, comments, pull requests, and other outputs without giving the agentic portion of the workflow write permissions.

## Overview (`safe-outputs:`)

The `safe-outputs:` element of your workflow's frontmatter declares that your agentic workflow should conclude with optional automated actions based on the agentic workflow's output. This enables your workflow to write content that is then automatically processed to create GitHub issues, comments, pull requests, or add labels—all without giving the agentic portion of the workflow any write permissions.

**How It Works:**
1. The agentic part of your workflow runs with minimal read-only permissions. It is given additional prompting to write its output to the special known files
2. The compiler automatically generates additional jobs that read this output and perform the requested actions
3. Only these generated jobs receive the necessary write permissions

For example:

```yaml
safe-outputs:
  create-issue:
  create-discussion:
  add-issue-comment:
```

This declares that the workflow should create at most one new issue, at most one new discussion, and add at most one comment to the triggering issue or pull request based on the agentic workflow's output. To create multiple issues, discussions, or comments, use the `max` parameter.

## Available Output Types

### New Issue Creation (`create-issue:`)

Adding issue creation to the `safe-outputs:` section declares that the workflow should conclude with the creation of GitHub issues based on the workflow's output.

**Basic Configuration:**
```yaml
safe-outputs:
  create-issue:
```

**With Configuration:**
```yaml
safe-outputs:
  create-issue:
    title-prefix: "[ai] "            # Optional: prefix for issue titles
    labels: [automation, agentic]    # Optional: labels to attach to issues
    max: 5                           # Optional: maximum number of issues (default: 1)
```

The agentic part of your workflow should describe the issue(s) it wants created.

**Example markdown to generate the output:**

```yaml
# Code Analysis Agent

Analyze the latest commit and provide insights.
Create new issues with your findings. For each issue, provide a title starting with "AI Code Analysis" and detailed description of the analysis findings.
```

The compiled workflow will have additional prompting describing that, to create issues, it should write the issue details to a file.

### New Discussion Creation (`create-discussion:`)

Adding discussion creation to the `safe-outputs:` section declares that the workflow should conclude with the creation of GitHub discussions based on the workflow's output.

**Basic Configuration:**
```yaml
safe-outputs:
  create-discussion:
```

**With Configuration:**
```yaml
safe-outputs:
  create-discussion:
    title-prefix: "[ai] "            # Optional: prefix for discussion titles
    category-id: "DIC_kwDOGFsHUM4BsUn3"  # Optional: specific discussion category ID
    max: 3                           # Optional: maximum number of discussions (default: 1)
```

The agentic part of your workflow should describe the discussion(s) it wants created.

**Example markdown to generate the output:**

```yaml
# Research Discussion Agent

Research the latest developments in AI and create discussions to share findings.
Create new discussions with your research findings. For each discussion, provide a title starting with "AI Research Update" and detailed summary of the findings.
```

The compiled workflow will have additional prompting describing that, to create discussions, it should write the discussion details to a file.

**Note:** If no `category-id` is specified, the workflow will use the first available discussion category in the repository. Discussions require the `discussions: write` permission.

### Issue Comment Creation (`add-issue-comment:`)

Adding comment creation to the `safe-outputs:` section declares that the workflow should conclude with posting comments based on the workflow's output. By default, comments are posted on the triggering issue or pull request, but this can be configured using the `target` option.

**Basic Configuration:**
```yaml
safe-outputs:
  add-issue-comment:
```

**With Configuration:**
```yaml
safe-outputs:
  add-issue-comment:
    max: 3                          # Optional: maximum number of comments (default: 1)
    target: "*"                     # Optional: target for comments
                                    # "triggering" (default) - only comment on triggering issue/PR
                                    # "*" - allow comments on any issue (requires issue_number in agent output)
                                    # explicit number - comment on specific issue number
```

The agentic part of your workflow should describe the comment(s) it wants posted.

**Example natural language to generate the output:**

```markdown
# Issue/PR Analysis Agent

Analyze the issue or pull request and provide feedback.
Create issue comments on the triggering issue or PR with your analysis findings. Each comment should provide specific insights about different aspects of the issue.
```

The compiled workflow will have additional prompting describing that, to create comments, it should write the comment content to a special file.

### Pull Request Creation (`create-pull-request:`)

Adding pull request creation to the `safe-outputs:` section declares that the workflow should conclude with the creation of a pull request containing code changes generated by the workflow.

```yaml
safe-outputs:
  create-pull-request:
```

**With Configuration:**
```yaml
safe-outputs:
  create-pull-request:               # Creates exactly one pull request
    title-prefix: "[ai] "            # Optional: prefix for PR titles
    labels: [automation, agentic]    # Optional: labels to attach to PRs
    draft: true                      # Optional: create as draft PR (defaults to true)
```

At most one pull request is currently supported.

The agentic part of your workflow should instruct to:
1. **Make code changes**: Make any code changes in the working directory—these are automatically collected using `git add -A` and committed
2. **Create pull request**: Describe the pull request title and body content you want

**Example natural language to generate the output:**

```markdown
# Code Improvement Agent

Analyze the latest commit and suggest improvements.

1. Make any file changes directly in the working directory
2. Create a pull request for your improvements, with a descriptive title and detailed description of the changes made
```

### Label Addition (`add-issue-label:`)

Adding `add-issue-label:` to the `safe-outputs:` section of your workflow declares that the workflow should conclude with adding labels to the current issue or pull request based on the coding agent's analysis.

```yaml
safe-outputs:
  add-issue-label:
```

or with further configuration:

```yaml
safe-outputs:
  add-issue-label:
    allowed: [triage, bug, enhancement] # Optional: allowed labels for addition.
    max: 3                              # Optional: maximum number of labels to add (default: 3)
```

The agentic part of your workflow should analyze the issue content and determine appropriate labels. 

**Example of natural language to generate the output:**

```markdown
# Issue Labeling Agent

Analyze the issue content and add appropriate labels to the issue.
```

The agentic part of your workflow will have implicit additional prompting saying that, to add labels to a GitHub issue, you must write labels to a special file, one label per line.

### Issue Updates (`update-issue:`)

Adding `update-issue:` to the `safe-outputs:` section declares that the workflow should conclude with updating GitHub issues based on the coding agent's analysis. You can configure which fields are allowed to be updated.

**Basic Configuration:**
```yaml
safe-outputs:
  update-issue:
```

**With Configuration:**
```yaml
safe-outputs:
  update-issue:
    status:                             # Optional: presence indicates status can be updated (open/closed)
    target: "*"                         # Optional: target for updates
                                        # "triggering" (default) - only update triggering issue
                                        # "*" - allow updates to any issue (requires issue_number in agent output)
                                        # explicit number - update specific issue number
    title:                              # Optional: presence indicates title can be updated
    body:                               # Optional: presence indicates body can be updated
    max: 3                              # Optional: maximum number of issues to update (default: 1)
```

The agentic part of your workflow should analyze the issue and determine what updates to make.

**Example natural language to generate the output:**

```markdown
# Issue Update Agent

Analyze the issue and update its status, title, or body as needed.
Update the issue based on your analysis. You can change the title, body content, or status (open/closed).
```

**Safety Features:**

- Only explicitly enabled fields (`status`, `title`, `body`) can be updated
- Status values are validated (must be "open" or "closed")
- Empty or invalid field values are rejected
- Target configuration controls which issues can be updated for security
- Update count is limited by `max` setting (default: 1)
- Only GitHub's `issues.update` API endpoint is used

### Push to Branch (`push-to-branch:`)

Adding `push-to-branch:` to the `safe-outputs:` section declares that the workflow should conclude with pushing changes to a specific branch based on the agentic workflow's output. This is useful for applying code changes directly to a designated branch within pull requests.

**Basic Configuration:**
```yaml
safe-outputs:
  push-to-branch:
```

**With Configuration:**
```yaml
safe-outputs:
  push-to-branch:
    branch: feature-branch               # Optional: the branch to push changes to (default: "triggering")
    target: "*"                          # Optional: target for push operations
                                         # "triggering" (default) - only push in triggering PR context
                                         # "*" - allow pushes to any pull request (requires pull_request_number in agent output)
                                         # explicit number - push for specific pull request number
```

The agentic part of your workflow should describe the changes to be pushed and optionally provide a commit message.

**Example natural language to generate the output:**

```markdown
# Code Update Agent

Analyze the pull request and make necessary code improvements.

1. Make any file changes directly in the working directory  
2. Push changes to the feature branch with a descriptive commit message
```

**Safety Features:**

- Changes are applied via git patches generated from the workflow's modifications
- Only the specified branch can be modified
- Target configuration controls which pull requests can trigger pushes for security
- Push operations are limited to one per workflow execution
- Requires valid patch content to proceed (empty patches are rejected)

**Safety Features:**

- Empty lines in coding agent output are ignored
- Lines starting with `-` are rejected (no removal operations allowed)
- Duplicate labels are automatically removed
- If `allowed` is provided, all requested labels must be in the `allowed` list or the job fails with a clear error message. If `allowed` is not provided then any labels are allowed (including creating new labels).
- Label count is limited by `max` setting (default: 3) - exceeding this limit causes job failure
- Only GitHub's `issues.addLabels` API endpoint is used (no removal endpoints)

When `create-pull-request` or `push-to-branch` are enabled in the `safe-outputs` configuration, the system automatically adds the following additional Claude tools to enable file editing and pull request creation:

## Automatically Added Tools

When `create-pull-request` or `push-to-branch` are configured, these Claude tools are automatically added:

- **Edit**: Allows editing existing files
- **MultiEdit**: Allows making multiple edits to files in a single operation
- **Write**: Allows creating new files or overwriting existing files
- **NotebookEdit**: Allows editing Jupyter notebook files

Along with the file editing tools, these Git commands are also automatically whitelisted:

- `git checkout:*`
- `git branch:*`
- `git switch:*`
- `git add:*`
- `git rm:*`
- `git commit:*`
- `git merge:*`

## Security and Sanitization

All coding agent output is automatically sanitized for security before being processed:

- **XML Character Escaping**: Special characters (`<`, `>`, `&`, `"`, `'`) are escaped to prevent injection attacks
- **URI Protocol Filtering**: Only HTTPS URIs are allowed; other protocols (HTTP, FTP, file://, javascript:, etc.) are replaced with "(redacted)"
- **Domain Allowlisting**: HTTPS URIs are checked against the `allowed-domains` list. Unlisted domains are replaced with "(redacted)"
- **Default Allowed Domains**: When `allowed-domains` is not specified, safe GitHub domains are used by default:
  - `github.com`
  - `github.io`
  - `githubusercontent.com`
  - `githubassets.com`
  - `github.dev`
  - `codespaces.new`
- **Length and Line Limits**: Content is truncated if it exceeds safety limits (0.5MB or 65,000 lines)
- **Control Character Removal**: Non-printable characters and ANSI escape sequences are stripped

**Configuration:**

```yaml
safe-outputs:
  allowed-domains:                    # Optional: domains allowed in coding agent output URIs
    - github.com                      # Default GitHub domains are always included
    - api.github.com                  # Additional trusted domains can be specified
    - trusted-domain.com              # URIs from unlisted domains are replaced with "(redacted)"
```

## Related Documentation

- [Frontmatter Options](frontmatter.md) - All configuration options for workflows
- [Workflow Structure](workflow-structure.md) - Directory layout and organization
- [Command Triggers](command-triggers.md) - Special /mention triggers and context text
- [Commands](commands.md) - CLI commands for workflow management
