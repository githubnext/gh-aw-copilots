package workflow

import (
	"strings"
	"testing"
)

func TestNetworkHookGenerator(t *testing.T) {
	generator := &NetworkHookGenerator{}

	t.Run("GenerateNetworkHookScript", func(t *testing.T) {
		allowedDomains := []string{"example.com", "*.trusted.com", "api.service.org"}
		script := generator.GenerateNetworkHookScript(allowedDomains)

		// Check that script contains the expected domains
		if !strings.Contains(script, `"example.com"`) {
			t.Error("Script should contain example.com")
		}
		if !strings.Contains(script, `"*.trusted.com"`) {
			t.Error("Script should contain *.trusted.com")
		}
		if !strings.Contains(script, `"api.service.org"`) {
			t.Error("Script should contain api.service.org")
		}

		// Check for required Python imports and functions
		if !strings.Contains(script, "import json") {
			t.Error("Script should import json")
		}
		if !strings.Contains(script, "import urllib.parse") {
			t.Error("Script should import urllib.parse")
		}
		if !strings.Contains(script, "def extract_domain") {
			t.Error("Script should define extract_domain function")
		}
		if !strings.Contains(script, "def is_domain_allowed") {
			t.Error("Script should define is_domain_allowed function")
		}
	})

	t.Run("GenerateNetworkHookWorkflowStep", func(t *testing.T) {
		allowedDomains := []string{"api.github.com", "*.trusted.com"}
		step := generator.GenerateNetworkHookWorkflowStep(allowedDomains)

		stepStr := strings.Join(step, "\n")

		// Check that the step contains proper YAML structure
		if !strings.Contains(stepStr, "name: Generate Network Permissions Hook") {
			t.Error("Step should have correct name")
		}
		if !strings.Contains(stepStr, "network_permissions.js") {
			t.Error("Step should create hook file in correct location")
		}
		if !strings.Contains(stepStr, "actions/github-script") {
			t.Error("Step should use actions/github-script")
		}

		// Check that domains are included in the hook
		if !strings.Contains(stepStr, "api.github.com") {
			t.Error("Step should contain api.github.com domain")
		}
		if !strings.Contains(stepStr, "*.trusted.com") {
			t.Error("Step should contain *.trusted.com domain")
		}
	})

	t.Run("EmptyDomainsGeneration", func(t *testing.T) {
		allowedDomains := []string{} // Empty list means deny-all
		script := generator.GenerateNetworkHookScript(allowedDomains)

		// Should still generate a valid script
		if !strings.Contains(script, "ALLOWED_DOMAINS = []") {
			t.Error("Script should handle empty domains list (deny-all policy)")
		}
		if !strings.Contains(script, "def is_domain_allowed") {
			t.Error("Script should still define required functions")
		}
	})
}

func TestShouldEnforceNetworkPermissions(t *testing.T) {
	t.Run("nil permissions", func(t *testing.T) {
		if ShouldEnforceNetworkPermissions(nil) {
			t.Error("Should not enforce permissions when nil")
		}
	})

	t.Run("valid permissions with domains", func(t *testing.T) {
		permissions := &NetworkPermissions{
			Allowed: []string{"example.com", "*.trusted.com"},
		}
		if !ShouldEnforceNetworkPermissions(permissions) {
			t.Error("Should enforce permissions when provided")
		}
	})

	t.Run("empty permissions (deny-all)", func(t *testing.T) {
		permissions := &NetworkPermissions{
			Allowed: []string{}, // Empty list means deny-all
		}
		if !ShouldEnforceNetworkPermissions(permissions) {
			t.Error("Should enforce permissions even with empty allowed list (deny-all policy)")
		}
	})
}

func TestGetAllowedDomains(t *testing.T) {
	t.Run("nil permissions", func(t *testing.T) {
		domains := GetAllowedDomains(nil)
		if domains == nil {
			t.Error("Should return default whitelist when permissions are nil")
		}
		if len(domains) == 0 {
			t.Error("Expected default whitelist domains for nil permissions, got empty list")
		}
	})

	t.Run("empty permissions (deny-all)", func(t *testing.T) {
		permissions := &NetworkPermissions{
			Allowed: []string{}, // Empty list means deny-all
		}
		domains := GetAllowedDomains(permissions)
		if domains == nil {
			t.Error("Should return empty slice, not nil, for deny-all policy")
		}
		if len(domains) != 0 {
			t.Errorf("Expected 0 domains for deny-all policy, got %d", len(domains))
		}
	})

	t.Run("valid permissions with domains", func(t *testing.T) {
		permissions := &NetworkPermissions{
			Allowed: []string{"example.com", "*.trusted.com", "api.service.org"},
		}
		domains := GetAllowedDomains(permissions)
		expectedDomains := []string{"example.com", "*.trusted.com", "api.service.org"}
		if len(domains) != len(expectedDomains) {
			t.Fatalf("Expected %d domains, got %d", len(expectedDomains), len(domains))
		}

		for i, expected := range expectedDomains {
			if domains[i] != expected {
				t.Errorf("Expected domain %d to be '%s', got '%s'", i, expected, domains[i])
			}
		}
	})
}

func TestDeprecatedHasNetworkPermissions(t *testing.T) {
	t.Run("deprecated function always returns false", func(t *testing.T) {
		// Test that the deprecated function always returns false
		if HasNetworkPermissions(nil) {
			t.Error("Deprecated HasNetworkPermissions should always return false")
		}

		config := &EngineConfig{ID: "claude"}
		if HasNetworkPermissions(config) {
			t.Error("Deprecated HasNetworkPermissions should always return false")
		}
	})
}

func TestEngineConfigParsing(t *testing.T) {
	compiler := &Compiler{}

	t.Run("ParseNetworkPermissions", func(t *testing.T) {
		frontmatter := map[string]any{
			"network": map[string]any{
				"allowed": []any{"example.com", "*.trusted.com", "api.service.org"},
			},
		}

		networkPermissions := compiler.extractNetworkPermissions(frontmatter)

		if networkPermissions == nil {
			t.Fatal("Network permissions should not be nil")
		}

		expectedDomains := []string{"example.com", "*.trusted.com", "api.service.org"}
		if len(networkPermissions.Allowed) != len(expectedDomains) {
			t.Fatalf("Expected %d domains, got %d", len(expectedDomains), len(networkPermissions.Allowed))
		}

		for i, expected := range expectedDomains {
			if networkPermissions.Allowed[i] != expected {
				t.Errorf("Expected domain %d to be '%s', got '%s'", i, expected, networkPermissions.Allowed[i])
			}
		}
	})

	t.Run("ParseWithoutNetworkPermissions", func(t *testing.T) {
		frontmatter := map[string]any{
			"engine": map[string]any{
				"id":    "claude",
				"model": "claude-3-5-sonnet-20241022",
			},
		}

		networkPermissions := compiler.extractNetworkPermissions(frontmatter)

		if networkPermissions != nil {
			t.Error("Network permissions should be nil when not specified")
		}
	})

	t.Run("ParseEmptyNetworkPermissions", func(t *testing.T) {
		frontmatter := map[string]any{
			"network": map[string]any{
				"allowed": []any{}, // Empty list means deny-all
			},
		}

		networkPermissions := compiler.extractNetworkPermissions(frontmatter)

		if networkPermissions == nil {
			t.Fatal("Network permissions should not be nil")
		}

		if len(networkPermissions.Allowed) != 0 {
			t.Errorf("Expected 0 domains for deny-all policy, got %d", len(networkPermissions.Allowed))
		}
	})
}
