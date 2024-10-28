const fs = require('fs-extra');
const path = require('path');

async function moveFile(file, destination) {
  try {
    const destPath = path.join(destination, path.basename(file));
    await fs.move(file, destPath, { overwrite: true });
    console.log(`Moved ${file} to ${destPath}`);
  } catch (error) {
    console.error(`Error moving file: ${error.message}`);
  }
}

module.exports = { moveFile };
