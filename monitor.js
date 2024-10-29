const chokidar = require('chokidar');
const { getVersionedBackupPath, updateMetadata } = require('./utils');
const fs = require('fs-extra');
const path = require('path');

/**
 * Monitor files for changes using chokidar
 * @param {Array<string>} paths - List of file paths to monitor
 * @param {string} backupPath - Path to save backups
 * @param {string} metadataPath - Path to metadata.json
 */
function monitorFiles(paths, backupPath, metadataPath) {
  const watcher = chokidar.watch(paths, { persistent: true });

  watcher
    .on('change', async (file) => await handleFileChange(file, backupPath, metadataPath))
    .on('error', (error) => console.error(`Watcher error: ${error}`));

  console.log('Started monitoring files...');
}

/**
 * Handle file change events by creating a versioned backup.
 * @param {string} file - The file that changed.
 * @param {string} backupDir - Directory to store backups.
 * @param {string} metadataPath - Path to metadata.json.
 */
async function handleFileChange(file, backupDir, metadataPath) {
  const versionedBackupPath = getVersionedBackupPath(file, backupDir);

  try {
    await fs.copy(file, versionedBackupPath);
    console.log(`Backed up ${file} to ${versionedBackupPath}`);

    await updateMetadata(metadataPath, file, versionedBackupPath);
  } catch (error) {
    console.error(`Error backing up ${file}: ${error.message}`);
  }
}

module.exports = { monitorFiles };
