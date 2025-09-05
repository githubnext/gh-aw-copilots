---
name: AI Inference with GitHub Models
on:
  workflow_dispatch:
    inputs:
      prompt:
        description: 'The prompt to send to the AI model'
        required: false
        default: 'Write a simple "Hello, World!" program in Python'
      model:
        description: 'The GitHub model to use'
        required: false
        default: 'gpt-4o-mini'
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
        # Use the specified model or default to gpt-4o-mini
        model: ${{ github.event.inputs.model || 'gpt-4o-mini' }}
        # Use the provided prompt or create one based on the event
        prompt-file: ${{ env.GITHUB_AW_PROMPT }}
        # Configure the AI inference settings
        max_tokens: 1000
        temperature: 0.7
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

    - name: Process AI Response
      run: |
        echo "AI Inference completed successfully"
        echo "Response from GitHub Model (${{ github.event.inputs.model || 'gpt-4o-mini' }}):"
        echo "================================="
        echo "${{ steps.ai_inference.outputs.response }}"
        echo "================================="
        
        # Save the response to a file for potential further processing
        echo "${{ steps.ai_inference.outputs.response }}" > ai_response.txt
        
        # Create a summary for GitHub Actions
        echo "## AI Inference Results" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "**Model Used:** ${{ github.event.inputs.model || 'gpt-4o-mini' }}" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "**Response:**" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY
        echo "${{ steps.ai_inference.outputs.response }}" >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY

    - name: Create Issue Comment (for issue events)
      if: github.event_name == 'issues'
      run: |
        # Generate safe output for issue comment
        echo '{"type": "add-issue-comment", "body": "## 🤖 AI Analysis\n\nI'\''ve analyzed this issue using GitHub'\''s AI models. Here'\''s my assessment:\n\n${{ steps.ai_inference.outputs.response }}\n\n---\n*This response was generated using GitHub Models via the AI Inference action.*"}' >> $GITHUB_AW_SAFE_OUTPUTS

safe-outputs:
  add-issue-comment:
    max: 1
    target: "*"
---

# AI Inference with GitHub Models

This agentic workflow demonstrates integration with GitHub's AI inference capabilities using the `actions/ai-inference` custom action and GitHub Models. It provides AI-powered responses to GitHub events like new issues and supports manual execution with custom prompts.

## Purpose

This workflow showcases how to leverage GitHub's AI infrastructure within the agentic workflow framework, demonstrating:
- Integration with GitHub Models through the AI inference action
- Safe output handling for AI-generated responses  
- Event-driven AI analysis of repository activity
- Configurable AI model selection and parameters

## Features

- **Multi-Model Support**: Configure different GitHub Models (GPT-4o-mini, GPT-4o, GPT-3.5-turbo)
- **Event-Driven**: Automatically responds to new issues with AI analysis
- **Manual Execution**: Supports custom prompts via workflow dispatch
- **Safe Outputs**: Uses safe output system to post AI responses as issue comments
- **Configurable Parameters**: Adjustable temperature (0.7) and max_tokens (1000)

## Usage Examples

### Manual Execution with Custom Prompt
1. Navigate to Actions tab in your repository
2. Select "AI Inference with GitHub Models" workflow  
3. Click "Run workflow"
4. Configure inputs:
   - **Prompt**: "Analyze the security implications of using environment variables in GitHub Actions"
   - **Model**: "gpt-4o" (for more advanced analysis)
5. Click "Run workflow" to execute

### Automatic Issue Analysis
When a new issue is created, the workflow automatically:
- Analyzes the issue title and description using AI
- Generates potential root causes and resolution suggestions  
- Posts the AI analysis as a comment on the issue via safe outputs

## Available Models

- **`gpt-4o-mini`** (default) - Fast and cost-effective for most tasks
- **`gpt-4o`** - More capable model for complex analysis and reasoning
- **`gpt-3.5-turbo`** - Basic analysis and simple responses

## Configuration Options

### Model Parameters
- **Temperature**: 0.7 (controls response creativity/randomness)
- **Max Tokens**: 1000 (maximum response length)
- **Model**: Configurable via workflow dispatch input

### Customization
Edit this workflow file to:
- Modify default prompts for different events
- Change model parameters (temperature, max_tokens)  
- Add additional processing steps
- Adjust safe output configurations
- Add new trigger events

## Implementation Details

### Custom Engine
Uses the custom engine with these key components:
- **AI Inference Action**: `actions/ai-inference@v1` for GitHub Models integration
- **Prompt File**: Uses `${{ env.GITHUB_AW_PROMPT }}` for dynamic prompt handling
- **Response Processing**: Formats AI output and creates GitHub Actions summaries

### Safe Outputs
Implements safe output handling for:
- **add-issue-comment**: Posts AI responses as issue comments (max: 1)
- **Target**: "*" (applies to any issue)

### Security Features
- **Minimal Permissions**: Only `contents: read` and `models: read`
- **GitHub's AI Infrastructure**: Uses secure GitHub Models backend
- **Content Filtering**: Built-in safety measures in GitHub Models
- **Safe Output System**: Prevents unauthorized repository actions