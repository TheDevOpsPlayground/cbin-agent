const chokidar = require('chokidar');
const fs = require('fs-extra');
const path = require('path');

/**
 * Monitor files for changes using chokidar
 * @param {Array<string>} paths - List of file paths to monitor
 * @param {string} backupPath - Path to save backups
 */
function monitorFiles(paths, backupPath) {
  const watcher = chokidar.watch(paths, { persistent: true });

  watcher
    .on('add', (file) => console.log(`File added: ${file}`))
    .on('change', (file) => handleFileChange(file, backupPath))
    .on('unlink', (file) => console.log(`File removed: ${file}`))
    .on('error', (error) => console.error(`Watcher error: ${error}`));

  console.log('Started monitoring files...');
}

/**
 * Handle file change events by creating a backup of the modified file.
 * @param {string} file - The file that changed
 * @param {string} backupPath - Path to store the backup
 */
async function handleFileChange(file, backupPath) {
  const backupFile = path.join(backupPath, path.basename(file));

  try {
    await fs.copy(file, backupFile);
    console.log(`Backed up ${file} to ${backupFile}`);
  } catch (error) {
    console.error(`Error backing up ${file}: ${error.message}`);
  }
}

module.exports = { monitorFiles };
