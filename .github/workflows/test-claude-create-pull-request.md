---
on:
  workflow_dispatch:

engine: 
  id: claude

safe-outputs:
  create-pull-request:
    title-prefix: "[claude-test] "
    labels: [claude, automation, bot]
---

Add a file "TEST.md" with content "Hello from Claude"

Add a log file "foo.log" containing the current time. This is just a log file and isn't meant to go in the pull request.

Create a pull request with title "A Pull Request from Claude" and body "World"

Add a haiku about GitHub Actions and AI to the PR body.