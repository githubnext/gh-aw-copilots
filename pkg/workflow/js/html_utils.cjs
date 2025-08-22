/**
 * Shared utility function to strip HTML tags from text
 * @param {string} text - The text to strip HTML tags from
 * @returns {string} - The text with HTML tags removed
 */
function stripHtmlTags(text) {
  // Remove HTML tags using regex
  return text.replace(/<[^>]*>/g, '');
}

/**
 * Check if HTML content should be stripped based on environment variable
 * @param {string} envVarName - The name of the environment variable to check
 * @returns {boolean} - True if HTML should be stripped, false otherwise
 */
function shouldStripHTML(envVarName) {
  const allowHTML = process.env[envVarName];
  return allowHTML === 'false';
}

module.exports = {
  stripHtmlTags,
  shouldStripHTML
};