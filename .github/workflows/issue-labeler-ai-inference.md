---
on:
  issues:
    types: [opened]
  reaction: eyes

permissions:
  issues: read
  models: read

safe-outputs:
  add-issue-label:
    max: 5

engine:
  id: custom
  steps:
    - name: Analyze Issue with AI Inference
      id: analyze_issue
      uses: actions/ai-inference@v1
      with:
        model: gpt-4
        prompt: |
          Analyze this GitHub issue and suggest appropriate labels based on the content.
          
          Issue Title: ${{ github.event.issue.title }}
          Issue Body: ${{ github.event.issue.body }}
          
          Based on the content, suggest 1-3 relevant labels from common categories like:
          - Type: bug, feature, enhancement, documentation, question
          - Priority: low, medium, high, critical
          - Component: frontend, backend, api, ui, tests, ci
          - Status: needs-triage, needs-info, ready-to-work
          
          Respond with only a JSON array of label names, no additional text.
          Example: ["bug", "high", "backend"]
        temperature: 0.1

    - name: Generate Issue Labels Output
      run: |
        # Get the AI inference result
        AI_LABELS='${{ steps.analyze_issue.outputs.response }}'
        
        # Parse the JSON array and create the safe output
        echo "AI suggested labels: $AI_LABELS"
        
        # Extract labels from JSON array and create safe output
        if [ -n "$AI_LABELS" ] && [ "$AI_LABELS" != "null" ]; then
          # Create the safe output for adding labels
          echo '{"type": "add-issue-label", "labels": '$AI_LABELS'}' >> $GITHUB_AW_SAFE_OUTPUTS
          echo "Generated safe output for labels: $AI_LABELS"
        else
          echo "No valid labels received from AI inference, adding default triage label"
          echo '{"type": "add-issue-label", "labels": ["needs-triage"]}' >> $GITHUB_AW_SAFE_OUTPUTS
        fi

    - name: Log Analysis Results
      run: |
        echo "Issue analysis completed for issue #${{ github.event.issue.number }}"
        echo "Issue Title: ${{ github.event.issue.title }}"
        echo "AI Response: ${{ steps.analyze_issue.outputs.response }}"
        
        # Display the safe outputs file content
        if [ -f "$GITHUB_AW_SAFE_OUTPUTS" ]; then
          echo "Generated safe outputs:"
          cat "$GITHUB_AW_SAFE_OUTPUTS"
        else
          echo "No safe outputs file found"
        fi
---

# AI-Powered Issue Labeling Workflow

This workflow automatically analyzes newly created GitHub issues using the `actions/ai-inference` action and applies appropriate labels based on the AI's analysis of the issue content.

## How it Works

1. **Trigger**: Runs when a new issue is opened
2. **AI Analysis**: Uses `actions/ai-inference` with GPT-4 to analyze the issue title and body
3. **Label Generation**: The AI suggests relevant labels from common categories
4. **Automatic Labeling**: Uses safe outputs to automatically apply the suggested labels

## AI Categories

The AI is prompted to suggest labels from these categories:
- **Type**: bug, feature, enhancement, documentation, question
- **Priority**: low, medium, high, critical  
- **Component**: frontend, backend, api, ui, tests, ci
- **Status**: needs-triage, needs-info, ready-to-work

## Permissions Required

- `issues: read` - To analyze issue content
- `models: read` - Required for `actions/ai-inference` to access LLM models

## Fallback Behavior

If the AI inference fails or returns invalid results, the workflow will automatically add a `needs-triage` label to ensure the issue is flagged for human review.