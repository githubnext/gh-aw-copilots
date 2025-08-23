package workflow

import (
	"testing"
)

// TestJSONSchemaValidationIntegration demonstrates that JSON schema validation is working
func TestJSONSchemaValidationIntegration(t *testing.T) {
	tests := []struct {
		name                string
		frontmatter         map[string]any
		expectSchemaError   bool
		expectCustomError   bool
	}{
		{
			name: "valid frontmatter passes both validations",
			frontmatter: map[string]any{
				"on":     "push",
				"engine": "claude",
				"tools": map[string]any{
					"github": map[string]any{
						"allowed": []string{"create_issue"},
					},
				},
			},
			expectSchemaError: false,
			expectCustomError: false,
		},
		{
			name: "invalid engine caught by both validations",
			frontmatter: map[string]any{
				"on":     "push",
				"engine": "invalid-engine",
			},
			expectSchemaError: true,
			expectCustomError: true,
		},
		{
			name: "additional properties caught by schema validation",
			frontmatter: map[string]any{
				"on":              "push",
				"engine":          "claude",
				"invalid_property": "value",
			},
			expectSchemaError: true,
			expectCustomError: false, // Custom validation doesn't check for additional properties
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

			// Test with JSON schema validation
			schemaErrors := validator.ValidateFrontmatterWithOptions(tt.frontmatter, ValidationOptions{UseJSONSchema: true})
			if tt.expectSchemaError && len(schemaErrors) == 0 {
				t.Errorf("Expected schema validation error, got none")
			}
			if !tt.expectSchemaError && len(schemaErrors) > 0 {
				t.Errorf("Expected no schema validation error, got %d errors: %v", len(schemaErrors), schemaErrors)
			}

			// Test with custom validation
			customErrors := validator.ValidateFrontmatterWithOptions(tt.frontmatter, ValidationOptions{UseJSONSchema: false})
			if tt.expectCustomError && len(customErrors) == 0 {
				t.Errorf("Expected custom validation error, got none")
			}
			if !tt.expectCustomError && len(customErrors) > 0 {
				t.Errorf("Expected no custom validation error, got %d errors: %v", len(customErrors), customErrors)
			}

			// Log errors for debugging
			if len(schemaErrors) > 0 {
				t.Logf("Schema errors: %v", schemaErrors)
			}
			if len(customErrors) > 0 {
				t.Logf("Custom errors: %v", customErrors)
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
	
	// Test custom validation
	errors := validator.ValidateFrontmatterWithOptions(frontmatter, ValidationOptions{UseJSONSchema: false})
	
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