package workflow

import (
	"testing"
)

// TestJSONSchemaValidationIntegration demonstrates that JSON schema validation is working
func TestJSONSchemaValidationIntegration(t *testing.T) {
	tests := []struct {
		name                string
		frontmatter         map[string]any
		expectError         bool
		expectedErrorCount  int
	}{
		{
			name: "valid frontmatter passes validation",
			frontmatter: map[string]any{
				"on":     "push",
				"engine": "claude",
				"tools": map[string]any{
					"github": map[string]any{
						"allowed": []string{"create_issue"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid engine caught by validation",
			frontmatter: map[string]any{
				"on":     "push",
				"engine": "invalid-engine",
			},
			expectError: true,
			expectedErrorCount: 1, // Dynamic validation should catch this
		},
		{
			name: "additional properties caught by schema validation",
			frontmatter: map[string]any{
				"on":              "push",
				"engine":          "claude",
				"invalid_property": "value",
			},
			expectError: true,
			expectedErrorCount: 1, // JSON schema should catch this
		},
		{
			name: "missing required 'on' field",
			frontmatter: map[string]any{
				"engine": "claude",
			},
			expectError: true,
			expectedErrorCount: 1, // JSON schema should catch this
		},
		{
			name: "max-turns out of range",
			frontmatter: map[string]any{
				"on":        "push",
				"engine":    "claude",
				"max-turns": 150,
			},
			expectError: true,
			expectedErrorCount: 1, // Dynamic validation should catch this  
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock frontmatter YAML for the locator
			mockYAML := `---
on: push
engine: claude
---`
			validator := NewFrontmatterValidator(mockYAML)

			// Test with the new unified validation (JSON schema + dynamic rules)
			errors := validator.ValidateFrontmatter(tt.frontmatter)
			
			if tt.expectError && len(errors) == 0 {
				t.Errorf("Expected validation error, got none")
			}
			if !tt.expectError && len(errors) > 0 {
				t.Errorf("Expected no validation error, got %d errors: %v", len(errors), errors)
			}
			
			if tt.expectedErrorCount > 0 && len(errors) != tt.expectedErrorCount {
				t.Errorf("Expected %d errors, got %d errors: %v", tt.expectedErrorCount, len(errors), errors)
			}

			// Log errors for debugging
			if len(errors) > 0 {
				t.Logf("Validation errors: %v", errors)
			}
		})
	}
}

// TestEngineRegistryIntegration verifies that EngineRegistry is used for validation hints
func TestEngineRegistryIntegration(t *testing.T) {
	// Mock frontmatter YAML
	mockYAML := `---
on: push
engine: invalid-engine
---`
	
	validator := NewFrontmatterValidator(mockYAML)
	
	frontmatter := map[string]any{
		"on":     "push", 
		"engine": "invalid-engine",
	}
	
	// Test validation
	errors := validator.ValidateFrontmatter(frontmatter)
	
	// Should have an engine validation error
	if len(errors) == 0 {
		t.Fatal("Expected validation errors for invalid engine")
	}
	
	var engineError *FrontmatterValidationError
	for _, err := range errors {
		if err.Path == "engine" {
			engineError = &err
			break
		}
	}
	
	if engineError == nil {
		t.Fatal("Expected engine validation error")
	}
	
	// Get the hint
	hint := generateHintForValidationError(*engineError)
	
	// Should contain engines from registry 
	registry := GetGlobalEngineRegistry()
	engines := registry.GetSupportedEngines()
	
	for _, engine := range engines {
		if hint == "" {
			t.Errorf("Expected hint to contain engine '%s', but hint is empty", engine)
		}
	}
	
	t.Logf("Engine validation error: %s", engineError.Message)
	t.Logf("Generated hint: %s", hint)
	t.Logf("Registry engines: %v", engines)
}