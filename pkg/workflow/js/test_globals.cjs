// Test file to validate that all @actions/github-script globals are properly typed
// This file exercises all the globals provided by the github-script action environment

async function testGlobals() {
  // Test core functionality
  core.info('Testing core functionality');
  core.setOutput('test-output', 'test-value');
  
  // Test github/octokit functionality
  console.log('GitHub API available:', typeof github);
  console.log('Octokit API available:', typeof octokit);
  console.log('Context available:', typeof context);
  console.log('Repository:', context.repo.owner + '/' + context.repo.repo);
  
  // Test exec functionality  
  console.log('Exec available:', typeof exec);
  console.log('Exec.exec available:', typeof exec.exec);
  
  // Test glob functionality
  console.log('Glob available:', typeof glob);
  console.log('Glob.create available:', typeof glob.create);
  
  // Test io functionality
  console.log('IO available:', typeof io);
  console.log('IO.mkdirP available:', typeof io.mkdirP);
  
  // Test require functionality
  console.log('Require available:', typeof require);
  console.log('Original require available:', typeof __original_require__);
  
  // Test Node.js globals
  console.log('Process available:', typeof process);
  console.log('Console available:', typeof console);
  console.log('Module available:', typeof module);
  console.log('Exports available:', typeof exports);
  
  return 'All globals tested successfully';
}

// Note: This is a test file and won't be executed in actual workflows
module.exports = testGlobals;