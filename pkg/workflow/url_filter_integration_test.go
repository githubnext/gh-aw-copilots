package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestURLFilteringIntegration(t *testing.T) {
	tests := []struct {
		name             string
		frontmatter      string
		markdown         string
		expectedInOutput []string
		notInOutput      []string
	}{
		{
			name: "URL filtering with no allow-domains config",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
---`,
			markdown: "# Test\nThis workflow tests URL filtering.",
			expectedInOutput: []string{
				"const allowDomains = process.env.GH_AW_ALLOW_DOMAINS ? process.env.GH_AW_ALLOW_DOMAINS.split(',') : null;",
				"function filterURLs(content, allowDomains)",
				"const urlFilterResult = filterURLs(sanitized, allowDomains);",
				"sanitized = urlFilterResult.filteredContent;",
			},
			notInOutput: []string{},
		},
		{
			name: "URL filtering with allow-domains config",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
allow-domains:
  - github.com
  - example.org
---`,
			markdown: "# Test\nThis workflow tests URL filtering with domains.",
			expectedInOutput: []string{
				"const allowDomains = process.env.GH_AW_ALLOW_DOMAINS ? process.env.GH_AW_ALLOW_DOMAINS.split(',') : [\"github.com\",\"example.org\"];",
				"function filterURLs(content, allowDomains)",
				"const urlFilterResult = filterURLs(sanitized, allowDomains);",
				"console.log('Filtered URLs:', urlFilterResult.removedURLs);",
			},
			notInOutput: []string{},
		},
		{
			name: "URL filtering with single allow-domain",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
allow-domains: github.com
---`,
			markdown: "# Test\nThis workflow tests URL filtering with single domain.",
			expectedInOutput: []string{
				"const allowDomains = process.env.GH_AW_ALLOW_DOMAINS ? process.env.GH_AW_ALLOW_DOMAINS.split(',') : [\"github.com\"];",
				"function filterURLs(content, allowDomains)",
			},
			notInOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "url-filter-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create .github/workflows directory
			workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
			err = os.MkdirAll(workflowsDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create workflows directory: %v", err)
			}

			// Write markdown file
			markdownPath := filepath.Join(workflowsDir, "test.md")
			content := tt.frontmatter + "\n\n" + tt.markdown
			err = os.WriteFile(markdownPath, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to write markdown file: %v", err)
			}

			// Create compiler and compile
			compiler := NewCompiler(false, "", "test")
			compiler.SetSkipValidation(true) // Skip validation for tests

			err = compiler.CompileWorkflow(markdownPath)
			if err != nil {
				t.Fatalf("Failed to compile workflow: %v", err)
			}

			// Read generated lock file
			lockFile := filepath.Join(workflowsDir, "test.lock.yml")
			lockContent, err := os.ReadFile(lockFile)
			if err != nil {
				t.Fatalf("Failed to read generated lock file: %v", err)
			}

			lockContentStr := string(lockContent)

			// Check expected content
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(lockContentStr, expected) {
					t.Errorf("Expected content not found in lock file: %q", expected)
				}
			}

			// Check content that should not be present
			for _, notExpected := range tt.notInOutput {
				if strings.Contains(lockContentStr, notExpected) {
					t.Errorf("Unexpected content found in lock file: %q", notExpected)
				}
			}
		})
	}
}

func TestAllowDomainsExtraction(t *testing.T) {
	tests := []struct {
		name           string
		frontmatter    map[string]any
		expectedDomains []string
	}{
		{
			name:           "no allow-domains",
			frontmatter:    map[string]any{},
			expectedDomains: nil,
		},
		{
			name: "single string domain",
			frontmatter: map[string]any{
				"allow-domains": "github.com",
			},
			expectedDomains: []string{"github.com"},
		},
		{
			name: "array of domains",
			frontmatter: map[string]any{
				"allow-domains": []any{"github.com", "example.org"},
			},
			expectedDomains: []string{"github.com", "example.org"},
		},
		{
			name: "empty array",
			frontmatter: map[string]any{
				"allow-domains": []any{},
			},
			expectedDomains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(false, "", "test")
			result := compiler.extractStringArray(tt.frontmatter, "allow-domains")
			
			if len(result) != len(tt.expectedDomains) {
				t.Errorf("Expected %d domains, got %d", len(tt.expectedDomains), len(result))
				return
			}
			
			for i, expected := range tt.expectedDomains {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected domain[%d] = %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}