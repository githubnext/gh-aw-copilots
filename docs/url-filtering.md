# URL Domain Filtering

The `allow-domains` feature provides security by filtering URLs in generated agentic output (issue bodies, comments, PR descriptions).

## Behavior

- **Always**: Remove any URL that is not HTTPS
- **Default restrictions**: If `allow-domains` is not configured, only GitHub-owned domains (`github.com`, `github.io`, `githubusercontent.com`, `githubassets.com`, `githubapp.com`, `github.dev`) are allowed
- **Custom restrictions**: If `allow-domains` is configured, only HTTPS URLs matching those domain patterns are preserved

## Configuration

### Frontmatter Configuration

```yaml
---
name: My Secure Workflow
on: push
engine: claude
allow-domains:
  - github.com
  - example.org
---
```

### Single Domain

```yaml
allow-domains: github.com
```

### Environment Variable Override

Set the `GH_AW_ALLOW_DOMAINS` environment variable (comma-separated) to override frontmatter settings:

```bash
export GH_AW_ALLOW_DOMAINS="github.com,example.org"
```

## Examples

### Input Content
```
Visit https://github.com for code and http://malicious.com for bad stuff.
Also check [API docs](https://api.github.com/docs) and [bad link](https://evil.example.com).
```

### With default behavior (no `allow-domains` configured)
```
Visit https://github.com for code and [filtered] for bad stuff.
Also check [API docs](https://api.github.com/docs) and bad link [filtered].
```

### With `allow-domains: [github.com]`
```
Visit https://github.com for code and [filtered] for bad stuff.
Also check [API docs](https://api.github.com/docs) and bad link [filtered].
```

### Domain Matching Rules

- **Exact match**: `github.com` matches `https://github.com`
- **Subdomain match**: `github.com` matches `https://api.github.com`
- **Case insensitive**: `GitHub.COM` matches `github.com`
- **No partial match**: `github.com` does NOT match `https://mygithub.com`

## Filtered URL Logging

Filtered URLs are logged to the workflow console for audit purposes:
```
Filtered URLs: ['http://malicious.com', 'https://evil.example.com']
```