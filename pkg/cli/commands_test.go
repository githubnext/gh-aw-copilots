package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test the CLI functions that are exported from this package

func TestListWorkflows(t *testing.T) {
	// Test the ListWorkflows function (which includes listAgenticEngines)
	err := ListWorkflows(false)

	// Should return nil (no error) and print table-formatted output
	if err != nil {
		t.Errorf("ListWorkflows should not return an error for valid input, got: %v", err)
	}
}

func TestListWorkflowsVerbose(t *testing.T) {
	// Test the ListWorkflows function in verbose mode
	err := ListWorkflows(true)

	// Should return nil (no error) and print table-formatted output with descriptions
	if err != nil {
		t.Errorf("ListWorkflows verbose mode should not return an error for valid input, got: %v", err)
	}
}

func TestAddWorkflow(t *testing.T) {
	// Clean up any existing .github/workflows for this test
	defer os.RemoveAll(".github")

	tests := []struct {
		name        string
		workflow    string
		number      int
		expectError bool
	}{
		{
			name:        "nonexistent workflow",
			workflow:    "nonexistent-workflow",
			number:      1,
			expectError: true,
		},
		{
			name:        "empty workflow name",
			workflow:    "",
			number:      1,
			expectError: false, // AddWorkflow shows help when workflow is empty, doesn't error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddWorkflowWithTracking(tt.workflow, tt.number, false, "", "", false, nil)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', got nil", tt.name)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
			}
		})
	}
}

func TestAddWorkflowForce(t *testing.T) {
	// This test verifies that the force flag works correctly
	// Note: This is a unit test to verify the function signature and basic logic
	// It doesn't test the actual file system operations

	// Test that force=false fails when a file "exists" (simulated by empty workflow name which triggers help)
	err := AddWorkflowWithTracking("", 1, false, "", "", false, nil)
	if err != nil {
		t.Errorf("Expected no error for empty workflow (shows help), got: %v", err)
	}

	// Test that force=true works with same parameters
	err = AddWorkflowWithTracking("", 1, false, "", "", true, nil)
	if err != nil {
		t.Errorf("Expected no error for empty workflow with force=true, got: %v", err)
	}
}

func TestCompileWorkflows(t *testing.T) {
	// Clean up any existing .github/workflows for this test
	defer os.RemoveAll(".github")

	tests := []struct {
		name         string
		markdownFile string
		expectError  bool
	}{
		{
			name:         "nonexistent specific file",
			markdownFile: "nonexistent.md",
			expectError:  true, // Should error when file doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []string
			if tt.markdownFile != "" {
				args = []string{tt.markdownFile}
			}
			err := CompileWorkflows(args, false, "", false, false, false, false)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', got nil", tt.name)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
			}
		})
	}
}

