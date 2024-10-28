const { Inotify } = require('inotify');
const fs = require('fs-extra');
const path = require('path');

const inotify = new Inotify(); 

function monitorFiles(paths, backupPath) {
  paths.forEach((filePath) => {
    if (!fs.existsSync(filePath)) {
      console.log(`File not found: ${filePath}`);
      return;
    }

    inotify.addWatch({
      path: filePath,
      watch_for: Inotify.IN_MODIFY | Inotify.IN_CREATE | Inotify.IN_DELETE,
      callback: (event) => handleEvent(event, filePath, backupPath),
    });

    console.log(`Monitoring ${filePath} for changes...`);
  });
}

function handleEvent(event, filePath, backupPath) {
  const eventType = event.mask & Inotify.IN_MODIFY ? 'modified' : 
                    event.mask & Inotify.IN_DELETE ? 'deleted' : 'created';

  console.log(`File ${eventType}: ${filePath}`);
  
  if (eventType === 'modified') {
    const backupFile = path.join(backupPath, path.basename(filePath));
    fs.copy(filePath, backupFile)
      .then(() => console.log(`Backed up ${filePath} to ${backupFile}`))
      .catch((err) => console.error(`Backup failed: ${err.message}`));
  }
}

module.exports = { monitorFiles };
