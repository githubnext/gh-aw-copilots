---
name: AI Inference with GitHub Models
on:
  workflow_dispatch:
    inputs:
      issue_number:
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
          ISSUE_NUMBER="${{ github.event.inputs.issue_number }}"
        fi
        
        # Generate safe output for issue comment
        echo "{\"type\": \"add-issue-comment\", \"issue_number\": \"$ISSUE_NUMBER\", \"body\": \"## 🤖 AI Analysis\\n\\nI've analyzed this issue using GitHub's AI models. Here's my assessment:\\n\\n${{ steps.ai_inference.outputs.response }}\\n\\n---\\n*This response was generated using GitHub Models via the AI Inference action.*\"}" >> $GITHUB_AW_SAFE_OUTPUTS

safe-outputs:
  add-issue-comment:
    max: 1
    target: "*"
---

Summarize the issue #${{ github.event.issue.number }}