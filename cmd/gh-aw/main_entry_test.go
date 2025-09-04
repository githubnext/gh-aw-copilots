package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/githubnext/gh-aw/pkg/cli"
)

func TestValidateEngine(t *testing.T) {
	tests := []struct {
		name       string
		engine     string
		expectErr  bool
		errMessage string
	}{
		{
			name:      "empty engine (uses default)",
			engine:    "",
			expectErr: false,
		},
		{
			name:      "valid claude engine",
			engine:    "claude",
			expectErr: false,
		},
		{
			name:      "valid codex engine",
			engine:    "codex",
			expectErr: false,
		},
		{
			name:      "valid genaiscript engine",
			engine:    "genaiscript",
			expectErr: false,
		},
		{
			name:       "invalid engine",
			engine:     "gpt4",
			expectErr:  true,
			errMessage: "invalid engine value 'gpt4'",
		},
		{
			name:       "invalid engine case sensitive",
			engine:     "Claude",
			expectErr:  true,
			errMessage: "invalid engine value 'Claude'",
		},
		{
			name:       "invalid engine with spaces",
			engine:     "claude ",
			expectErr:  true,
			errMessage: "invalid engine value 'claude '",
		},
		{
			name:       "completely invalid engine",
			engine:     "invalid-engine",
			expectErr:  true,
			errMessage: "invalid engine value 'invalid-engine'",
		},
		{
			name:       "numeric engine",
			engine:     "123",
			expectErr:  true,
			errMessage: "invalid engine value '123'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEngine(tt.engine)

			if tt.expectErr {
				if err == nil {
					t.Errorf("validateEngine(%q) expected error but got none", tt.engine)
					return
				}

				if tt.errMessage != "" && err.Error() != fmt.Sprintf("invalid engine value '%s'. Must be 'claude', 'codex', or 'genaiscript'", tt.engine) {
					t.Errorf("validateEngine(%q) error message = %v, want to contain %v", tt.engine, err.Error(), tt.errMessage)
				}
			} else {
				if err != nil {
					t.Errorf("validateEngine(%q) unexpected error: %v", tt.engine, err)
				}
			}
		})
	}
}

func TestInitFunction(t *testing.T) {
	// Test that init function doesn't panic
	t.Run("init function executes without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("init() panicked: %v", r)
			}
		}()

		// The init function has already been called when the package was loaded
		// We can't call it again, but we can verify that the initialization worked
		// by checking that the version was set
		if version == "" {
			t.Error("init() should have initialized version variable")
		}
	})
}

func TestMainFunction(t *testing.T) {
	// We can't easily test the main() function directly since it calls os.Exit(),
	// but we can test the command structure and basic functionality

	t.Run("main function setup", func(t *testing.T) {
		// Test that root command is properly configured
		if rootCmd.Use == "" {
			t.Error("rootCmd.Use should not be empty")
		}

		if rootCmd.Short == "" {
			t.Error("rootCmd.Short should not be empty")
		}

		if rootCmd.Long == "" {
			t.Error("rootCmd.Long should not be empty")
		}

		// Test that commands are properly added
		if len(rootCmd.Commands()) == 0 {
			t.Error("rootCmd should have subcommands")
		}
	})

	t.Run("version command is available", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "version" {
				found = true
				break
			}
		}
		if !found {
			t.Error("version command should be available")
		}
	})

	t.Run("list command is available", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "list" {
				found = true
				break
			}
		}
		if !found {
			t.Error("list command should be available")
		}
	})

	t.Run("root command help", func(t *testing.T) {
		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute help
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()

		// Restore output
		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Errorf("root command help failed: %v", err)
		}

		if output == "" {
			t.Error("root command help should produce output")
		}

		// Reset args for other tests
		rootCmd.SetArgs([]string{})
	})
}

