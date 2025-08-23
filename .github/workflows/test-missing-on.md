---
# Missing required 'on' field completely
engine: claude
max-turns: 5
permissions:
  contents: read
tools:
  github:
    allowed: [create_comment]
---

# Missing Required Field Test

This workflow tests validation of missing required fields.

## Validation Error Expected

- **Missing 'on' field**
  - Path: `on`
  - Message: `missing required field 'on'`
  - Hint: `Add an 'on' field to specify when the workflow should run (e.g., 'on: push')`
  - No source span (field doesn't exist)

## Job Description

This workflow intentionally omits the required `on` field to test error handling for missing required fields in frontmatter validation.