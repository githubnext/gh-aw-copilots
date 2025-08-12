## 📖 Reference

### 🛠️ Workflow Management

```bash
# List all available natural language action workflows
gh aw list

# List installed workflow packages
gh aw list --packages

# Create a new workflow from template
gh aw new my-workflow

# Create a new workflow with force overwrite
gh aw new issue-handler --force

# Add a new workflow from a sample
gh aw add samples/weekly-research.md -r githubnext/agentics

# Compile existing workflows after modifying the markdown file
gh aw compile

# Compile workflows with verbose output
gh aw compile --verbose

# Compile workflows and generate auto-compile workflow for automatic compilation
gh aw compile --auto-compile

# Compile workflows with both verbose output and auto-compile workflow generation
gh aw compile --verbose --auto-compile

# Watch for changes and automatically recompile workflows
gh aw compile --watch

# Watch with verbose output for development
gh aw compile --watch --verbose

# Watch with auto-compile workflow generation
gh aw compile --watch --auto-compile --verbose

# Remove a workflow
gh aw remove WorkflowName

# Run a workflow immediately
gh aw run WorkflowName

# Show status of all natural language action workflows
gh aw status

# Show status of workflows matching a pattern
gh aw status WorkflowPrefix
gh aw status path/to/workflow.lock.yml

# Disable all natural language action workflows
gh aw disable

# Disable workflows matching a pattern
gh aw disable WorkflowPrefix
gh aw disable path/to/workflow.lock.yml

# Enable all natural language action workflows
gh aw enable

# Enable workflows matching a pattern
gh aw enable WorkflowPrefix
gh aw enable path/to/workflow.lock.yml

# Download and analyze workflow logs
gh aw logs

# Download logs for specific workflow
gh aw logs weekly-research

# Download last 10 runs with date filtering
gh aw logs -c 10 --start-date 2024-01-01 --end-date 2024-01-31

# Download logs to custom directory
gh aw logs -o ./my-logs

# Inspect MCP servers in workflows
gh aw inspect

# Inspect MCP servers in a specific workflow
gh aw inspect weekly-research

# Inspect only specific MCP servers
gh aw inspect weekly-research --server repo-mind

# Verbose inspection with connection details
gh aw inspect weekly-research -v

# Launch the official MCP inspector tool
gh aw inspect weekly-research --inspector
```

### 🔍 MCP Server Inspection

The `inspect` command allows you to analyze and troubleshoot Model Context Protocol (MCP) servers configured in your workflows. This command connects to MCP servers, validates configurations, and displays available tools, resources, and capabilities.

```bash
# List all workflows that contain MCP server configurations
gh aw inspect

# Inspect all MCP servers in a specific workflow
gh aw inspect workflow-name

# Filter inspection to specific servers by name
gh aw inspect workflow-name --server server-name

# Enable verbose output with connection details
gh aw inspect workflow-name --verbose

# Launch the official @modelcontextprotocol/inspector web interface
gh aw inspect workflow-name --inspector
```

**Features:**
- **Server Discovery**: Automatically finds and lists workflows with MCP configurations
- **Connection Testing**: Validates that MCP servers can be reached and authenticated
- **Capability Inspection**: Lists available tools, resources, and roots from each server
- **Permission Analysis**: Shows which tools are allowed/blocked based on workflow configuration
- **Multi-Protocol Support**: Works with stdio, Docker container, and HTTP MCP servers
- **Secret Validation**: Checks that required environment variables and tokens are available
- **Web Inspector Integration**: Launches the official MCP inspector tool for interactive debugging

**Supported MCP Server Types:**
- **Stdio**: Direct command execution (`command` + `args`)
- **Docker**: Containerized servers (`container` field)
- **HTTP**: Remote MCP servers accessible via URL

### 🔄 Auto-Compile Workflow Management

The `--auto-compile` flag enables automatic compilation of agentic workflows when markdown files change.

```bash
# Generate auto-compile workflow that triggers on markdown file changes
gh aw compile --auto-compile

# Auto-compile workflow features:
# - Triggers when .github/workflows/*.md files are modified
# - Automatically compiles markdown files to .lock.yml files
# - Commits and pushes the compiled workflow files
# - Uses locally built gh-aw extension for development workflows
```

