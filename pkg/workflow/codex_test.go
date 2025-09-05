package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexAIConfiguration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "codex-ai-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	tests := []struct {
		name          string
		frontmatter   string
		expectedAI    string
		expectCodex   bool
		expectWarning bool
	}{
		{
			name: "default claude ai",
			frontmatter: `---
tools:
  github:
    allowed: [list_issues]
---`,
			expectedAI:    "claude",
			expectCodex:   false,
			expectWarning: false,
		},
		{
			name: "explicit claude ai",
			frontmatter: `---
engine: claude
tools:
  github:
    allowed: [list_issues]
---`,
			expectedAI:    "claude",
			expectCodex:   false,
			expectWarning: false,
		},
		{
			name: "codex ai",
			frontmatter: `---
engine: codex
tools:
  github:
    allowed: [list_issues]
---`,
			expectedAI:    "codex",
			expectCodex:   true,
			expectWarning: true,
		},
		{
			name: "codex ai without tools",
			frontmatter: `---
engine: codex
---`,
			expectedAI:    "codex",
			expectCodex:   true,
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testContent := tt.frontmatter + `

# Test Workflow

This is a test workflow.
`

			testFile := filepath.Join(tmpDir, tt.name+"-workflow.md")
			if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile the workflow
			err := compiler.CompileWorkflow(testFile)
			if err != nil {
				t.Fatalf("Unexpected error compiling workflow: %v", err)
			}

			// Read the generated lock file
			lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
			content, err := os.ReadFile(lockFile)
			if err != nil {
				t.Fatalf("Failed to read lock file: %v", err)
			}

			lockContent := string(content)

			if tt.expectCodex {
				// Check that Node.js setup is present for codex
				if !strings.Contains(lockContent, "Setup Node.js") {
					t.Errorf("Expected lock file to contain 'Setup Node.js' step for codex but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "actions/setup-node@v4") {
					t.Errorf("Expected lock file to contain Node.js setup action for codex but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that codex installation is present
				if !strings.Contains(lockContent, "Install Codex") {
					t.Errorf("Expected lock file to contain 'Install Codex' step but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "npm install -g @openai/codex") {
					t.Errorf("Expected lock file to contain codex installation command but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that codex command is present
				if !strings.Contains(lockContent, "Run Codex") {
					t.Errorf("Expected lock file to contain 'Run Codex' step but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "codex exec") {
					t.Errorf("Expected lock file to contain 'codex exec' command but it didn't.\nContent:\n%s", lockContent)
				}
				// Check for correct model based on AI setting
				if !strings.Contains(lockContent, "model=o4-mini") {
					t.Errorf("Expected lock file to contain 'model=o4-mini' for codex but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "OPENAI_API_KEY") {
					t.Errorf("Expected lock file to contain 'OPENAI_API_KEY' for codex but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that CODEX_HOME is set
				if !strings.Contains(lockContent, "export CODEX_HOME=/tmp/mcp-config") {
					t.Errorf("Expected lock file to contain 'export CODEX_HOME=/tmp/mcp-config' but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that config.toml is generated (not mcp-servers.json)
				if !strings.Contains(lockContent, "cat > /tmp/mcp-config/config.toml") {
					t.Errorf("Expected lock file to contain config.toml generation for codex but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "[mcp_servers.github]") {
					t.Errorf("Expected lock file to contain '[mcp_servers.github]' section in config.toml but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that history configuration is present
				if !strings.Contains(lockContent, "[history]") {
					t.Errorf("Expected lock file to contain '[history]' section in config.toml but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "persistence = \"none\"") {
					t.Errorf("Expected lock file to contain 'persistence = \"none\"' in config.toml but it didn't.\nContent:\n%s", lockContent)
				}
				// Ensure it does NOT contain mcp-servers.json
				if strings.Contains(lockContent, "mcp-servers.json") {
					t.Errorf("Expected lock file to NOT contain 'mcp-servers.json' when using codex.\nContent:\n%s", lockContent)
				}
				// Check that prompt printing step is present (regardless of engine)
				if !strings.Contains(lockContent, "Print prompt to step summary") {
					t.Errorf("Expected lock file to contain 'Print prompt to step summary' step but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "cat $GITHUB_AW_PROMPT >> $GITHUB_STEP_SUMMARY") {
					t.Errorf("Expected lock file to contain prompt printing command but it didn't.\nContent:\n%s", lockContent)
				}
				// Ensure it does NOT contain Claude Code
				if strings.Contains(lockContent, "Execute Claude Code Action") {
					t.Errorf("Expected lock file to NOT contain 'Execute Claude Code Action' step when using codex.\nContent:\n%s", lockContent)
				}
			} else {
				// Check that Claude Code is present
				if !strings.Contains(lockContent, "Execute Claude Code Action") {
					t.Errorf("Expected lock file to contain 'Execute Claude Code Action' step but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, fmt.Sprintf("anthropics/claude-code-base-action@%s", DefaultClaudeActionVersion)) {
					t.Errorf("Expected lock file to contain Claude Code action but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that prompt printing step is present
				if !strings.Contains(lockContent, "Print prompt to step summary") {
					t.Errorf("Expected lock file to contain 'Print prompt to step summary' step but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "cat $GITHUB_AW_PROMPT >> $GITHUB_STEP_SUMMARY") {
					t.Errorf("Expected lock file to contain prompt printing command but it didn't.\nContent:\n%s", lockContent)
				}
				// Check that mcp-servers.json is generated (not config.toml)
				if !strings.Contains(lockContent, "cat > /tmp/mcp-config/mcp-servers.json") {
					t.Errorf("Expected lock file to contain mcp-servers.json generation for claude but it didn't.\nContent:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "\"mcpServers\":") {
					t.Errorf("Expected lock file to contain '\"mcpServers\":' section in mcp-servers.json but it didn't.\nContent:\n%s", lockContent)
				}
				// Ensure it does NOT contain codex
				if strings.Contains(lockContent, "codex exec") {
					t.Errorf("Expected lock file to NOT contain 'codex exec' when using claude.\nContent:\n%s", lockContent)
				}
				// Ensure it does NOT contain config.toml
				if strings.Contains(lockContent, "config.toml") {
					t.Errorf("Expected lock file to NOT contain 'config.toml' when using claude.\nContent:\n%s", lockContent)
				}
				// Ensure it does NOT contain CODEX_HOME
				if strings.Contains(lockContent, "CODEX_HOME") {
					t.Errorf("Expected lock file to NOT contain 'CODEX_HOME' when using claude.\nContent:\n%s", lockContent)
				}
			}
		})
	}
}

func TestCodexMCPConfigGeneration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "codex-mcp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := NewCompiler(false, "", "test")

	tests := []struct {
		name                 string
		frontmatter          string
		expectedAI           string
		expectConfigToml     bool
		expectMcpServersJson bool
		expectCodexHome      bool
	}{
		{
			name: "codex with github tools generates config.toml",
			frontmatter: `---
engine: codex
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "codex",
			expectConfigToml:     true,
			expectMcpServersJson: false,
			expectCodexHome:      true,
		},
		{
			name: "claude with github tools generates mcp-servers.json",
			frontmatter: `---
engine: claude
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "claude",
			expectConfigToml:     false,
			expectMcpServersJson: true,
			expectCodexHome:      false,
		},
		{
			name: "codex with docker github tools generates config.toml",
			frontmatter: `---
engine: codex
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "codex",
			expectConfigToml:     true,
			expectMcpServersJson: false,
			expectCodexHome:      true,
		},
		{
			name: "claude with docker github tools generates mcp-servers.json",
			frontmatter: `---
engine: claude
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "claude",
			expectConfigToml:     false,
			expectMcpServersJson: true,
			expectCodexHome:      false,
		},
		{
			name: "codex with services github tools generates config.toml",
			frontmatter: `---
engine: codex
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "codex",
			expectConfigToml:     true,
			expectMcpServersJson: false,
			expectCodexHome:      true,
		},
		{
			name: "claude with services github tools generates mcp-servers.json",
			frontmatter: `---
engine: claude
tools:
  github:
    allowed: [get_issue, create_issue]
---`,
			expectedAI:           "claude",
			expectConfigToml:     false,
			expectMcpServersJson: true,
			expectCodexHome:      false,
		},
		{
			name: "codex with custom MCP tools generates config.toml",
			frontmatter: `---
engine: codex
tools:
  github:
    allowed: [get_issue, create_issue]
  custom-server:
    mcp:
      type: stdio
      command: "python"
      args: ["-m", "my_server"]
      env:
        API_KEY: "${{ secrets.API_KEY }}"
    allowed: ["*"]
---`,
			expectedAI:           "codex",
			expectConfigToml:     true,
			expectMcpServersJson: false,
			expectCodexHome:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testContent := tt.frontmatter + `

# Test MCP Configuration

This is a test workflow for MCP configuration with different AI engines.
`

			testFile := filepath.Join(tmpDir, tt.name+"-workflow.md")
			if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Compile the workflow
			err := compiler.CompileWorkflow(testFile)
			if err != nil {
				t.Fatalf("Unexpected error compiling workflow: %v", err)
			}

			// Read the generated lock file
			lockFile := strings.TrimSuffix(testFile, ".md") + ".lock.yml"
			content, err := os.ReadFile(lockFile)
			if err != nil {
				t.Fatalf("Failed to read lock file: %v", err)
			}

			lockContent := string(content)

			// Test config.toml generation
			if tt.expectConfigToml {
				if !strings.Contains(lockContent, "cat > /tmp/mcp-config/config.toml") {
					t.Errorf("Expected config.toml generation but didn't find it in:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "[mcp_servers.github]") {
					t.Errorf("Expected [mcp_servers.github] section but didn't find it in:\n%s", lockContent)
				}

				if !strings.Contains(lockContent, "command = \"docker\"") {
					t.Errorf("Expected docker command in config.toml but didn't find it in:\n%s", lockContent)
				}
				// Check for custom MCP server if test includes it
				if strings.Contains(tt.name, "custom MCP") {
					if !strings.Contains(lockContent, "[mcp_servers.custom-server]") {
						t.Errorf("Expected [mcp_servers.custom-server] section but didn't find it in:\n%s", lockContent)
					}
					if !strings.Contains(lockContent, "command = \"python\"") {
						t.Errorf("Expected python command for custom server but didn't find it in:\n%s", lockContent)
					}
					if !strings.Contains(lockContent, "\"API_KEY\" = \"${{ secrets.API_KEY }}\"") {
						t.Errorf("Expected API_KEY env var for custom server but didn't find it in:\n%s", lockContent)
					}
				}
				// Should NOT have services section (services mode removed)
				if strings.Contains(lockContent, "services:") {
					t.Errorf("Expected NO services section in workflow but found it in:\n%s", lockContent)
				}
			} else {
				if strings.Contains(lockContent, "config.toml") {
					t.Errorf("Expected NO config.toml but found it in:\n%s", lockContent)
				}
			}

			// Test mcp-servers.json generation
			if tt.expectMcpServersJson {
				if !strings.Contains(lockContent, "cat > /tmp/mcp-config/mcp-servers.json") {
					t.Errorf("Expected mcp-servers.json generation but didn't find it in:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "\"mcpServers\":") {
					t.Errorf("Expected mcpServers section but didn't find it in:\n%s", lockContent)
				}
				if !strings.Contains(lockContent, "\"github\":") {
					t.Errorf("Expected github section in JSON but didn't find it in:\n%s", lockContent)
				}

				if !strings.Contains(lockContent, "\"command\": \"docker\"") {
					t.Errorf("Expected docker command in mcp-servers.json but didn't find it in:\n%s", lockContent)
				}
				// Should NOT have services section (services mode removed)
				if strings.Contains(lockContent, "services:") {
					t.Errorf("Expected NO services section in workflow but found it in:\n%s", lockContent)
				}
			} else {
				if strings.Contains(lockContent, "mcp-servers.json") {
					t.Errorf("Expected NO mcp-servers.json but found it in:\n%s", lockContent)
				}
			}

			// Test CODEX_HOME setting
			if tt.expectCodexHome {
				if !strings.Contains(lockContent, "export CODEX_HOME=/tmp/mcp-config") {
					t.Errorf("Expected CODEX_HOME export but didn't find it in:\n%s", lockContent)
				}
			} else {
				if strings.Contains(lockContent, "CODEX_HOME") {
					t.Errorf("Expected NO CODEX_HOME but found it in:\n%s", lockContent)
				}
			}

			// Verify AI type
			if tt.expectedAI == "codex" {
				if !strings.Contains(lockContent, "codex exec") {
					t.Errorf("Expected codex exec command but didn't find it in:\n%s", lockContent)
				}
				if strings.Contains(lockContent, "claude-code-base-action") {
					t.Errorf("Expected NO claude action but found it in:\n%s", lockContent)
				}
			} else {
				if !strings.Contains(lockContent, "claude-code-base-action") {
					t.Errorf("Expected claude action but didn't find it in:\n%s", lockContent)
				}
				if strings.Contains(lockContent, "codex exec") {
					t.Errorf("Expected NO codex exec but found it in:\n%s", lockContent)
				}
			}
		})
	}
}
