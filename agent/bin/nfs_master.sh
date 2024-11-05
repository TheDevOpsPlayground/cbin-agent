#!/bin/bash

# Define variables
EXPORT_DIR="/mnt/recyclebin"
CLIENT_IP="192.168.122.98"  # Replace with Server B's IP address
PERMISSIONS="rw,sync,no_subtree_check"

# Create directory to share
sudo mkdir -p $EXPORT_DIR
sudo chown nobody:nogroup $EXPORT_DIR
sudo chmod 755 $EXPORT_DIR

# Add export entry to /etc/exports
echo "$EXPORT_DIR $CLIENT_IP($PERMISSIONS)" | sudo tee -a /etc/exports

# Restart NFS service
sudo exportfs -a
sudo systemctl restart nfs-kernel-server
sudo systemctl enable nfs-kernel-server

echo "NFS Server setup complete. Shared directory: $EXPORT_DIR"