### 👀 Watch Mode for Development
The `--watch` flag provides automatic recompilation during workflow development, monitoring for file changes in real-time.
```bash
# Watch all workflow files in .github/workflows/ for changes
gh aw compile --watch

# Watch with verbose output for detailed compilation feedback
gh aw compile --watch --verbose

# Watch with auto-compile workflow generation
gh aw compile --watch --auto-compile --verbose

# Watch mode features:
# - Real-time monitoring of .github/workflows/*.md files
# - Automatic recompilation when markdown files are modified, created, or deleted
# - Debounced file system events (300ms) to prevent excessive compilation
# - Selective compilation - only recompiles changed files for better performance
# - Automatic cleanup of .lock.yml files when corresponding .md files are deleted
# - Graceful shutdown with Ctrl+C (SIGINT/SIGTERM handling)
# - Enhanced error handling with console formatting
# - Immediate feedback with success/error messages using emojis
```

### 📦 Package Management

```bash
# Install workflow packages globally (default)
gh aw install org/repo

# Install packages locally in current project
gh aw install org/repo --local

# Install a specific version, branch, or commit
gh aw install org/repo@v1.0.0
gh aw install org/repo@main --local
gh aw install org/repo@commit-sha

# Uninstall a workflow package globally
gh aw uninstall org/repo

# Uninstall a workflow package locally
gh aw uninstall org/repo --local

# List all installed packages (global and local)
gh aw list --packages

# List only local packages
gh aw list --packages --local
```

**Package Management Features:**

- **Install from GitHub**: Download workflow packages from any GitHub repository's `workflows/` directory
- **Version Control**: Specify exact versions, branches, or commits using `@version` syntax
- **Global Storage**: Global packages are stored in `~/.aw/packages/org/repo/` directory structure
- **Local Storage**: Local packages are stored in `.aw/packages/org/repo/` directory structure
- **Flexible Installation**: Choose between global (shared across projects) or local (project-specific) installations

**Note**: The `disable`, `enable`, and `status` commands require:

- GitHub CLI (`gh`) to be installed and authenticated
- The command to be run from within a git repository
- The workflows to be already updated (`.lock.yml` files must exist)

**Package Installation Requirements:**

- GitHub CLI (`gh`) to be installed and authenticated with access to the target repository
- Network access to download from GitHub repositories
- Target repository must have a `workflows/` directory containing `.md` files

### 📝 Creating New Workflows

The `gh aw new` command creates a new workflow markdown file with comprehensive template content and examples.

```bash
# Create a new workflow with example configuration
gh aw new my-custom-workflow

# Create a new workflow, overwriting if it exists
gh aw new issue-handler --force
```

**New Workflow Features:**

- **Template Generation**: Creates a comprehensive markdown file with commented examples
- **All Options Covered**: Includes examples of all trigger types, permissions, tools, and frontmatter options
- **Ready to Use**: Generated file serves as both documentation and working example
- **Customizable**: Easy to modify for specific use cases

**Generated Template Includes:**
- Complete frontmatter examples with all available options
- Trigger event configurations (issues, pull requests, schedule, etc.)
- Permissions and security settings
- AI processor configurations (Claude, Codex)
- Tools configuration (GitHub, MCP servers, etc.)
- Example workflow instructions

### 📊 Workflow Logs and Analysis

The `gh aw logs` command downloads and analyzes workflow execution logs with aggregated metrics and cost analysis.

```bash
# Download logs for all agentic workflows in the repository
gh aw logs

# Download logs for a specific workflow
gh aw logs weekly-research

# Limit the number of runs and filter by date
gh aw logs -c 10 --start-date 2024-01-01 --end-date 2024-01-31

# Download to custom directory
gh aw logs -o ./workflow-logs
```

**Workflow Logs Features:**

