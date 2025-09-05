package workflow

import (
	"strings"
	"testing"
)

func TestAllowedDomainsParsing(t *testing.T) {
	tests := []struct {
		name            string
		frontmatter     map[string]any
		expectedDomains []string
	}{
		{
			name: "no output config",
			frontmatter: map[string]any{
				"engine": "claude",
			},
			expectedDomains: nil,
		},
		{
			name: "output config with allowed-domains",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"allowed-domains": []any{"example.com", "trusted.org"},
				},
			},
			expectedDomains: []string{"example.com", "trusted.org"},
		},
		{
			name: "output config with create-issue and allowed-domains",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-issue": map[string]any{
						"title-prefix": "[auto] ",
					},
					"allowed-domains": []any{"github.com", "api.github.com"},
				},
			},
			expectedDomains: []string{"github.com", "api.github.com"},
		},
		{
			name: "output config without allowed-domains",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-issue": map[string]any{
						"title-prefix": "[auto] ",
					},
				},
			},
			expectedDomains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler(false, "", "test")
			config := c.extractSafeOutputsConfig(tt.frontmatter)

			if tt.expectedDomains == nil {
				if config == nil {
					return // expected case
				}
				if len(config.AllowedDomains) == 0 {
					return // expected case
				}
				t.Errorf("Expected no allowed domains, but got %v", config.AllowedDomains)
				return
			}

			if config == nil {
				t.Errorf("Expected output config, but got nil")
				return
			}

			if len(config.AllowedDomains) != len(tt.expectedDomains) {
				t.Errorf("Expected %d allowed domains, but got %d", len(tt.expectedDomains), len(config.AllowedDomains))
				return
			}

			for i, expected := range tt.expectedDomains {
				if config.AllowedDomains[i] != expected {
					t.Errorf("Expected domain %s at index %d, but got %s", expected, i, config.AllowedDomains[i])
				}
			}
		})
	}
}

func TestAllowedDomainsInWorkflow(t *testing.T) {
	// Create a test compiler with verbose output to check generated workflow
	c := NewCompiler(true, "", "test")

	// Test workflow with allowed domains
	frontmatter := map[string]any{
		"engine": "claude",
		"safe-outputs": map[string]any{
			"allowed-domains": []any{"example.com", "trusted.org"},
		},
	}

	config := c.extractSafeOutputsConfig(frontmatter)
	if config == nil {
		t.Fatal("Expected output config, but got nil")
	}

	if len(config.AllowedDomains) != 2 {
		t.Errorf("Expected 2 allowed domains, but got %d", len(config.AllowedDomains))
	}

	expectedDomains := []string{"example.com", "trusted.org"}
	for i, expected := range expectedDomains {
		if config.AllowedDomains[i] != expected {
			t.Errorf("Expected domain %s at index %d, but got %s", expected, i, config.AllowedDomains[i])
		}
	}
}

