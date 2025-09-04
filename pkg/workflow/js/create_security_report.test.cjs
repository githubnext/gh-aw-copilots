import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import fs from "fs";
import path from "path";

// Mock the GitHub Actions core module
const mockCore = {
  setOutput: vi.fn(),
  summary: {
    addRaw: vi.fn().mockReturnThis(),
    write: vi.fn().mockResolvedValue(undefined),
  },
};

// Mock the context
const mockContext = {
  runId: "12345",
  repo: {
    owner: "test-owner",
    repo: "test-repo",
  },
  payload: {
    repository: {
      html_url: "https://github.com/test-owner/test-repo",
    },
  },
};

// Set up globals
global.core = mockCore;
global.context = mockContext;

// Read the security report script
const securityReportScript = fs.readFileSync(
  path.join(import.meta.dirname, "create_security_report.cjs"),
  "utf8"
);

describe("create_security_report.cjs", () => {
  beforeEach(() => {
    // Reset mocks
    mockCore.setOutput.mockClear();
    mockCore.summary.addRaw.mockClear();
    mockCore.summary.write.mockClear();

    // Set up basic environment
    process.env.GITHUB_AW_AGENT_OUTPUT = "";
    delete process.env.GITHUB_AW_SECURITY_REPORT_MAX;
  });

  afterEach(() => {
    // Clean up any created files
    try {
      const sarifFile = path.join(process.cwd(), "security-report.sarif");
      if (fs.existsSync(sarifFile)) {
        fs.unlinkSync(sarifFile);
      }
    } catch (e) {
      // Ignore cleanup errors
    }
  });

  describe("main function", () => {
    it("should handle missing environment variable", async () => {
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "No GITHUB_AW_AGENT_OUTPUT environment variable found"
      );

      consoleSpy.mockRestore();
    });

    it("should handle empty agent output", async () => {
      process.env.GITHUB_AW_AGENT_OUTPUT = "   ";
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith("Agent output content is empty");

      consoleSpy.mockRestore();
    });

    it("should handle invalid JSON", async () => {
      process.env.GITHUB_AW_AGENT_OUTPUT = "invalid json";
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "Error parsing agent output JSON:",
        expect.any(String)
      );

      consoleSpy.mockRestore();
    });

    it("should handle missing items array", async () => {
      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify({
        status: "success",
      });
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "No valid items found in agent output"
      );

      consoleSpy.mockRestore();
    });

    it("should handle no security report items", async () => {
      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify({
        items: [
          { type: "create-issue", title: "Test Issue" },
        ],
      });
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith(
        "No create-security-report items found in agent output"
      );

      consoleSpy.mockRestore();
    });

    it("should create SARIF file for valid security findings", async () => {
      const securityFindings = {
        items: [
          {
            type: "create-security-report",
            file: "src/app.js",
            line: 42,
            severity: "error",
            message: "SQL injection vulnerability detected",
          },
          {
            type: "create-security-report", 
            file: "src/utils.js",
            line: 15,
            severity: "warning",
            message: "Potential XSS vulnerability",
          },
        ],
      };

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(securityFindings);
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      // Check that SARIF file was created
      const sarifFile = path.join(process.cwd(), "security-report.sarif");
      expect(fs.existsSync(sarifFile)).toBe(true);

      // Check SARIF content
      const sarifContent = JSON.parse(fs.readFileSync(sarifFile, "utf8"));
      expect(sarifContent.version).toBe("2.1.0");
      expect(sarifContent.runs).toHaveLength(1);
      expect(sarifContent.runs[0].results).toHaveLength(2);

      // Check first finding
      const firstResult = sarifContent.runs[0].results[0];
      expect(firstResult.message.text).toBe("SQL injection vulnerability detected");
      expect(firstResult.level).toBe("error");
      expect(firstResult.locations[0].physicalLocation.artifactLocation.uri).toBe("src/app.js");
      expect(firstResult.locations[0].physicalLocation.region.startLine).toBe(42);

      // Check second finding  
      const secondResult = sarifContent.runs[0].results[1];
      expect(secondResult.message.text).toBe("Potential XSS vulnerability");
      expect(secondResult.level).toBe("warning");
      expect(secondResult.locations[0].physicalLocation.artifactLocation.uri).toBe("src/utils.js");
      expect(secondResult.locations[0].physicalLocation.region.startLine).toBe(15);

      // Check outputs were set
      expect(mockCore.setOutput).toHaveBeenCalledWith("sarif_file", sarifFile);
      expect(mockCore.setOutput).toHaveBeenCalledWith("findings_count", 2);

      // Check summary was written
      expect(mockCore.summary.addRaw).toHaveBeenCalled();
      expect(mockCore.summary.write).toHaveBeenCalled();

      consoleSpy.mockRestore();
    });

    it("should respect max findings limit", async () => {
      process.env.GITHUB_AW_SECURITY_REPORT_MAX = "1";

      const securityFindings = {
        items: [
          {
            type: "create-security-report",
            file: "src/app.js", 
            line: 42,
            severity: "error",
            message: "First finding",
          },
          {
            type: "create-security-report",
            file: "src/utils.js",
            line: 15, 
            severity: "warning",
            message: "Second finding",
          },
        ],
      };

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(securityFindings);
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      // Check that SARIF file was created with only 1 finding
      const sarifFile = path.join(process.cwd(), "security-report.sarif");
      expect(fs.existsSync(sarifFile)).toBe(true);

      const sarifContent = JSON.parse(fs.readFileSync(sarifFile, "utf8"));
      expect(sarifContent.runs[0].results).toHaveLength(1);
      expect(sarifContent.runs[0].results[0].message.text).toBe("First finding");

      // Check output reflects the limit
      expect(mockCore.setOutput).toHaveBeenCalledWith("findings_count", 1);

      consoleSpy.mockRestore();
    });

    it("should validate and filter invalid security findings", async () => {
      const mixedFindings = {
        items: [
          {
            type: "create-security-report",
            file: "src/valid.js",
            line: 10,
            severity: "error", 
            message: "Valid finding",
          },
          {
            type: "create-security-report",
            // Missing file
            line: 20,
            severity: "error",
            message: "Invalid - no file",
          },
          {
            type: "create-security-report",
            file: "src/invalid.js",
            // Missing line
            severity: "error",
            message: "Invalid - no line",
          },
          {
            type: "create-security-report",
            file: "src/invalid2.js", 
            line: "not-a-number",
            severity: "error",
            message: "Invalid - bad line",
          },
          {
            type: "create-security-report",
            file: "src/invalid3.js",
            line: 30,
            severity: "invalid-severity",
            message: "Invalid - bad severity",
          },
        ],
      };

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(mixedFindings);
      const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

      await eval(`(async () => { ${securityReportScript} })()`);

      // Check that SARIF file was created with only the 1 valid finding
      const sarifFile = path.join(process.cwd(), "security-report.sarif");
      expect(fs.existsSync(sarifFile)).toBe(true);

      const sarifContent = JSON.parse(fs.readFileSync(sarifFile, "utf8"));
      expect(sarifContent.runs[0].results).toHaveLength(1);
      expect(sarifContent.runs[0].results[0].message.text).toBe("Valid finding");

      // Check outputs
      expect(mockCore.setOutput).toHaveBeenCalledWith("findings_count", 1);

      consoleSpy.mockRestore();
    });
  });
});