- **Automated Download**: Downloads logs and artifacts from GitHub Actions
- **Metrics Analysis**: Extracts execution time, token usage, and cost information
- **Aggregated Reporting**: Provides summary statistics across multiple runs
- **Flexible Filtering**: Filter by date range and limit number of runs
- **Cost Tracking**: Analyzes AI model usage costs when available
- **Custom Output**: Specify custom output directory for organized storage

**Log Analysis Includes:**
- Execution duration and performance metrics
- AI model token consumption and costs
- Success/failure rates and error patterns
- Workflow run frequency and patterns
- Artifact and log file organization

### 📋 Workflow Structure

#### Directory Structure

Agentic workflows are stored in a unified location:

- **`.github/workflows/`**: Contains both your markdown workflow definitions (source files) and the generated GitHub Actions YAML files (.lock.yml files)
- **`.gitattributes`**: Automatically created/updated to mark `.lock.yml` files as generated code using `linguist-generated=true`

Create markdown files in `.github/workflows/` with the following structure:

```markdown
---
on:
  issues:
    types: [opened]

permissions:
  issues: write

tools:
  github:
    allowed: [add_issue_comment]
---

# Workflow Description

Read the issue #${{ github.event.issue.number }}. Add a comment to the issue listing useful resources and links.
```

### ⚙️ Frontmatter Options for GitHub Actions

The YAML frontmatter supports standard GitHub Actions properties:

- `on`: Trigger events for the workflow.
- `permissions`: Required permissions for the workflow
- `run-name`: Name of the workflow run
- `runs-on`: Runner environment for the workflow
- `timeout_minutes`: Workflow timeout
- `concurrency`: Concurrency settings for the workflow
- `env`: Environment variables for the workflow
- `if`: Conditional execution of the workflow
- `steps`: Custom steps for the job

Additional properties are:

- `ai`: AI executor to use (`claude`, `codex`, `ai-inference`, defaults to `claude`)
- `tools`: Tools (e.g. `github`, `Bash`, custom MCP servers)
- `max-runs`: Maximum number of workflow runs before automatically disabling (optional, prevents cost overrun)
- `stop-time`: Deadline timestamp when workflow should stop running (optional, format: "YYYY-MM-DD HH:MM:SS")
- `ai-reaction`: Emoji reaction to add/remove on triggering GitHub item (optional, defaults to `eyes`)
- `cache`: Cache configuration for workflow dependencies (object or array)

### `on:`

A standard GitHub Actions `on:` trigger section. For example:

```markdown
on:
  issues:
    types: [opened]
```

Defaults to a "semi-active agent" - triggering frequently and when there is any activity in the repository.

```
on:
  # Start either every 10 minutes, or when some kind of human event occurs.
  # Because of the implicit "concurrency" section, only one instance of this
  # workflow will run at a time.
  schedule:
    - cron: "0/10 * * * *"
  issues:
    types: [opened, edited, closed]
  issue_comment:
    types: [created, edited]
  pull_request:
    types: [opened, edited, closed]
  push:
    branches:
      - main
  workflow_dispatch:
```

### `on: alias`

You can use the non-standard `on: alias` to create workflows that respond to `@mentions` in issues and comments:

```yaml
on:
  alias:
    name: my-bot  # Optional: defaults to filename without .md extension
```

This is equivalent to:
```yaml
on:
  issues:
    types: [opened, edited, reopened]
  issue_comment:
    types: [created, edited]
  pull_request:
    types: [opened, edited, reopened]
```

And adds a conditional `if` statement that matches `@alias-name` in issue bodies or comments:

```yaml
if: contains(github.event.issue.body, '@my-bot') || contains(github.event.comment.body, '@my-bot')
```

#### Combining alias with other events

The `alias` trigger can be combined with other compatible events:

```yaml
on:
  alias:
    name: my-bot
  workflow_dispatch:
  schedule:
    - cron: "0 9 * * 1"
```

This will create a workflow that responds to `@my-bot` mentions AND can be triggered manually or on a schedule.

**Note**: You cannot combine `alias` with `issues`, `issue_comment`, or `pull_request` events as these would conflict with the automatic alias events.

**Example alias workflow:**

