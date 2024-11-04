#!/bin/bash

# Define variables
INSTALL_DIR="/opt/cbin"
CBIN_PATH="/usr/local/bin/cbin"
HEALTHCHECKER_PATH="/usr/local/bin/health"
CONFIG_DIR="/etc/cbin"
LOG_DIR="/var/log/cbin"
MOUNT_POINT="/mnt/recyclebin"
CBINSYSTEMD_FILE="/etc/systemd/system/cbin.service"
HEALTHCHECKERSYSTEMD_FILE="/etc/systemd/system/healthchecker.service"
CURRENT_DIR="$(pwd)"   
ENV_FILE_SRC="$CURRENT_DIR/env" 
ENV_FILE="$INSTALL_DIR/env" 

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root." 
   exit 1
fi

# Check if source env file exists
if [[ ! -f "$ENV_FILE_SRC" ]]; then
    echo "Error: Environment file $ENV_FILE_SRC not found."
    exit 1
fi

# Create necessary directories
echo "Creating installation directories..."
sudo mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR" "$MOUNT_POINT"

cp "$ENV_FILE_SRC" "$ENV_FILE"
source "$ENV_FILE"

if [[ -z "$master_ip" || -z "$client_ip" ]]; then
    echo "Error: Missing required environment variables in $ENV_FILE."
    exit 1
fi

mount -o rw,sync,nfsvers=4 "$master_ip:/mnt/check/$client_ip" "$MOUNT_POINT"
if [[ $? -ne 0 ]]; then
   echo "Error: Failed to mount NFS. Installation aborted."
   exit 1
fi

# Download binary from GitHub and place it in the installation directory
echo "Downloading cbin binary..."
sudo curl -L "https://github.com/Toymakerftw/recycler/raw/refs/heads/cbin/agent/bin/cbin" -o "$INSTALL_DIR/cbin"
sudo chmod +x "$INSTALL_DIR/cbin"
|
echo "Downloading health-checker binary..."
sudo curl -L "https://github.com/Toymakerftw/recycler/raw/refs/heads/cbin/agent/bin/health" -o "$INSTALL_DIR/health"
sudo chmod +x "$INSTALL_DIR/health"

# Create symbolic link to make it globally accessible
echo "Creating symbolic link for global access..."
sudo ln -sf "$INSTALL_DIR/cbin" "$CBIN_PATH"


# Set up default configuration file
echo "Setting up default configuration file..."
cat <<EOL | sudo tee "$CONFIG_DIR/config.conf" > /dev/null
{
  "recycleBinDir": "$MOUNT_POINT",
  "numWorkers": 4
}
EOL

# Adjust ownership and permissions for the log and recycle bin directories
echo "Adjusting permissions for log and recycle bin directories..."
sudo chown -R "$USER":"$USER" "$LOG_DIR" "$MOUNT_POINT"
sudo chmod 755 "$LOG_DIR" "$MOUNT_POINT"

# Set up systemd service for background running (optional)
echo "Creating systemd service..."
cat <<EOL | sudo tee "$CBINSYSTEMD_FILE" > /dev/null
[Unit]
Description=Recycler CLI Service
After=network.target

[Service]
ExecStart=$CBIN_PATH
User=$USER
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

# Set up systemd service for background running (optional)
echo "Creating systemd service..."
cat <<EOL | sudo tee "$HEALTHCHECKERSYSTEMD_FILE" > /dev/null
[Unit]
Description=Recycler CLI Service
After=network.target

[Service]
ExecStart=$HEALTHCHECKER_PATH
User=$USER
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

# Reload systemd, enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable cbin
sudo systemctl start cbin

# Reload systemd, enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable health
sudo systemctl start health

# Add alias to global bashrc for 'rm' command replacement
if ! grep -Fxq "alias rm='$CBIN_PATH'" /etc/bash.bashrc; then
   echo "alias rm='$CBIN_PATH'" | sudo tee -a /etc/bash.bashrc > /dev/null
fi

# Display configuration and log file locations, along with usage information
echo -e "\nInstallation Complete!"
echo -e "---------------------------------------------------------------"
echo -e "cbin is now installed and globally accessible as 'cbin'"
echo -e "Default Configuration File: $CONFIG_DIR/config.conf"
echo -e "Log File: $LOG_DIR/cbin.log"
echo -e "Recycle Bin Directory: $MOUNT_POINT"
echo -e "---------------------------------------------------------------"
echo -e "Usage:"
echo -e "  cbin -f <file1,file2,...>     # Move specified files to recycle bin"
echo -e "  cbin -h                       # Display help message"
echo -e "Example:"
echo -e "  cbin -f file1.txt,file2.log"
echo -e "---------------------------------------------------------------"
echo -e "The tool will automatically recycle the specified files to the configured directory.\n"