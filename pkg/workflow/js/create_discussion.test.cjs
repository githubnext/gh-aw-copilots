import { describe, it, expect, beforeEach, vi } from 'vitest';
import fs from 'fs';
import path from 'path';

// Mock the global objects that GitHub Actions provides
const mockCore = {
  setOutput: vi.fn(),
  summary: {
    addRaw: vi.fn().mockReturnThis(),
    write: vi.fn()
  }
};

const mockGithub = {
  graphql: vi.fn()
};

const mockContext = {
  runId: 12345,
  repo: {
    owner: 'testowner',
    repo: 'testrepo'
  },
  payload: {
    repository: {
      html_url: 'https://github.com/testowner/testrepo'
    }
  }
};

// Set up global variables
global.core = mockCore;
global.github = mockGithub;
global.context = mockContext;

describe('create_discussion.cjs', () => {
  let createDiscussionScript;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();
    
    // Reset environment variables
    delete process.env.GITHUB_AW_AGENT_OUTPUT;
    delete process.env.GITHUB_AW_DISCUSSION_TITLE_PREFIX;
    delete process.env.GITHUB_AW_DISCUSSION_CATEGORY_ID;
    
    // Read the script content
    const scriptPath = path.join(process.cwd(), 'pkg/workflow/js/create_discussion.cjs');
    createDiscussionScript = fs.readFileSync(scriptPath, 'utf8');
  });

  it('should handle missing GITHUB_AW_AGENT_OUTPUT environment variable', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    expect(consoleSpy).toHaveBeenCalledWith('No GITHUB_AW_AGENT_OUTPUT environment variable found');
    consoleSpy.mockRestore();
  });

  it('should handle empty agent output', async () => {
    process.env.GITHUB_AW_AGENT_OUTPUT = '   ';  // Use spaces instead of empty string
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    expect(consoleSpy).toHaveBeenCalledWith('Agent output content is empty');
    consoleSpy.mockRestore();
  });

  it('should handle invalid JSON in agent output', async () => {
    process.env.GITHUB_AW_AGENT_OUTPUT = 'invalid json';
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    // Check that it logs the content length first, then the error
    expect(consoleSpy).toHaveBeenCalledWith('Agent output content length:', 12);
    expect(consoleSpy).toHaveBeenCalledWith(
      'Error parsing agent output JSON:',
      expect.stringContaining('Unexpected token')
    );
    consoleSpy.mockRestore();
  });

  it('should handle missing create-discussion items', async () => {
    const validOutput = {
      items: [
        { type: 'create-issue', title: 'Test Issue', body: 'Test body' }
      ]
    };
    process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(validOutput);
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    expect(consoleSpy).toHaveBeenCalledWith('No create-discussion items found in agent output');
    consoleSpy.mockRestore();
  });

  it('should create discussions successfully with basic configuration', async () => {
    // Mock the GraphQL responses
    mockGithub.graphql
      .mockResolvedValueOnce({
        // Repository query response
        repository: {
          id: 'R_test123',
          discussionCategories: {
            nodes: [
              { id: 'DIC_test456', name: 'General', slug: 'general' }
            ]
          }
        }
      })
      .mockResolvedValueOnce({
        // Create discussion mutation response
        createDiscussion: {
          discussion: {
            id: 'D_test789',
            number: 1,
            title: 'Test Discussion',
            url: 'https://github.com/testowner/testrepo/discussions/1'
          }
        }
      });

    const validOutput = {
      items: [
        { type: 'create-discussion', title: 'Test Discussion', body: 'Test discussion body' }
      ]
    };
    process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(validOutput);
    
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    // Verify GraphQL calls
    expect(mockGithub.graphql).toHaveBeenCalledTimes(2);
    
    // Verify repository query
    expect(mockGithub.graphql).toHaveBeenNthCalledWith(1, 
      expect.stringContaining('query GetRepository'),
      { owner: 'testowner', name: 'testrepo' }
    );
    
    // Verify create discussion mutation
    expect(mockGithub.graphql).toHaveBeenNthCalledWith(2,
      expect.stringContaining('mutation CreateDiscussion'),
      {
        repositoryId: 'R_test123',
        categoryId: 'DIC_test456',
        title: 'Test Discussion',
        body: expect.stringContaining('Test discussion body')
      }
    );
    
    // Verify outputs were set
    expect(mockCore.setOutput).toHaveBeenCalledWith('discussion_number', 1);
    expect(mockCore.setOutput).toHaveBeenCalledWith('discussion_url', 'https://github.com/testowner/testrepo/discussions/1');
    
    // Verify summary was written
    expect(mockCore.summary.addRaw).toHaveBeenCalledWith(
      expect.stringContaining('## GitHub Discussions')
    );
    expect(mockCore.summary.write).toHaveBeenCalled();
    
    consoleSpy.mockRestore();
  });

  it('should apply title prefix when configured', async () => {
    // Mock the GraphQL responses
    mockGithub.graphql
      .mockResolvedValueOnce({
        repository: {
          id: 'R_test123',
          discussionCategories: {
            nodes: [
              { id: 'DIC_test456', name: 'General', slug: 'general' }
            ]
          }
        }
      })
      .mockResolvedValueOnce({
        createDiscussion: {
          discussion: {
            id: 'D_test789',
            number: 1,
            title: '[ai] Test Discussion',
            url: 'https://github.com/testowner/testrepo/discussions/1'
          }
        }
      });

    const validOutput = {
      items: [
        { type: 'create-discussion', title: 'Test Discussion', body: 'Test discussion body' }
      ]
    };
    process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(validOutput);
    process.env.GITHUB_AW_DISCUSSION_TITLE_PREFIX = '[ai] ';
    
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    // Verify the title was prefixed
    expect(mockGithub.graphql).toHaveBeenNthCalledWith(2,
      expect.stringContaining('mutation CreateDiscussion'),
      expect.objectContaining({
        title: '[ai] Test Discussion'
      })
    );
    
    consoleSpy.mockRestore();
  });

  it('should use specified category ID when configured', async () => {
    // Mock the GraphQL responses
    mockGithub.graphql
      .mockResolvedValueOnce({
        repository: {
          id: 'R_test123',
          discussionCategories: {
            nodes: [
              { id: 'DIC_test456', name: 'General', slug: 'general' },
              { id: 'DIC_custom789', name: 'Custom', slug: 'custom' }
            ]
          }
        }
      })
      .mockResolvedValueOnce({
        createDiscussion: {
          discussion: {
            id: 'D_test789',
            number: 1,
            title: 'Test Discussion',
            url: 'https://github.com/testowner/testrepo/discussions/1'
          }
        }
      });

    const validOutput = {
      items: [
        { type: 'create-discussion', title: 'Test Discussion', body: 'Test discussion body' }
      ]
    };
    process.env.GITHUB_AW_AGENT_OUTPUT = JSON.stringify(validOutput);
    process.env.GITHUB_AW_DISCUSSION_CATEGORY_ID = 'DIC_custom789';
    
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    
    // Execute the script
    await eval(`(async () => { ${createDiscussionScript} })()`);
    
    // Verify the specified category was used
    expect(mockGithub.graphql).toHaveBeenNthCalledWith(2,
      expect.stringContaining('mutation CreateDiscussion'),
      expect.objectContaining({
        categoryId: 'DIC_custom789'
      })
    );
    
    consoleSpy.mockRestore();
  });
});