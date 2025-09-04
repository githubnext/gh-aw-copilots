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
    const parsedLog = parseGenAIScriptLog(content);

    if (parsedLog) {
      core.summary.addRaw(parsedLog).write();
      console.log("GenAIScript log parsed successfully");
    } else {
      console.log("Failed to parse GenAIScript log");
    }
  } catch (error) {
    core.setFailed(error.message);
  }
}

function parseGenAIScriptLog(logContent) {
  try {
    let markdown = "## GenAIScript Execution Log\n\n";
    const lines = logContent.split("\n");
    
    let hasContent = false;
    let inCodeBlock = false;
    let tokenCount = 0;
    let model = "";
    let errors = [];
    let warnings = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      
      if (!line) continue;

      // Extract model information
      const modelMatch = line.match(/model[:\s]+([^\s,]+)/i);
      if (modelMatch) {
        model = modelMatch[1];
      }

      // Extract token usage
      const tokenMatches = [
        line.match(/total[_\s]tokens[:\s]+(\d+)/i),
        line.match(/tokens[:\s]+(\d+)/i),
        line.match(/completion[_\s]tokens[:\s]+(\d+)/i)
      ];
      
      for (const match of tokenMatches) {
        if (match) {
          const tokens = parseInt(match[1]);
          if (tokens > tokenCount) {
            tokenCount = tokens;
          }
          break;
        }
      }

      // Collect errors and warnings
      const lowerLine = line.toLowerCase();
      if (lowerLine.includes("error")) {
        errors.push(line);
      }
      if (lowerLine.includes("warning")) {
        warnings.push(line);
      }

      // Check for code blocks or output sections
      if (line.startsWith("```") || line.includes("genai.md")) {
        hasContent = true;
        inCodeBlock = !inCodeBlock;
      }

      // Add interesting lines to the markdown
      if (line.includes("Running script") || 
          line.includes("Completed") ||
          line.includes("Output:") ||
          inCodeBlock ||
          errors.length > 0 ||
          warnings.length > 0) {
        markdown += `${line}\n`;
        hasContent = true;
      }
    }

    // Add summary information
    if (model) {
      markdown += `\n**Model:** ${model}\n`;
    }
    
    if (tokenCount > 0) {
      markdown += `**Tokens Used:** ${tokenCount.toLocaleString()}\n`;
    }

    if (errors.length > 0) {
      markdown += `\n### Errors (${errors.length})\n`;
      errors.forEach(error => {
        markdown += `- ${error}\n`;
      });
    }

    if (warnings.length > 0) {
      markdown += `\n### Warnings (${warnings.length})\n`;
      warnings.forEach(warning => {
        markdown += `- ${warning}\n`;
      });
    }

    // Check if output files were generated
    if (typeof require !== "undefined") {
      // Only check filesystem if we're in Node.js environment
      const outputDir = "/tmp/genaiscript-output";
      try {
        const fs = require("fs");
        if (fs.existsSync(outputDir)) {
          const outputFiles = fs.readdirSync(outputDir);
          if (outputFiles.length > 0) {
            markdown += `\n### Generated Files\n`;
            outputFiles.forEach(file => {
              markdown += `- ${file}\n`;
            });
          }
        }
      } catch (fsError) {
        // Ignore filesystem errors in test environment
      }
    }

    return hasContent ? markdown : null;
  } catch (error) {
    console.error("Error parsing GenAIScript log:", error);
    return `## GenAIScript Log Parse Error\n\n${error.message}\n\n\`\`\`\n${logContent}\n\`\`\``;
  }
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { main, parseGenAIScriptLog };
} else if (typeof main === "function") {
  main();
}