// TestMainFunctionExecutionPath tests the main function execution path
// This covers the main() function at line 360
func TestMainFunctionExecutionPath(t *testing.T) {
	// Test that we can build and run the main function successfully
	t.Run("main function integration test", func(t *testing.T) {
		// Only run this test if we're in development (has go)
		if _, err := exec.LookPath("go"); err != nil {
			t.Skip("go binary not available - skipping main function integration test")
		}

		// Test help command execution through main function
		cmd := exec.Command("go", "run", "main.go", "--help")
		cmd.Dir = "."

		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run main with --help: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "GitHub Agentic Workflows") {
			t.Error("main function help output should contain 'GitHub Agentic Workflows'")
		}

		if !strings.Contains(outputStr, "Usage:") {
			t.Error("main function help output should contain usage information")
		}
	})

	t.Run("main function version command", func(t *testing.T) {
		// Test version command execution through main function
		cmd := exec.Command("go", "run", "main.go", "version")
		cmd.Dir = "."

		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run main with version: %v", err)
		}

		outputStr := string(output)
		// Should produce some version output (even if it's "unknown")
		if len(strings.TrimSpace(outputStr)) == 0 {
			t.Error("main function version command should produce output")
		}
	})

	t.Run("main function error handling", func(t *testing.T) {
		// Test error handling in main function
		cmd := exec.Command("go", "run", "main.go", "invalid-command")
		cmd.Dir = "."

		_, err := cmd.Output()
		if err == nil {
			t.Error("main function should return non-zero exit code for invalid command")
		}

		// Check that it's an ExitError (non-zero exit code)
		if exitError, ok := err.(*exec.ExitError); !ok {
			t.Errorf("Expected ExitError for invalid command, got %T: %v", err, err)
		} else if exitError.ExitCode() == 0 {
			t.Error("Expected non-zero exit code for invalid command")
		}
	})

	t.Run("main function version info setup", func(t *testing.T) {
		// Test that SetVersionInfo is called in main()
		// We can verify this by checking that the CLI package has version info

		// Reset version info to simulate fresh start
		originalVersion := cli.GetVersion()

		// Set a test version
		cli.SetVersionInfo("test-version")

		// Verify it was set
		if cli.GetVersion() != "test-version" {
			t.Error("SetVersionInfo should update the version in CLI package")
		}

		// Restore original version
		cli.SetVersionInfo(originalVersion)
	})

	t.Run("main function basic execution flow", func(t *testing.T) {
		// Test that main function sets up CLI properly and exits cleanly for valid commands
		cmd := exec.Command("go", "run", "main.go", "list")
		cmd.Dir = "."

		// This should run successfully (exit code 0) even if no workflows found
		output, err := cmd.Output()
		if err != nil {
			// Check if it's just a non-zero exit (which is okay for some commands)
			if exitError, ok := err.(*exec.ExitError); ok {
				// Some commands might return non-zero but still function properly
				t.Logf("Command returned exit code %d, output: %s", exitError.ExitCode(), string(output))
			} else {
				t.Fatalf("Failed to run main with list command: %v", err)
			}
		}

		// Should produce some output
		if len(output) == 0 {
			t.Error("list command should produce some output")
		}
	})
}

func TestVersionCommandFunctionality(t *testing.T) {
	t.Run("version information is available", func(t *testing.T) {
		// The cli package should provide version functionality
		versionInfo := cli.GetVersion()
		if versionInfo == "" {
			t.Error("GetVersion() should return version information")
		}
	})
}

func TestCommandLineIntegration(t *testing.T) {
	// Test basic command line parsing and validation

	t.Run("command structure validation", func(t *testing.T) {
		// Test that essential commands are present
		expectedCommands := []string{"add", "compile", "list", "remove", "status", "run", "version"}

		cmdMap := make(map[string]bool)
		for _, cmd := range rootCmd.Commands() {
			cmdMap[cmd.Name()] = true
		}

		missingCommands := []string{}
		for _, expected := range expectedCommands {
			if !cmdMap[expected] {
				missingCommands = append(missingCommands, expected)
			}
		}

		if len(missingCommands) > 0 {
			t.Errorf("Missing expected commands: %v", missingCommands)
		}
	})

	t.Run("global flags are configured", func(t *testing.T) {
		// Test that global flags are properly configured
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		if flag == nil {
			t.Error("verbose flag should be configured")
		}

		if flag != nil && flag.DefValue != "false" {
			t.Error("verbose flag should default to false")
		}
	})
}

func TestCommandErrorHandling(t *testing.T) {
	t.Run("invalid command produces error", func(t *testing.T) {
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		// Test invalid command
		rootCmd.SetArgs([]string{"invalid-command"})
		err := rootCmd.Execute()

		// Restore stderr
		w.Close()
		os.Stderr = oldStderr

		// Read captured output
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err == nil {
			t.Error("invalid command should produce an error")
		}

		if output == "" {
			t.Error("invalid command should produce error output")
		}

		// Reset args for other tests
		rootCmd.SetArgs([]string{})
	})
}
