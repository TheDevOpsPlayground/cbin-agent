# Recycler CLI

Recycler CLI is a **safe alternative to `rm`**. Instead of permanently deleting files, it **moves them to a mounted disk** (recycle bin). Additionally, the tool can **monitor specific files** for changes using **Chokidar** and **create backups** when modifications occur.

---

## Features

- **Recycle Instead of Delete**: Safely move files to a designated recycle bin directory.
- **File Monitoring**: Watch specific files for changes (create, modify, delete).
- **Automated Backups**: Backup modified files to a configured directory.
- **Configurable via `config.json`**: Easily toggle features and paths for monitoring.

---

## Prerequisites

- **Node.js** v12+ installed on your system.
- A **mounted disk** for the recycle bin and backup locations (e.g., `/mnt/recycle-bin`).
- Ensure you have Chokidar available (installed via `npm install chokidar`).

---

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/Toymakerftw/recycler recycler-cli
   cd recycler-cli
   ```

2. Install dependencies:

   ```bash
   npm install
   ```

3. Make the CLI executable:

   ```bash
   chmod +x index.js
   ```

4. Optionally, link the CLI globally:

   ```bash
   sudo npm link
   ```

---

## Usage

### 1. **Move a File to the Recycle Bin**

```bash
recycler move <file>
```

Example:

```bash
recycler move /home/user/sample.txt
```

This moves the file to the recycle bin path defined in `config.json`.

### 2. **Monitor Files for Changes**

```bash
recycler monitor
```

This command monitors the files listed in `config.json` and backs them up on modification.

---

## Configuration

Modify the `config.json` to customize paths and toggle features.

```json
{
  "recycleBinPath": "/mnt/recycle-bin",
  "backupPath": "/mnt/backup",
  "monitoring": {
    "enabled": true,
    "paths": [
      "/home/user/important-file.txt",
      "/home/user/another-file.txt"
    ]
  }
}
```

- **`recycleBinPath`**: The directory where files will be moved instead of being deleted.
- **`backupPath`**: Directory where modified files will be backed up.
- **`monitoring.enabled`**: Toggle file monitoring on or off.
- **`monitoring.paths`**: List of files to be monitored for changes.

For production use, place the `config.json` file at `/etc/recycler-cli/config.json`.

---

## Add a Bash Alias

To use **`recycler`** as a safer alternative to `rm`, add the following to your **`.bashrc`**:

```bash
alias rm=recycler
```

Reload the `.bashrc`:

```bash
source ~/.bashrc
```

Now, when you run:

```bash
rm some-file.txt
```

It will invoke the `recycler` tool and move the file to the recycle bin.

---

## Example Commands

1. **Move a File**:

   ```bash
   recycler move /home/user/document.txt
   ```

2. **Monitor Files for Changes**:

   ```bash
   recycler monitor
   ```

---

## To Do

- **Logging**: Add logging for all events (e.g., file movements, backups) using a logging library like **winston** or **console.log** inside the utility functions for better debugging.
- **Run as a Daemon**: Use `pm2` or `systemd` to keep the monitoring process running in the background.
- **Notifications**: Integrate with a notification service to alert on critical events.
- **Unit Tests**: Write tests with **Jest** or **Mocha** for better code quality.

---

## Production Deployment

To deploy the Recycler CLI tool in a production environment:

1. **Package the Application**: Ensure the CLI tool is installed globally using:
   ```bash
   sudo npm install -g .
   ```

2. **Create a Config File for Production**: Store the production config at `/etc/recycler-cli/config.json`.

3. **Use `systemd` to Manage the Service**:
   - Create a systemd service file at `/etc/systemd/system/recycler.service` with the following content:
     ```ini
     [Unit]
     Description=Recycler CLI Monitor Service
     After=network.target

     [Service]
     ExecStart=/usr/local/bin/recycler monitor -c /etc/recycler-cli/config.json
     Restart=always
     User=<your-username>
     Environment=NODE_ENV=production
     StandardOutput=journal
     StandardError=journal

     [Install]
     WantedBy=multi-user.target
     ```

   - Start the service and enable it to run on boot:
     ```bash
     sudo systemctl daemon-reload
     sudo systemctl start recycler
     sudo systemctl enable recycler
     ```

4. **Monitor the Service**: Use `journalctl` to check logs and ensure the tool is functioning correctly:
   ```bash
   sudo journalctl -u recycler -f
   ```

---

## Uninstallation

To uninstall the Recycler CLI tool and remove the alias:

1. If you linked the tool globally, you can unlink it with:

   ```bash
   sudo npm unlink recycler-cli
   ```

2. To remove the alias, open your `.bashrc` in a text editor and remove the line:

   ```bash
   alias rm=recycler
   ```

3. Reload the `.bashrc` to apply changes:

   ```bash
   source ~/.bashrc
   ```

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

## Contributing

Feel free to open issues or submit pull requests if you’d like to improve this project!

---

## Author

Developed by Anandhraman. 🚀