---
engine: invalid-ai-engine
on:
  schedule:
    - cron: "0 9 * * 1"
max-turns: 150
timeout_minutes: 30
permissions:
  issues: write
  contents: read
tools:
  - name: github
    allowed: [create_issue]
  - type: shell
    allowed: ["curl", "wget"]
---

# Test Workflow with Schema Validation Errors

This workflow intentionally contains validation errors for testing purposes:

1. Invalid engine: `invalid-ai-engine` (should be `claude` or `codex`)
2. Max-turns too high: `150` (should be between 1-100)
3. Tool missing name: The shell tool has no `name` field

## Job Description

This is a test workflow designed to validate error reporting and source location mapping.

The workflow should fail validation and report precise error locations using the JSONPath to source span mapping system.

### Expected Validation Errors

1. **Engine Error**: Line 2, Column 9-26
   - Path: `engine`
   - Message: `unsupported engine 'invalid-ai-engine', must be one of: claude, codex`
   - Hint: `Supported engines: claude, codex`

2. **Max-turns Error**: Line 5, Column 12-14  
   - Path: `max-turns`
   - Message: `max-turns must be between 1 and 100, got 150`
   - Hint: `max-turns should be a number between 1 and 100`

3. **Tools Name Error**: Around Line 13
   - Path: `tools[1].name`
   - Message: `tool must have a 'name' field`
   - Hint: `Each tool must have a 'name' field specifying the tool identifier`

This workflow demonstrates the schema validation and error location mapping capabilities of the gh-aw system.