```markdown
---
on:
  alias:
    name: summarize-issue
permissions:
  issues: write
tools:
  github:
    allowed: [add_issue_comment]
---

# Issue Summarizer

When someone mentions @summarize-issue in an issue or comment, 
analyze and provide a helpful summary.

The current context text is: "${{ needs.task.outputs.text }}"
```

### `text` Output Variable

Workflows using the `alias` trigger (and all other workflows) have access to a computed `text` output variable from the preamble `task` job that provides the current context text based on the triggering event:

```markdown
---
on:
  alias:
    name: my-bot
---

# My Bot

Analyze the current text: "${{ needs.task.outputs.text }}"
```

**How the `text` variable is computed:**

- **Issues**: `title + "\n\n" + body` (issue title and body combined)
- **Pull Requests**: `title + "\n\n" + body` (PR title and body combined)  
- **Issue Comments**: `comment.body` (comment content only)
- **PR Review Comments**: `comment.body` (comment content only)
- **PR Reviews**: `review.body` (review content only)
- **Other events**: Empty string

**Benefits of using `${{ needs.task.outputs.text }}`:**

- **Unified interface**: Works consistently across all event types
- **Context-aware**: Automatically provides the most relevant text for each event
- **Maintainable**: Single source of truth for current context text
- **Alias-friendly**: Perfect for workflows that respond to @mentions in various contexts

**Migration from hardcoded event references:**

Instead of using event-specific references:
```markdown
# Before - event-specific
- Issue body: ${{ github.event.issue.body }}
- Comment content: ${{ github.event.comment.body }}  
- Pull request body: ${{ github.event.pull_request.body }}

# After - unified approach
The current text: "${{ needs.task.outputs.text }}"
```

### `permissions:`

See [GitHub Actions permissions](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions).

If you specify the access for any of these permissions, all of those that are not specified are set to none.

You can use the following syntax to define one of read-all or write-all access for all of the available permissions:

```
permissions: read-all
```

or

```
permissions: write-all
```

You can use the following syntax to disable permissions for all of the available permissions:


```
permissions: {}
```

