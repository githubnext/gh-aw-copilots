/**
 * Sanitization function for adversarial LLM outputs with enhanced security
 * Provides XML character escaping, URI filtering, and domain allowlisting
 */

function sanitizeContent(content, options = {}) {
  if (!content || typeof content !== 'string') {
    return '';
  }

  // Default allowed domains for GitHub infrastructure
  const defaultAllowedDomains = [
    'github.com',
    'github.io',
    'githubusercontent.com',
    'githubassets.com',
    'github.dev',
    'codespaces.new'
  ];

  const allowedDomains = options.allowedDomains || defaultAllowedDomains;

  let sanitized = content;

  // Remove control characters (except newlines and tabs)
  sanitized = sanitized.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '');

  // XML character escaping
  sanitized = sanitized
    .replace(/&/g, '&amp;')   // Must be first to avoid double-escaping
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');

  // URI filtering - replace non-https protocols with "(redacted)"
  // This regex matches common URI schemes at word boundaries
  sanitized = sanitized.replace(/\b(?:http|ftp|file|data|javascript|vbscript|mailto|tel|ssh|ldap|jdbc|chrome|edge|safari|firefox|opera):[^\s\])}'"<>&\x00-\x1f]+/gi, '(redacted)');

  // Domain filtering for HTTPS URIs
  // Match https:// URIs and check if domain is in allowlist
  sanitized = sanitized.replace(/\bhttps:\/\/([^\/\s\])}'"<>&\x00-\x1f]+)/gi, (match, domain) => {
    // Extract the hostname part (before first slash, colon, or other delimiter)
    const hostname = domain.split(/[\/:\?#]/)[0].toLowerCase();
    
    // Check if this domain or any parent domain is in the allowlist
    const isAllowed = allowedDomains.some(allowedDomain => {
      const normalizedAllowed = allowedDomain.toLowerCase();
      return hostname === normalizedAllowed || hostname.endsWith('.' + normalizedAllowed);
    });

    return isAllowed ? match : '(redacted)';
  });

  // Limit total length to prevent DoS (0.5MB max)
  const maxLength = 524288;
  if (sanitized.length > maxLength) {
    sanitized = sanitized.substring(0, maxLength) + '\n[Content truncated due to length]';
  }

  // Limit number of lines to prevent log flooding (65k max)
  const lines = sanitized.split('\n');
  const maxLines = 65000;
  if (lines.length > maxLines) {
    sanitized = lines.slice(0, maxLines).join('\n') + '\n[Content truncated due to line count]';
  }

  // Remove ANSI escape sequences
  sanitized = sanitized.replace(/\x1b\[[0-9;]*[mGKH]/g, '');

  // Trim excessive whitespace
  return sanitized.trim();
}

module.exports = { sanitizeContent };