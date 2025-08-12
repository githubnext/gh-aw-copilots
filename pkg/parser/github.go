package parser

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetGitHubToken attempts to get GitHub token from environment or gh CLI
func GetGitHubToken() (string, error) {
	// First try environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token, nil
	}

	// Fall back to gh auth token command
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("GITHUB_TOKEN environment variable not set and 'gh auth token' failed: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN environment variable not set and 'gh auth token' returned empty token")
	}

	return token, nil
}
