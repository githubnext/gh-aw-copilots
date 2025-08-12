package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	// This should work in the current workspace since it's a git repo
	root, err := findGitRoot()
	if err != nil {
		t.Fatalf("Expected to find git root, but got error: %v", err)
	}

	if root == "" {
		t.Fatal("Expected non-empty git root")
	}

	// Check that the returned path exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Fatalf("Git root path does not exist: %s", root)
	}

	// Check that .git directory exists in the root
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf(".git directory does not exist in reported git root: %s", gitDir)
	}

	t.Logf("Git root found: %s", root)
}
