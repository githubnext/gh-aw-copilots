package workflow

import (
	"testing"
)

// TestJSONSchemaValidationIntegration demonstrates that JSON schema validation is working
func TestJSONSchemaValidationIntegration(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter map[string]any
		expectError bool
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
		},
		{
			name: "additional properties caught by schema validation",
			frontmatter: map[string]any{
				"on":               "push",
				"engine":           "claude",
				"invalid_property": "value",
			},
			expectError: true,
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
			errors := validator.ValidateFrontmatter(tt.frontmatter)

			if tt.expectError && len(errors) == 0 {
				t.Errorf("Expected validation error, got none")
			}
			if !tt.expectError && len(errors) > 0 {
				t.Errorf("Expected no validation error, got %d errors: %v", len(errors), errors)
			}

			// Log errors for debugging
			if len(errors) > 0 {
				t.Logf("Validation errors: %v", errors)
			}
		})
	}
}

// TestEngineRegistryIntegration verifies engine validation works with JSON schema
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

	// Should have validation errors for invalid engine
	if len(errors) == 0 {
		t.Fatal("Expected validation errors for invalid engine")
	}

	// Just check that validation catches invalid engines
	hasError := len(errors) > 0
	if !hasError {
		t.Error("Expected validation to catch invalid engine")
	}

	t.Logf("Validation errors: %v", errors)
}