func TestCompileWorkflowsWithNoEmit(t *testing.T) {
	defer os.RemoveAll(".github")
	
	// Create test directory and workflow
	err := os.MkdirAll(".github/workflows", 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a simple test workflow
	workflowContent := `---
on:
  workflow_dispatch:
permissions:
  contents: read
---

# Test Workflow for No Emit

This is a test workflow to verify the --no-emit flag functionality.`

	err = os.WriteFile(".github/workflows/no-emit-test.md", []byte(workflowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test workflow file: %v", err)
	}

	// Test compilation with noEmit = false (should create lock file)
	err = CompileWorkflows([]string{"no-emit-test"}, false, "", false, false, false, false)
	if err != nil {
		t.Errorf("CompileWorkflows with noEmit=false should not error, got: %v", err)
	}

	// Verify lock file was created
	if _, err := os.Stat(".github/workflows/no-emit-test.lock.yml"); os.IsNotExist(err) {
		t.Error("Lock file should have been created when noEmit=false")
	}

	// Remove lock file
	os.Remove(".github/workflows/no-emit-test.lock.yml")

	// Test compilation with noEmit = true (should NOT create lock file)
	err = CompileWorkflows([]string{"no-emit-test"}, false, "", false, false, false, true)
	if err != nil {
		t.Errorf("CompileWorkflows with noEmit=true should not error, got: %v", err)
	}

	// Verify lock file was NOT created
	if _, err := os.Stat(".github/workflows/no-emit-test.lock.yml"); !os.IsNotExist(err) {
		t.Error("Lock file should NOT have been created when noEmit=true")
	}
}

func TestRemoveWorkflows(t *testing.T) {
	err := RemoveWorkflows("test-pattern", false)

	// Should not error since it's a stub implementation
	if err != nil {
		t.Errorf("RemoveWorkflows should not return error for valid input, got: %v", err)
	}
}

func TestStatusWorkflows(t *testing.T) {
	err := StatusWorkflows("test-pattern", false)

	// Should not error since it's a stub implementation
	if err != nil {
		t.Errorf("StatusWorkflows should not return error for valid input, got: %v", err)
	}
}

func TestEnableWorkflows(t *testing.T) {
	err := EnableWorkflows("test-pattern")

	// Should not error since it's a stub implementation
	if err != nil {
		t.Errorf("EnableWorkflows should not return error for valid input, got: %v", err)
	}
}

func TestDisableWorkflows(t *testing.T) {
	err := DisableWorkflows("test-pattern")

	// Should not error since it's a stub implementation
	if err != nil {
		t.Errorf("DisableWorkflows should not return error for valid input, got: %v", err)
	}
}

func TestRunWorkflowOnGitHub(t *testing.T) {
	// Test with empty workflow name
	err := RunWorkflowOnGitHub("", false)
	if err == nil {
		t.Error("RunWorkflowOnGitHub should return error for empty workflow name")
	}

	// Test with nonexistent workflow (this will fail but gracefully)
	err = RunWorkflowOnGitHub("nonexistent-workflow", false)
	if err == nil {
		t.Error("RunWorkflowOnGitHub should return error for non-existent workflow")
	}
}

func TestRunWorkflowsOnGitHub(t *testing.T) {
	// Test with empty workflow list
	err := RunWorkflowsOnGitHub([]string{}, 0, false)
	if err == nil {
		t.Error("RunWorkflowsOnGitHub should return error for empty workflow list")
	}

	// Test with workflow list containing empty name
	err = RunWorkflowsOnGitHub([]string{"valid-workflow", ""}, 0, false)
	if err == nil {
		t.Error("RunWorkflowsOnGitHub should return error for workflow list containing empty name")
	}

	// Test with nonexistent workflows (this will fail but gracefully)
	err = RunWorkflowsOnGitHub([]string{"nonexistent-workflow1", "nonexistent-workflow2"}, 0, false)
	if err == nil {
		t.Error("RunWorkflowsOnGitHub should return error for non-existent workflows")
	}

	// Test with negative repeat seconds (should work as 0)
	err = RunWorkflowsOnGitHub([]string{"nonexistent-workflow"}, -1, false)
	if err == nil {
		t.Error("RunWorkflowsOnGitHub should return error for non-existent workflow regardless of repeat value")
	}
}

func TestGetLatestWorkflowRunWithTimestamp(t *testing.T) {
	// Test with non-existent workflow - should handle gracefully
	url, createdAt, err := getLatestWorkflowRunWithTimestamp("nonexistent-workflow.lock.yml", false)
	if err == nil {
		t.Error("getLatestWorkflowRunWithTimestamp should return error for non-existent workflow")
	}
	if url != "" {
		t.Error("getLatestWorkflowRunWithTimestamp should return empty URL for non-existent workflow")
	}
	if !createdAt.IsZero() {
		t.Error("getLatestWorkflowRunWithTimestamp should return zero time for non-existent workflow")
	}
}

// func TestGetLatestWorkflowRunURLWithRetry(t *testing.T) {
// 	// Test with non-existent workflow - should handle gracefully and return error after retries
// 	url, err := getLatestWorkflowRunURLWithRetry("nonexistent-workflow.lock.yml", false)
// 	if err == nil {
// 		t.Error("getLatestWorkflowRunURLWithRetry should return error for non-existent workflow")
// 	}
// 	if url != "" {
// 		t.Error("getLatestWorkflowRunURLWithRetry should return empty URL for non-existent workflow")
// 	}

// 	// The error message should indicate multiple attempts were made
// 	if !strings.Contains(err.Error(), "attempts") {
// 		t.Errorf("Error message should mention retry attempts, got: %v", err)
// 	}
// }

func TestAllCommandsExist(t *testing.T) {
	defer os.RemoveAll(".github")

	// Test that all expected functions exist and can be called
	// This helps ensure the interface is stable

	// Test structure: function, expected to error
	tests := []struct {
		fn          func() error
		expectError bool
		name        string
	}{
		{func() error { return ListWorkflows(false) }, false, "ListWorkflows"},
		{func() error { return AddWorkflowWithTracking("", 1, false, "", "", false, nil) }, false, "AddWorkflowWithTracking (empty name)"}, // Shows help when empty, doesn't error
		{func() error { return CompileWorkflows([]string{}, false, "", false, false, false, false) }, false, "CompileWorkflows"},                  // Should compile existing markdown files successfully
		{func() error { return RemoveWorkflows("test", false) }, false, "RemoveWorkflows"},                                                 // Should handle missing directory gracefully
		{func() error { return StatusWorkflows("test", false) }, false, "StatusWorkflows"},                                                 // Should handle missing directory gracefully
		{func() error { return EnableWorkflows("test") }, false, "EnableWorkflows"},                                                        // Should handle missing directory gracefully
		{func() error { return DisableWorkflows("test") }, false, "DisableWorkflows"},                                                      // Should handle missing directory gracefully
		{func() error { return RunWorkflowOnGitHub("", false) }, true, "RunWorkflowOnGitHub"},                                              // Should error with empty workflow name
		{func() error { return RunWorkflowsOnGitHub([]string{}, 0, false) }, true, "RunWorkflowsOnGitHub"},                                 // Should error with empty workflow list
	}

	for _, test := range tests {
		err := test.fn()
		if test.expectError && err == nil {
			t.Errorf("%s: expected error but got nil", test.name)
		} else if !test.expectError && err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}
	}
}

func TestAddWorkflowWithPR(t *testing.T) {
	// Clean up any existing .github/workflows for this test
	defer os.RemoveAll(".github")

	// Test with nonexistent workflow (should fail early due to workflow not found or repo access)
	err := AddWorkflowWithRepoAndPR("nonexistent-workflow", 1, false, "", "", "", false)
	if err == nil {
		t.Error("AddWorkflowWithRepoAndPR should return an error for nonexistent workflow or missing git setup")
	}

	// The error could be either:
	// 1. GitHub CLI not available
	// 2. Not in a git repository
	// 3. Repository access check failure
	// 4. Working directory not clean
	// 5. Workflow not found
	// All of these are expected in the test environment
	t.Logf("Expected error for PR creation: %v", err)
}

// TestInstallPackage tests the InstallPackage function
func TestInstallPackage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gh-aw-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the getPackagesDir function by temporarily changing directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		repoSpec    string
		local       bool
		verbose     bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid repo spec",
			repoSpec:    "invalid",
			local:       true,
			verbose:     false,
			expectError: true,
			errorMsg:    "invalid repository specification",
		},
		{
			name:        "empty repo spec",
			repoSpec:    "",
			local:       true,
			verbose:     false,
			expectError: true,
			errorMsg:    "invalid repository specification",
		},
		{
			name:        "valid repo spec but download will fail",
			repoSpec:    "nonexistent/repo",
			local:       true,
			verbose:     true,
			expectError: true,
			errorMsg:    "failed to download workflows",
		},
		{
			name:        "valid repo spec with version but download will fail",
			repoSpec:    "nonexistent/repo@v1.0.0",
			local:       false,
			verbose:     false,
			expectError: true,
			errorMsg:    "failed to download workflows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InstallPackage(tt.repoSpec, tt.local, tt.verbose)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', got nil", tt.name)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
			}

			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestUninstallPackage tests the UninstallPackage function
