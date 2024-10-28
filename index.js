#!/usr/bin/env node

const { Command } = require('commander');
const { moveFile } = require('./utils');
const { monitorFiles } = require('./monitor');
const fs = require('fs');

// Default config path (can be overridden by CLI flag)
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

const program = new Command();
program
  .option('-c, --config <path>', 'Path to config.json', DEFAULT_CONFIG_PATH);

program
  .command('move <file>')
  .description('Move a file to the recycle bin (mounted disk)')
  .action((file, options) => {
    const config = loadConfig(program.opts().config);
    moveFile(file, config.recycleBinPath);
  });

program
  .command('monitor')
  .description('Monitor files for changes based on config.json')
  .action(() => {
    const config = loadConfig(program.opts().config);

    if (config.monitoring.enabled) {
      monitorFiles(config.monitoring.paths, config.backupPath);
    } else {
      console.log('File monitoring is disabled in config.json');
    }
  });

program.parse(process.argv);
