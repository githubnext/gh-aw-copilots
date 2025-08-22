package sanitizer

import (
	"strings"
	"testing"
)

func TestFilterURLs(t *testing.T) {
	tests := []struct {
		name                string
		content             string
		config              *FilterURLsConfig
		expectedContent     string
		expectedRemovedURLs []string
	}{
		{
			name:                "empty content",
			content:             "",
			config:              nil,
			expectedContent:     "",
			expectedRemovedURLs: nil,
		},
		{
			name:                "no URLs",
			content:             "This is just plain text without any URLs.",
			config:              nil,
			expectedContent:     "This is just plain text without any URLs.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "HTTPS URL with no domain restrictions",
			content: "Check out https://example.com for more info.",
			config:  nil,
			expectedContent:     "Check out https://example.com for more info.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "HTTP URL always filtered",
			content: "Visit http://example.com for details.",
			config:  nil,
			expectedContent:     "Visit [filtered] for details.",
			expectedRemovedURLs: []string{"http://example.com"},
		},
		{
			name:    "FTP URL always filtered",
			content: "Download from ftp://files.example.com/file.zip",
			config:  nil,
			expectedContent:     "Download from [filtered]",
			expectedRemovedURLs: []string{"ftp://files.example.com/file.zip"},
		},
		{
			name:    "HTTPS URL with allowed domain - exact match",
			content: "Visit https://github.com for code.",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Visit https://github.com for code.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "HTTPS URL with allowed domain - subdomain match",
			content: "API at https://api.github.com is available.",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "API at https://api.github.com is available.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "HTTPS URL with disallowed domain",
			content: "Don't visit https://malicious.example.com",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com", "example.org"}},
			expectedContent:     "Don't visit [filtered]",
			expectedRemovedURLs: []string{"https://malicious.example.com"},
		},
		{
			name:    "markdown link with allowed URL",
			content: "Check out [GitHub](https://github.com) for code repositories.",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Check out [GitHub](https://github.com) for code repositories.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "markdown link with disallowed URL",
			content: "Don't click [this link](https://malicious.example.com).",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Don't click this link [filtered].",
			expectedRemovedURLs: []string{"https://malicious.example.com"},
		},
		{
			name:    "markdown link with empty text and disallowed URL",
			content: "Visit [](https://malicious.example.com) for details.",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Visit [filtered] for details.",
			expectedRemovedURLs: []string{"https://malicious.example.com"},
		},
		{
			name:    "markdown link with HTTP URL",
			content: "Old site: [Legacy Site](http://old.example.com)",
			config:  &FilterURLsConfig{AllowDomains: []string{"example.com"}},
			expectedContent:     "Old site: Legacy Site [filtered]",
			expectedRemovedURLs: []string{"http://old.example.com"},
		},
		{
			name:    "mixed URLs in content",
			content: "Visit https://github.com and also check http://example.com and [link](https://api.github.com)",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Visit https://github.com and also check [filtered] and [link](https://api.github.com)",
			expectedRemovedURLs: []string{"http://example.com"},
		},
		{
			name:    "multiple disallowed URLs",
			content: "Don't visit https://bad1.com or https://bad2.com or [bad link](https://bad3.com)",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Don't visit [filtered] or [filtered] or bad link [filtered]",
			expectedRemovedURLs: []string{"https://bad3.com", "https://bad1.com", "https://bad2.com"},
		},
		{
			name:    "case insensitive domain matching",
			content: "Visit https://API.GITHUB.COM for details.",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "Visit https://API.GITHUB.COM for details.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "case insensitive allowed domains",
			content: "Visit https://api.github.com for details.",
			config:  &FilterURLsConfig{AllowDomains: []string{"GITHUB.COM"}},
			expectedContent:     "Visit https://api.github.com for details.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "URLs with paths and query params",
			content: "API: https://api.github.com/repos/owner/repo?param=value",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			expectedContent:     "API: https://api.github.com/repos/owner/repo?param=value",
			expectedRemovedURLs: nil,
		},
		{
			name:    "invalid URLs should be filtered due to non-HTTPS",
			content: "Invalid URL: http://example.com",
			config:  &FilterURLsConfig{AllowDomains: []string{"example.com"}},
			expectedContent:     "Invalid URL: [filtered]",
			expectedRemovedURLs: []string{"http://example.com"},
		},
		{
			name:    "multiple allowed domains",
			content: "Visit https://github.com and https://example.org but not https://bad.com",
			config:  &FilterURLsConfig{AllowDomains: []string{"github.com", "example.org"}},
			expectedContent:     "Visit https://github.com and https://example.org but not [filtered]",
			expectedRemovedURLs: []string{"https://bad.com"},
		},
		{
			name:    "empty allowed domains list",
			content: "Visit https://example.com for details.",
			config:  &FilterURLsConfig{AllowDomains: []string{}},
			expectedContent:     "Visit https://example.com for details.",
			expectedRemovedURLs: nil,
		},
		{
			name:    "whitespace in allowed domains",
			content: "Visit https://example.com for details.",
			config:  &FilterURLsConfig{AllowDomains: []string{" example.com ", "\tgithub.com\n"}},
			expectedContent:     "Visit https://example.com for details.",
			expectedRemovedURLs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterURLs(tt.content, tt.config)
			
			if result.FilteredContent != tt.expectedContent {
				t.Errorf("FilterURLs() filteredContent = %q, want %q", result.FilteredContent, tt.expectedContent)
			}
			
			if len(result.RemovedURLs) != len(tt.expectedRemovedURLs) {
				t.Errorf("FilterURLs() removedURLs count = %d, want %d", len(result.RemovedURLs), len(tt.expectedRemovedURLs))
				t.Errorf("Got: %v", result.RemovedURLs)
				t.Errorf("Want: %v", tt.expectedRemovedURLs)
				return
			}
			
			for i, removed := range result.RemovedURLs {
				if removed != tt.expectedRemovedURLs[i] {
					t.Errorf("FilterURLs() removedURLs[%d] = %q, want %q", i, removed, tt.expectedRemovedURLs[i])
				}
			}
		})
	}
}