func TestUninstallPackage(t *testing.T) {
	tests := []struct {
		name        string
		repoSpec    string
		local       bool
		verbose     bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid repo spec",
			repoSpec:    "invalid",
			local:       true,
			verbose:     false,
			expectError: true,
			errorMsg:    "invalid repository specification",
		},
		{
			name:        "empty repo spec",
			repoSpec:    "",
			local:       true,
			verbose:     false,
			expectError: true,
			errorMsg:    "invalid repository specification",
		},
		{
			name:        "valid repo spec - package not installed",
			repoSpec:    "nonexistent/repo",
			local:       true,
			verbose:     true,
			expectError: false,
		},
		{
			name:        "valid repo spec with version - package not installed",
			repoSpec:    "nonexistent/repo@v1.0.0",
			local:       false,
			verbose:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UninstallPackage(tt.repoSpec, tt.local, tt.verbose)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', got nil", tt.name)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
			}

			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestListPackages tests the ListPackages function
func TestListPackages(t *testing.T) {
	tests := []struct {
		name        string
		local       bool
		verbose     bool
		expectError bool
	}{
		{
			name:        "list local packages",
			local:       true,
			verbose:     false,
			expectError: false, // Should not error even if directory doesn't exist
		},
		{
			name:        "list global packages",
			local:       false,
			verbose:     true,
			expectError: false, // Should not error even if directory doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ListPackages(tt.local, tt.verbose)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', got nil", tt.name)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
			}
		})
	}
}

