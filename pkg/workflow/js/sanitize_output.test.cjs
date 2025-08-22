import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import fs from 'fs';
import path from 'path';

// Mock the global objects that GitHub Actions provides
const mockCore = {
  setOutput: vi.fn()
};

// Set up global variables
global.core = mockCore;

describe('sanitize_output.cjs', () => {
  let sanitizeScript;
  let sanitizeContentFunction;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();
    
    // Reset environment variables
    delete process.env.GITHUB_AW_OUTPUT;
    delete process.env.GITHUB_AW_ALLOWED_DOMAINS;
    
    // Read the script content
    const scriptPath = path.join(process.cwd(), 'pkg/workflow/js/sanitize_output.cjs');
    sanitizeScript = fs.readFileSync(scriptPath, 'utf8');
    
    // Extract sanitizeContent function for unit testing
    // We need to eval the script to get access to the function
    const scriptWithExport = sanitizeScript.replace(
      'await main();',
      'global.testSanitizeContent = sanitizeContent;'
    );
    eval(scriptWithExport);
    sanitizeContentFunction = global.testSanitizeContent;
  });

  describe('sanitizeContent function', () => {
    it('should handle null and undefined inputs', () => {
      expect(sanitizeContentFunction(null)).toBe('');
      expect(sanitizeContentFunction(undefined)).toBe('');
      expect(sanitizeContentFunction('')).toBe('');
    });

    it('should neutralize @mentions by wrapping in backticks', () => {
      const input = 'Hello @user and @org/team';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('`@user`');
      expect(result).toContain('`@org/team`');
    });

    it('should not neutralize @mentions inside code blocks', () => {
      const input = 'Check `@user` in code and @realuser outside';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('`@user`'); // Already in backticks, stays as is
      expect(result).toContain('`@realuser`'); // Gets wrapped
    });

    it('should neutralize bot trigger phrases', () => {
      const input = 'This fixes #123 and closes #456. Also resolves #789';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('`fixes #123`');
      expect(result).toContain('`closes #456`');
      expect(result).toContain('`resolves #789`');
    });

    it('should remove control characters except newlines and tabs', () => {
      const input = 'Hello\x00world\x0C\nNext line\t\x1Fbad';
      const result = sanitizeContentFunction(input);
      expect(result).not.toContain('\x00');
      expect(result).not.toContain('\x0C');
      expect(result).not.toContain('\x1F');
      expect(result).toContain('\n');
      expect(result).toContain('\t');
    });

    it('should escape XML characters', () => {
      const input = '<script>alert("test")</script> & more';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('&lt;script&gt;');
      expect(result).toContain('&quot;test&quot;');
      expect(result).toContain('&amp; more');
    });

    it('should block HTTP URLs while preserving HTTPS URLs', () => {
      const input = 'HTTP: http://bad.com and HTTPS: https://github.com';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('(redacted)'); // HTTP URL blocked
      expect(result).toContain('https://github.com'); // HTTPS URL preserved
      expect(result).not.toContain('http://bad.com');
    });

    it('should block various unsafe protocols', () => {
      const input = 'Bad: ftp://file.com javascript:alert(1) file://local data:text/html,<script>';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('(redacted)');
      expect(result).not.toContain('ftp://');
      expect(result).not.toContain('javascript:');
      expect(result).not.toContain('file://');
      expect(result).not.toContain('data:');
    });

    it('should preserve HTTPS URLs for allowed domains', () => {
      const input = 'Links: https://github.com/user/repo https://github.io/page https://githubusercontent.com/file';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('https://github.com/user/repo');
      expect(result).toContain('https://github.io/page');
      expect(result).toContain('https://githubusercontent.com/file');
    });

    it('should block HTTPS URLs for disallowed domains', () => {
      const input = 'Bad: https://evil.com/malware Good: https://github.com/repo';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('(redacted)'); // evil.com blocked
      expect(result).toContain('https://github.com/repo'); // github.com allowed
      expect(result).not.toContain('https://evil.com');
    });

    it('should respect custom allowed domains from environment', () => {
      process.env.GITHUB_AW_ALLOWED_DOMAINS = 'example.com,trusted.org';
      
      // Re-run the script setup to pick up env variable
      const scriptWithExport = sanitizeScript.replace(
        'await main();',
        'global.testSanitizeContent = sanitizeContent;'
      );
      eval(scriptWithExport);
      const customSanitize = global.testSanitizeContent;
      
      const input = 'Links: https://example.com/page https://trusted.org/file https://github.com/repo';
      const result = customSanitize(input);
      expect(result).toContain('https://example.com/page');
      expect(result).toContain('https://trusted.org/file');
      expect(result).toContain('(redacted)'); // github.com now blocked
      expect(result).not.toContain('https://github.com/repo');
    });

    it('should handle subdomain matching correctly', () => {
      const input = 'Subdomains: https://api.github.com/v1 https://docs.github.com/guide';
      const result = sanitizeContentFunction(input);
      expect(result).toContain('https://api.github.com/v1');
      expect(result).toContain('https://docs.github.com/guide');
    });

    it('should truncate content that exceeds maximum length', () => {
      const longContent = 'x'.repeat(600000); // Exceeds 524288 limit
      const result = sanitizeContentFunction(longContent);
      expect(result.length).toBeLessThan(600000);
      expect(result).toContain('[Content truncated due to length]');
    });

    it('should truncate content that exceeds maximum lines', () => {
      const manyLines = '\n'.repeat(70000); // Exceeds 65000 limit
      const result = sanitizeContentFunction(manyLines);
      const lines = result.split('\n');
      expect(lines.length).toBeLessThanOrEqual(65001); // +1 for truncation message
      expect(result).toContain('[Content truncated due to line count]');
    });

    it('should remove ANSI escape sequences', () => {
      const input = '\x1b[31mRed text\x1b[0m \x1b[1;32mBold green\x1b[m';
      const result = sanitizeContentFunction(input);
      expect(result).not.toContain('\x1b[');
      expect(result).toContain('Red text');
      expect(result).toContain('Bold green');
    });

    it('should handle complex mixed content correctly', () => {
      const input = `
# Issue Report by @user

This fixes #123 and has links:
- HTTP: http://bad.com (should be blocked)
- HTTPS: https://github.com/repo (should be preserved)
- JavaScript: javascript:alert('xss') (should be blocked)

<script>alert("xss")</script>

Special chars: \x00\x1F & "quotes" 'apostrophes'
      `.trim();
      
      const result = sanitizeContentFunction(input);
      
      // Check @mention neutralization
      expect(result).toContain('`@user`');
      
      // Check bot trigger neutralization
      expect(result).toContain('`fixes #123`');
      
      // Check URL filtering
      expect(result).toContain('(redacted)'); // HTTP and JavaScript URLs
      expect(result).toContain('https://github.com/repo');
      expect(result).not.toContain('http://bad.com');
      expect(result).not.toContain('javascript:alert');
      
      // Check XML escaping
      expect(result).toContain('&lt;script&gt;');
      expect(result).toContain('&quot;quotes&quot;');
      expect(result).toContain('&apos;apostrophes&apos;');
      expect(result).toContain('&amp;');
      
      // Check control character removal
      expect(result).not.toContain('\x00');
      expect(result).not.toContain('\x1F');
    });

    it('should trim excessive whitespace', () => {
      const input = '   \n\n  Content with spacing  \n\n  ';
      const result = sanitizeContentFunction(input);
      expect(result).toBe('Content with spacing');
    });

    it('should handle empty environment variable gracefully', () => {
      process.env.GITHUB_AW_ALLOWED_DOMAINS = '  ,  ,  ';
      
      const scriptWithExport = sanitizeScript.replace(
        'await main();',
        'global.testSanitizeContent = sanitizeContent;'
      );
      eval(scriptWithExport);
      const customSanitize = global.testSanitizeContent;
      
      const input = 'Link: https://github.com/repo';
      const result = customSanitize(input);
      // With empty allowedDomains array, all HTTPS URLs get blocked
      expect(result).toContain('(redacted)');
      expect(result).not.toContain('https://github.com/repo');
    });
  });

  describe('main function', () => {
    beforeEach(() => {
      // Clean up any test files
      const testFile = '/tmp/test-output.txt';
      if (fs.existsSync(testFile)) {
        fs.unlinkSync(testFile);
      }
      
      // Make fs available globally for the evaluated script
      global.fs = fs;
    });

    afterEach(() => {
      // Clean up global fs
      delete global.fs;
    });

    it('should handle missing GITHUB_AW_OUTPUT environment variable', async () => {
      delete process.env.GITHUB_AW_OUTPUT;
      
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);
      
      expect(consoleSpy).toHaveBeenCalledWith('GITHUB_AW_OUTPUT not set, no output to collect');
      expect(mockCore.setOutput).toHaveBeenCalledWith('output', '');
      
      consoleSpy.mockRestore();
    });

    it('should handle non-existent output file', async () => {
      process.env.GITHUB_AW_OUTPUT = '/tmp/non-existent-file.txt';
      
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);
      
      expect(consoleSpy).toHaveBeenCalledWith('Output file does not exist:', '/tmp/non-existent-file.txt');
      expect(mockCore.setOutput).toHaveBeenCalledWith('output', '');
      
      consoleSpy.mockRestore();
    });

    it('should handle empty output file', async () => {
      const testFile = '/tmp/test-empty-output.txt';
      fs.writeFileSync(testFile, '   \n  \t  \n  ');
      process.env.GITHUB_AW_OUTPUT = testFile;
      
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);
      
      expect(consoleSpy).toHaveBeenCalledWith('Output file is empty');
      expect(mockCore.setOutput).toHaveBeenCalledWith('output', '');
      
      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it('should process and sanitize output file content', async () => {
      const testContent = 'Hello @user! This fixes #123. Link: http://bad.com and https://github.com/repo';
      const testFile = '/tmp/test-output.txt';
      fs.writeFileSync(testFile, testContent);
      process.env.GITHUB_AW_OUTPUT = testFile;
      
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'Collected agentic output (sanitized):',
        expect.stringContaining('`@user`')
      );
      
      const outputCall = mockCore.setOutput.mock.calls.find(call => call[0] === 'output');
      expect(outputCall).toBeDefined();
      const sanitizedOutput = outputCall[1];
      
      // Verify sanitization occurred
      expect(sanitizedOutput).toContain('`@user`');
      expect(sanitizedOutput).toContain('`fixes #123`');
      expect(sanitizedOutput).toContain('(redacted)'); // HTTP URL
      expect(sanitizedOutput).toContain('https://github.com/repo'); // HTTPS URL preserved
      
      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it('should truncate log output for very long content', async () => {
      const longContent = 'x'.repeat(250); // More than 200 chars to trigger truncation
      const testFile = '/tmp/test-long-output.txt';
      fs.writeFileSync(testFile, longContent);
      process.env.GITHUB_AW_OUTPUT = testFile;
      
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);
      
      const logCalls = consoleSpy.mock.calls;
      const outputLogCall = logCalls.find(call => 
        call[0] && call[0].includes('Collected agentic output (sanitized):')
      );
      
      expect(outputLogCall).toBeDefined();
      expect(outputLogCall[1]).toContain('...');
      expect(outputLogCall[1].length).toBeLessThan(longContent.length);
      
      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });
  });
});