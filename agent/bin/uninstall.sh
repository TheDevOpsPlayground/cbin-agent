#!/bin/bash

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root."
    exit 1
fi

# Stop and disable services
echo "Stopping and disabling services..."
systemctl stop cbin
systemctl disable cbin
systemctl stop health
systemctl disable health

# Remove systemd service files
echo "Removing systemd service files..."
rm -f /etc/systemd/system/cbin.service
rm -f /etc/systemd/system/health.service

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Remove symbolic links
echo "Removing symbolic links..."
rm -f /usr/local/bin/cbin
rm -f /usr/local/bin/health

# Remove binaries
echo "Removing binaries..."
rm -rf /opt/cbin

# Remove configuration files
echo "Removing configuration files..."
rm -rf /etc/cbin

# Remove log files
echo "Removing log files..."
rm -rf /var/log/cbin

# Remove mount point
echo "Unmounting and removing mount point..."
umount /mnt/recyclebin
rm -rf /mnt/recyclebin

# Remove alias from bash.bashrc
echo "Removing alias from bash.bashrc..."
sed -i '/alias rm/d' /etc/bash.bashrc

# Remove NFS mount from /etc/fstab
echo "Removing NFS mount from /etc/fstab..."
sed -i '/nfs/d' /etc/fstab

# Remove environment files
echo "Removing environment files..."
rm -f /etc/cbin/env
rm -f $(pwd)/.env

# Remove install directory
echo "Removing install directory..."
rm -rf /opt/cbin

echo "Uninstallation complete."