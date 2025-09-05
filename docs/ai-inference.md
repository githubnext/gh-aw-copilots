# AI Inference with GitHub Models - Example Usage

This document provides examples of how to use the `ai-inference-github-models` agentic workflow.

## Workflow File Location
`.github/workflows/ai-inference-github-models.md`

## Example Usage Scenarios

### 1. Manual Execution with Custom Prompt
To run the workflow manually with a custom prompt:

1. Navigate to Actions tab in your repository
2. Select "AI Inference with GitHub Models" workflow
3. Click "Run workflow"
4. Configure inputs:
   - **Prompt**: "Analyze the security implications of using environment variables in GitHub Actions"
   - **Model**: "gpt-4o" (for more advanced analysis)
5. Click "Run workflow"

### 2. Automatic Issue Analysis
When a new issue is created, the workflow automatically:
- Analyzes the issue title and description
- Provides potential root causes and resolution suggestions
- Posts the AI analysis as a comment on the issue

## Configuration Options

### Available Models
- `gpt-4o-mini` (default) - Fast and cost-effective
- `gpt-4o` - More capable for complex tasks
- `gpt-3.5-turbo` - Basic analysis

### Customization
Edit `.github/workflows/ai-inference-github-models.md` to:
- Modify the default prompts for different events
- Change model parameters (temperature, max_tokens)
- Add additional processing steps
- Adjust safe output configurations

## Security Features
- Uses GitHub's secure AI inference infrastructure
- Implements safe outputs to prevent unauthorized actions
- Applies content filtering and safety measures
- Scoped permissions for minimal access requirements