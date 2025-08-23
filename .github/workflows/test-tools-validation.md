---
engine: codex
on: push
max-turns: 0
tools:
  - description: "First tool without name"
    type: file
    allowed: ["read", "write"]
  - name: github
    type: api
    allowed: [create_issue]
  - type: shell
    allowed: ["git", "curl"]
  - description: "Another unnamed tool"
    type: http
    allowed: ["GET", "POST"]
---

# Multiple Tools Validation Errors

This workflow tests validation of tools array with multiple missing name fields.

## Expected Validation Errors

1. **Max-turns Error**: Line 4
   - Path: `max-turns`
   - Message: `max-turns must be between 1 and 100, got 0`
   - Hint: `max-turns should be a number between 1 and 100`

2. **Tools[0] Name Error**: Around Line 6
   - Path: `tools[0].name`
   - Message: `tool must have a 'name' field`
   - Hint: `Each tool must have a 'name' field specifying the tool identifier`

3. **Tools[2] Name Error**: Around Line 12
   - Path: `tools[2].name`
   - Message: `tool must have a 'name' field`
   - Hint: `Each tool must have a 'name' field specifying the tool identifier`

4. **Tools[3] Name Error**: Around Line 15
   - Path: `tools[3].name`
   - Message: `tool must have a 'name' field`
   - Hint: `Each tool must have a 'name' field specifying the tool identifier`

## Job Description

This workflow tests complex array validation scenarios with multiple missing name fields in the tools array.