func TestSafeOutputsConfigGeneration(t *testing.T) {
	tests := []struct {
		name               string
		frontmatter        map[string]any
		expectedInConfig   []string
		unexpectedInConfig []string
	}{
		{
			name: "create-discussion config",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-discussion": map[string]any{
						"title-prefix": "[discussion] ",
						"max":          2,
					},
				},
			},
			expectedInConfig: []string{"create-discussion"},
		},
		{
			name: "create-pull-request-review-comment config",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-pull-request-review-comment": map[string]any{
						"max": 5,
					},
				},
			},
			expectedInConfig: []string{"create-pull-request-review-comment"},
		},
		{
			name: "create-security-report config",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-security-report": map[string]any{},
				},
			},
			expectedInConfig: []string{"create-security-report"},
		},
		{
			name: "multiple safe outputs including previously missing ones",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-issue":                       map[string]any{"max": 1},
					"create-discussion":                  map[string]any{"max": 3},
					"create-pull-request-review-comment": map[string]any{"max": 10},
					"create-security-report":             map[string]any{},
					"add-issue-comment":                  map[string]any{},
				},
			},
			expectedInConfig: []string{
				"create-issue",
				"create-discussion",
				"create-pull-request-review-comment",
				"create-security-report",
				"add-issue-comment",
			},
		},
		{
			name: "no safe outputs config",
			frontmatter: map[string]any{
				"engine": "claude",
			},
			expectedInConfig:   []string{},
			unexpectedInConfig: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(false, "", "test")
			config := compiler.extractSafeOutputsConfig(tt.frontmatter)

			// Test the config generation in generateOutputCollectionStep by creating a mock workflow data
			workflowData := &WorkflowData{
				SafeOutputs: config,
			}

			// Use the compiler's generateOutputCollectionStep to test the GITHUB_AW_SAFE_OUTPUTS_CONFIG generation
			var yamlBuilder strings.Builder
			compiler.generateOutputCollectionStep(&yamlBuilder, workflowData)
			generatedYAML := yamlBuilder.String()

			// Look specifically for the GITHUB_AW_SAFE_OUTPUTS_CONFIG environment variable line
			configLinePresent := strings.Contains(generatedYAML, "GITHUB_AW_SAFE_OUTPUTS_CONFIG:")

			if len(tt.expectedInConfig) > 0 {
				// If we expect items in config, the config line should be present
				if !configLinePresent {
					t.Errorf("Expected GITHUB_AW_SAFE_OUTPUTS_CONFIG environment variable to be present, but it was not found")
					return
				}

				// Extract the config line to check its contents
				configLine := ""
				lines := strings.Split(generatedYAML, "\n")
				for _, line := range lines {
					if strings.Contains(line, "GITHUB_AW_SAFE_OUTPUTS_CONFIG:") {
						configLine = line
						break
					}
				}

				// Check expected items are present in the config line
				for _, expected := range tt.expectedInConfig {
					if !strings.Contains(configLine, expected) {
						t.Errorf("Expected %q to be in GITHUB_AW_SAFE_OUTPUTS_CONFIG, but it was not found in config line: %s", expected, configLine)
					}
				}

				// Check unexpected items are not present in the config line
				for _, unexpected := range tt.unexpectedInConfig {
					if strings.Contains(configLine, unexpected) {
						t.Errorf("Did not expect %q to be in GITHUB_AW_SAFE_OUTPUTS_CONFIG, but it was found in config line: %s", unexpected, configLine)
					}
				}
			}
			// If we don't expect any items and no unexpected items specified,
			// the config line may or may not be present (depending on whether SafeOutputs is nil)
			// This is acceptable behavior
		})
	}
}

func TestCreateDiscussionConfigParsing(t *testing.T) {
	tests := []struct {
		name                string
		frontmatter         map[string]any
		expectedTitlePrefix string
		expectedCategoryId  string
		expectedMax         int
		expectConfig        bool
	}{
		{
			name: "no create-discussion config",
			frontmatter: map[string]any{
				"engine": "claude",
			},
			expectConfig: false,
		},
		{
			name: "basic create-discussion config",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-discussion": map[string]any{},
				},
			},
			expectedTitlePrefix: "",
			expectedCategoryId:  "",
			expectedMax:         1, // default
			expectConfig:        true,
		},
		{
			name: "create-discussion with title-prefix",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-discussion": map[string]any{
						"title-prefix": "[ai] ",
					},
				},
			},
			expectedTitlePrefix: "[ai] ",
			expectedCategoryId:  "",
			expectedMax:         1,
			expectConfig:        true,
		},
		{
			name: "create-discussion with category-id",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-discussion": map[string]any{
						"category-id": "DIC_kwDOGFsHUM4BsUn3",
					},
				},
			},
			expectedTitlePrefix: "",
			expectedCategoryId:  "DIC_kwDOGFsHUM4BsUn3",
			expectedMax:         1,
			expectConfig:        true,
		},
		{
			name: "create-discussion with all options",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-discussion": map[string]any{
						"title-prefix": "[research] ",
						"category-id":  "DIC_kwDOGFsHUM4BsUn3",
						"max":          3,
					},
				},
			},
			expectedTitlePrefix: "[research] ",
			expectedCategoryId:  "DIC_kwDOGFsHUM4BsUn3",
			expectedMax:         3,
			expectConfig:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler(false, "", "test")
			config := c.extractSafeOutputsConfig(tt.frontmatter)

			if !tt.expectConfig {
				if config != nil && config.CreateDiscussions != nil {
					t.Errorf("Expected no create-discussion config, but got one")
				}
				return
			}

			if config == nil || config.CreateDiscussions == nil {
				t.Errorf("Expected create-discussion config, but got nil")
				return
			}

			discussionConfig := config.CreateDiscussions

			if discussionConfig.TitlePrefix != tt.expectedTitlePrefix {
				t.Errorf("Expected title prefix %q, but got %q", tt.expectedTitlePrefix, discussionConfig.TitlePrefix)
			}

			if discussionConfig.CategoryId != tt.expectedCategoryId {
				t.Errorf("Expected category ID %q, but got %q", tt.expectedCategoryId, discussionConfig.CategoryId)
			}

			if discussionConfig.Max != tt.expectedMax {
				t.Errorf("Expected max %d, but got %d", tt.expectedMax, discussionConfig.Max)
			}
		})
	}
}
