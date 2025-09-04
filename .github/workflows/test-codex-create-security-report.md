---
name: Test Codex Security Report
on:
  issues:
    types: [opened, reopened]
  reaction: eyes

engine: 
  id: codex

safe-outputs:
  create-security-report:
    max: 10
---

If the title of the issue #${{ github.event.issue.number }} contains "Security Scan Codex" then analyze the repository codebase for security vulnerabilities and create security reports.

For each security finding you identify, specify:
- The file path relative to the repository root
- The line number where the issue occurs
- Optional column number for precise location
- The severity level (error, warning, info, or note)
- A detailed description of the security issue
- Optionally, a custom rule ID suffix for meaningful SARIF rule identifiers

Focus on common security issues like:
- Hardcoded secrets or credentials
- SQL injection vulnerabilities
- Cross-site scripting (XSS) issues
- Insecure file operations
- Authentication bypasses
- Input validation problems
