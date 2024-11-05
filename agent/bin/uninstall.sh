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
BASHRC_FILE="/etc/bash.bashrc"

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root."
   exit 1
fi

# Stop and disable systemd services
echo "Stopping and disabling systemd services..."
sudo systemctl stop cbin 2>/dev/null || true
sudo systemctl disable cbin 2>/dev/null || true
sudo systemctl stop health 2>/dev/null || true
sudo systemctl disable health 2>/dev/null || true

# Remove systemd service files
echo "Removing systemd service files..."
sudo rm -f "$CBINSYSTEMD_FILE" "$HEALTHCHECKERSYSTEMD_FILE"

# Reload systemd to apply changes
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Remove symbolic links
echo "Removing symbolic links..."
sudo rm -f "$CBIN_PATH" "$HEALTHCHECKER_PATH"

# Remove installation directories and files
echo "Removing installation directories and files..."
sudo rm -rf "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR" "$MOUNT_POINT"

# Remove alias from global bashrc if it exists
echo "Removing alias from bashrc..."
sudo sed -i.bak '/alias rm=/d' "$BASHRC_FILE"

# Confirmation message
echo -e "\nUninstallation Complete!"
echo -e "All files, directories, and services related to cbin and health-checker have been removed."