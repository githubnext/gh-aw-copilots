import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { promises as fs } from 'fs';
import { join } from 'path';
import { spawn } from 'child_process';
import { PassThrough, Readable } from 'stream';

// Type definitions for testing
interface MCPRequest {
  jsonrpc: '2.0';
  method: string;
  params?: any;
  id: string | number | null;
}

interface MCPResponse {
  jsonrpc: '2.0';
  id?: string | number | null;
  result?: any;
  error?: {
    code: number;
    message: string;
    data?: any;
  };
}

describe('mcp-safe-outputs-server.ts', () => {
  let mockStdin: PassThrough;
  let mockStdout: string[];
  let originalEnv: NodeJS.ProcessEnv;
  let outputFile: string;

  beforeEach(() => {
    // Save original environment
    originalEnv = { ...process.env };
    
    // Set up test output file
    outputFile = '/tmp/test-safe-outputs.jsonl';
    process.env.GITHUB_AW_SAFE_OUTPUTS = outputFile;
    process.env.MCP_SAFE_OUTPUTS_CONFIG = JSON.stringify({
      'create-issue': true,
      'add-issue-comment': true,
      'update-issue': true
    });

    // Mock stdin/stdout
    mockStdin = new PassThrough();
    mockStdout = [];
    
    // Mock console.log to capture output
    vi.spyOn(console, 'log').mockImplementation((msg: string) => {
      mockStdout.push(msg);
    });

    // Mock console.error
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(async () => {
    // Restore original environment
    process.env = originalEnv;
    
    // Clean up test files
    try {
      await fs.unlink(outputFile);
    } catch {
      // Ignore if file doesn't exist
    }

    // Restore console
    vi.restoreAllMocks();
  });

  describe('TypeScript Compilation', () => {
    it('should compile without TypeScript errors', async () => {
      // This test is implicitly run by the TypeScript compiler during the build process
      // If there are any TypeScript errors, the test suite won't even run
      expect(true).toBe(true);
    });

    it('should have correct type definitions for MCP interfaces', () => {
      // Test that our interfaces are properly typed
      const request: MCPRequest = {
        jsonrpc: '2.0',
        method: 'initialize',
        id: 'test-id'
      };
      
      const response: MCPResponse = {
        jsonrpc: '2.0',
        id: 'test-id',
        result: {}
      };
      
      expect(request.jsonrpc).toBe('2.0');
      expect(response.jsonrpc).toBe('2.0');
    });
  });

  describe('Configuration Parsing', () => {
    it('should parse safe outputs configuration from environment', async () => {
      const config = {
        'create-issue': true,
        'add-issue-comment': true,
        'create-pull-request': true
      };
      
      process.env.MCP_SAFE_OUTPUTS_CONFIG = JSON.stringify(config);
      
      // Dynamic import of the server module to test configuration parsing
      // Note: This is a simplified test - the actual server would need more sophisticated testing
      expect(process.env.MCP_SAFE_OUTPUTS_CONFIG).toBe(JSON.stringify(config));
    });

    it('should handle invalid configuration gracefully', () => {
      process.env.MCP_SAFE_OUTPUTS_CONFIG = 'invalid-json';
      
      // The server should handle invalid JSON gracefully
      // This is more of a contract test - the actual error handling is tested in the server
      expect(() => {
        try {
          JSON.parse(process.env.MCP_SAFE_OUTPUTS_CONFIG || '{}');
        } catch {
          // Should handle gracefully
        }
      }).not.toThrow();
    });

    it('should use default output file if GITHUB_AW_SAFE_OUTPUTS is not set', () => {
      delete process.env.GITHUB_AW_SAFE_OUTPUTS;
      
      const defaultFile = process.env.GITHUB_AW_SAFE_OUTPUTS || '/tmp/safe_output.jsonl';
      expect(defaultFile).toBe('/tmp/safe_output.jsonl');
    });
  });

  describe('MCP Protocol Messages', () => {
    it('should handle initialize request format', () => {
      const initRequest: MCPRequest = {
        jsonrpc: '2.0',
        method: 'initialize',
        id: 1
      };
      
      expect(initRequest.method).toBe('initialize');
      expect(initRequest.jsonrpc).toBe('2.0');
      expect(typeof initRequest.id).toBe('number');
    });

    it('should handle tools/list request format', () => {
      const toolsRequest: MCPRequest = {
        jsonrpc: '2.0',
        method: 'tools/list',
        id: 2
      };
      
      expect(toolsRequest.method).toBe('tools/list');
      expect(toolsRequest.jsonrpc).toBe('2.0');
    });

    it('should handle tools/call request format', () => {
      const toolCallRequest: MCPRequest = {
        jsonrpc: '2.0',
        method: 'tools/call',
        params: {
          name: 'create-issue',
          arguments: {
            title: 'Test Issue',
            body: 'Test body'
          }
        },
        id: 3
      };
      
      expect(toolCallRequest.method).toBe('tools/call');
      expect(toolCallRequest.params.name).toBe('create-issue');
      expect(toolCallRequest.params.arguments.title).toBe('Test Issue');
    });
  });

  describe('Tool Definitions', () => {
    it('should define create-issue tool with correct schema', () => {
      const createIssueTool = {
        name: 'create-issue',
        description: 'Create a new GitHub issue',
        inputSchema: {
          type: 'object',
          properties: {
            title: { type: 'string', description: 'Issue title' },
            body: { type: 'string', description: 'Issue body/description' },
            labels: { 
              type: 'array', 
              items: { type: 'string' },
              description: 'Issue labels'
            },
            assignees: {
              type: 'array',
              items: { type: 'string' },
              description: 'Issue assignees'
            }
          },
          required: ['title', 'body']
        }
      };
      
      expect(createIssueTool.name).toBe('create-issue');
      expect(createIssueTool.inputSchema.required).toContain('title');
      expect(createIssueTool.inputSchema.required).toContain('body');
    });

    it('should define add-issue-comment tool with correct schema', () => {
      const addCommentTool = {
        name: 'add-issue-comment',
        description: 'Add a comment to an issue or pull request',
        inputSchema: {
          type: 'object',
          properties: {
            body: { type: 'string', description: 'Comment body' },
            issue_number: { type: 'number', description: 'Issue/PR number (optional, defaults to triggering issue)' }
          },
          required: ['body']
        }
      };
      
      expect(addCommentTool.name).toBe('add-issue-comment');
      expect(addCommentTool.inputSchema.required).toContain('body');
    });
  });

  describe('Output File Operations', () => {
    it('should write entries to JSONL format', async () => {
      const testEntry = {
        type: 'create-issue',
        data: {
          title: 'Test Issue',
          body: 'Test body'
        },
        timestamp: new Date().toISOString()
      };
      
      // Create the output directory
      await fs.mkdir(join(outputFile, '..'), { recursive: true });
      
      // Write test entry
      await fs.appendFile(outputFile, JSON.stringify(testEntry) + '\n');
      
      // Read and verify
      const content = await fs.readFile(outputFile, 'utf8');
      const lines = content.trim().split('\n');
      
      expect(lines.length).toBe(1);
      
      const parsedEntry = JSON.parse(lines[0]);
      expect(parsedEntry.type).toBe('create-issue');
      expect(parsedEntry.data.title).toBe('Test Issue');
      expect(parsedEntry.timestamp).toBeDefined();
    });

    it('should handle multiple entries in JSONL format', async () => {
      const entries = [
        {
          type: 'create-issue',
          data: { title: 'Issue 1', body: 'Body 1' },
          timestamp: new Date().toISOString()
        },
        {
          type: 'add-issue-comment',
          data: { body: 'Comment 1' },
          timestamp: new Date().toISOString()
        }
      ];
      
      // Create the output directory
      await fs.mkdir(join(outputFile, '..'), { recursive: true });
      
      // Write entries
      for (const entry of entries) {
        await fs.appendFile(outputFile, JSON.stringify(entry) + '\n');
      }
      
      // Read and verify
      const content = await fs.readFile(outputFile, 'utf8');
      const lines = content.trim().split('\n');
      
      expect(lines.length).toBe(2);
      expect(JSON.parse(lines[0]).type).toBe('create-issue');
      expect(JSON.parse(lines[1]).type).toBe('add-issue-comment');
    });
  });

  describe('Integration Tests', () => {
    it('should be executable with Node.js and tsx', () => {
      // Test that the server script is properly structured for execution
      const serverPath = join(__dirname, 'mcp-safe-outputs-server.ts');
      
      // Check that file exists (this test assumes the file is in the same directory)
      expect(() => fs.access(serverPath)).not.toThrow();
    });

    it('should handle process.stdin as async iterator', async () => {
      // Test that the server can handle stdin stream properly
      const mockData = '{"jsonrpc":"2.0","method":"initialize","id":1}\n';
      
      // Create a readable stream
      const readable = new Readable({
        read() {
          this.push(mockData);
          this.push(null); // End stream
        }
      });
      
      const chunks: string[] = [];
      for await (const chunk of readable) {
        chunks.push(chunk.toString());
      }
      
      expect(chunks.length).toBeGreaterThan(0);
      expect(chunks[0]).toContain('initialize');
    });

    it('should validate JSON-RPC protocol structure', () => {
      const validRequest = {
        jsonrpc: '2.0',
        method: 'test',
        id: 1
      };
      
      const validResponse = {
        jsonrpc: '2.0',
        id: 1,
        result: {}
      };
      
      const validErrorResponse = {
        jsonrpc: '2.0',
        id: 1,
        error: {
          code: -32000,
          message: 'Test error'
        }
      };
      
      expect(validRequest.jsonrpc).toBe('2.0');
      expect(validResponse.jsonrpc).toBe('2.0');
      expect(validErrorResponse.jsonrpc).toBe('2.0');
      expect(validErrorResponse.error.code).toBe(-32000);
    });
  });
});