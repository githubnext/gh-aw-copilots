import { describe, it, expect, beforeEach, vi } from 'vitest';
import fs from 'fs';
import path from 'path';

// Mock the global objects that GitHub Actions provides
const mockCore = {
  setFailed: vi.fn(),
  setOutput: vi.fn(),
  summary: {
    addRaw: vi.fn().mockReturnThis(),
    write: vi.fn()
  }
};

const mockGithub = {
  request: vi.fn()
};

const mockContext = {
  eventName: 'issues',
  repo: {
    owner: 'testowner',
    repo: 'testrepo'
  },
  payload: {
    issue: {
      number: 123
    }
  }
};

// Set up global variables
global.core = mockCore;
global.github = mockGithub;
global.context = mockContext;

describe('add_reaction.cjs', () => {
  let addReactionScript;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();
    
    // Reset environment variables
    delete process.env.GITHUB_AW_REACTION_MODE;
    delete process.env.GITHUB_AW_REACTION;
    delete process.env.GITHUB_AW_REACTION_ID;
    
    // Reset context to default
    global.context = {
      eventName: 'issues',
      repo: {
        owner: 'testowner',
        repo: 'testrepo'
      },
      payload: {
        issue: {
          number: 123
        }
      }
    };

    // Load the script content
    const scriptPath = path.join(process.cwd(), 'pkg/workflow/js/add_reaction.cjs');
    addReactionScript = fs.readFileSync(scriptPath, 'utf8');
  });

  describe('Environment variable validation', () => {
    it('should use default values when environment variables are not set', async () => {
      mockGithub.request.mockResolvedValue({
        data: { id: 123, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Reaction mode:', 'add');
      expect(consoleSpy).toHaveBeenCalledWith('Reaction type:', 'eyes');
      
      consoleSpy.mockRestore();
    });

    it('should fail with invalid reaction type', async () => {
      process.env.GITHUB_AW_REACTION = 'invalid';

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockCore.setFailed).toHaveBeenCalledWith(
        'Invalid reaction type: invalid. Valid reactions are: +1, -1, laugh, confused, heart, hooray, rocket, eyes'
      );
    });

    it('should accept all valid reaction types', async () => {
      const validReactions = ['+1', '-1', 'laugh', 'confused', 'heart', 'hooray', 'rocket', 'eyes'];
      
      for (const reaction of validReactions) {
        vi.clearAllMocks();
        process.env.GITHUB_AW_REACTION = reaction;
        
        mockGithub.request.mockResolvedValue({
          data: { id: 123, content: reaction }
        });

        await eval(`(async () => { ${addReactionScript} })()`);

        expect(mockCore.setFailed).not.toHaveBeenCalled();
        expect(mockCore.setOutput).toHaveBeenCalledWith('reaction-id', '123');
      }
    });
  });

  describe('Event context handling', () => {
    it('should handle issues event', async () => {
      global.context.eventName = 'issues';
      global.context.payload = { issue: { number: 123 } };
      
      mockGithub.request.mockResolvedValue({
        data: { id: 456, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('API endpoint:', '/repos/testowner/testrepo/issues/123/reactions');
      expect(mockGithub.request).toHaveBeenCalledWith('POST /repos/testowner/testrepo/issues/123/reactions', {
        content: 'eyes',
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      
      consoleSpy.mockRestore();
    });

    it('should handle issue_comment event', async () => {
      global.context.eventName = 'issue_comment';
      global.context.payload = { comment: { id: 789 } };
      
      mockGithub.request.mockResolvedValue({
        data: { id: 456, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('API endpoint:', '/repos/testowner/testrepo/issues/comments/789/reactions');
      expect(mockGithub.request).toHaveBeenCalledWith('POST /repos/testowner/testrepo/issues/comments/789/reactions', {
        content: 'eyes',
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      
      consoleSpy.mockRestore();
    });

    it('should handle pull_request event', async () => {
      global.context.eventName = 'pull_request';
      global.context.payload = { pull_request: { number: 456 } };
      
      mockGithub.request.mockResolvedValue({
        data: { id: 789, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('API endpoint:', '/repos/testowner/testrepo/issues/456/reactions');
      expect(mockGithub.request).toHaveBeenCalledWith('POST /repos/testowner/testrepo/issues/456/reactions', {
        content: 'eyes',
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      
      consoleSpy.mockRestore();
    });

    it('should handle pull_request_review_comment event', async () => {
      global.context.eventName = 'pull_request_review_comment';
      global.context.payload = { comment: { id: 321 } };
      
      mockGithub.request.mockResolvedValue({
        data: { id: 654, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('API endpoint:', '/repos/testowner/testrepo/pulls/comments/321/reactions');
      expect(mockGithub.request).toHaveBeenCalledWith('POST /repos/testowner/testrepo/pulls/comments/321/reactions', {
        content: 'eyes',
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      
      consoleSpy.mockRestore();
    });

    it('should fail on unsupported event type', async () => {
      global.context.eventName = 'unsupported';

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockCore.setFailed).toHaveBeenCalledWith('Unsupported event type: unsupported');
    });

    it('should fail when issue number is missing', async () => {
      global.context.eventName = 'issues';
      global.context.payload = {};

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockCore.setFailed).toHaveBeenCalledWith('Issue number not found in event payload');
    });

    it('should fail when comment ID is missing', async () => {
      global.context.eventName = 'issue_comment';
      global.context.payload = {};

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockCore.setFailed).toHaveBeenCalledWith('Comment ID not found in event payload');
    });
  });

  describe('Add reaction functionality', () => {
    it('should successfully add reaction with direct response', async () => {
      process.env.GITHUB_AW_REACTION = 'heart';
      
      mockGithub.request.mockResolvedValue({
        data: { id: 123, content: 'heart' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Successfully added reaction: heart (id: 123)');
      expect(mockCore.setOutput).toHaveBeenCalledWith('reaction-id', '123');
      
      consoleSpy.mockRestore();
    });

    it('should fallback to list when create response has no ID', async () => {
      process.env.GITHUB_AW_REACTION = 'rocket';
      
      // First call (create) returns no ID
      mockGithub.request.mockResolvedValueOnce({
        data: { content: 'rocket' }
      });
      
      // Second call (list) returns reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 456, content: 'rocket', user: { login: 'github-actions[bot]' } }
        ]
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Could not get reaction ID from create response, falling back to list...');
      expect(consoleSpy).toHaveBeenCalledWith('Found existing reaction: rocket (id: 456)');
      expect(mockCore.setOutput).toHaveBeenCalledWith('reaction-id', '456');
      
      consoleSpy.mockRestore();
    });

    it('should fallback to list when create fails', async () => {
      process.env.GITHUB_AW_REACTION = 'hooray';
      
      // First call (create) fails
      mockGithub.request.mockRejectedValueOnce(new Error('Reaction already exists'));
      
      // Second call (list) returns reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 789, content: 'hooray', user: { login: 'github-actions[bot]' } }
        ]
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Create reaction failed, trying to find existing reaction...');
      expect(consoleSpy).toHaveBeenCalledWith('Found existing reaction: hooray (id: 789)');
      expect(mockCore.setOutput).toHaveBeenCalledWith('reaction-id', '789');
      
      consoleSpy.mockRestore();
    });

    it('should warn when reaction is not found in list', async () => {
      process.env.GITHUB_AW_REACTION = 'confused';
      
      // Create fails
      mockGithub.request.mockRejectedValueOnce(new Error('Failed'));
      
      // List returns no matching reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 999, content: 'different', user: { login: 'other-user' } }
        ]
      });

      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Warning: could not determine reaction id; cleanup will list/filter.');
      expect(mockCore.setOutput).toHaveBeenCalledWith('reaction-id', '');
      
      consoleSpy.mockRestore();
    });
  });

  describe('Remove reaction functionality', () => {
    it('should remove reaction by ID when provided', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'remove';
      process.env.GITHUB_AW_REACTION = 'heart';
      process.env.GITHUB_AW_REACTION_ID = '123';
      
      mockGithub.request.mockResolvedValue({});

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockGithub.request).toHaveBeenCalledWith('DELETE /reactions/{reaction_id}', {
        reaction_id: 123,
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      expect(consoleSpy).toHaveBeenCalledWith('Successfully removed reaction by ID: 123');
      
      consoleSpy.mockRestore();
    });

    it('should fallback to list when remove by ID fails', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'remove';
      process.env.GITHUB_AW_REACTION = 'laugh';
      process.env.GITHUB_AW_REACTION_ID = '456';
      
      // First call (delete by ID) fails
      mockGithub.request.mockRejectedValueOnce(new Error('Not found'));
      
      // Second call (list) returns reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 789, content: 'laugh', user: { login: 'github-actions[bot]' } }
        ]
      });
      
      // Third call (delete by ID from list) succeeds
      mockGithub.request.mockResolvedValueOnce({});

      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Failed to remove reaction by ID 456, falling back to list method...');
      expect(consoleLogSpy).toHaveBeenCalledWith('Successfully removed reaction: laugh (id: 789)');
      
      consoleSpy.mockRestore();
      consoleLogSpy.mockRestore();
    });

    it('should remove multiple matching reactions when listing', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'remove';
      process.env.GITHUB_AW_REACTION = '+1';
      
      // List returns multiple matching reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 111, content: '+1', user: { login: 'github-actions[bot]' } },
          { id: 222, content: '+1', user: { login: 'github-actions[bot]' } },
          { id: 333, content: '+1', user: { login: 'other-user' } }
        ]
      });
      
      // Delete calls succeed
      mockGithub.request.mockResolvedValue({});

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockGithub.request).toHaveBeenCalledWith('DELETE /reactions/{reaction_id}', {
        reaction_id: 111,
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      expect(mockGithub.request).toHaveBeenCalledWith('DELETE /reactions/{reaction_id}', {
        reaction_id: 222,
        headers: { 'Accept': 'application/vnd.github+json' }
      });
      expect(consoleSpy).toHaveBeenCalledWith('Successfully removed reaction: +1 (id: 111)');
      expect(consoleSpy).toHaveBeenCalledWith('Successfully removed reaction: +1 (id: 222)');
      
      consoleSpy.mockRestore();
    });

    it('should handle when no matching reactions are found', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'remove';
      process.env.GITHUB_AW_REACTION = 'eyes';
      
      // List returns no matching reactions
      mockGithub.request.mockResolvedValueOnce({
        data: [
          { id: 444, content: 'different', user: { login: 'github-actions[bot]' } }
        ]
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('No matching reactions found to remove for: eyes');
      
      consoleSpy.mockRestore();
    });
  });

  describe('Error handling', () => {
    it('should fail with invalid mode', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'invalid';

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(mockCore.setFailed).toHaveBeenCalledWith("Invalid mode: invalid. Must be 'add' or 'remove'");
    });

    it('should handle API errors gracefully during add', async () => {
      // Mock the GitHub request to fail both on create and list
      mockGithub.request.mockRejectedValue(new Error('API Error'));

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Failed to process reaction:', 'API Error');
      expect(mockCore.setFailed).toHaveBeenCalledWith('Failed to process reaction: API Error');
      
      consoleSpy.mockRestore();
      consoleWarnSpy.mockRestore();
    });

    it('should handle non-Error objects in catch block during add', async () => {
      // Mock the GitHub request to fail both on create and list
      mockGithub.request.mockRejectedValue('String error');

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Failed to process reaction:', 'String error');
      expect(mockCore.setFailed).toHaveBeenCalledWith('Failed to process reaction: String error');
      
      consoleSpy.mockRestore();
      consoleWarnSpy.mockRestore();
    });
  });

  describe('Output and logging', () => {
    it('should log reaction mode and type', async () => {
      process.env.GITHUB_AW_REACTION_MODE = 'add';
      process.env.GITHUB_AW_REACTION = 'rocket';
      
      mockGithub.request.mockResolvedValue({
        data: { id: 123, content: 'rocket' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('Reaction mode:', 'add');
      expect(consoleSpy).toHaveBeenCalledWith('Reaction type:', 'rocket');
      
      consoleSpy.mockRestore();
    });

    it('should log API endpoint', async () => {
      mockGithub.request.mockResolvedValue({
        data: { id: 123, content: 'eyes' }
      });

      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      await eval(`(async () => { ${addReactionScript} })()`);

      expect(consoleSpy).toHaveBeenCalledWith('API endpoint:', '/repos/testowner/testrepo/issues/123/reactions');
      
      consoleSpy.mockRestore();
    });
  });
});