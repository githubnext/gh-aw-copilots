async function main() {
  /** @type {typeof import("fs")} */
  const fs = require("fs");
  const { execSync } = require("child_process");

  // Environment validation - fail early if required variables are missing
  const branchName = process.env.GITHUB_AW_PUSH_BRANCH;
  if (!branchName) {
    core.setFailed("GITHUB_AW_PUSH_BRANCH environment variable is required");
    return;
  }

  const outputContent = process.env.GITHUB_AW_AGENT_OUTPUT || "";
  if (outputContent.trim() === "") {
    console.log("Agent output content is empty");
    return;
  }

  const target = process.env.GITHUB_AW_PUSH_TARGET || "triggering";

  // Check if patch file exists and has valid content
  if (!fs.existsSync("/tmp/aw.patch")) {
    core.setFailed("No patch file found - cannot push without changes");
    return;
  }

  const patchContent = fs.readFileSync("/tmp/aw.patch", "utf8");

  // Check for actual error conditions (but allow empty patches as valid noop)
  if (patchContent.includes("Failed to generate patch")) {
    core.setFailed(
      "Patch file contains error message - cannot push without changes"
    );
    return;
  }

  // Empty patch is valid - it means no changes (noop operation)
  const isEmpty = !patchContent || !patchContent.trim();
  if (isEmpty) {
    console.log("Patch file is empty - no changes to apply (noop operation)");
  }

  console.log("Agent output content length:", outputContent.length);
  if (!isEmpty) {
    console.log("Patch content validation passed");
  }
  console.log("Target branch:", branchName);
  console.log("Target configuration:", target);

  // Parse the validated output JSON
  let validatedOutput;
  try {
    validatedOutput = JSON.parse(outputContent);
  } catch (error) {
    console.log(
      "Error parsing agent output JSON:",
      error instanceof Error ? error.message : String(error)
    );
    return;
  }

  if (!validatedOutput.items || !Array.isArray(validatedOutput.items)) {
    console.log("No valid items found in agent output");
    return;
  }

  // Find the push-to-branch item
  const pushItem = validatedOutput.items.find(
    /** @param {any} item */ item => item.type === "push-to-branch"
  );
  if (!pushItem) {
    console.log("No push-to-branch item found in agent output");
    return;
  }

  console.log("Found push-to-branch item");

  // Validate target configuration for pull request context
  if (target !== "*" && target !== "triggering") {
    // If target is a specific number, validate it's a valid pull request number
    const targetNumber = parseInt(target, 10);
    if (isNaN(targetNumber)) {
      core.setFailed(
        'Invalid target configuration: must be "triggering", "*", or a valid pull request number'
      );
      return;
    }
  }

  // Check if we're in a pull request context when required
  if (target === "triggering" && !context.payload.pull_request) {
    core.setFailed(
      'push-to-branch with target "triggering" requires pull request context'
    );
    return;
  }

  // Configure git (required for commits)
  execSync('git config --global user.email "action@github.com"', {
    stdio: "inherit",
  });
  execSync('git config --global user.name "GitHub Action"', {
    stdio: "inherit",
  });

  // Switch to or create the target branch
  console.log("Switching to branch:", branchName);
  try {
    // Try to checkout existing branch first
    execSync("git fetch origin", { stdio: "inherit" });
    execSync(`git checkout ${branchName}`, { stdio: "inherit" });
    console.log("Checked out existing branch:", branchName);
  } catch (error) {
    // Branch doesn't exist, create it
    console.log("Branch does not exist, creating new branch:", branchName);
    execSync(`git checkout -b ${branchName}`, { stdio: "inherit" });
  }

  // Apply the patch using git CLI (skip if empty)
  if (!isEmpty) {
    console.log("Applying patch...");
    try {
      execSync("git apply /tmp/aw.patch", { stdio: "inherit" });
      console.log("Patch applied successfully");
    } catch (error) {
      console.error(
        "Failed to apply patch:",
        error instanceof Error ? error.message : String(error)
      );
      core.setFailed("Failed to apply patch");
      return;
    }
  } else {
    console.log("Skipping patch application (empty patch)");
  }

  // Commit and push the changes
  execSync("git add .", { stdio: "inherit" });

  // Check if there are changes to commit
  let hasChanges = false;
  try {
    execSync("git diff --cached --exit-code", { stdio: "ignore" });
    console.log("No changes to commit - noop operation completed successfully");
    hasChanges = false;
  } catch (error) {
    // Exit code != 0 means there are changes to commit, which is what we want
    hasChanges = true;
  }

  let commitSha;
  if (hasChanges) {
    const commitMessage = pushItem.message || "Apply agent changes";
    execSync(`git commit -m "${commitMessage}"`, { stdio: "inherit" });
    execSync(`git push origin ${branchName}`, { stdio: "inherit" });
    console.log("Changes committed and pushed to branch:", branchName);
    commitSha = execSync("git rev-parse HEAD", { encoding: "utf8" }).trim();
  } else {
    // For noop operations, get the current HEAD commit
    commitSha = execSync("git rev-parse HEAD", { encoding: "utf8" }).trim();
  }

  // Get commit SHA and push URL
  const pushUrl = context.payload.repository
    ? `${context.payload.repository.html_url}/tree/${branchName}`
    : `https://github.com/${context.repo.owner}/${context.repo.repo}/tree/${branchName}`;

  // Set outputs
  core.setOutput("branch_name", branchName);
  core.setOutput("commit_sha", commitSha);
  core.setOutput("push_url", pushUrl);

  // Write summary to GitHub Actions summary
  const summaryTitle = hasChanges
    ? "Push to Branch"
    : "Push to Branch (No Changes)";
  const summaryContent = hasChanges
    ? `
## ${summaryTitle}
- **Branch**: \`${branchName}\`
- **Commit**: [${commitSha.substring(0, 7)}](${pushUrl})
- **URL**: [${pushUrl}](${pushUrl})
`
    : `
## ${summaryTitle}
- **Branch**: \`${branchName}\`
- **Status**: No changes to apply (noop operation)
- **URL**: [${pushUrl}](${pushUrl})
`;

  await core.summary.addRaw(summaryContent).write();
}

await main();
