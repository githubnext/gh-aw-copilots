#!/usr/bin/env node

/**
 * MCP Server for Safe Outputs
 * This server exposes MCP tools for each configured safe output type
 * and writes outputs to the safe_output.jsonl file
 */

import { promises as fs } from 'fs';
import { join } from 'path';

// MCP Server Protocol types
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

interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, any>;
    required?: string[];
  };
}

interface SafeOutputEntry {
  type: string;
  data: any;
  timestamp: string;
}

class SafeOutputsMCPServer {
  private outputFile: string;
  private availableTools: MCPTool[] = [];
  private safeOutputsConfig: any;

  constructor() {
    // Get the output file path from environment variable
    this.outputFile = process.env.GITHUB_AW_SAFE_OUTPUTS || '/tmp/safe_output.jsonl';
    
    // Parse safe outputs configuration from environment
    const configEnv = process.env.MCP_SAFE_OUTPUTS_CONFIG;
    if (configEnv) {
      try {
        this.safeOutputsConfig = JSON.parse(configEnv);
        this.initializeTools();
      } catch (error) {
        console.error('Failed to parse safe outputs config:', error);
        this.safeOutputsConfig = {};
      }
    } else {
      this.safeOutputsConfig = {};
    }
  }

  private initializeTools() {
    const config = this.safeOutputsConfig;

    // Create-issue tool
    if (config['create-issue']) {
      this.availableTools.push({
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
      });
    }

    // Add-issue-comment tool
    if (config['add-issue-comment']) {
      this.availableTools.push({
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
      });
    }

    // Create-discussion tool
    if (config['create-discussion']) {
      this.availableTools.push({
        name: 'create-discussion',
        description: 'Create a new GitHub discussion',
        inputSchema: {
          type: 'object',
          properties: {
            title: { type: 'string', description: 'Discussion title' },
            body: { type: 'string', description: 'Discussion body' },
            category_id: { type: 'string', description: 'Discussion category ID (optional)' }
          },
          required: ['title', 'body']
        }
      });
    }

    // Create-pull-request tool
    if (config['create-pull-request']) {
      this.availableTools.push({
        name: 'create-pull-request',
        description: 'Create a new pull request',
        inputSchema: {
          type: 'object',
          properties: {
            title: { type: 'string', description: 'PR title' },
            body: { type: 'string', description: 'PR description' },
            branch: { type: 'string', description: 'Source branch name' },
            base: { type: 'string', description: 'Base branch (optional, defaults to main)' }
          },
          required: ['title', 'body', 'branch']
        }
      });
    }

    // Add-issue-label tool
    if (config['add-issue-label']) {
      this.availableTools.push({
        name: 'add-issue-label',
        description: 'Add labels to an issue or pull request',
        inputSchema: {
          type: 'object',
          properties: {
            labels: { 
              type: 'array', 
              items: { type: 'string' },
              description: 'Labels to add'
            },
            issue_number: { type: 'number', description: 'Issue/PR number (optional, defaults to triggering issue)' }
          },
          required: ['labels']
        }
      });
    }

    // Update-issue tool
    if (config['update-issue']) {
      this.availableTools.push({
        name: 'update-issue',
        description: 'Update an existing issue',
        inputSchema: {
          type: 'object',
          properties: {
            title: { type: 'string', description: 'New issue title' },
            body: { type: 'string', description: 'New issue body' },
            state: { type: 'string', enum: ['open', 'closed'], description: 'Issue state' },
            issue_number: { type: 'number', description: 'Issue number (optional, defaults to triggering issue)' }
          }
        }
      });
    }

    // Push-to-branch tool
    if (config['push-to-branch']) {
      this.availableTools.push({
        name: 'push-to-branch',
        description: 'Push changes to a branch',
        inputSchema: {
          type: 'object',
          properties: {
            branch: { type: 'string', description: 'Target branch name' },
            files: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  path: { type: 'string', description: 'File path' },
                  content: { type: 'string', description: 'File content' }
                },
                required: ['path', 'content']
              },
              description: 'Files to update'
            },
            commit_message: { type: 'string', description: 'Commit message' }
          },
          required: ['branch', 'files', 'commit_message']
        }
      });
    }

    // Create-security-report tool
    if (config['create-security-report']) {
      this.availableTools.push({
        name: 'create-security-report',
        description: 'Create a security report (SARIF)',
        inputSchema: {
          type: 'object',
          properties: {
            title: { type: 'string', description: 'Report title' },
            description: { type: 'string', description: 'Report description' },
            findings: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  file: { type: 'string', description: 'File path' },
                  line: { type: 'number', description: 'Line number' },
                  severity: { type: 'string', enum: ['error', 'warning', 'note'], description: 'Finding severity' },
                  message: { type: 'string', description: 'Finding message' }
                },
                required: ['file', 'line', 'severity', 'message']
              },
              description: 'Security findings'
            }
          },
          required: ['title', 'findings']
        }
      });
    }

    // Missing-tool tool
    if (config['missing-tool']) {
      this.availableTools.push({
        name: 'missing-tool',
        description: 'Report missing tools or functionality',
        inputSchema: {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Name of the missing tool' },
            description: { type: 'string', description: 'Description of what the tool should do' },
            reason: { type: 'string', description: 'Why this tool is needed' }
          },
          required: ['name', 'description']
        }
      });
    }
  }

  private async writeOutput(type: string, data: any): Promise<void> {
    const entry: SafeOutputEntry = {
      type,
      data,
      timestamp: new Date().toISOString()
    };

    try {
      // Ensure output directory exists
      const outputDir = join(this.outputFile, '..');
      await fs.mkdir(outputDir, { recursive: true });
      
      // Append to JSONL file
      await fs.appendFile(this.outputFile, JSON.stringify(entry) + '\n');
    } catch (error) {
      console.error('Failed to write output:', error);
      throw error;
    }
  }

  private async handleToolCall(toolName: string, args: any): Promise<any> {
    try {
      await this.writeOutput(toolName, args);
      return { success: true, message: `${toolName} output written successfully` };
    } catch (error) {
      throw new Error(`Failed to execute ${toolName}: ${error}`);
    }
  }

  private createResponse(id: string | number | null, result?: any, error?: any): MCPResponse {
    const response: MCPResponse = { jsonrpc: '2.0', id };
    if (error) {
      response.error = {
        code: -32000,
        message: error.message || String(error),
        data: error
      };
    } else {
      response.result = result;
    }
    return response;
  }

  private async handleRequest(request: MCPRequest): Promise<MCPResponse> {
    const { method, params, id } = request;

    try {
      switch (method) {
        case 'initialize':
          return this.createResponse(id, {
            protocolVersion: '2024-11-05',
            capabilities: {
              tools: {}
            },
            serverInfo: {
              name: 'safe-outputs-mcp-server',
              version: '1.0.0'
            }
          });

        case 'tools/list':
          return this.createResponse(id, { tools: this.availableTools });

        case 'tools/call':
          const { name: toolName, arguments: toolArgs } = params;
          if (!this.availableTools.find(tool => tool.name === toolName)) {
            throw new Error(`Tool ${toolName} not found`);
          }
          const result = await this.handleToolCall(toolName, toolArgs);
          return this.createResponse(id, { content: [{ type: 'text', text: JSON.stringify(result) }] });

        default:
          throw new Error(`Unknown method: ${method}`);
      }
    } catch (error) {
      return this.createResponse(id, undefined, error);
    }
  }

  async start() {
    console.error('Starting Safe Outputs MCP Server...');
    console.error(`Output file: ${this.outputFile}`);
    console.error(`Available tools: ${this.availableTools.map(t => t.name).join(', ')}`);

    // Set up stdin/stdout communication
    process.stdin.setEncoding('utf8');

    for await (const chunk of process.stdin) {
      const lines = chunk.toString().split('\n').filter((line: string) => line.trim());
      
      for (const line of lines) {
        try {
          const request = JSON.parse(line) as MCPRequest;
          const response = await this.handleRequest(request);
          console.log(JSON.stringify(response));
        } catch (error) {
          console.log(JSON.stringify({
            jsonrpc: '2.0',
            id: null,
            error: {
              code: -32700,
              message: 'Parse error',
              data: error
            }
          }));
        }
      }
    }
  }
}

// Start the server
const server = new SafeOutputsMCPServer();
server.start().catch(error => {
  console.error('Server error:', error);
  process.exit(1);
});