func TestShouldFilterURL(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		config      *FilterURLsConfig
		shouldFilter bool
	}{
		{
			name:        "HTTPS URL with no restrictions",
			rawURL:      "https://example.com",
			config:      nil,
			shouldFilter: false,
		},
		{
			name:        "HTTP URL always filtered",
			rawURL:      "http://example.com",
			config:      nil,
			shouldFilter: true,
		},
		{
			name:        "FTP URL always filtered",
			rawURL:      "ftp://files.example.com",
			config:      nil,
			shouldFilter: true,
		},
		{
			name:        "invalid URL filtered",
			rawURL:      "not-a-url",
			config:      nil,
			shouldFilter: true,
		},
		{
			name:        "HTTPS URL with allowed domain",
			rawURL:      "https://github.com",
			config:      &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			shouldFilter: false,
		},
		{
			name:        "HTTPS URL with disallowed domain",
			rawURL:      "https://malicious.com",
			config:      &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			shouldFilter: true,
		},
		{
			name:        "subdomain of allowed domain",
			rawURL:      "https://api.github.com",
			config:      &FilterURLsConfig{AllowDomains: []string{"github.com"}},
			shouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldFilterURL(tt.rawURL, tt.config)
			if result != tt.shouldFilter {
				t.Errorf("shouldFilterURL() = %v, want %v", result, tt.shouldFilter)
			}
		})
	}
}

func TestIsHostnameAllowed(t *testing.T) {
	tests := []struct {
		name           string
		hostname       string
		allowedDomains []string
		expected       bool
	}{
		{
			name:           "exact match",
			hostname:       "github.com",
			allowedDomains: []string{"github.com"},
			expected:       true,
		},
		{
			name:           "subdomain match",
			hostname:       "api.github.com",
			allowedDomains: []string{"github.com"},
			expected:       true,
		},
		{
			name:           "no match",
			hostname:       "malicious.com",
			allowedDomains: []string{"github.com"},
			expected:       false,
		},
		{
			name:           "case insensitive match",
			hostname:       "API.GITHUB.COM",
			allowedDomains: []string{"github.com"},
			expected:       true,
		},
		{
			name:           "multiple domains - first match",
			hostname:       "github.com",
			allowedDomains: []string{"github.com", "example.org"},
			expected:       true,
		},
		{
			name:           "multiple domains - second match",
			hostname:       "api.example.org",
			allowedDomains: []string{"github.com", "example.org"},
			expected:       true,
		},
		{
			name:           "no domains allowed",
			hostname:       "github.com",
			allowedDomains: []string{},
			expected:       false,
		},
		{
			name:           "empty hostname",
			hostname:       "",
			allowedDomains: []string{"github.com"},
			expected:       false,
		},
		{
			name:           "whitespace in allowed domains",
			hostname:       "github.com",
			allowedDomains: []string{" github.com ", "\texample.org\n"},
			expected:       true,
		},
		{
			name:           "partial domain match should not work",
			hostname:       "badgithub.com",
			allowedDomains: []string{"github.com"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHostnameAllowed(tt.hostname, tt.allowedDomains)
			if result != tt.expected {
				t.Errorf("isHostnameAllowed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateJavaScriptURLFilter(t *testing.T) {
	tests := []struct {
		name         string
		allowDomains []string
		shouldContain []string
	}{
		{
			name:         "basic JavaScript generation",
			allowDomains: []string{"github.com", "example.org"},
			shouldContain: []string{
				"function filterURLs(content, allowDomains)",
				"function shouldFilterURL(rawURL)",
				"markdownLinkRegex",
				"urlRegex",
				"return { filteredContent, removedURLs };",
			},
		},
		{
			name:         "empty domains",
			allowDomains: []string{},
			shouldContain: []string{
				"function filterURLs(content, allowDomains)",
				"if (!allowDomains || allowDomains.length === 0)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := GenerateJavaScriptURLFilter(tt.allowDomains)
			
			if js == "" {
				t.Error("GenerateJavaScriptURLFilter() returned empty string")
				return
			}
			
			for _, expected := range tt.shouldContain {
				if !strings.Contains(js, expected) {
					t.Errorf("GenerateJavaScriptURLFilter() does not contain %q", expected)
				}
			}
		})
	}
}