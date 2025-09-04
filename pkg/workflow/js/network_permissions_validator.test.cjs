import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import fs from "fs";
import path from "path";

// Import the functions from the network permissions validator
const {
  extractDomain,
  isDomainAllowed,
  validateNetworkAccess
} = require("./network_permissions_validator.cjs");

describe("network_permissions_validator.cjs", () => {
  describe("extractDomain function", () => {
    it("should return null for empty or null input", () => {
      expect(extractDomain(null)).toBe(null);
      expect(extractDomain(undefined)).toBe(null);
      expect(extractDomain("")).toBe(null);
    });

    it("should extract domain from HTTP URLs", () => {
      expect(extractDomain("http://example.com/path")).toBe("example.com");
      expect(extractDomain("http://subdomain.example.com")).toBe("subdomain.example.com");
      expect(extractDomain("http://example.com:8080/path")).toBe("example.com");
    });

    it("should extract domain from HTTPS URLs", () => {
      expect(extractDomain("https://github.com/user/repo")).toBe("github.com");
      expect(extractDomain("https://api.github.com/repos")).toBe("api.github.com");
      expect(extractDomain("https://www.example.com/")).toBe("www.example.com");
    });

    it("should extract domain from site: search queries", () => {
      expect(extractDomain("site:github.com")).toBe("github.com");
      expect(extractDomain("search query site:stackoverflow.com more text")).toBe("stackoverflow.com");
      expect(extractDomain("site:api.example.org")).toBe("api.example.org");
    });

    it("should handle malformed URLs gracefully", () => {
      expect(extractDomain("https://")).toBe(null);
      expect(extractDomain("http://")).toBe(null);
      expect(extractDomain("not-a-url")).toBe(null);
    });

    it("should return lowercase domains", () => {
      expect(extractDomain("https://GITHUB.COM/user")).toBe("github.com");
      expect(extractDomain("site:EXAMPLE.COM")).toBe("example.com");
    });
  });

  describe("isDomainAllowed function", () => {
    it("should handle null domain with empty allowed domains (deny-all)", () => {
      expect(isDomainAllowed(null, [], 'WebSearch')).toBe(false);
    });

    it("should handle null domain with non-empty allowed domains", () => {
      expect(isDomainAllowed(null, ["example.com"], 'WebSearch')).toBe(true);
    });

    it("should always block WebFetch with null domain", () => {
      expect(isDomainAllowed(null, ["example.com"], 'WebFetch')).toBe(false);
      expect(isDomainAllowed(null, [], 'WebFetch')).toBe(false);
    });

    it("should deny all domains when allowed list is empty", () => {
      expect(isDomainAllowed("github.com", [])).toBe(false);
      expect(isDomainAllowed("example.com", [])).toBe(false);
    });

    it("should allow exact domain matches", () => {
      const allowedDomains = ["github.com", "example.com"];
      expect(isDomainAllowed("github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("example.com", allowedDomains)).toBe(true);
    });

    it("should deny non-matching domains", () => {
      const allowedDomains = ["github.com", "example.com"];
      expect(isDomainAllowed("malicious.com", allowedDomains)).toBe(false);
      expect(isDomainAllowed("evil.org", allowedDomains)).toBe(false);
    });

    it("should handle wildcard patterns", () => {
      const allowedDomains = ["*.github.com", "*.example.org"];
      expect(isDomainAllowed("api.github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("raw.github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("sub.example.org", allowedDomains)).toBe(true);
    });

    it("should not allow parent domains with wildcard patterns", () => {
      const allowedDomains = ["*.github.com"];
      expect(isDomainAllowed("github.com", allowedDomains)).toBe(false);
    });

    it("should handle complex wildcard patterns", () => {
      const allowedDomains = ["*.githubusercontent.com", "github.*"];
      expect(isDomainAllowed("raw.githubusercontent.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("github.io", allowedDomains)).toBe(true);
    });

    it("should handle mixed exact and wildcard patterns", () => {
      const allowedDomains = ["github.com", "*.example.com", "trusted.org"];
      expect(isDomainAllowed("github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("api.example.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("trusted.org", allowedDomains)).toBe(true);
      expect(isDomainAllowed("evil.com", allowedDomains)).toBe(false);
    });
  });

  describe("validateNetworkAccess function", () => {
    it("should allow non-network tools", () => {
      const data = { tool_name: "WriteFile", tool_input: {} };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(true);
      expect(result.messages).toEqual([]);
    });

    it("should allow tools with empty tool_name", () => {
      const data = { tool_name: "", tool_input: {} };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(true);
      expect(result.messages).toEqual([]);
    });

    it("should validate WebFetch requests", () => {
      const data = {
        tool_name: "WebFetch",
        tool_input: { url: "https://github.com/user/repo" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(true);
      expect(result.messages).toEqual([]);
    });

    it("should block WebFetch requests to unauthorized domains", () => {
      const data = {
        tool_name: "WebFetch",
        tool_input: { url: "https://malicious.com/data" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked for domain: malicious.com");
      expect(result.messages).toContain("Allowed domains: github.com");
    });

    it("should validate WebSearch requests with site: queries", () => {
      const data = {
        tool_name: "WebSearch",
        tool_input: { query: "programming site:stackoverflow.com" }
      };
      const result = validateNetworkAccess(data, ["stackoverflow.com"]);
      expect(result.allowed).toBe(true);
      expect(result.messages).toEqual([]);
    });

    it("should block WebSearch requests with unauthorized site: queries", () => {
      const data = {
        tool_name: "WebSearch",
        tool_input: { query: "search site:malicious.com" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked for domain: malicious.com");
    });

    it("should block general WebSearch requests when domain allowlist is configured", () => {
      const data = {
        tool_name: "WebSearch",
        tool_input: { query: "general search without site" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked for WebSearch: no specific domain detected");
      expect(result.messages).toContain("Allowed domains: github.com");
    });

    it("should apply deny-all policy for WebSearch with empty allowed domains", () => {
      const data = {
        tool_name: "WebSearch",
        tool_input: { query: "any search" }
      };
      const result = validateNetworkAccess(data, []);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked: deny-all policy in effect");
      expect(result.messages).toContain("No domains are allowed for WebSearch");
    });

    it("should handle missing tool_input", () => {
      const data = { tool_name: "WebFetch" };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked for domain: null");
    });

    it("should handle missing url and query in tool_input", () => {
      const data = {
        tool_name: "WebFetch",
        tool_input: { other_field: "value" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
      expect(result.messages).toContain("Network access blocked for domain: null");
    });

    it("should work with wildcard domain patterns", () => {
      const data = {
        tool_name: "WebFetch",
        tool_input: { url: "https://api.github.com/repos" }
      };
      const result = validateNetworkAccess(data, ["*.github.com"]);
      expect(result.allowed).toBe(true);
      expect(result.messages).toEqual([]);
    });
  });

  describe("edge cases and error handling", () => {
    it("should handle malformed input data gracefully", () => {
      const data = null;
      expect(() => validateNetworkAccess(data, ["github.com"])).not.toThrow();
    });

    it("should handle undefined tool_input", () => {
      const data = { tool_name: "WebFetch", tool_input: undefined };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(false);
    });

    it("should handle special characters in domains", () => {
      const allowedDomains = ["test.example-site.com", "under_score.org"];
      expect(isDomainAllowed("test.example-site.com", allowedDomains)).toBe(true);
      // Note: underscores in domains are technically invalid but we test the regex handling
    });

    it("should be case insensitive for domain matching", () => {
      const data = {
        tool_name: "WebFetch",
        tool_input: { url: "https://GITHUB.COM/user" }
      };
      const result = validateNetworkAccess(data, ["github.com"]);
      expect(result.allowed).toBe(true);
    });

    it("should handle empty strings in allowed domains", () => {
      const allowedDomains = ["", "github.com", ""];
      expect(isDomainAllowed("github.com", allowedDomains)).toBe(true);
      expect(isDomainAllowed("other.com", allowedDomains)).toBe(false);
    });
  });

  describe("integration scenarios", () => {
    it("should handle real-world GitHub API scenarios", () => {
      const allowedDomains = ["api.github.com", "*.githubusercontent.com", "github.com"];
      
      const scenarios = [
        {
          data: { tool_name: "WebFetch", tool_input: { url: "https://api.github.com/repos/owner/repo" }},
          expected: true
        },
        {
          data: { tool_name: "WebFetch", tool_input: { url: "https://raw.githubusercontent.com/owner/repo/main/README.md" }},
          expected: true
        },
        {
          data: { tool_name: "WebSearch", tool_input: { query: "github repository site:github.com" }},
          expected: true
        },
        {
          data: { tool_name: "WebFetch", tool_input: { url: "https://evil.com/steal-data" }},
          expected: false
        }
      ];

      for (const scenario of scenarios) {
        const result = validateNetworkAccess(scenario.data, allowedDomains);
        expect(result.allowed).toBe(scenario.expected);
      }
    });

    it("should handle deny-all policy correctly", () => {
      const denyAllDomains = [];
      
      const scenarios = [
        { tool_name: "WriteFile", tool_input: {} }, // Should be allowed (not network tool)
        { tool_name: "WebFetch", tool_input: { url: "https://github.com" }}, // Should be blocked
        { tool_name: "WebSearch", tool_input: { query: "any search" }} // Should be blocked
      ];

      const [allowedResult, blockedFetch, blockedSearch] = scenarios.map(data => 
        validateNetworkAccess(data, denyAllDomains)
      );

      expect(allowedResult.allowed).toBe(true);
      expect(blockedFetch.allowed).toBe(false);
      expect(blockedSearch.allowed).toBe(false);
      expect(blockedSearch.messages).toContain("Network access blocked: deny-all policy in effect");
    });
  });
});