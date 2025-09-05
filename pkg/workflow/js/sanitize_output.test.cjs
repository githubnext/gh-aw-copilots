import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import fs from "fs";
import path from "path";

// Mock the global objects that GitHub Actions provides
const mockCore = {
  setOutput: vi.fn(),
  warning: vi.fn(),
  error: vi.fn(),
};

// Set up global variables
global.core = mockCore;

describe("sanitize_output.cjs", () => {
  let sanitizeScript;
  let sanitizeContentFunction;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Reset environment variables
    delete process.env.GITHUB_AW_SAFE_OUTPUTS;
    delete process.env.GITHUB_AW_ALLOWED_DOMAINS;

    // Read the script content
    const scriptPath = path.join(
      process.cwd(),
      "pkg/workflow/js/sanitize_output.cjs"
    );
    sanitizeScript = fs.readFileSync(scriptPath, "utf8");

    // Extract sanitizeContent function for unit testing
    // We need to eval the script to get access to the function
    const scriptWithExport = sanitizeScript.replace(
      "await main();",
      "global.testSanitizeContent = sanitizeContent;"
    );
    eval(scriptWithExport);
    sanitizeContentFunction = global.testSanitizeContent;
  });

  describe("sanitizeContent function", () => {
    it("should handle null and undefined inputs", () => {
      expect(sanitizeContentFunction(null)).toBe("");
      expect(sanitizeContentFunction(undefined)).toBe("");
      expect(sanitizeContentFunction("")).toBe("");
    });

    it("should neutralize @mentions by wrapping in backticks", () => {
      const input = "Hello @user and @org/team";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`@user`");
      expect(result).toContain("`@org/team`");
    });

    it("should not neutralize @mentions inside code blocks", () => {
      const input = "Check `@user` in code and @realuser outside";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`@user`"); // Already in backticks, stays as is
      expect(result).toContain("`@realuser`"); // Gets wrapped
    });

    it("should neutralize bot trigger phrases", () => {
      const input = "This fixes #123 and closes #456. Also resolves #789";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`fixes #123`");
      expect(result).toContain("`closes #456`");
      expect(result).toContain("`resolves #789`");
    });

    it("should remove control characters except newlines and tabs", () => {
      const input = "Hello\x00world\x0C\nNext line\t\x1Fbad";
      const result = sanitizeContentFunction(input);
      expect(result).not.toContain("\x00");
      expect(result).not.toContain("\x0C");
      expect(result).not.toContain("\x1F");
      expect(result).toContain("\n");
      expect(result).toContain("\t");
    });

    it("should escape XML characters", () => {
      const input = '<script>alert("test")</script> & more';
      const result = sanitizeContentFunction(input);
      expect(result).toContain("&lt;script&gt;");
      expect(result).toContain("&quot;test&quot;");
      expect(result).toContain("&amp; more");
    });

    it("should block HTTP URLs while preserving HTTPS URLs", () => {
      const input = "HTTP: http://bad.com and HTTPS: https://github.com";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("(redacted)"); // HTTP URL blocked
      expect(result).toContain("https://github.com"); // HTTPS URL preserved
      expect(result).not.toContain("http://bad.com");
    });

    it("should block various unsafe protocols", () => {
      const input =
        "Bad: ftp://file.com javascript:alert(1) file://local data:text/html,<script>";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("(redacted)");
      expect(result).not.toContain("ftp://");
      expect(result).not.toContain("javascript:");
      expect(result).not.toContain("file://");
      expect(result).not.toContain("data:");
    });

    it("should preserve HTTPS URLs for allowed domains", () => {
      const input =
        "Links: https://github.com/user/repo https://github.io/page https://githubusercontent.com/file";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("https://github.com/user/repo");
      expect(result).toContain("https://github.io/page");
      expect(result).toContain("https://githubusercontent.com/file");
    });

    it("should block HTTPS URLs for disallowed domains", () => {
      const input =
        "Bad: https://evil.com/malware Good: https://github.com/repo";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("(redacted)"); // evil.com blocked
      expect(result).toContain("https://github.com/repo"); // github.com allowed
      expect(result).not.toContain("https://evil.com");
    });

    it("should respect custom allowed domains from environment", () => {
      process.env.GITHUB_AW_ALLOWED_DOMAINS = "example.com,trusted.org";

      // Re-run the script setup to pick up env variable
      const scriptWithExport = sanitizeScript.replace(
        "await main();",
        "global.testSanitizeContent = sanitizeContent;"
      );
      eval(scriptWithExport);
      const customSanitize = global.testSanitizeContent;

      const input =
        "Links: https://example.com/page https://trusted.org/file https://github.com/repo";
      const result = customSanitize(input);
      expect(result).toContain("https://example.com/page");
      expect(result).toContain("https://trusted.org/file");
      expect(result).toContain("(redacted)"); // github.com now blocked
      expect(result).not.toContain("https://github.com/repo");
    });

    it("should handle subdomain matching correctly", () => {
      const input =
        "Subdomains: https://api.github.com/v1 https://docs.github.com/guide";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("https://api.github.com/v1");
      expect(result).toContain("https://docs.github.com/guide");
    });

    it("should truncate content that exceeds maximum length", () => {
      const longContent = "x".repeat(600000); // Exceeds 524288 limit
      const result = sanitizeContentFunction(longContent);
      expect(result.length).toBeLessThan(600000);
      expect(result).toContain("[Content truncated due to length]");
    });

    it("should truncate content that exceeds maximum lines", () => {
      const manyLines = "\n".repeat(70000); // Exceeds 65000 limit
      const result = sanitizeContentFunction(manyLines);
      const lines = result.split("\n");
      expect(lines.length).toBeLessThanOrEqual(65001); // +1 for truncation message
      expect(result).toContain("[Content truncated due to line count]");
    });

    it("should remove ANSI escape sequences", () => {
      const input = "\x1b[31mRed text\x1b[0m \x1b[1;32mBold green\x1b[m";
      const result = sanitizeContentFunction(input);
      expect(result).not.toContain("\x1b[");
      expect(result).toContain("Red text");
      expect(result).toContain("Bold green");
    });

    it("should handle complex mixed content correctly", () => {
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
      expect(result).toContain("`@user`");

      // Check bot trigger neutralization
      expect(result).toContain("`fixes #123`");

      // Check URL filtering
      expect(result).toContain("(redacted)"); // HTTP and JavaScript URLs
      expect(result).toContain("https://github.com/repo");
      expect(result).not.toContain("http://bad.com");
      expect(result).not.toContain("javascript:alert");

      // Check XML escaping
      expect(result).toContain("&lt;script&gt;");
      expect(result).toContain("&quot;quotes&quot;");
      expect(result).toContain("&apos;apostrophes&apos;");
      expect(result).toContain("&amp;");

      // Check control character removal
      expect(result).not.toContain("\x00");
      expect(result).not.toContain("\x1F");
    });

    it("should trim excessive whitespace", () => {
      const input = "   \n\n  Content with spacing  \n\n  ";
      const result = sanitizeContentFunction(input);
      expect(result).toBe("Content with spacing");
    });

    it("should handle empty environment variable gracefully", () => {
      process.env.GITHUB_AW_ALLOWED_DOMAINS = "  ,  ,  ";

      const scriptWithExport = sanitizeScript.replace(
        "await main();",
        "global.testSanitizeContent = sanitizeContent;"
      );
      eval(scriptWithExport);
      const customSanitize = global.testSanitizeContent;

      const input = "Link: https://github.com/repo";
      const result = customSanitize(input);
      // With empty allowedDomains array, all HTTPS URLs get blocked
      expect(result).toContain("(redacted)");
      expect(result).not.toContain("https://github.com/repo");
    });

    it("should handle @mentions with various formats", () => {
      const input =
        "Contact @user123, @org-name/team_name, @a, and @normalname";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`@user123`");
      expect(result).toContain("`@org-name/team_name`");
      expect(result).toContain("`@a`");
      expect(result).toContain("`@normalname`");
    });

    it("should not neutralize @mentions at start of backticked expressions", () => {
      const input = "Code: `@user.method()` and normal @user mention";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`@user.method()`"); // Should remain unchanged
      expect(result).toContain("`@user`"); // Should be neutralized
    });

    it("should handle various bot trigger phrase formats", () => {
      const input =
        "Fix #123, close #abc, FIXES #XYZ, resolves #456, fixes    #789";
      const result = sanitizeContentFunction(input);
      expect(result).toContain("`Fix #123`");
      expect(result).toContain("`close #abc`");
      expect(result).toContain("`FIXES #XYZ`");
      expect(result).toContain("`resolves #456`"); // With space
      expect(result).toContain("`fixes #789`"); // Multiple spaces normalized to single
    });

    it("should handle edge cases in protocol filtering", () => {
      const input = `
        Protocols: HTTP://CAPS.COM, https://github.com/path?query=value#fragment
        More: mailto:user@domain.com tel:+1234567890 ssh://server:22/path
        Edge: ://malformed http:// https:// 
        Nested: (https://github.com) [http://bad.com] "ftp://files.com"
      `;
      const result = sanitizeContentFunction(input);

      // Check case insensitive protocol blocking
      expect(result).toContain("(redacted)"); // HTTP://CAPS.COM
      expect(result).toContain("https://github.com/path?query=value#fragment");
      expect(result).toContain("(redacted)"); // mailto, tel, ssh, http, ftp
      expect(result).not.toContain("HTTP://CAPS.COM");
      expect(result).not.toContain("mailto:user@domain.com");
      expect(result).not.toContain("tel:+1234567890");
      expect(result).not.toContain("ssh://server:22/path");
    });

    it("should preserve HTTPS URLs in various contexts", () => {
      const input = `
        Links in text: Visit https://github.com/user/repo for details.
        In parentheses: (https://github.io/docs)
        In brackets: [https://githubusercontent.com/file.txt]
        Multiple: https://github.com https://github.io https://githubassets.com
      `;
      const result = sanitizeContentFunction(input);

      expect(result).toContain("https://github.com/user/repo");
      expect(result).toContain("https://github.io/docs");
      expect(result).toContain("https://githubusercontent.com/file.txt");
      expect(result).toContain("https://github.com");
      expect(result).toContain("https://github.io");
      expect(result).toContain("https://githubassets.com");
    });

    it("should handle complex domain matching scenarios", () => {
      const input = `
        Valid: https://api.github.com/v4/graphql https://docs.github.com/en/
        Invalid: https://github.com.evil.com https://notgithub.com
        Edge: https://github.com.attacker.com https://sub.github.io.fake.com
      `;
      const result = sanitizeContentFunction(input);

      // Valid subdomains should be preserved
      expect(result).toContain("https://api.github.com/v4/graphql");
      expect(result).toContain("https://docs.github.com/en/");

      // Invalid domains should be blocked
      expect(result).toContain("(redacted)");
      expect(result).not.toContain("github.com.evil.com");
      expect(result).not.toContain("notgithub.com");
      expect(result).not.toContain("github.com.attacker.com");
      expect(result).not.toContain("sub.github.io.fake.com");
    });

    it("should handle URLs with special characters and edge cases", () => {
      const input = `
        URLs: https://github.com/user/repo-name_with.dots
        Query: https://github.com/search?q=test&type=code
        Fragment: https://github.com/user/repo#readme
        Port: https://github.dev:443/workspace
        Auth: https://github.com/repo (user info stripped by domain parsing)
      `;
      const result = sanitizeContentFunction(input);

      expect(result).toContain("https://github.com/user/repo-name_with.dots");
      expect(result).toContain(
        "https://github.com/search?q=test&amp;type=code"
      ); // & escaped
      expect(result).toContain("https://github.com/user/repo#readme");
      expect(result).toContain("https://github.dev:443/workspace");
      expect(result).toContain("https://github.com/repo");
    });

    it("should handle length truncation at exact boundary", () => {
      const exactLength = 524288;
      const input = "x".repeat(exactLength);
      const result = sanitizeContentFunction(input);
      expect(result.length).toBe(exactLength);
      expect(result).not.toContain("[Content truncated due to length]");

      const overLength = "x".repeat(exactLength + 100); // Significantly longer
      const overResult = sanitizeContentFunction(overLength);
      // The result should be truncated and contain the truncation message
      expect(overResult).toContain("[Content truncated due to length]");
      // The result should be shorter than the original due to truncation
      expect(overResult.length).toBeLessThan(overLength.length);
    });

    it("should handle line truncation at exact boundary", () => {
      const exactLines = 65000;
      // Create content with exactly 65000 lines (65000 newlines = 65001 elements when split)
      const input = Array(exactLines).fill("line").join("\n");
      const result = sanitizeContentFunction(input);
      const lines = result.split("\n");
      expect(lines.length).toBe(exactLines);
      expect(result).not.toContain("[Content truncated due to line count]");

      // Test with more than 65000 lines
      const overLines = Array(exactLines + 1)
        .fill("line")
        .join("\n");
      const overResult = sanitizeContentFunction(overLines);
      const overResultLines = overResult.split("\n");
      expect(overResultLines.length).toBeLessThanOrEqual(exactLines + 1); // +1 for truncation message
      expect(overResult).toContain("[Content truncated due to line count]");
    });

    it("should handle various ANSI escape sequence patterns", () => {
      const input = `
        Color: \x1b[31mRed\x1b[0m \x1b[1;32mBold Green\x1b[m
        Cursor: \x1b[2J\x1b[H Clear and home
        Other: \x1b[?25h Show cursor \x1b[K Clear line
        Complex: \x1b[38;5;196mTrueColor\x1b[0m
      `;
      const result = sanitizeContentFunction(input);

      expect(result).not.toContain("\x1b[");
      expect(result).toContain("Red");
      expect(result).toContain("Bold Green");
      expect(result).toContain("Clear and home");
      expect(result).toContain("Show cursor");
      expect(result).toContain("Clear line");
      expect(result).toContain("TrueColor");
    });

    it("should handle XML escaping in complex nested content", () => {
      const input = `
        <xml attr="value & 'quotes'">
          <![CDATA[<script>alert("xss")</script>]]>
          <!-- comment with "quotes" & 'apostrophes' -->
        </xml>
      `;
      const result = sanitizeContentFunction(input);

      expect(result).toContain(
        "&lt;xml attr=&quot;value &amp; &apos;quotes&apos;&quot;&gt;"
      );
      expect(result).toContain(
        "&lt;![CDATA[&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;]]&gt;"
      );
      expect(result).toContain(
        "&lt;!-- comment with &quot;quotes&quot; &amp; &apos;apostrophes&apos; --&gt;"
      );
      expect(result).toContain("&lt;/xml&gt;");
    });

    it("should handle non-string inputs robustly", () => {
      expect(sanitizeContentFunction(123)).toBe("");
      expect(sanitizeContentFunction({})).toBe("");
      expect(sanitizeContentFunction([])).toBe("");
      expect(sanitizeContentFunction(true)).toBe("");
      expect(sanitizeContentFunction(false)).toBe("");
    });

    it("should preserve line breaks and tabs in content structure", () => {
      const input = `Line 1
\t\tIndented line
\n\nDouble newline

\tTab at start`;
      const result = sanitizeContentFunction(input);

      expect(result).toContain("\n");
      expect(result).toContain("\t");
      expect(result.split("\n").length).toBeGreaterThan(1);
      expect(result).toContain("Line 1");
      expect(result).toContain("Indented line");
      expect(result).toContain("Tab at start");
    });

    it("should handle simultaneous protocol and domain filtering", () => {
      const input = `
        Good HTTPS: https://github.com/repo
        Bad HTTPS: https://evil.com/malware  
        Bad HTTP allowed domain: http://github.com/repo
        Mixed: https://evil.com/path?goto=https://github.com/safe
      `;
      const result = sanitizeContentFunction(input);

      expect(result).toContain("https://github.com/repo");
      expect(result).toContain("(redacted)"); // For evil.com and http://github.com
      expect(result).not.toContain("https://evil.com");
      expect(result).not.toContain("http://github.com");

      // The safe URL in query param should still be preserved
      expect(result).toContain("https://github.com/safe");
    });
  });

  describe("main function", () => {
    beforeEach(() => {
      // Clean up any test files
      const testFile = "/tmp/test-output.txt";
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

    it("should handle missing GITHUB_AW_SAFE_OUTPUTS environment variable", async () => {
      delete process.env.GITHUB_AW_SAFE_OUTPUTS;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "GITHUB_AW_SAFE_OUTPUTS not set, no output to collect"
      );
      expect(mockCore.setOutput).toHaveBeenCalledWith("output", "");

      consoleSpy.mockRestore();
    });

    it("should handle non-existent output file", async () => {
      process.env.GITHUB_AW_SAFE_OUTPUTS = "/tmp/non-existent-file.txt";

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "Output file does not exist:",
        "/tmp/non-existent-file.txt"
      );
      expect(mockCore.setOutput).toHaveBeenCalledWith("output", "");

      consoleSpy.mockRestore();
    });

    it("should handle empty output file", async () => {
      const testFile = "/tmp/test-empty-output.txt";
      fs.writeFileSync(testFile, "   \n  \t  \n  ");
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith("Output file is empty");
      expect(mockCore.setOutput).toHaveBeenCalledWith("output", "");

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should process and sanitize output file content", async () => {
      const testContent =
        "Hello @user! This fixes #123. Link: http://bad.com and https://github.com/repo";
      const testFile = "/tmp/test-output.txt";
      fs.writeFileSync(testFile, testContent);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "Collected agentic output (sanitized):",
        expect.stringContaining("`@user`")
      );

      const outputCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "output"
      );
      expect(outputCall).toBeDefined();
      const sanitizedOutput = outputCall[1];

      // Verify sanitization occurred
      expect(sanitizedOutput).toContain("`@user`");
      expect(sanitizedOutput).toContain("`fixes #123`");
      expect(sanitizedOutput).toContain("(redacted)"); // HTTP URL
      expect(sanitizedOutput).toContain("https://github.com/repo"); // HTTPS URL preserved

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should truncate log output for very long content", async () => {
      const longContent = "x".repeat(250); // More than 200 chars to trigger truncation
      const testFile = "/tmp/test-long-output.txt";
      fs.writeFileSync(testFile, longContent);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      const logCalls = consoleSpy.mock.calls;
      const outputLogCall = logCalls.find(
        call =>
          call[0] && call[0].includes("Collected agentic output (sanitized):")
      );

      expect(outputLogCall).toBeDefined();
      expect(outputLogCall[1]).toContain("...");
      expect(outputLogCall[1].length).toBeLessThan(longContent.length);

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should handle file read errors gracefully", async () => {
      // Create a file and then remove read permissions
      const testFile = "/tmp/test-no-read.txt";
      fs.writeFileSync(testFile, "test content");

      // Mock readFileSync to throw an error
      const originalReadFileSync = fs.readFileSync;
      const readFileSyncSpy = vi
        .spyOn(fs, "readFileSync")
        .mockImplementation(() => {
          throw new Error("Permission denied");
        });

      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      let thrownError = null;
      try {
        // Execute the script - it should throw but we catch it
        await eval(`(async () => { ${sanitizeScript} })()`);
      } catch (error) {
        thrownError = error;
      }

      expect(thrownError).toBeTruthy();
      expect(thrownError.message).toContain("Permission denied");

      // Restore spies
      readFileSyncSpy.mockRestore();
      consoleSpy.mockRestore();

      // Clean up
      if (fs.existsSync(testFile)) {
        fs.unlinkSync(testFile);
      }
    });

    it("should handle binary file content", async () => {
      const binaryData = Buffer.from([0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd]);
      const testFile = "/tmp/test-binary.txt";
      fs.writeFileSync(testFile, binaryData);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      // Should handle binary data gracefully
      const outputCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "output"
      );
      expect(outputCall).toBeDefined();

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should handle content with only whitespace", async () => {
      const whitespaceContent = "   \n\n\t\t  \r\n  ";
      const testFile = "/tmp/test-whitespace.txt";
      fs.writeFileSync(testFile, whitespaceContent);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith("Output file is empty");
      expect(mockCore.setOutput).toHaveBeenCalledWith("output", "");

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should handle very large files with mixed content", async () => {
      // Create content that will trigger both length and line truncation
      const lineContent =
        'This is a line with @user and https://evil.com plus <script>alert("xss")</script>\n';
      const repeatedContent = lineContent.repeat(70000); // Will exceed line limit

      const testFile = "/tmp/test-large-mixed.txt";
      fs.writeFileSync(testFile, repeatedContent);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      const outputCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "output"
      );
      expect(outputCall).toBeDefined();
      const result = outputCall[1];

      // Should be truncated (could be due to line count or length limit)
      expect(result).toMatch(
        /\[Content truncated due to (line count|length)\]/
      );

      // But should still sanitize what it processes
      expect(result).toContain("`@user`");
      expect(result).toContain("(redacted)"); // evil.com
      expect(result).toContain("&lt;script&gt;"); // XML escaping

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });

    it("should preserve log message format for short content", async () => {
      const shortContent = "Short message with @user";
      const testFile = "/tmp/test-short.txt";
      fs.writeFileSync(testFile, shortContent);
      process.env.GITHUB_AW_SAFE_OUTPUTS = testFile;

      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      // Execute the script
      await eval(`(async () => { ${sanitizeScript} })()`);

      const logCalls = consoleSpy.mock.calls;
      const outputLogCall = logCalls.find(
        call =>
          call[0] && call[0].includes("Collected agentic output (sanitized):")
      );

      expect(outputLogCall).toBeDefined();
      // Should not have ... for short content
      expect(outputLogCall[1]).not.toContain("...");
      expect(outputLogCall[1]).toContain("`@user`");

      consoleSpy.mockRestore();
      fs.unlinkSync(testFile);
    });
  });
});
