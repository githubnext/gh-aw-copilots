const fs = require("fs");

function neutralizeMentions(s) {
  // Replace @name or @org/team outside code with `@name`
  return s.replace(/(^|[^\w`])@([A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?(?:\/[A-Za-z0-9._-]+)?)/g,
                   (_m, p1, p2) => `${p1}\`@${p2}\``);
}

function neutralizeBotTriggers(s) {
  // Neutralize common bot trigger phrases like "fixes #123", "closes #asdfs", etc.
  return s.replace(/\b(fixes?|closes?|resolves?|fix|close|resolve)\s+#(\w+)/gi, 
                   (match, action, ref) => `\`${action} #${ref}\``);
}

function sanitizeContent(content) {
  if (!content || typeof content !== 'string') {
    return '';
  }

  // Read allowed domains from environment variable
  const allowedDomainsEnv = process.env.GITHUB_AW_ALLOWED_DOMAINS;
  const defaultAllowedDomains = [
    'github.com',
    'github.io',
    'githubusercontent.com',
    'githubassets.com',
    'github.dev',
    'codespaces.new'
  ];

  const allowedDomains = allowedDomainsEnv 
    ? allowedDomainsEnv.split(',').map(d => d.trim()).filter(d => d)
    : defaultAllowedDomains;

  let sanitized = content;

  // Neutralize @mentions to prevent unintended notifications
  sanitized = neutralizeMentions(sanitized);

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
  // Step 1: Temporarily mark HTTPS URLs to protect them
  const httpsPlaceholder = '___HTTPS_PLACEHOLDER___';
  const httpsUrls = [];
  sanitized = sanitized.replace(/\bhttps:\/\/[^\s\])}'"<>&\x00-\x1f]+/gi, (match) => {
    httpsUrls.push(match);
    return httpsPlaceholder + (httpsUrls.length - 1);
  });
  
  // Step 2: Replace other protocols with "(redacted)"
  sanitized = sanitized.replace(/\b(?:http:\/\/|ftp:|file:|data:|javascript:|vbscript:|mailto:|tel:|ssh:|ldap:|jdbc:|chrome:|edge:|safari:|firefox:|opera:)[^\s\])}'"<>&\x00-\x1f]+/gi, '(redacted)');
  
  // Step 3: Restore HTTPS URLs
  sanitized = sanitized.replace(new RegExp(httpsPlaceholder + '(\\d+)', 'g'), (match, index) => {
    return httpsUrls[parseInt(index)];
  });

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

  // Neutralize common bot trigger phrases
  sanitized = neutralizeBotTriggers(sanitized);

  // Trim excessive whitespace
  return sanitized.trim();
}

async function main() {
  const outputFile = process.env.GITHUB_AW_OUTPUT;
  if (!outputFile) {
    console.log('GITHUB_AW_OUTPUT not set, no output to collect');
    core.setOutput('output', '');
    return;
  }
  
  if (!fs.existsSync(outputFile)) {
    console.log('Output file does not exist:', outputFile);
    core.setOutput('output', '');
    return;
  }
  
  const outputContent = fs.readFileSync(outputFile, 'utf8');
  if (outputContent.trim() === '') {
    console.log('Output file is empty');
    core.setOutput('output', '');
  } else {
    const sanitizedContent = sanitizeContent(outputContent);
    console.log('Collected agentic output (sanitized):', sanitizedContent.substring(0, 200) + (sanitizedContent.length > 200 ? '...' : ''));
    core.setOutput('output', sanitizedContent);
  }
}

await main();