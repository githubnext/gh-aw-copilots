// Type definitions for GitHub Actions github-script action
// These globals are provided by the github-script action environment
// Based on @actions/github-script AsyncFunctionArguments interface

import * as actionsCore from '@actions/core';
import * as actionsExec from '@actions/exec';
import * as actionsGithub from '@actions/github';
import * as actionsGlob from '@actions/glob';
import * as actionsIo from '@actions/io';
import { Context } from '@actions/github/lib/context';
import { GitHub } from '@actions/github/lib/utils';

declare global {
  /**
   * GitHub API client instance provided by github-script action
   * This is an authenticated Octokit instance with pagination plugins
   */
  const github: InstanceType<typeof GitHub>;
  
  /**
   * Alternative name for the github client (same as github)
   * Provided for backward compatibility
   */
  const octokit: InstanceType<typeof GitHub>;
  
  /**
   * GitHub Actions context object provided by github-script action
   * Contains information about the workflow run context
   */
  const context: Context;
  
  /**
   * Actions core utilities provided by github-script action
   * For setting outputs, logging, and other workflow operations
   */
  const core: typeof actionsCore;
  
  /**
   * Actions exec utilities provided by github-script action
   * For executing shell commands and tools
   */
  const exec: typeof actionsExec;
  
  /**
   * Actions glob utilities provided by github-script action
   * For file pattern matching and globbing
   */
  const glob: typeof actionsGlob;
  
  /**
   * Actions io utilities provided by github-script action
   * For file and directory operations
   */
  const io: typeof actionsIo;
  
  /**
   * Console object for logging (available in Node.js environment)
   */
  const console: Console;
  
  /**
   * Process object for environment variables and utilities
   */
  const process: NodeJS.Process;
  
  /**
   * Enhanced require function for CommonJS modules
   * This is a proxy wrapper around the normal Node.js require
   * that enables requiring relative paths and npm packages
   */
  const require: NodeRequire;
  
  /**
   * Original require function without the github-script wrapper
   * Use this if you need the non-wrapped require functionality
   */
  const __original_require__: NodeRequire;
  
  /**
   * Global exports object for CommonJS modules
   */
  var exports: any;
  
  /**
   * Global module object for CommonJS modules
   */
  var module: NodeJS.Module;
}

export {};
