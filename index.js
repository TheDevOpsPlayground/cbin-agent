#!/usr/bin/env node

const { Command } = require('commander');
const { moveFile } = require('./utils');
const { monitorFiles } = require('./monitor');
const config = require('./config.json');
const program = new Command();

program
  .command('move <file>')
  .description('Move a file to the recycle bin (mounted disk)')
  .action((file) => {
    const destination = config.recycleBinPath;
    moveFile(file, destination);
  });

program
  .command('monitor')
  .description('Monitor files for changes based on config.json')
  .action(() => {
    if (config.monitoring.enabled) {
      monitorFiles(config.monitoring.paths, config.backupPath);
    } else {
      console.log('File monitoring is disabled in config.json');
    }
  });

program.parse(process.argv);
