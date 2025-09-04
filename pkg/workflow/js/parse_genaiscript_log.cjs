function main() {
  const fs = require("fs");

  try {
    const logFile = process.env.AGENT_LOG_FILE;
    if (!logFile) {
      console.log("No agent log file specified");
      return;
    }

    if (!fs.existsSync(logFile)) {
      console.log(`Log file not found: ${logFile}`);
      return;
    }

    const content = fs.readFileSync(logFile, "utf8");

    // Send log content directly to summary without parsing
    core.summary
      .addRaw(`## GenAIScript Output\n\n\`\`\`\n${content}\n\`\`\``)
      .write();
    console.log("GenAIScript log sent to summary successfully");
  } catch (error) {
    core.setFailed(error.message);
  }
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { main };
} else if (typeof main === "function") {
  main();
}
