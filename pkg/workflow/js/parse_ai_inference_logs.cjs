function main() {
  const fs = require("fs");

  // Parse command line arguments
  const args = process.argv.slice(2);
  const logFilePath = args[0];
  const verbose = args.includes("--verbose");

  if (!logFilePath) {
    console.error("Usage: node parse_ai_inference_logs.cjs <log_file_path> [--verbose]");
    process.exit(1);
  }

  if (!fs.existsSync(logFilePath)) {
    console.error(`Log file not found: ${logFilePath}`);
    process.exit(1);
  }

  // Read the log file
  const logContent = fs.readFileSync(logFilePath, 'utf8');
  
  // Initialize metrics
  let totalTokens = 0;
  let estimatedCost = 0.0;
  let errorCount = 0;
  let warningCount = 0;
  
  // Parse log line by line
  const lines = logContent.split('\n');
  
  for (const line of lines) {
    const trimmed = line.trim();
    
    // Skip empty lines
    if (!trimmed) continue;
    
    // Count errors and warnings
    if (trimmed.toLowerCase().includes('error')) {
      errorCount++;
      if (verbose) {
        console.log(`Found error: ${trimmed}`);
      }
    }
    
    if (trimmed.toLowerCase().includes('warning') || trimmed.toLowerCase().includes('warn')) {
      warningCount++;
      if (verbose) {
        console.log(`Found warning: ${trimmed}`);
      }
    }
    
    // Try to parse JSON lines for metrics
    if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
      try {
        const jsonData = JSON.parse(trimmed);
        
        // Extract token usage from various fields
        const tokenFields = ['tokens', 'token_count', 'input_tokens', 'output_tokens', 'total_tokens'];
        for (const field of tokenFields) {
          if (jsonData[field] && typeof jsonData[field] === 'number') {
            totalTokens += jsonData[field];
            if (verbose) {
              console.log(`Found ${field}: ${jsonData[field]}`);
            }
          }
        }
        
        // Check nested usage object (common in OpenAI/GPT responses)
        if (jsonData.usage) {
          if (jsonData.usage.total_tokens) {
            totalTokens += jsonData.usage.total_tokens;
            if (verbose) {
              console.log(`Found usage.total_tokens: ${jsonData.usage.total_tokens}`);
            }
          } else {
            // Sum input and output tokens if total not available
            const inputTokens = jsonData.usage.prompt_tokens || jsonData.usage.input_tokens || 0;
            const outputTokens = jsonData.usage.completion_tokens || jsonData.usage.output_tokens || 0;
            if (inputTokens || outputTokens) {
              const tokens = inputTokens + outputTokens;
              totalTokens += tokens;
              if (verbose) {
                console.log(`Found input_tokens: ${inputTokens}, output_tokens: ${outputTokens}, total: ${tokens}`);
              }
            }
          }
        }
        
        // Extract cost information
        const costFields = ['cost', 'price', 'amount', 'total_cost', 'estimated_cost'];
        for (const field of costFields) {
          if (jsonData[field] && typeof jsonData[field] === 'number') {
            estimatedCost += jsonData[field];
            if (verbose) {
              console.log(`Found ${field}: ${jsonData[field]}`);
            }
          }
        }
        
        // Check for model information
        if (verbose && (jsonData.model || jsonData.Model)) {
          console.log(`Found model: ${jsonData.model || jsonData.Model}`);
        }
        
      } catch (parseError) {
        // Not valid JSON, continue parsing
        if (verbose) {
          console.log(`Failed to parse JSON line: ${trimmed.substring(0, 100)}...`);
        }
      }
    }
  }
  
  // Output metrics as JSON
  const metrics = {
    TokenUsage: totalTokens,
    EstimatedCost: estimatedCost,
    ErrorCount: errorCount,
    WarningCount: warningCount
  };
  
  console.log(JSON.stringify(metrics));
  
  if (verbose) {
    console.log(`\nParsing complete:`);
    console.log(`  Total tokens: ${totalTokens}`);
    console.log(`  Estimated cost: $${estimatedCost.toFixed(4)}`);
    console.log(`  Errors: ${errorCount}`);
    console.log(`  Warnings: ${warningCount}`);
  }
}

main();