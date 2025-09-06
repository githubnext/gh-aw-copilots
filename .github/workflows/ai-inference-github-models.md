---
name: AI Inference with GitHub Models
on:
  workflow_dispatch:
    inputs:
      input_number:
        description: 'The number of the issue to analyze'
        required: true
  issues:
    types: [opened]

permissions:
  contents: read
  models: read

engine:
  id: custom
  max-turns: 3
  steps:
    - name: Setup AI Inference with GitHub Models
      uses: actions/ai-inference@v1
      id: ai_inference
      with:
        # Use gpt-4o-mini model
        model: gpt-4o-mini
        # Use the provided prompt or create one based on the event
        prompt-file: ${{ env.GITHUB_AW_PROMPT }}
        # Configure the AI inference settings
        max_tokens: 1000
        temperature: 0.7
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

    - name: Create Issue Comment
      run: |
        # Determine issue number based on event type
        if [ "${{ github.event_name }}" == "issues" ]; then
          ISSUE_NUMBER="${{ github.event.issue.number }}"
        else
          ISSUE_NUMBER="${{ github.event.inputs.input_number }}"
        fi
        
        # Generate safe output for issue comment
        echo "{\"type\": \"add-issue-comment\", \"issue_number\": \"$ISSUE_NUMBER\", \"body\": \"## 🤖 AI Analysis\\n\\nI've analyzed this issue using GitHub's AI models. Here's my assessment:\\n\\n${{ steps.ai_inference.outputs.response }}\\n\\n---\\n*This response was generated using GitHub Models via the AI Inference action.*\"}" >> $GITHUB_AW_SAFE_OUTPUTS

safe-outputs:
  add-issue-comment:
    max: 1
    target: "*"
---

Summarize the issue

## Purpose

This workflow showcases how to leverage GitHub's AI infrastructure to summarize issues, demonstrating:
- Integration with GitHub Models through the AI inference action
- Safe output handling for AI-generated responses  
- Event-driven AI analysis of repository issues
- Support for manual execution on specific issues

## Features

- **Issue Analysis**: Automatically responds to new issues with AI-generated summaries
- **Manual Execution**: Analyze any specific issue by providing its number
- **Safe Outputs**: Uses safe output system to post AI responses as issue comments
- **Fixed Model**: Uses gpt-4o-mini for consistent, fast responses

## Usage Examples

### Manual Execution for Specific Issue
1. Navigate to Actions tab in your repository
2. Select "AI Inference with GitHub Models" workflow  
3. Click "Run workflow"
4. Enter the issue number to analyze
5. Click "Run workflow" to execute

### Automatic Issue Analysis
When a new issue is created, the workflow automatically:
- Analyzes the issue title and description using AI
- Generates a summary of the issue content  
- Posts the AI analysis as a comment on the issue via safe outputs

## Configuration Options

### Model Parameters
- **Model**: gpt-4o-mini (fixed for consistent performance)
- **Temperature**: 0.7 (controls response creativity/randomness)
- **Max Tokens**: 1000 (maximum response length)

### Customization
Edit this workflow file to:
- Modify the prompt for issue analysis
- Change model parameters (temperature, max_tokens)  
- Add additional processing steps
- Adjust safe output configurations

## Implementation Details

### Custom Engine
Uses the custom engine with these key components:
- **AI Inference Action**: `actions/ai-inference@v1` for GitHub Models integration
- **Prompt File**: Uses `${{ env.GITHUB_AW_PROMPT }}` for dynamic prompt handling
- **Issue Number Handling**: Supports both automatic (from issue events) and manual (from workflow dispatch) execution

### Safe Outputs
Implements safe output handling for:
- **add-issue-comment**: Posts AI responses as issue comments (max: 1)
- **Target**: "*" (applies to any issue)

### Security Features
- **Minimal Permissions**: Only `contents: read` and `models: read`
- **GitHub's AI Infrastructure**: Uses secure GitHub Models backend
- **Content Filtering**: Built-in safety measures in GitHub Models
- **Safe Output System**: Prevents unauthorized repository actions