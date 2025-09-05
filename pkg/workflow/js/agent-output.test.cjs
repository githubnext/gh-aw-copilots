import { describe, test, expect, beforeAll } from "vitest";
import fs from "fs";
import path from "path";
import Ajv from "ajv";

// Test the agent output JSON schema
describe("Agent Output JSON Schema", () => {
  let ajv;
  let schema;

  beforeAll(() => {
    // Load the schema
    const schemaPath = path.join(
      __dirname,
      "../../../schemas/agent-output.json"
    );
    schema = JSON.parse(fs.readFileSync(schemaPath, "utf8"));

    // Initialize AJV validator
    ajv = new Ajv({ strict: false });
  });

  test("should validate a valid agent output with create-issue", () => {
    const validOutput = {
      items: [
        {
          type: "create-issue",
          title: "Test Issue",
          body: "This is a test issue body",
          labels: ["bug", "needs-investigation"],
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });

  test("should validate a valid agent output with multiple types", () => {
    const validOutput = {
      items: [
        {
          type: "create-issue",
          title: "Test Issue",
          body: "Test body",
        },
        {
          type: "add-issue-comment",
          body: "This is a comment",
        },
        {
          type: "missing-tool",
          tool: "git-blame",
          reason: "Need to analyze code authorship",
        },
      ],
      errors: ["Line 4: Invalid JSON - parsing failed"],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });

  test("should reject output missing required fields", () => {
    const invalidOutput = {
      items: [
        {
          type: "create-issue",
          title: "Test Issue",
          // Missing required 'body' field
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(invalidOutput);

    expect(isValid).toBe(false);
    expect(validate.errors).toBeDefined();
  });

  test("should reject output with invalid type", () => {
    const invalidOutput = {
      items: [
        {
          type: "invalid-type",
          data: "some data",
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(invalidOutput);

    expect(isValid).toBe(false);
    expect(validate.errors).toBeDefined();
  });

  test("should validate update-issue with status field", () => {
    const validOutput = {
      items: [
        {
          type: "update-issue",
          status: "closed",
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });

  test("should validate create-security-report with object sarif", () => {
    const validOutput = {
      items: [
        {
          type: "create-security-report",
          sarif: {
            version: "2.1.0",
            runs: [],
          },
          category: "security-audit",
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });

  test("should validate create-security-report with string sarif", () => {
    const validOutput = {
      items: [
        {
          type: "create-security-report",
          sarif: '{"version": "2.1.0", "runs": []}',
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });

  test("should validate create-pull-request-review-comment with line numbers", () => {
    const validOutput = {
      items: [
        {
          type: "create-pull-request-review-comment",
          path: "src/main.js",
          line: 42,
          body: "Consider using const instead of let here.",
          start_line: 40,
          side: "RIGHT",
        },
      ],
      errors: [],
    };

    const validate = ajv.compile(schema);
    const isValid = validate(validOutput);

    if (!isValid) {
      console.log("Validation errors:", validate.errors);
    }

    expect(isValid).toBe(true);
  });
});
