---
on:
  workflow_dispatch:
  issues:
    types: [opened, reopened, closed]
  pull_request:
    types: [opened, reopened, synchronize, closed]
  push:
    branches: [main]
  schedule:
    - cron: "0 12 * * 1"  # Weekly on Mondays at noon

engine:
  id: custom
  steps:
    - name: Setup test environment
      run: |
        echo "Testing all safe outputs with custom engine"
        echo "Creating sample outputs for each safe output type..."
        
        # Create directory for safe outputs
        mkdir -p safe-outputs
        
    - name: Generate Create Issue Output
      run: |
        cat > safe-outputs/create-issue.md << 'EOF'
        # Test Issue Created by Custom Engine
        
        This issue was automatically created by the test-safe-outputs-custom-engine workflow to validate the create-issue safe output functionality.
        
        **Test Details:**
        - Engine: Custom
        - Trigger: ${{ github.event_name }}
        - Repository: ${{ github.repository }}
        - Run ID: ${{ github.run_id }}
        
        This is a test issue and can be closed after verification.
        EOF
        
    - name: Generate Add Issue Comment Output
      run: |
        cat > safe-outputs/add-issue-comment.md << 'EOF'
        ## Test Comment from Custom Engine
        
        This comment was automatically posted by the test-safe-outputs-custom-engine workflow to validate the add-issue-comment safe output functionality.
        
        **Test Information:**
        - Workflow: test-safe-outputs-custom-engine
        - Engine Type: Custom (GitHub Actions steps)
        - Execution Time: $(date)
        - Event: ${{ github.event_name }}
        
        ✅ Safe output testing in progress...
        EOF
        
    - name: Generate Add Issue Labels Output
      run: |
        cat > safe-outputs/add-issue-labels.txt << 'EOF'
        test-safe-outputs
        automation
        custom-engine
        EOF
        
    - name: Generate Update Issue Output
      run: |
        cat > safe-outputs/update-issue.json << 'EOF'
        {
          "title": "[UPDATED] Test Issue - Custom Engine Safe Output Test",
          "body": "# Updated Issue Body\n\nThis issue has been updated by the test-safe-outputs-custom-engine workflow to validate the update-issue safe output functionality.\n\n**Update Details:**\n- Updated by: Custom Engine\n- Update time: $(date)\n- Original trigger: ${{ github.event_name }}\n\n**Test Status:** ✅ Update functionality verified",
          "status": "open"
        }
        EOF
        
    - name: Generate Create Pull Request Output
      run: |
        # Create a test file change
        echo "# Test file created by custom engine safe output test" > test-custom-engine-$(date +%Y%m%d-%H%M%S).md
        echo "This file was created to test the create-pull-request safe output." >> test-custom-engine-$(date +%Y%m%d-%H%M%S).md
        echo "Generated at: $(date)" >> test-custom-engine-$(date +%Y%m%d-%H%M%S).md
        
        # Create PR description
        cat > safe-outputs/create-pull-request.md << 'EOF'
        # Test Pull Request - Custom Engine Safe Output
        
        This pull request was automatically created by the test-safe-outputs-custom-engine workflow to validate the create-pull-request safe output functionality.
        
        ## Changes Made
        - Created test file with timestamp
        - Demonstrates custom engine file creation capabilities
        
        ## Test Information
        - Engine: Custom (GitHub Actions steps)
        - Workflow: test-safe-outputs-custom-engine
        - Trigger Event: ${{ github.event_name }}
        - Run ID: ${{ github.run_id }}
        
        This PR can be merged or closed after verification of the safe output functionality.
        EOF
        
    - name: Generate Create Discussion Output
      run: |
        cat > safe-outputs/create-discussion.md << 'EOF'
        # Test Discussion - Custom Engine Safe Output
        
        This discussion was automatically created by the test-safe-outputs-custom-engine workflow to validate the create-discussion safe output functionality.
        
        ## Purpose
        This discussion serves as a test of the safe output system's ability to create GitHub discussions through custom engine workflows.
        
        ## Test Details
        - **Engine Type:** Custom (GitHub Actions steps)  
        - **Workflow:** test-safe-outputs-custom-engine
        - **Created:** $(date)
        - **Trigger:** ${{ github.event_name }}
        - **Repository:** ${{ github.repository }}
        
        ## Discussion Points
        1. Custom engine successfully executed
        2. Safe output file generation completed
        3. Discussion creation triggered
        
        Feel free to participate in this test discussion or archive it after verification.
        EOF
        
    - name: Generate PR Review Comment Output
      run: |
        cat > safe-outputs/create-pull-request-review-comment.json << 'EOF'
        {
          "path": "README.md",
          "line": 1,
          "body": "## Custom Engine Review Comment Test\n\nThis review comment was automatically created by the test-safe-outputs-custom-engine workflow to validate the create-pull-request-review-comment safe output functionality.\n\n**Review Details:**\n- Generated by: Custom Engine\n- Test time: $(date)\n- Workflow: test-safe-outputs-custom-engine\n\n✅ PR review comment safe output test completed."
        }
        EOF
        
    - name: Generate Push to Branch Output
      run: |
        # Create another test file for branch push
        echo "# Branch Push Test File" > branch-push-test-$(date +%Y%m%d-%H%M%S).md
        echo "This file tests the push-to-branch safe output functionality." >> branch-push-test-$(date +%Y%m%d-%H%M%S).md
        echo "Created by custom engine at: $(date)" >> branch-push-test-$(date +%Y%m%d-%H%M%S).md
        
        cat > safe-outputs/push-to-branch.md << 'EOF'
        Custom engine test: Push to branch functionality
        
        This commit was generated by the test-safe-outputs-custom-engine workflow to validate the push-to-branch safe output functionality.
        
        Files created:
        - branch-push-test-[timestamp].md
        
        Test executed at: $(date)
        EOF
        
    - name: Generate Missing Tool Output
      run: |
        cat > safe-outputs/missing-tool.json << 'EOF'
        {
          "tool_name": "example-missing-tool",
          "reason": "This is a test of the missing-tool safe output functionality. No actual tool is missing.",
          "alternatives": "This is a simulated missing tool report generated by the custom engine test workflow.",
          "context": "test-safe-outputs-custom-engine workflow validation"
        }
        EOF
        
    - name: List generated outputs
      run: |
        echo "Generated safe output files:"
        find safe-outputs -type f -exec echo "- {}" \; -exec head -3 {} \; -exec echo "" \;
        
        echo "Additional test files created:"
        ls -la *.md 2>/dev/null || echo "No additional .md files found"