func TestNewWorkflow(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-new-workflow-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Logf("Warning: Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	tests := []struct {
		name          string
		workflowName  string
		force         bool
		expectedError bool
		setup         func(t *testing.T)
	}{
		{
			name:          "create new workflow",
			workflowName:  "test-workflow",
			force:         false,
			expectedError: false,
		},
		{
			name:          "fail to overwrite existing workflow without force",
			workflowName:  "existing-workflow",
			force:         false,
			expectedError: true,
			setup: func(t *testing.T) {
				// Create an existing workflow file
				os.MkdirAll(".github/workflows", 0755)
				os.WriteFile(".github/workflows/existing-workflow.md", []byte("test"), 0644)
			},
		},
		{
			name:          "overwrite existing workflow with force",
			workflowName:  "force-workflow",
			force:         true,
			expectedError: false,
			setup: func(t *testing.T) {
				// Create an existing workflow file
				if err := os.MkdirAll(".github/workflows", 0755); err != nil {
					t.Fatalf("Failed to create workflows directory: %v", err)
				}
				if err := os.WriteFile(".github/workflows/force-workflow.md", []byte("old content"), 0644); err != nil {
					t.Fatalf("Failed to create existing workflow file: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup if needed
			if test.setup != nil {
				test.setup(t)
			}

			// Run the function
			err := NewWorkflow(test.workflowName, false, test.force)

			// Check error expectation
			if test.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !test.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If no error expected, verify the file was created
			if !test.expectedError {
				filePath := ".github/workflows/" + test.workflowName + ".md"
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected workflow file was not created: %s", filePath)
				}

				// Verify the content contains expected template elements
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read created workflow file: %v", err)
				} else {
					contentStr := string(content)
					// Check for key template elements
					expectedElements := []string{
						"# Trigger - when should this workflow run?",
						"on:",
						"permissions:",
						"safe-outputs:",
						"# " + test.workflowName,
						"workflow_dispatch:",
					}
					for _, element := range expectedElements {
						if !strings.Contains(contentStr, element) {
							t.Errorf("Template missing expected element: %s", element)
						}
					}
				}
			}

			// Clean up for next test
			os.RemoveAll(".github")
		})
	}
}

// Test SetVersionInfo and GetVersion functions
func TestSetVersionInfo(t *testing.T) {
	// Save original version to restore after test
	originalVersion := GetVersion()
	defer SetVersionInfo(originalVersion)

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "normal version",
			version: "1.0.0",
		},
		{
			name:    "empty version",
			version: "",
		},
		{
			name:    "version with pre-release",
			version: "2.0.0-beta.1",
		},
		{
			name:    "version with build metadata",
			version: "1.2.3+20240808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersionInfo(tt.version)
			got := GetVersion()
			if got != tt.version {
				t.Errorf("SetVersionInfo(%q) -> GetVersion() = %q, want %q", tt.version, got, tt.version)
			}
		})
	}
}

// Test AddWorkflowWithRepo function
func TestAddWorkflowWithRepo(t *testing.T) {
	// Clean up any existing .github/workflows for this test
	defer os.RemoveAll(".github")

	tests := []struct {
		name        string
		workflow    string
		repo        string
		expectError bool
		description string
	}{
		{
			name:        "empty workflow and repo",
			workflow:    "",
			repo:        "",
			expectError: false, // Should show help message, not error
			description: "empty workflow shows help",
		},
		{
			name:        "nonexistent workflow without repo",
			workflow:    "nonexistent-workflow",
			repo:        "",
			expectError: true,
			description: "nonexistent workflow should fail",
		},
		{
			name:        "workflow with invalid repo format",
			workflow:    "test-workflow",
			repo:        "invalid-repo-format",
			expectError: true,
			description: "invalid repo format should fail during installation",
		},
		{
			name:        "workflow with nonexistent repo",
			workflow:    "test-workflow",
			repo:        "nonexistent/nonexistent-repo",
			expectError: true,
			description: "nonexistent repo should fail during installation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddWorkflowWithRepo(tt.workflow, 1, false, "", tt.repo, "", false)

			if tt.expectError {
				if err == nil {
					t.Errorf("AddWorkflowWithRepo(%q, %q) expected error (%s), but got none", tt.workflow, tt.repo, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("AddWorkflowWithRepo(%q, %q) unexpected error (%s): %v", tt.workflow, tt.repo, tt.description, err)
				}
			}
		})
	}
}

