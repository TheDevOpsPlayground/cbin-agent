#!/bin/bash

# Define variables
INSTALL_DIR="/opt/cbin"
CBIN_PATH="/usr/local/bin/cbin"
HEALTHCHECKER_PATH="/usr/local/bin/health"
CONFIG_DIR="/etc/cbin"
LOG_DIR="/var/log/cbin"
MOUNT_POINT="/mnt/recyclebin"
CBINSYSTEMD_FILE="/etc/systemd/system/cbin.service"
HEALTHCHECKERSYSTEMD_FILE="/etc/systemd/system/health.service"
CURRENT_DIR="$(pwd)"
ENV_FILE_SRC="$CURRENT_DIR/env"
ENV_FILE="$INSTALL_DIR/env"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root."
    exit 1
fi

# Verify environment file exists
if [[ ! -f "$ENV_FILE_SRC" ]]; then
    echo "Error: Environment file $ENV_FILE_SRC not found."
    exit 1
fi

# Create necessary directories with correct permissions
echo "Creating directories..."
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"
chown root:root "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"
chmod 755 "$INSTALL_DIR"

# Copy and source environment file
cp "$ENV_FILE_SRC" "$ENV_FILE"
source "$ENV_FILE"

# Ensure required environment variables are set
if [[ -z "$master_ip" || -z "$client_ip" ]]; then
    echo "Error: Missing required environment variables in $ENV_FILE."
    exit 1
fi

# Mount NFS directory and add to /etc/fstab
#echo "Mounting NFS directory..."
#mount -t nfs "$master_ip:$REMOTE_DIR" "$MOUNT_POINT"
#if [[ $? -ne 0 ]]; then
#    echo "Error: Failed to mount NFS. Installation aborted."
#    exit 1
#fi
#echo "$master_ip:$REMOTE_DIR $MOUNT_POINT nfs defaults 0 0" >> /etc/fstab

# Download binaries
echo "Downloading cbin and health-checker binaries..."
curl -L "https://github.com/Toymakerftw/recycler/raw/refs/heads/cbin/agent/bin/cbin" -o "$INSTALL_DIR/cbin"
curl -L "https://github.com/Toymakerftw/recycler/raw/refs/heads/cbin/agent/bin/health" -o "$INSTALL_DIR/health"
chmod +x "$INSTALL_DIR/cbin" "$INSTALL_DIR/health"

# Create symbolic links
ln -sf "$INSTALL_DIR/cbin" "$CBIN_PATH"
ln -sf "$INSTALL_DIR/health" "$HEALTHCHECKER_PATH"

# Configure default settings
echo "Creating default configuration file..."
cat <<EOL > "$CONFIG_DIR/config.conf"
{
  "recycleBinDir": "$MOUNT_POINT",
  "numWorkers": 4
}
EOL
chmod 644 "$CONFIG_DIR/config.conf"

# Systemd service setup
echo "Creating systemd services..."
cat <<EOL > "$CBINSYSTEMD_FILE"
[Unit]
Description=Recycler CLI Service
After=network.target

[Service]
ExecStart=$CBIN_PATH
User=root
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

cat <<EOL > "$HEALTHCHECKERSYSTEMD_FILE"
[Unit]
Description=Recycler Health Checker Service
After=network.target

[Service]
ExecStart=$HEALTHCHECKER_PATH
User=root
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

# Reload systemd, enable, and start services
systemctl daemon-reload
systemctl enable cbin
systemctl start cbin
systemctl enable health
systemctl start health

# Set up alias for rm command replacement
#if ! grep -Fxq "alias rm='$CBIN_PATH'" /etc/bash.bashrc; then
#    echo "alias rm='$CBIN_PATH'" >> /etc/bash.bashrc
#fi

# Completion message
echo -e "\nInstallation Complete!"
echo -e "---------------------------------------------------------------"
echo -e "cbin is now installed and globally accessible as 'cbin'"
echo -e "Configuration File: $CONFIG_DIR/config.conf"
echo -e "Log Directory: $LOG_DIR"
echo -e "Recycle Bin Directory: $MOUNT_POINT"
echo -e "---------------------------------------------------------------"
echo -e "Usage:"
echo -e "  cbin -f <file1,file2,...>     # Move specified files to recycle bin"
echo -e "  cbin -h                       # Display help message"
echo -e "Example:"
echo -e "  cbin -f file1.txt,file2.log"
echo -e "---------------------------------------------------------------"