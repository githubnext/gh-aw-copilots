# PR Creation Feature Documentation

The `--pr` flag for the `gh aw add` command automatically creates a pull request when adding workflows.

## Usage

```bash
gh aw add <workflow-name> --pr
```

## What it does

When you use the `--pr` flag, the command will:

1. **Check prerequisites:**
   - Verify GitHub CLI is available and authenticated
   - Confirm you're in a git repository
   - Check if you have write access to the repo (or can fork)
   - Ensure working directory is clean (no uncommitted changes)

2. **Create workflow branch:**
   - Get current branch name for later restoration
   - Create a temporary branch named `add-workflow-<workflow-name>`
   - Switch to the new branch

3. **Add workflow files:**
   - Use existing workflow installation logic
   - Install any required packages if `-r` flag is used
   - Copy workflow files to `.github/workflows/`
   - Stage changes to git

4. **Create pull request:**
   - Commit changes with descriptive message
   - Push branch to remote (with upstream tracking)
   - Create PR using GitHub CLI
   - Display PR URL to user

5. **Clean up:**
   - Switch back to original branch
   - Leave the new branch available for further changes if needed

## Example

```bash
# Add a workflow and create PR in one command
gh aw add weekly-research --pr

# With verbose output to see all steps
gh aw add weekly-research --pr --verbose

# Combined with repository installation
gh aw add weekly-research -r githubnext/agentics --pr
```

## Error Scenarios

The command will fail gracefully in these scenarios:
- GitHub CLI not installed or not authenticated
- Not in a git repository
- Working directory has uncommitted changes
- No write access and unable to fork
- Workflow not found
- Git operations fail (network issues, permissions)

## Benefits

- **Streamlined workflow:** No need to manually create branches, commit, push, and create PR
- **Consistent process:** Same branch naming and commit message format every time
- **Safety checks:** Prevents common mistakes like uncommitted changes or missing authentication
- **Flexibility:** Works with all existing `gh aw add` flags and options