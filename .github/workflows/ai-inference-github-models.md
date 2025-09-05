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

This agentic workflow demonstrates how to use the actions/ai-inference custom action with GitHub Models to provide AI-powered responses to various GitHub events.

## Features

- **Multi-Model Support**: Configure different GitHub Models (GPT-4o-mini, GPT-4, etc.)
- **Event-Driven**: Responds to workflow dispatch and new issues
- **Context-Aware**: Provides relevant AI responses based on the triggering event
- **Safe Outputs**: Automatically posts AI responses as comments on issues
- **Configurable**: Allows customization of prompts and models via workflow dispatch

## Usage

### Manual Execution
1. Go to the Actions tab in your repository
2. Select "AI Inference with GitHub Models"
3. Click "Run workflow"
4. Optionally customize the prompt and model
5. Click "Run workflow" to execute

### Automatic Execution
- **New Issues**: When an issue is opened, the AI analyzes it and provides suggestions

## Models Available
- `gpt-4o-mini` (default) - Fast and efficient for most tasks
- `gpt-4o` - More capable model for complex analysis
- `gpt-3.5-turbo` - Cost-effective option for simple tasks

## Customization

You can customize the behavior by:
1. Modifying the `model` parameter in the workflow dispatch
2. Updating the prompts in the engine steps
3. Adjusting the `max_tokens` and `temperature` parameters
4. Adding additional processing steps

## Security

This workflow uses:
- GitHub's secure AI inference infrastructure
- Safe outputs to prevent unauthorized actions
- Proper permission scoping for GitHub API access
- Content filtering and safety measures built into GitHub Models