safe-outputs:
  create-issue:
    title-prefix: "[Custom Engine Test] "
    labels: [test-safe-outputs, automation, custom-engine]
    max: 1
  add-issue-comment:
    max: 1
    target: "*"
  create-pull-request:
    title-prefix: "[Custom Engine Test] "
    labels: [test-safe-outputs, automation, custom-engine]
    draft: true
  add-issue-label:
    allowed: [test-safe-outputs, automation, custom-engine, bug, enhancement, documentation]
    max: 3
  update-issue:
    status:
    title:
    body:
    target: "*"
    max: 1
  push-to-branch:
    target: "*"
  missing-tool:
    max: 5
  create-discussion:
    title-prefix: "[Custom Engine Test] "
    max: 1
  create-pull-request-review-comment:
    max: 1
    side: "RIGHT"

permissions:
  contents: read
  issues: write
  pull-requests: write
  discussions: write
---

# Test Safe Outputs - Custom Engine

This workflow validates all safe output types using the custom engine implementation. It demonstrates the ability to use GitHub Actions steps directly in agentic workflows while leveraging the safe output processing system.

## Purpose

This is a comprehensive test workflow that exercises every available safe output type:

- **create-issue**: Creates test issues with custom engine
- **add-issue-comment**: Posts comments on issues/PRs
- **create-pull-request**: Creates PRs with code changes
- **add-issue-label**: Adds labels to issues/PRs
- **update-issue**: Updates issue properties
- **push-to-branch**: Pushes changes to branches
- **missing-tool**: Reports missing functionality (test simulation)
- **create-discussion**: Creates repository discussions
- **create-pull-request-review-comment**: Creates PR review comments

## Custom Engine Implementation

The workflow uses the custom engine with GitHub Actions steps to generate all the required safe output files. Each step creates the appropriate output file with test content that demonstrates the functionality.

## Test Content

All generated content is clearly marked as test data and includes:
- Timestamp information
- Trigger event details
- Workflow identification
- Clear indication that it's test data

The content can be safely created and cleaned up as part of testing the safe output functionality.