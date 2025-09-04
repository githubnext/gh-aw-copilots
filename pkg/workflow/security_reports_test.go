package workflow

import (
	"strings"
	"testing"
)

// TestSecurityReportsConfig tests the parsing of create-security-report configuration
func TestSecurityReportsConfig(t *testing.T) {
	compiler := NewCompiler(false, "", "test-version")

	tests := []struct {
		name           string
		frontmatter    map[string]any
		expectedConfig *CreateSecurityReportsConfig
	}{
		{
			name: "basic security report configuration",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-security-report": nil,
				},
			},
			expectedConfig: &CreateSecurityReportsConfig{Max: 0}, // 0 means unlimited
		},
		{
			name: "security report with max configuration",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-security-report": map[string]any{
						"max": 50,
					},
				},
			},
			expectedConfig: &CreateSecurityReportsConfig{Max: 50},
		},
		{
			name: "no security report configuration",
			frontmatter: map[string]any{
				"safe-outputs": map[string]any{
					"create-issue": nil,
				},
			},
			expectedConfig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := compiler.extractSafeOutputsConfig(tt.frontmatter)

			if tt.expectedConfig == nil {
				if config == nil || config.CreateSecurityReports == nil {
					return // Expected no config
				}
				t.Errorf("Expected no CreateSecurityReports config, but got: %+v", config.CreateSecurityReports)
				return
			}

			if config == nil || config.CreateSecurityReports == nil {
				t.Errorf("Expected CreateSecurityReports config, but got nil")
				return
			}

			if config.CreateSecurityReports.Max != tt.expectedConfig.Max {
				t.Errorf("Expected Max=%d, got Max=%d", tt.expectedConfig.Max, config.CreateSecurityReports.Max)
			}
		})
	}
}

// TestBuildCreateOutputSecurityReportJob tests the creation of security report job
func TestBuildCreateOutputSecurityReportJob(t *testing.T) {
	compiler := NewCompiler(false, "", "test-version")

	// Test valid configuration
	data := &WorkflowData{
		SafeOutputs: &SafeOutputsConfig{
			CreateSecurityReports: &CreateSecurityReportsConfig{Max: 0},
		},
	}

	job, err := compiler.buildCreateOutputSecurityReportJob(data, "main_job")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if job.Name != "create_security_report" {
		t.Errorf("Expected job name 'create_security_report', got '%s'", job.Name)
	}

	if job.TimeoutMinutes != 10 {
		t.Errorf("Expected timeout 10 minutes, got %d", job.TimeoutMinutes)
	}

	if len(job.Depends) != 1 || job.Depends[0] != "main_job" {
		t.Errorf("Expected dependency on 'main_job', got %v", job.Depends)
	}

	// Check that job has necessary permissions
	if !strings.Contains(job.Permissions, "security-events: write") {
		t.Errorf("Expected security-events: write permission in job, got: %s", job.Permissions)
	}

	// Check that steps include SARIF upload
	stepsStr := strings.Join(job.Steps, "")
	if !strings.Contains(stepsStr, "Upload SARIF") {
		t.Errorf("Expected SARIF upload steps in job")
	}

	if !strings.Contains(stepsStr, "codeql-action/upload-sarif") {
		t.Errorf("Expected CodeQL SARIF upload action in job")
	}

	// Test with max configuration
	dataWithMax := &WorkflowData{
		SafeOutputs: &SafeOutputsConfig{
			CreateSecurityReports: &CreateSecurityReportsConfig{Max: 25},
		},
	}

	jobWithMax, err := compiler.buildCreateOutputSecurityReportJob(dataWithMax, "main_job")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stepsWithMaxStr := strings.Join(jobWithMax.Steps, "")
	if !strings.Contains(stepsWithMaxStr, "GITHUB_AW_SECURITY_REPORT_MAX: 25") {
		t.Errorf("Expected max configuration in environment variables")
	}

	// Test error case - no configuration
	dataNoConfig := &WorkflowData{SafeOutputs: nil}
	_, err = compiler.buildCreateOutputSecurityReportJob(dataNoConfig, "main_job")
	if err == nil {
		t.Errorf("Expected error when no SafeOutputs config provided")
	}
}

// TestParseSecurityReportsConfig tests the parsing function directly
func TestParseSecurityReportsConfig(t *testing.T) {
	compiler := NewCompiler(false, "", "test-version")

	tests := []struct {
		name        string
		outputMap   map[string]any
		expectedMax int
		expectNil   bool
	}{
		{
			name: "basic configuration",
			outputMap: map[string]any{
				"create-security-report": nil,
			},
			expectedMax: 0,
			expectNil:   false,
		},
		{
			name: "configuration with max",
			outputMap: map[string]any{
				"create-security-report": map[string]any{
					"max": 100,
				},
			},
			expectedMax: 100,
			expectNil:   false,
		},
		{
			name: "no configuration",
			outputMap: map[string]any{
				"other-config": nil,
			},
			expectedMax: 0,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := compiler.parseSecurityReportsConfig(tt.outputMap)

			if tt.expectNil {
				if config != nil {
					t.Errorf("Expected nil config, got: %+v", config)
				}
				return
			}

			if config == nil {
				t.Errorf("Expected config, got nil")
				return
			}

			if config.Max != tt.expectedMax {
				t.Errorf("Expected Max=%d, got Max=%d", tt.expectedMax, config.Max)
			}
		})
	}
}
