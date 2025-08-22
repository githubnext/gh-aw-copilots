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
			name: "URL filtering with no output.allowed-domains config",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
output:
  comment: {}
---`,
			markdown: "# Test\nThis workflow tests URL filtering.",
			expectedInOutput: []string{
				"GITHUB_AW_AGENT_OUTPUT:", // Environment variable should be set
				"actions/github-script@v7", // Should use github-script action
				"function filterURLs(content, allowDomains)", // Should have embedded filtering function
				"const defaultGitHubDomains = ['github.com', 'github.io', 'githubusercontent.com', 'githubassets.com', 'githubapp.com', 'github.dev']",
			},
			notInOutput: []string{
				"GH_AW_ALLOW_DOMAINS:", // Should not have custom domains env var when not configured
			},
		},
		{
			name: "URL filtering with output.allowed-domains config",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
output:
  comment: {}
  allowed-domains:
    - github.com
    - example.org
---`,
			markdown: "# Test\nThis workflow tests URL filtering with domains.",
			expectedInOutput: []string{
				"GITHUB_AW_AGENT_OUTPUT:",
				"GH_AW_ALLOW_DOMAINS: \"github.com,example.org\"", // Should set custom domains env var
				"function filterURLs(content, allowDomains)",
				"const defaultGitHubDomains = ['github.com', 'github.io', 'githubusercontent.com', 'githubassets.com', 'githubapp.com', 'github.dev']",
			},
			notInOutput: []string{},
		},
		{
			name: "URL filtering with single output.allowed-domain",
			frontmatter: `---
name: Test Workflow
on: push
engine: claude
output:
  issue: {}
  allowed-domains: github.com
---`,
			markdown: "# Test\nThis workflow tests URL filtering with single domain.",
			expectedInOutput: []string{
				"GITHUB_AW_AGENT_OUTPUT:",
				"GH_AW_ALLOW_DOMAINS: \"github.com\"", // Should set single domain env var
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
		name            string
		frontmatter     map[string]any
		expectedDomains []string
	}{
		{
			name:            "no output.allowed-domains",
			frontmatter:     map[string]any{},
			expectedDomains: nil,
		},
		{
			name: "single string domain in output",
			frontmatter: map[string]any{
				"output": map[string]any{
					"allowed-domains": "github.com",
				},
			},
			expectedDomains: []string{"github.com"},
		},
		{
			name: "array of domains in output",
			frontmatter: map[string]any{
				"output": map[string]any{
					"allowed-domains": []any{"github.com", "example.org"},
				},
			},
			expectedDomains: []string{"github.com", "example.org"},
		},
		{
			name: "empty array in output",
			frontmatter: map[string]any{
				"output": map[string]any{
					"allowed-domains": []any{},
				},
			},
			expectedDomains: nil,
		},
		{
			name: "output section without allowed-domains",
			frontmatter: map[string]any{
				"output": map[string]any{
					"comment": map[string]any{},
				},
			},
			expectedDomains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(false, "", "test")
			outputConfig := compiler.extractOutputConfig(tt.frontmatter)
			
			var result []string
			if outputConfig != nil && len(outputConfig.AllowedDomains) > 0 {
				result = outputConfig.AllowedDomains
			}

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
