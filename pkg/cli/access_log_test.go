package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAccessLogParsing(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Create test access.log content
	testLogContent := `1701234567.123    180 192.168.1.100 TCP_MISS/200 1234 GET http://example.com/api/data - HIER_DIRECT/93.184.216.34 text/html
1701234568.456    250 192.168.1.100 TCP_DENIED/403 0 CONNECT github.com:443 - HIER_NONE/- -
1701234569.789    120 192.168.1.100 TCP_HIT/200 5678 GET http://api.github.com/repos - HIER_DIRECT/140.82.112.6 application/json
1701234570.012    0 192.168.1.100 TCP_DENIED/403 0 GET http://malicious.site/evil - HIER_NONE/- -`

	// Write test log file
	accessLogPath := filepath.Join(tempDir, "access.log")
	err := os.WriteFile(accessLogPath, []byte(testLogContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test access.log: %v", err)
	}

	// Test parsing
	analysis, err := parseSquidAccessLog(accessLogPath, false)
	if err != nil {
		t.Fatalf("Failed to parse access log: %v", err)
	}

	// Verify results
	if analysis.TotalRequests != 4 {
		t.Errorf("Expected 4 total requests, got %d", analysis.TotalRequests)
	}

	if analysis.AllowedCount != 2 {
		t.Errorf("Expected 2 allowed requests, got %d", analysis.AllowedCount)
	}

	if analysis.DeniedCount != 2 {
		t.Errorf("Expected 2 denied requests, got %d", analysis.DeniedCount)
	}

	// Check allowed domains
	expectedAllowed := []string{"api.github.com", "example.com"}
	if len(analysis.AllowedDomains) != len(expectedAllowed) {
		t.Errorf("Expected %d allowed domains, got %d", len(expectedAllowed), len(analysis.AllowedDomains))
	}

	// Check denied domains
	expectedDenied := []string{"github.com", "malicious.site"}
	if len(analysis.DeniedDomains) != len(expectedDenied) {
		t.Errorf("Expected %d denied domains, got %d", len(expectedDenied), len(analysis.DeniedDomains))
	}
}

func TestAnalyzeAccessLogs(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Test case 1: No access.log file
	analysis, err := analyzeAccessLogs(tempDir, false)
	if err != nil {
		t.Errorf("Unexpected error when no access.log exists: %v", err)
	}
	if analysis != nil {
		t.Errorf("Expected nil analysis when no access.log exists, got %v", analysis)
	}

	// Test case 2: With access.log file
	testLogContent := `1701234567.123    180 192.168.1.100 TCP_MISS/200 1234 GET http://example.com/ - HIER_DIRECT/93.184.216.34 text/html`
	accessLogPath := filepath.Join(tempDir, "access.log")
	err = os.WriteFile(accessLogPath, []byte(testLogContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test access.log: %v", err)
	}

	analysis, err = analyzeAccessLogs(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to analyze access logs: %v", err)
	}
	if analysis == nil {
		t.Fatal("Expected analysis result, got nil")
	}

	if analysis.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", analysis.TotalRequests)
	}
}

func TestExtractDomainFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://example.com/path", "example.com"},
		{"https://api.github.com/repos", "api.github.com"},
		{"github.com:443", "github.com"},
		{"malicious.site", "malicious.site"},
		{"http://sub.domain.com:8080/path", "sub.domain.com"},
	}

	for _, test := range tests {
		result := extractDomainFromURL(test.url)
		if result != test.expected {
			t.Errorf("extractDomainFromURL(%q) = %q, expected %q", test.url, result, test.expected)
		}
	}
}