import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import fs from "fs";
import path from "path";

describe("missing_tool.cjs", () => {
  let mockCore;
  let missingToolScript;
  let originalConsole;

  beforeEach(() => {
    // Save original console before mocking
    originalConsole = global.console;

    // Mock console methods
    global.console = {
      log: vi.fn(),
      error: vi.fn(),
    };

    // Mock core actions methods
    mockCore = {
      setOutput: vi.fn(),
      info: vi.fn(),
      warning: vi.fn(),
      error: vi.fn(),
    };
    global.core = mockCore;

    // Mock require
    global.require = vi.fn().mockImplementation(module => {
      if (module === "fs") {
        return fs;
      }
      if (module === "@actions/core") {
        return mockCore;
      }
      throw new Error(`Module not found: ${module}`);
    });

    // Read the script file
    const scriptPath = path.join(__dirname, "missing_tool.cjs");
    missingToolScript = fs.readFileSync(scriptPath, "utf8");
  });

  afterEach(() => {
    // Clean up environment variables
    delete process.env.GITHUB_AW_AGENT_OUTPUT;
    delete process.env.GITHUB_AW_MISSING_TOOL_MAX;

    // Restore original console
    global.console = originalConsole;

    // Clean up globals
    delete global.core;
    delete global.require;
  });

  const runScript = async () => {
    const scriptFunction = new Function(missingToolScript);
    return scriptFunction();
  };

  describe("JSON Array Input Format", () => {
    it("should parse JSON array with missing-tool entries", async () => {
      const testData = [
        {
          type: "missing-tool",
          tool: "docker",
          reason: "Need containerization support",
          alternatives: "Use VM or manual setup",
        },
        {
          type: "missing-tool",
          tool: "kubectl",
          reason: "Kubernetes cluster management required",
        },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "2");
      const toolsReportedCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "tools_reported"
      );
      expect(toolsReportedCall).toBeDefined();

      const reportedTools = JSON.parse(toolsReportedCall[1]);
      expect(reportedTools).toHaveLength(2);
      expect(reportedTools[0].tool).toBe("docker");
      expect(reportedTools[0].reason).toBe("Need containerization support");
      expect(reportedTools[0].alternatives).toBe("Use VM or manual setup");
      expect(reportedTools[1].tool).toBe("kubectl");
      expect(reportedTools[1].reason).toBe(
        "Kubernetes cluster management required"
      );
      expect(reportedTools[1].alternatives).toBe(null);
    });

    it("should filter out non-missing-tool entries", async () => {
      const testData = [
        {
          type: "missing-tool",
          tool: "docker",
          reason: "Need containerization",
        },
        {
          type: "other-type",
          data: "should be ignored",
        },
        {
          type: "missing-tool",
          tool: "kubectl",
          reason: "Need k8s support",
        },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "2");
      const toolsReportedCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "tools_reported"
      );
      const reportedTools = JSON.parse(toolsReportedCall[1]);
      expect(reportedTools).toHaveLength(2);
      expect(reportedTools[0].tool).toBe("docker");
      expect(reportedTools[1].tool).toBe("kubectl");
    });
  });

  describe("Validation", () => {
    it("should skip entries missing tool field", async () => {
      const testData = [
        {
          type: "missing-tool",
          reason: "No tool specified",
        },
        {
          type: "missing-tool",
          tool: "valid-tool",
          reason: "This should work",
        },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "1");
      expect(mockCore.warning).toHaveBeenCalledWith(
        `missing-tool entry missing 'tool' field: ${JSON.stringify(testData[0])}`
      );
    });

    it("should skip entries missing reason field", async () => {
      const testData = [
        {
          type: "missing-tool",
          tool: "some-tool",
        },
        {
          type: "missing-tool",
          tool: "valid-tool",
          reason: "This should work",
        },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "1");
      expect(mockCore.warning).toHaveBeenCalledWith(
        `missing-tool entry missing 'reason' field: ${JSON.stringify(testData[0])}`
      );
    });
  });

  describe("Max Reports Limit", () => {
    it("should respect max reports limit", async () => {
      const testData = [
        { type: "missing-tool", tool: "tool1", reason: "reason1" },
        { type: "missing-tool", tool: "tool2", reason: "reason2" },
        { type: "missing-tool", tool: "tool3", reason: "reason3" },
        { type: "missing-tool", tool: "tool4", reason: "reason4" },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);
      process.env.GITHUB_AW_MISSING_TOOL_MAX = "2";

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "2");
      expect(mockCore.info).toHaveBeenCalledWith(
        "Reached maximum number of missing tool reports (2)"
      );

      const toolsReportedCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "tools_reported"
      );
      const reportedTools = JSON.parse(toolsReportedCall[1]);
      expect(reportedTools).toHaveLength(2);
      expect(reportedTools[0].tool).toBe("tool1");
      expect(reportedTools[1].tool).toBe("tool2");
    });

    it("should work without max limit", async () => {
      const testData = [
        { type: "missing-tool", tool: "tool1", reason: "reason1" },
        { type: "missing-tool", tool: "tool2", reason: "reason2" },
        { type: "missing-tool", tool: "tool3", reason: "reason3" },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);
      // No GITHUB_AW_MISSING_TOOL_MAX set

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "3");
    });
  });

  describe("Edge Cases", () => {
    it("should handle empty agent output", async () => {
      process.env.GITHUB_AW_AGENT_OUTPUT = "";

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "0");
      expect(mockCore.info).toHaveBeenCalledWith("No agent output to process");
    });

    it("should handle missing environment variables", async () => {
      // Don't set any environment variables

      await runScript();

      expect(mockCore.setOutput).toHaveBeenCalledWith("total_count", "0");
    });

    it("should add timestamp to reported tools", async () => {
      const testData = [
        {
          type: "missing-tool",
          tool: "test-tool",
          reason: "testing timestamp",
        },
      ];

      process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(testData);

      const beforeTime = new Date();
      await runScript();
      const afterTime = new Date();

      const toolsReportedCall = mockCore.setOutput.mock.calls.find(
        call => call[0] === "tools_reported"
      );
      const reportedTools = JSON.parse(toolsReportedCall[1]);
      expect(reportedTools).toHaveLength(1);

      const timestamp = new Date(reportedTools[0].timestamp);
      expect(timestamp).toBeInstanceOf(Date);
      expect(timestamp.getTime()).toBeGreaterThanOrEqual(beforeTime.getTime());
      expect(timestamp.getTime()).toBeLessThanOrEqual(afterTime.getTime());
    });
  });
});
