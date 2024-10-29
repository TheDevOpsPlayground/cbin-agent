const fs = require('fs-extra');
const path = require('path');

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

/**
 * Create the root directory structure if it doesn't exist.
 * @param {string} rootDir - Root directory path.
 */
function createRootDirectory(rootDir) {
  const directories = ['recycle-bin', 'backup'];
  try {
    directories.forEach((dir) => {
      const fullPath = path.join(rootDir, dir);
      fs.ensureDirSync(fullPath);
    });

    // Create metadata.json if not exists
    const metadataFile = path.join(rootDir, 'metadata.json');
    if (!fs.existsSync(metadataFile)) {
      fs.writeJsonSync(metadataFile, { files: [] }, { spaces: 2 });
    }

    console.log(`Directory structure created at ${rootDir}`);
  } catch (error) {
    console.error(`Error creating directory structure: ${error.message}`);
    process.exit(1);
  }
}

module.exports = { moveFile, createRootDirectory };
