# URL Domain Filtering

The `allow-domains` feature provides security by filtering URLs in generated agentic output (issue bodies, comments, PR descriptions).

## Behavior

- **Always**: Remove any URL that is not HTTPS
- **No restrictions**: If `allow-domains` is not configured, all HTTPS URLs are preserved
- **Domain restrictions**: If `allow-domains` is configured, only HTTPS URLs matching allowed domain patterns are preserved

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