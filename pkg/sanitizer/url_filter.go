package sanitizer

import (
	"net/url"
	"regexp"
	"strings"
)

// URLFilterResult holds the result of URL filtering operations
type URLFilterResult struct {
	FilteredContent string   // Content with URLs filtered
	RemovedURLs     []string // List of URLs that were removed for audit/logging
}

// FilterURLsConfig holds configuration for URL filtering
type FilterURLsConfig struct {
	AllowDomains []string // List of allowed domain patterns
}

// FilterURLs filters URLs in the content based on configuration
// Always removes non-HTTPS URLs
// If allowDomains is provided, only keeps HTTPS URLs matching allowed domain patterns
func FilterURLs(content string, config *FilterURLsConfig) *URLFilterResult {
	if content == "" {
		return &URLFilterResult{FilteredContent: "", RemovedURLs: nil}
	}

	var removedURLs []string
	filteredContent := content

	// Regular expressions for different URL formats
	// Match URLs in markdown links: [text](url)
	markdownLinkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
	// Match URLs (including all protocols)
	urlRegex := regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://[^\s<>"'\[\]{}()]+`)

	// First pass: handle markdown links
	filteredContent = markdownLinkRegex.ReplaceAllStringFunc(filteredContent, func(match string) string {
		submatches := markdownLinkRegex.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		linkText := submatches[1]
		linkURL := submatches[2]

		if shouldFilterURL(linkURL, config) {
			removedURLs = append(removedURLs, linkURL)
			// Convert to plain text with filtered indicator
			if linkText != "" {
				return linkText + " [filtered]"
			}
			return "[filtered]"
		}

		return match
	})

	// Second pass: handle plain URLs
	filteredContent = urlRegex.ReplaceAllStringFunc(filteredContent, func(match string) string {
		if shouldFilterURL(match, config) {
			removedURLs = append(removedURLs, match)
			return "[filtered]"
		}
		return match
	})

	return &URLFilterResult{
		FilteredContent: filteredContent,
		RemovedURLs:     removedURLs,
	}
}

// getDefaultAllowedDomains returns the default GitHub-owned domains
func getDefaultAllowedDomains() []string {
	return []string{
		"github.com",
		"github.io",
		"githubusercontent.com",
		"githubassets.com",
		"githubapp.com",
		"github.dev",
	}
}

// shouldFilterURL determines if a URL should be filtered based on the configuration
func shouldFilterURL(rawURL string, config *FilterURLsConfig) bool {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// If we can't parse it, filter it out for safety
		return true
	}

	// Always filter non-HTTPS URLs
	if parsedURL.Scheme != "https" {
		return true
	}

	// Get allowed domains - use default GitHub domains if none configured
	var allowedDomains []string
	if config == nil || len(config.AllowDomains) == 0 {
		allowedDomains = getDefaultAllowedDomains()
	} else {
		allowedDomains = config.AllowDomains
	}

	// Check if hostname matches any allowed domain pattern
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return true
	}

	return !isHostnameAllowed(hostname, allowedDomains)
}

// isHostnameAllowed checks if a hostname matches any of the allowed domain patterns
func isHostnameAllowed(hostname string, allowedDomains []string) bool {
	hostname = strings.ToLower(hostname)

	for _, allowedDomain := range allowedDomains {
		allowedDomain = strings.ToLower(strings.TrimSpace(allowedDomain))
		if allowedDomain == "" {
			continue
		}

		// Exact match
		if hostname == allowedDomain {
			return true
		}

		// Subdomain match (e.g., "example.com" matches "api.example.com")
		if strings.HasSuffix(hostname, "."+allowedDomain) {
			return true
		}
	}

	return false
}