func TestCollectIncludeDependencies(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	workflowsDir := tempDir + "/workflows"
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows directory: %v", err)
	}

	// Create test files
	sharedDir := workflowsDir + "/shared"
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("Failed to create shared directory: %v", err)
	}

	// Create a shared file
	sharedFile := sharedDir + "/common.md"
	sharedContent := `# Common Content
This is shared content.
`
	if err := os.WriteFile(sharedFile, []byte(sharedContent), 0644); err != nil {
		t.Fatalf("Failed to create shared file: %v", err)
	}

	// Create another shared file for recursive testing
	recursiveFile := sharedDir + "/recursive.md"
	recursiveContent := `# Recursive Content
@include shared/common.md
More content here.
`
	if err := os.WriteFile(recursiveFile, []byte(recursiveContent), 0644); err != nil {
		t.Fatalf("Failed to create recursive file: %v", err)
	}

	tests := []struct {
		name              string
		content           string
		workflowPath      string
		expectedDepsCount int
		expectError       bool
		description       string
	}{
		{
			name:              "no_includes",
			content:           "# Simple Workflow\nNo includes here.",
			workflowPath:      workflowsDir + "/simple.md",
			expectedDepsCount: 0,
			expectError:       false,
			description:       "Content without includes should return no dependencies",
		},
		{
			name:              "single_include",
			content:           "# Workflow with Include\n@include shared/common.md\nMore content.",
			workflowPath:      workflowsDir + "/with-include.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Content with one include should return one dependency",
		},
		{
			name:              "multiple_includes",
			content:           "# Multiple Includes\n@include shared/common.md\n@include shared/recursive.md",
			workflowPath:      workflowsDir + "/multi-include.md",
			expectedDepsCount: 3,
			expectError:       false,
			description:       "Content with multiple includes should return multiple dependencies (including recursive ones)",
		},
		{
			name:              "recursive_includes",
			content:           "# Recursive Test\n@include shared/recursive.md",
			workflowPath:      workflowsDir + "/recursive-test.md",
			expectedDepsCount: 2,
			expectError:       false,
			description:       "Recursive includes should collect all dependencies",
		},
		{
			name:              "section_reference",
			content:           "# Section Reference\n@include shared/common.md#Section",
			workflowPath:      workflowsDir + "/section-ref.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Include with section reference should work",
		},
		{
			name:              "nonexistent_file",
			content:           "# Missing File\n@include shared/missing.md",
			workflowPath:      workflowsDir + "/missing.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Include of nonexistent file should still add dependency but not recurse",
		},
		{
			name:              "optional_include_existing",
			content:           "# Optional Include Existing\n@include? shared/common.md\nMore content.",
			workflowPath:      workflowsDir + "/optional-existing.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Optional include of existing file should work like regular include",
		},
		{
			name:              "optional_include_missing",
			content:           "# Optional Include Missing\n@include? shared/optional.md\nMore content.",
			workflowPath:      workflowsDir + "/optional-missing.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Optional include of missing file should still add dependency",
		},
		{
			name:              "mixed_includes",
			content:           "# Mixed\n@include shared/common.md\n@include? shared/optional.md\n@include shared/recursive.md",
			workflowPath:      workflowsDir + "/mixed.md",
			expectedDepsCount: 4, // common.md + optional.md + recursive.md + recursive.md->common.md
			expectError:       false,
			description:       "Mixed regular and optional includes should collect all dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, err := collectIncludeDependencies(tt.content, tt.workflowPath, workflowsDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("collectIncludeDependencies expected error (%s), but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("collectIncludeDependencies unexpected error (%s): %v", tt.description, err)
				return
			}

			if len(deps) != tt.expectedDepsCount {
				t.Errorf("collectIncludeDependencies expected %d dependencies (%s), got %d", tt.expectedDepsCount, tt.description, len(deps))
			}

			// Verify dependency structure
			for i, dep := range deps {
				if dep.SourcePath == "" {
					t.Errorf("Dependency %d has empty SourcePath", i)
				}
				if dep.TargetPath == "" {
					t.Errorf("Dependency %d has empty TargetPath", i)
				}
			}

			// Verify optional flag for specific test cases
			if tt.name == "optional_include_existing" || tt.name == "optional_include_missing" {
				if len(deps) > 0 && !deps[0].IsOptional {
					t.Errorf("Optional include dependency should have IsOptional=true")
				}
			}
			if tt.name == "mixed_includes" {
				optionalFound := false
				regularFound := false
				for _, dep := range deps {
					if strings.Contains(dep.TargetPath, "optional") && dep.IsOptional {
						optionalFound = true
					}
					if (strings.Contains(dep.TargetPath, "common") || strings.Contains(dep.TargetPath, "recursive")) && !dep.IsOptional {
						regularFound = true
					}
				}
				if !optionalFound {
					t.Errorf("Mixed includes should have at least one optional dependency")
				}
				if !regularFound {
					t.Errorf("Mixed includes should have at least one regular dependency")
				}
			}
		})
	}
}

