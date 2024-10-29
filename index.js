#!/usr/bin/env node

const { Command } = require('commander');
const { moveFile, createRootDirectory } = require('./utils');
const { monitorFiles } = require('./monitor');
const fs = require('fs');
const os = require('os');
const path = require('path');

// Default config path (can be overridden)
const DEFAULT_CONFIG_PATH = '/etc/recycler-cli/config.json';

/**
 * Load configuration from the specified config file.
 * @param {string} configPath - Path to the config file.
 * @returns {Object} - Parsed config object.
 */
function loadConfig(configPath) {
  try {
    const data = fs.readFileSync(configPath, 'utf8');
    return JSON.parse(data);
  } catch (error) {
    console.error(`Error reading config file: ${error.message}`);
    process.exit(1);
  }
}

/**
 * Generate a unique root directory based on private IP and hostname.
 * @param {string} basePath - Base path from config.
 * @returns {string} - Full root directory path.
 */
function getRootDirectory(basePath) {
  const hostname = os.hostname();
  const privateIP = Object.values(os.networkInterfaces())
    .flat()
    .filter((iface) => iface.family === 'IPv4' && !iface.internal)
    .map((iface) => iface.address)[0];

  return path.join(basePath, `${privateIP}_${hostname}`);
}

const program = new Command();
program
  .option('-c, --config <path>', 'Path to config.json', DEFAULT_CONFIG_PATH);

program
  .command('init')
  .description('Initialize the root directory structure')
  .action(() => {
    const config = loadConfig(program.opts().config);
    const rootDirectory = getRootDirectory(config.DEFAULT_METADATA_PATH);
    createRootDirectory(rootDirectory);
    console.log(`Root directory initialized at ${rootDirectory}`);
  });

program
  .command('move <file>')
  .description('Move a file to the recycle bin (mounted disk)')
  .action((file) => {
    const config = loadConfig(program.opts().config);
    const rootDirectory = getRootDirectory(config.DEFAULT_METADATA_PATH);
    moveFile(file, path.join(rootDirectory, 'recycle-bin'));
  });

program
  .command('monitor')
  .description('Monitor files for changes based on config.json')
  .action(() => {
    const config = loadConfig(program.opts().config);
    if (config.monitoring.enabled) {
      const rootDirectory = getRootDirectory(config.DEFAULT_METADATA_PATH);
      monitorFiles(config.monitoring.paths, path.join(rootDirectory, 'backup'));
    } else {
      console.log('File monitoring is disabled in config.json');
    }
  });

program
  .command('delete <file>')
  .description('Move deleted file to recycle bin instead of permanently deleting')
  .action((file) => {
    const config = loadConfig(program.opts().config);
    const rootDirectory = getRootDirectory(config.DEFAULT_METADATA_PATH);
    const destination = path.join(rootDirectory, 'recycle-bin', path.basename(file));

    try {
      moveFile(file, destination);
      console.log(`Moved ${file} to recycle bin at ${destination}`);
    } catch (error) {
      console.error(`Error moving ${file}: ${error.message}`);
    }
  });

program.parse(process.argv);