// GenerateJavaScriptURLFilter generates JavaScript code for URL filtering
// This creates the JavaScript function that can be embedded in GitHub Actions workflows
func GenerateJavaScriptURLFilter(allowDomains []string) string {
	var js strings.Builder

	js.WriteString("// URL filtering function\n")
	js.WriteString("function filterURLs(content, allowDomains) {\n")
	js.WriteString("  if (!content || typeof content !== 'string') {\n")
	js.WriteString("    return { filteredContent: '', removedURLs: [] };\n")
	js.WriteString("  }\n")
	js.WriteString("  \n")
	js.WriteString("  let removedURLs = [];\n")
	js.WriteString("  let filteredContent = content;\n")
	js.WriteString("  \n")
	js.WriteString("  // Default GitHub-owned domains when no domains are configured\n")
	js.WriteString("  const defaultGitHubDomains = ['github.com', 'github.io', 'githubusercontent.com', 'githubassets.com', 'githubapp.com', 'github.dev'];\n")
	js.WriteString("  \n")
	js.WriteString("  // Helper function to determine if URL should be filtered\n")
	js.WriteString("  function shouldFilterURL(rawURL) {\n")
	js.WriteString("    try {\n")
	js.WriteString("      const url = new URL(rawURL);\n")
	js.WriteString("      \n")
	js.WriteString("      // Always filter non-HTTPS URLs\n")
	js.WriteString("      if (url.protocol !== 'https:') {\n")
	js.WriteString("        return true;\n")
	js.WriteString("      }\n")
	js.WriteString("      \n")
	js.WriteString("      // Use default GitHub domains if no domains are configured\n")
	js.WriteString("      const domainsToCheck = (allowDomains && allowDomains.length > 0) ? allowDomains : defaultGitHubDomains;\n")
	js.WriteString("      \n")
	js.WriteString("      // Check if hostname matches any allowed domain pattern\n")
	js.WriteString("      const hostname = url.hostname.toLowerCase();\n")
	js.WriteString("      if (!hostname) {\n")
	js.WriteString("        return true;\n")
	js.WriteString("      }\n")
	js.WriteString("      \n")
	js.WriteString("      for (const allowedDomain of domainsToCheck) {\n")
	js.WriteString("        const domain = allowedDomain.toLowerCase().trim();\n")
	js.WriteString("        if (!domain) continue;\n")
	js.WriteString("        \n")
	js.WriteString("        // Exact match\n")
	js.WriteString("        if (hostname === domain) {\n")
	js.WriteString("          return false;\n")
	js.WriteString("        }\n")
	js.WriteString("        \n")
	js.WriteString("        // Subdomain match\n")
	js.WriteString("        if (hostname.endsWith('.' + domain)) {\n")
	js.WriteString("          return false;\n")
	js.WriteString("        }\n")
	js.WriteString("      }\n")
	js.WriteString("      \n")
	js.WriteString("      return true;\n")
	js.WriteString("    } catch (error) {\n")
	js.WriteString("      // If we can't parse it, filter it out for safety\n")
	js.WriteString("      return true;\n")
	js.WriteString("    }\n")
	js.WriteString("  }\n")
	js.WriteString("  \n")
	js.WriteString("  // Handle markdown links: [text](url)\n")
	js.WriteString("  const markdownLinkRegex = /\\[([^\\]]*)\\]\\(([^)]+)\\)/g;\n")
	js.WriteString("  filteredContent = filteredContent.replace(markdownLinkRegex, (match, linkText, linkURL) => {\n")
	js.WriteString("    if (shouldFilterURL(linkURL)) {\n")
	js.WriteString("      removedURLs.push(linkURL);\n")
	js.WriteString("      return linkText ? linkText + ' [filtered]' : '[filtered]';\n")
	js.WriteString("    }\n")
	js.WriteString("    return match;\n")
	js.WriteString("  });\n")
	js.WriteString("  \n")
	js.WriteString("  // Handle plain URLs (including all protocols)\n")
	js.WriteString("  const urlRegex = /[a-zA-Z][a-zA-Z0-9+.-]*:\\/\\/[^\\s<>\"'\\[\\]{}()]+/g;\n")
	js.WriteString("  filteredContent = filteredContent.replace(urlRegex, (match) => {\n")
	js.WriteString("    if (shouldFilterURL(match)) {\n")
	js.WriteString("      removedURLs.push(match);\n")
	js.WriteString("      return '[filtered]';\n")
	js.WriteString("    }\n")
	js.WriteString("    return match;\n")
	js.WriteString("  });\n")
	js.WriteString("  \n")
	js.WriteString("  return { filteredContent, removedURLs };\n")
	js.WriteString("}\n")

	return js.String()
}