func TestCollectIncludesRecursive(t *testing.T) {
	// Create temporary test environment
	tempDir := t.TempDir()
	baseDir := tempDir
	workflowsDir := tempDir

	// Create test files
	file1 := tempDir + "/file1.md"
	file1Content := `# File 1
Content of file 1
`
	if err := os.WriteFile(file1, []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	file2 := tempDir + "/file2.md"
	file2Content := `# File 2
@include file1.md
Content of file 2
`
	if err := os.WriteFile(file2, []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	tests := []struct {
		name              string
		content           string
		expectedDepsCount int
		expectError       bool
		description       string
	}{
		{
			name:              "no_includes",
			content:           "# No Includes\nJust regular content.",
			expectedDepsCount: 0,
			expectError:       false,
			description:       "Content without includes should not add dependencies",
		},
		{
			name:              "single_include",
			content:           "# Single Include\n@include file1.md",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Single include should add one dependency",
		},
		{
			name:              "recursive_include",
			content:           "# Recursive\n@include file2.md",
			expectedDepsCount: 2,
			expectError:       false,
			description:       "Recursive include should collect all dependencies",
		},
		{
			name:              "whitespace_handling",
			content:           "# Whitespace\n@include    file1.md   \n",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Include with extra whitespace should work",
		},
		{
			name:              "section_reference",
			content:           "# Section\n@include file1.md#Header",
			expectedDepsCount: 1,
			expectError:       false,
			description:       "Section references should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dependencies []IncludeDependency
			seen := make(map[string]bool)

			err := collectIncludesRecursive(tt.content, baseDir, workflowsDir, &dependencies, seen)

			if tt.expectError {
				if err == nil {
					t.Errorf("collectIncludesRecursive expected error (%s), but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("collectIncludesRecursive unexpected error (%s): %v", tt.description, err)
				return
			}

			if len(dependencies) != tt.expectedDepsCount {
				t.Errorf("collectIncludesRecursive expected %d dependencies (%s), got %d", tt.expectedDepsCount, tt.description, len(dependencies))
			}

			// Verify all dependencies have proper paths
			for i, dep := range dependencies {
				if dep.SourcePath == "" {
					t.Errorf("Dependency %d has empty SourcePath", i)
				}
				if dep.TargetPath == "" {
					t.Errorf("Dependency %d has empty TargetPath", i)
				}
			}
		})
	}
}

func TestCollectIncludesRecursiveCircularReference(t *testing.T) {
	// Test circular reference detection
	tempDir := t.TempDir()
	baseDir := tempDir
	workflowsDir := tempDir

	// Create files with circular references
	file1 := tempDir + "/circular1.md"
	file1Content := `# Circular 1
@include circular2.md
`
	if err := os.WriteFile(file1, []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to create circular1: %v", err)
	}

	file2 := tempDir + "/circular2.md"
	file2Content := `# Circular 2  
@include circular1.md
`
	if err := os.WriteFile(file2, []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to create circular2: %v", err)
	}

	var dependencies []IncludeDependency
	seen := make(map[string]bool)

	content := "@include circular1.md"

	// This should not infinite loop due to the seen map
	err := collectIncludesRecursive(content, baseDir, workflowsDir, &dependencies, seen)

	// Should complete without error (circular references are prevented by seen map)
	if err != nil {
		t.Errorf("collectIncludesRecursive should handle circular references gracefully, got error: %v", err)
	}

	// Should have collected some dependencies but not infinite
	if len(dependencies) > 10 {
		t.Errorf("collectIncludesRecursive collected too many dependencies (%d), possible infinite loop", len(dependencies))
	}
}

// TestCleanupOrphanedIncludes tests that root workflow files are not removed as "orphaned" includes
func TestCleanupOrphanedIncludes(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-cleanup-orphaned")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create shared subdirectory
	sharedDir := filepath.Join(workflowsDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create root workflow files (these should NOT be considered orphaned)
	rootWorkflows := []string{"daily-plan.md", "weekly-research.md", "action-workflow-assessor.md"}
	for _, name := range rootWorkflows {
		content := fmt.Sprintf(`---
on:
  workflow_dispatch:
---

# %s

This is a root workflow.
`, strings.TrimSuffix(name, ".md"))
		if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create include files in shared/ directory (these should be considered orphaned if not used)
	includeFiles := []string{"shared/common.md", "shared/tools.md"}
	for _, name := range includeFiles {
		content := `---
tools:
  github:
    allowed: []
---

This is an include file.
`
		if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create one root workflow that actually uses an include
	workflowWithInclude := `---
on:
  workflow_dispatch:
---

# Workflow with Include

@include shared/common.md

This workflow uses an include.
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow-with-include.md"), []byte(workflowWithInclude), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory to simulate the git root
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Run cleanup
	err = cleanupOrphanedIncludes(true)
	if err != nil {
		t.Fatalf("cleanupOrphanedIncludes failed: %v", err)
	}

	// Verify that root workflow files still exist
	for _, name := range rootWorkflows {
		filePath := filepath.Join(workflowsDir, name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Root workflow file %s was incorrectly removed as orphaned", name)
		}
	}

	// Verify that workflow-with-include.md still exists
	if _, err := os.Stat(filepath.Join(workflowsDir, "workflow-with-include.md")); os.IsNotExist(err) {
		t.Error("Workflow with include was incorrectly removed")
	}

	// Verify that shared/common.md still exists (it's used by workflow-with-include.md)
	if _, err := os.Stat(filepath.Join(workflowsDir, "shared", "common.md")); os.IsNotExist(err) {
		t.Error("Used include file shared/common.md was incorrectly removed")
	}

	// Verify that shared/tools.md was removed (it's truly orphaned)
	if _, err := os.Stat(filepath.Join(workflowsDir, "shared", "tools.md")); !os.IsNotExist(err) {
		t.Error("Orphaned include file shared/tools.md was not removed")
	}
}

// TestPreviewOrphanedIncludes tests the preview functionality for orphaned includes
func TestPreviewOrphanedIncludes(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-preview-orphaned")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create shared subdirectory
	sharedDir := filepath.Join(workflowsDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow files
	workflow1 := `---
on:
  workflow_dispatch:
---

# Workflow 1

@include shared/common.md

This workflow uses common include.
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow1.md"), []byte(workflow1), 0644); err != nil {
		t.Fatal(err)
	}

	workflow2 := `---
on:
  workflow_dispatch:
---

# Workflow 2

@include shared/tools.md

This workflow uses tools include.
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow2.md"), []byte(workflow2), 0644); err != nil {
		t.Fatal(err)
	}

	workflow3 := `---
on:
  workflow_dispatch:
---

# Workflow 3

@include shared/common.md

This workflow also uses common include.
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow3.md"), []byte(workflow3), 0644); err != nil {
		t.Fatal(err)
	}

	// Create include files
	includeFiles := map[string]string{
		"shared/common.md": "Common include content",
		"shared/tools.md":  "Tools include content",
		"shared/unused.md": "Unused include content",
	}
	for name, content := range includeFiles {
		if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Change to the temporary directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test Case 1: Remove workflow2 - should orphan shared/tools.md but not shared/common.md
	// Use relative paths like the real RemoveWorkflows function does
	filesToRemove := []string{".github/workflows/workflow2.md"}
	orphaned, err := previewOrphanedIncludes(filesToRemove, false)
	if err != nil {
		t.Fatalf("previewOrphanedIncludes failed: %v", err)
	}

	expectedOrphaned := []string{"shared/tools.md", "shared/unused.md"}
	if len(orphaned) != len(expectedOrphaned) {
		t.Errorf("Expected %d orphaned includes, got %d: %v", len(expectedOrphaned), len(orphaned), orphaned)
	}

	for _, expected := range expectedOrphaned {
		found := false
		for _, actual := range orphaned {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be orphaned, but it wasn't found in: %v", expected, orphaned)
		}
	}

	// shared/common.md should NOT be orphaned as it's used by workflow1 and workflow3
	for _, include := range orphaned {
		if include == "shared/common.md" {
			t.Error("shared/common.md should not be orphaned as it's used by remaining workflows")
		}
	}

	// Test Case 2: Remove all workflows - should orphan all includes
	allFiles := []string{
		".github/workflows/workflow1.md",
		".github/workflows/workflow2.md",
		".github/workflows/workflow3.md",
	}
	orphaned, err = previewOrphanedIncludes(allFiles, false)
	if err != nil {
		t.Fatalf("previewOrphanedIncludes failed: %v", err)
	}

	expectedAllOrphaned := []string{"shared/common.md", "shared/tools.md", "shared/unused.md"}
	if len(orphaned) != len(expectedAllOrphaned) {
		t.Errorf("Expected %d orphaned includes when removing all workflows, got %d: %v", len(expectedAllOrphaned), len(orphaned), orphaned)
	}
}

// TestRemoveWorkflowsWithNoOrphansFlag tests that the --keep-orphans flag works correctly
func TestRemoveWorkflowsWithNoOrphansFlag(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-keep-orphans-flag")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create shared subdirectory
	sharedDir := filepath.Join(workflowsDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a workflow that uses an include
	workflowContent := `---
on:
  workflow_dispatch:
---

# Test Workflow

@include shared/common.md

This workflow uses an include.
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "test-workflow.md"), []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the include file
	includeContent := `This is a shared include file.`
	if err := os.WriteFile(filepath.Join(sharedDir, "common.md"), []byte(includeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test 1: Verify include file exists before removal
	if _, err := os.Stat(filepath.Join(sharedDir, "common.md")); os.IsNotExist(err) {
		t.Fatal("Include file should exist before removal")
	}

	// Test 2: Check preview shows orphaned includes when flag is not used
	filesToRemove := []string{".github/workflows/test-workflow.md"}
	orphaned, err := previewOrphanedIncludes(filesToRemove, false)
	if err != nil {
		t.Fatalf("previewOrphanedIncludes failed: %v", err)
	}

	if len(orphaned) != 1 || orphaned[0] != "shared/common.md" {
		t.Errorf("Expected shared/common.md to be orphaned, got: %v", orphaned)
	}

	// Note: We can't easily test the actual RemoveWorkflows function with user input
	// since it requires interactive confirmation. The logic is tested through
	// previewOrphanedIncludes and the flag handling is straightforward.
}

// TestExtractStopTimeFromLockFile tests the extractStopTimeFromLockFile function
func TestExtractStopTimeFromLockFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		content      string
		expectedTime string
	}{
		{
			name: "valid stop-time in lock file",
			content: `# This file was automatically generated by gh-aw
name: "Test Workflow"
jobs:
  test:
    steps:
      - name: Safety checks
        run: |
          STOP_TIME="2025-12-31 23:59:59"
          echo "Checking stop-time limit: $STOP_TIME"`,
			expectedTime: "2025-12-31 23:59:59",
		},
		{
			name: "no stop-time in lock file",
			content: `# This file was automatically generated by gh-aw
name: "Test Workflow"
jobs:
  test:
    steps:
      - name: Run tests
        run: echo "No stop time here"`,
			expectedTime: "",
		},
		{
			name: "malformed stop-time line",
			content: `# This file was automatically generated by gh-aw
name: "Test Workflow"
jobs:
  test:
    steps:
      - name: Safety checks
        run: |
          STOP_TIME=malformed-no-quotes
          echo "Invalid format"`,
			expectedTime: "",
		},
		{
			name: "multiple stop-time lines (should get first)",
			content: `# This file was automatically generated by gh-aw
name: "Test Workflow"
jobs:
  test:
    steps:
      - name: Safety checks
        run: |
          STOP_TIME="2025-06-01 12:00:00"
          echo "Checking stop-time limit: $STOP_TIME"
          STOP_TIME="2025-07-01 12:00:00"`,
			expectedTime: "2025-06-01 12:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test lock file
			lockFile := filepath.Join(tmpDir, tt.name+".lock.yml")
			err := os.WriteFile(lockFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test extraction
			result := extractStopTimeFromLockFile(lockFile)
			if result != tt.expectedTime {
				t.Errorf("extractStopTimeFromLockFile() = %q, want %q", result, tt.expectedTime)
			}
		})
	}

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		result := extractStopTimeFromLockFile("/non/existent/file.lock.yml")
		if result != "" {
			t.Errorf("extractStopTimeFromLockFile() for non-existent file = %q, want empty string", result)
		}
	})
}

// TestCalculateTimeRemaining tests the calculateTimeRemaining function
func TestCalculateTimeRemaining(t *testing.T) {
	tests := []struct {
		name        string
		stopTimeStr string
		expected    string
	}{
		{
			name:        "empty stop time",
			stopTimeStr: "",
			expected:    "N/A",
		},
		{
			name:        "invalid format",
			stopTimeStr: "invalid-date-format",
			expected:    "Invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTimeRemaining(tt.stopTimeStr)
			if result != tt.expected {
				t.Errorf("calculateTimeRemaining(%q) = %q, want %q", tt.stopTimeStr, result, tt.expected)
			}
		})
	}

	// Test with future time - this will test the logic but the exact result depends on current time
	t.Run("future time formatting", func(t *testing.T) {
		// Create a time 2 hours and 30 minutes in the future
		// Add a small buffer to account for execution time
		futureTime := time.Now().Add(2*time.Hour + 30*time.Minute + 1*time.Second)
		stopTimeStr := futureTime.Format("2006-01-02 15:04:05")

		result := calculateTimeRemaining(stopTimeStr)

		// Should contain "h" and "m" for hours and minutes
		if !strings.Contains(result, "h") || !strings.Contains(result, "m") {
			t.Errorf("calculateTimeRemaining() for future time should contain hours and minutes, got: %q", result)
		}

		// Should not be "Expired", "Invalid", or "N/A"
		if result == "Expired" || result == "Invalid" || result == "N/A" {
			t.Errorf("calculateTimeRemaining() for future time should not be %q", result)
		}
	})

	// Test with past time
	t.Run("past time - expired", func(t *testing.T) {
		// Create a time 1 hour in the past
		pastTime := time.Now().Add(-1 * time.Hour)
		stopTimeStr := pastTime.Format("2006-01-02 15:04:05")

		result := calculateTimeRemaining(stopTimeStr)
		if result != "Expired" {
			t.Errorf("calculateTimeRemaining() for past time = %q, want %q", result, "Expired")
		}
	})
}
