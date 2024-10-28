const fs = require('fs-extra');

/**
 * Move a file from source to destination.
 * @param {string} src - Source file path.
 * @param {string} dest - Destination file path.
 */
async function moveFile(src, dest) {
  try {
    await fs.move(src, dest, { overwrite: true });
    console.log(`Moved ${src} to ${dest}`);
  } catch (error) {
    console.error(`Failed to move ${src}: ${error.message}`);
    throw error;
  }
}

module.exports = { moveFile };