See also [How permissions are calculated for a workflow job](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#how-permissions-are-calculated-for-a-workflow-job)

### `run-name:`

Defaults to the workflow name, which can be overridden in the frontmatter.

### `runs-on:`

Defaults to `ubuntu-latest`, which can be overridden in the frontmatter to specify different runner environments like `windows-latest`, `macos-latest`, or specific runner labels.

### `timeout_minutes:`

Defaults to 15 minutes, which can be adjusted based on the expected duration of the workflow.

### `concurrency:`

Defaults to a single instance of the workflow running at a time, which can be adjusted based on the workflow's requirements.

### `env:`

Environment variables can be defined in the frontmatter to be used within the workflow. Defaults to empty.

### `if:`

Conditional execution of the workflow can be specified in the frontmatter. Defaults to empty, meaning the workflow always runs.

### `steps:`

Custom GitHub Action steps used before the agentic workflow.

### `engine:`

Specifies which AI engine to use for the workflow. Defaults to `claude`. Supports both simple string format and extended object format for advanced configuration.

#### Simple String Format

```yaml
engine: claude  # or codex, opencode, ai-inference
```

- `claude` (default): Uses Claude Code with full MCP tool support and allow-listing
- `codex` (**experimental**): Uses codex with OpenAI endpoints   
- `opencode` (**experimental**): Uses OpenCode AI coding assistant   
- `ai-inference`: Uses GitHub Models via actions/ai-inference with GitHub MCP support   

#### Extended Object Format

For advanced configuration, use the object format:

```yaml
engine:
  id: claude                        # Required: agent CLI identifier (claude, codex, opencode, ai-inference)
  version: beta                     # Optional: version of the action
  model: claude-3-5-sonnet-20241022 # Optional: LLM model to use
```

**Fields:**
- **`id`** (required): The agent CLI identifier - must be `claude`, `codex`, `opencode`, or `ai-inference`
- **`version`** (optional): Version of the action to use (e.g., `beta`, `stable`)
- **`model`** (optional): Specific LLM model to use (e.g., `claude-3-5-sonnet-20241022`, `gpt-4o`)

#### Examples

Simple format:
```yaml
---
on: issues
engine: claude
tools:
  github:
    allowed: [get_issue, add_issue_comment]
---
```

Extended format with model specification:
```yaml
---
on: issues
engine:
  id: claude
  version: beta
  model: claude-3-5-sonnet-20241022
tools:
  github:
    allowed: [get_issue, add_issue_comment]
---
```

Codex with specific model:
```yaml
---
on: issues
engine:
  id: codex
  model: gpt-4o
---
```

OpenCode configuration:
```yaml
---
on: issues
engine: opencode
tools:
  github:
    allowed: [get_issue, add_issue_comment]
---
```

AI Inference configuration:
```yaml
---
on: issues
engine: ai-inference
permissions:
  models: read
tools:
  github:
    allowed: [get_issue, add_issue_comment]
---
```

**Note:** Both Codex and OpenCode support are **experimental** and have different capabilities and limitations compared to Claude Code. 

**Note:** The AI Inference engine requires `models: read` permission to access GitHub Models and automatically enables GitHub MCP integration for tool support. 

### `max-runs:`

Sets a maximum number of successful workflow runs before the workflow automatically disables itself.

```yaml
---
max-runs: 10
tools:
  github:
    allowed: [create_issue, update_issue]
---
```

When the workflow runs, it will:
1. Count successful workflow runs that have produced a `workflow-complete.txt` artifact
2. If the count reaches or exceeds `max-runs`, disable the workflow using `gh workflow disable`
3. Allow the current run to complete but prevent future runs

**Note**: Only completed runs that have successfully generated the `workflow-complete.txt` artifact are counted toward the limit. This ensures that failed, incomplete or early-exiting runs don't contribute to cost overrun protection.

### `stop-time:`

Sets a deadline after which the workflow will automatically disable itself.

```yaml
---
stop-time: "2025-12-31 23:59:59"
tools:
  github:
    allowed: [create_issue, update_issue]
---
```

The workflow will:
1. Check if the current time has passed the `stop-time` deadline
2. If the deadline is reached, disable the workflow using `gh workflow disable`
3. Allow the current run to complete but prevent future runs

### `ai-reaction:`

Specifies an emoji reaction to automatically add and remove on the GitHub item (issue, PR, or comment) that triggered the workflow. This provides visual feedback that the agentic workflow is processing.

```yaml
---
ai-reaction: "eyes"
on:
  issues:
    types: [opened]
tools:
  github:
    allowed: [add_issue_comment]
---
```

**Available reactions:**
- `+1` (👍) - thumbs up
- `-1` (👎) - thumbs down  
- `laugh` (😄) - laugh
- `confused` (😕) - confused
- `heart` (❤️) - heart
- `hooray` (🎉) - hooray  
- `rocket` (🚀) - rocket
- `eyes` (👀) - eyes (default)

**Behavior:**
1. **Reaction Added**: When the workflow starts, the specified reaction is automatically added to the triggering GitHub item
2. **Reaction Removed**: When the workflow completes successfully, the reaction is automatically removed
3. **Default Value**: If not specified, defaults to `eyes` (👀)
4. **Supported Events**: Works with `issues`, `issue_comment`, `pull_request`, `pull_request_target`, and `pull_request_review_comment` events

**Example workflows with reactions:**

```markdown
---
ai-reaction: "rocket"
on:
  pull_request:
    types: [opened]
---

# PR Reviewer

Analyze the pull request and provide feedback.
```

```markdown
---
ai-reaction: "heart"
on:
  issues:
    types: [opened]
    labels: ["bug"]
---

# Bug Triage

Automatically triage and respond to bug reports.
```

**Note**: The reaction functionality requires no additional permissions as it uses the default `GITHUB_TOKEN` with automatic cleanup on workflow completion.

### `cache:`

Specifies cache configuration for workflow dependencies using the same syntax as the GitHub Actions `actions/cache` action. Cache configuration helps improve workflow performance by storing and reusing dependencies between runs.

**Single Cache:**
```yaml
cache:
  key: node-modules-${{ hashFiles('package-lock.json') }}
  path: node_modules
  restore-keys: |
    node-modules-
```

**Multiple Caches:**
```yaml
cache:
  - key: node-modules-${{ hashFiles('package-lock.json') }}
    path: node_modules
    restore-keys: |
      node-modules-
  - key: build-cache-${{ github.sha }}
    path: 
      - dist
      - .cache
    restore-keys:
      - build-cache-
    fail-on-cache-miss: false
```

**Supported Cache Parameters:**
- `key:` - Cache key (required)
- `path:` - Files/directories to cache (required, string or array)
- `restore-keys:` - Fallback keys (string or array)
- `upload-chunk-size:` - Chunk size for large files (integer)
- `fail-on-cache-miss:` - Fail if cache not found (boolean)
- `lookup-only:` - Only check cache existence (boolean)

Cache steps are automatically added to the workflow job and the cache configuration is removed from the final `.lock.yml` file.

### `tools:`

Tools can be defined in the frontmatter to specify which GitHub API calls and Bash commands are allowed. Defaults to empty.

All tools declared in included components are also included in the lock file.

#### `github:` tools

When you configure a `github` tool, you can specify which GitHub API operations are allowed. This allows you to control which actions the workflow can perform on GitHub repositories.
```yaml
tools:
  github:
    allowed: [create_issue, update_issue]
```

The system automatically includes a comprehensive set of default read-only GitHub MCP tools to provide broad repository access capabilities. These include:

* **Actions**: `download_workflow_run_artifact`, `get_job_logs`, `get_workflow_run`, `list_workflows`, etc.
* **Issues & PRs**: `get_issue`, `get_pull_request`, `list_issues`, `list_pull_requests`, `search_issues`, etc.  
* **Repository**: `get_commit`, `get_file_contents`, `list_branches`, `list_commits`, `search_code`, etc.
* **Security**: `get_code_scanning_alert`, `list_secret_scanning_alerts`, `get_dependabot_alert`, etc.
* **Users & Organizations**: `search_users`, `search_orgs`, `get_me`, etc.

These default tools are merged with any specific GitHub tools you declare in your `allowed` list, so you get both your custom tools and the comprehensive defaults.

##### GitHub MCP Configuration Options

The GitHub tool supports additional configuration options to control how MCP servers are executed:

```yaml
tools:
  github:
    docker_image_version: "sha-45e90ae" # Docker image version (default: "sha-45e90ae")
    allowed: [create_issue, update_issue]
```


#### `claude:` tools

When using `engine: claude`, you can define Claude-specific tools under the `claude:` section. This allows you to use Claude's capabilities like `Edit`, `Write`, `NotebookEdit`, etc. You can find a list of built-in tools avaliable in Claude Code in [its documentation](https://docs.anthropic.com/en/docs/claude-code/settings#tools-available-to-claude)

Group Claude-specific tools under a `claude:` section, and use a key for each:

```yaml
tools:
  claude:
    allowed:
      Edit:
      MultiEdit:
      Write:
      NotebookEdit:
      WebFetch:
      WebSearch:
      Bash: ["echo", "ls", "git status"]
```

#### Claude Bash Wildcards

The Claude Bash tool supports wildcard configurations for allowing bash commands:

```yaml
tools:
  claude:
    Bash:
      allowed: [":*"]  # Allow all bash commands - use with caution
```

**Wildcard Options:**
- **`:*`**: Allows **all bash commands** without restriction. When `:*` is present, all other commands in the `allowed` list are ignored.
- **`prefix:*`**: Allows **all bash commands starting with the prefix** without restriction.

**Security Note:** Using `:*` or `*` wildcards allows unrestricted bash access, which should only be used in trusted environments or when full bash capability is required.

**Example with mixed commands (`:*` takes precedence):**
```yaml
tools:
  claude:
    Bash:
      allowed: ["echo", "ls", ":*", "git status"]  # Only :* is effective
```

#### Default Claude Tools

When using `engine: claude` (the default) and including a `github` tool in your configuration, the following default Claude tools are automatically added to the `claude:` section to provide essential workflow capabilities:

- **`Task`**: Task management and workflow coordination
- **`Glob`**: File pattern matching and globbing operations  
- **`Grep`**: Text search and pattern matching within files
- **`LS`**: Directory listing and file system navigation
- **`Read`**: File reading operations
- **`NotebookRead`**: Jupyter notebook reading capabilities

These tools are added automatically and don't need to be explicitly declared in your `tools` section. If you already have any of these tools configured with custom settings, your configuration will be preserved.

**Note:** Default tools are only added when using `engine: claude`. When using `engine: codex`, the tools section is ignored entirely as codex doesn't support MCP tool allow-listing.

#### Custom MCP Tools

While we support `github` MCP server out of the box, you can also configure custom MCP servers directly in the `tools` section by specifying the appropriate MCP type:

```yaml
tools:
  trello:
    mcp:
      type: stdio
      command: "python"
      args: ["-m", "trello_mcp"]
      env:
        CUSTOM_TOKEN: "${secrets.CUSTOM_TOKEN}"
    allowed: ["create_card", "list_boards"]
```

**Docker Container Field:**

You can simplify MCP server definitions by using the `container` field, which automatically generates Docker run commands. This field cannot be combined with the `command` field:

```yaml
tools:
  notion:
    mcp:
      type: stdio
      container: "mcp/notion"
      env:
        NOTION_TOKEN: "${secrets.NOTION_TOKEN}"
    allowed: ["create_page", "search_pages"]
```

The `container` field automatically transforms into:
- **`command`**: `"docker"`  
- **`args`**: `["run", "--rm", "-i", "-e", "NOTION_TOKEN", "mcp/notion"]`

**Container Field Constraints:**
- **Mutually exclusive**: Cannot be combined with explicit `command` field
- **Auto environment**: Automatically adds `-e` flags for each environment variable

**JSON String Format:**
You can also specify MCP configuration as a JSON string, which is useful for complex configurations or when copying from existing JSON configurations:

```yaml
tools:
  trello:
    mcp: |
      { "type": "stdio",
        "command": "python",
        "args": ["-m", "trello_mcp"],
        "env":  { "CUSTOM_TOKEN: \"${secrets.CUSTOM_TOKEN}\" } }}
    allowed: ["create_card", "list_boards"]  
```

Both YAML object format and JSON string format are equivalent and will produce the same MCP server configuration.

**MCP Server Types:**
- **`stdio`**: Standard input/output communication (supported by both `codex` and `claude`)
- **`http`**: HTTP-based communication (supported by `claude` only)

MCP tools support wildcard access patterns to allow all tools from a particular server:

```yaml
tools:
  notion:
    mcp:
      ...
    allowed: ["*"]  # Allow all tools from the notion server
```

**Wildcard Behavior:**
- **`["*"]`**: Allows **all tools** from the MCP server. This generates `mcp__servername` as the allowed tool (e.g., `mcp__notion`)
- **Specific tools**: When listing specific tools like `["create_page", "search_pages"]`, each tool generates `mcp__servername__toolname` (e.g., `mcp__notion__create_page`, `mcp__notion__search_pages`)

**HTTP Server Example:**
```yaml

tools:
  notion:
    mcp:
      type: http
      url: "https://mcp.notion.com"
      headers:
        Authorization: "${CUSTOM_TOKEN}"
    allowed: ["*"]
```

### 📝 Include Directives

```markdown
@include relative/path/to/file.md -- Includes files relative to the current markdown file.
@include filename.md#Section -- Includes only a specific section from a markdown file.
```

The only frontmatter allowed in included files is `tools:`. The `allowed:` tools are merged across included files.

### 🔐 Secrets

Some secrets are automatically used by the generated workflow:

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- `ANTHROPIC_API_KEY`: For `claude` CLI
- `OPENAI_API_KEY`: For `codex` CLI

Additional secrets may not yet be defined in the frontmatter.

If the text of your MCP specifications or other frontmatter refer to secrets, these will be used in the specifications.