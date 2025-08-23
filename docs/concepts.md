# 📖 Introduction to Agentic Workflows

Now that you've got your first workflow running, let's dive deeper into the concepts and capabilities of GitHub Agentic Workflows.

Agentic workflows represent a new paradigm where AI agents can perform complex, multi-step tasks in your repository automatically. They combine the power of large language models with GitHub's collaboration platform to create truly intelligent automation.

## Core Concepts

### What makes a workflow "agentic"?

Traditional GitHub Actions follow pre-programmed steps. Agentic workflows use AI to:

- **Understand context** — Read and analyze repository content, issues, PRs, and discussions
- **Make decisions** — Determine what actions to take based on the current situation  
- **Use tools** — Interact with GitHub APIs, external services, and repository files
- **Generate content** — Create meaningful comments, documentation, and code changes
- **Learn and adapt** — Adjust behavior based on past action, feedback and outcomes
- **Productive ambiguity** — Interpret natural language instructions flexibly and productively

### The anatomy of an agentic workflow

Every agentic workflow has two main parts:

1. **Frontmatter (YAML)** — Configuration that defines triggers, permissions, and available tools
2. **Instructions (Markdown)** — Natural language description of what the AI should do

```markdown
---
# Configuration
on: { issues: { types: [opened] } }
permissions: { issues: write }
tools: { github: { allowed: [add_issue_comment] } }
---

# Natural Language Instructions
Analyze this issue and provide helpful triage comments...
```

See [Workflow Structure](workflow-structure.md) for details on file layout and security.

## Understanding AI Engines

Agentic workflows are powered by different AI engines:

- **Claude** (default) — Anthropic's AI model, excellent for reasoning and code analysis
- **Codex** (experimental) — OpenAI's code-focused model

The engine interprets your natural language instructions and executes them using the tools and permissions you've configured.

### Continuous AI Patterns

GitHub Agentic Workflows enable "Continuous AI" — the systematic application of AI to software collaboration:

- **Continuous Documentation** — Keep docs current and comprehensive
- **Continuous Code Improvement** — Incrementally enhance code quality
- **Continuous Triage** — Intelligent issue and PR management
- **Continuous Research** — Stay current with industry developments
- **Continuous Quality** — Automated code review and standards enforcement

### 📂 Available Demonstrator Workflows from "[The Agentics](https://github.com/githubnext/agentics?tab=readme-ov-file#-the-agentics)"

#### Research & Planning Workflows
- [📚 Weekly Research](https://github.com/githubnext/agentics?tab=readme-ov-file#-weekly-research) - Collect research updates and industry trends
- [👥 Daily Team Status](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-team-status) - Assess repository activity and create status reports
- [📋 Daily Plan](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-plan) - Update planning issues for team coordination
- [🏷️ Issue Triage](https://github.com/githubnext/agentics?tab=readme-ov-file#️-issue-triage) - Triage issues and pull requests

#### Coding & Development Workflows
- [📦 Daily Dependency Updater](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-dependency-updater) - Update dependencies and create pull requests
- [📖 Regular Documentation Update](https://github.com/githubnext/agentics?tab=readme-ov-file#-regular-documentation-update) - Update documentation automatically
- [🔍 Daily QA](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-qa) - Perform quality assurance tasks
- [🧪 Daily Test Coverage Improver](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-test-coverage-improver) - Improve test coverage by adding meaningful tests to under-tested areas
- [⚡ Daily Performance Improver](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-performance-improver) - Analyze and improve code performance through benchmarking and optimization
- [🔍 Daily Accessibility Review](https://github.com/githubnext/agentics?tab=readme-ov-file#-daily-accessibility-review) - Review application accessibility

## Advanced Concepts

### Lock Files and Compilation

When you modify a `.md` workflow file, you need to compile it:

```bash
gh aw compile
```

This generates a `.lock.yml` file containing the actual GitHub Actions workflow. Both files should be committed to your repository.

### Security and Permissions

Agentic workflows require careful security consideration:

- **Minimal permissions** — Grant only what the workflow needs
- **Tool allowlists** — Explicitly specify which tools the AI can use  
- **Input validation** — All inputs are automatically sanitized
- **Human oversight** — Critical actions can require human approval

See [Security Notes](security-notes.md) for comprehensive guidelines.

### Tools and MCPs

Workflows can use various tools through the Model Context Protocol (MCP):

- **GitHub tools** — Repository management, issue/PR operations
- **External APIs** — Integration with third-party services
- **File operations** — Read, write, and analyze repository files
- **Custom MCPs** — Build your own tool integrations

Learn more in [Tools Configuration](tools.md) and [MCPs](mcps.md).

## Building Effective Workflows

### Best Practices

1. **Start simple** — Begin with basic workflows and add complexity gradually
2. **Be specific** — Clear, detailed instructions produce better results
3. **Test iteratively** — Use `gh aw compile --watch` during development
4. **Monitor costs** — Use `gh aw logs` to track AI usage and optimize
5. **Review outputs** — Always verify AI-generated content before merging

### Common Patterns

- **Event-driven** — Respond to issues, PRs, pushes, etc.
- **Scheduled** — Regular maintenance and reporting tasks
- **Alias-triggered** — Activated by @mentions in comments
- **Conditional** — Use frontmatter logic to control execution

## Next Steps

Ready to build more sophisticated workflows? Explore:

- **[Workflow Structure](workflow-structure.md)** — Detailed file organization and security
- **[Frontmatter Options](frontmatter.md)** — Complete configuration reference
- **[Tools Configuration](tools.md)** — Available tools and permissions
- **[VS Code Integration](vscode.md)** — Enhanced authoring experience

The power of agentic workflows lies in their ability to understand context, make intelligent decisions, and take meaningful actions — all while maintaining the security and reliability you expect from GitHub Actions.