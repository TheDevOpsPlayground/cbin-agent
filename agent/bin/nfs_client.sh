#!/bin/bash

# Define variables
SERVER_IP="192.168.122.121"  # Replace with Server A's IP address
REMOTE_DIR="/mnt/recyclebin"
LOCAL_MOUNT_DIR="/mnt/recyclebin"

# Create mount point on the client
sudo mkdir -p $LOCAL_MOUNT_DIR

# Mount NFS directory
sudo mount -t nfs $SERVER_IP:$REMOTE_DIR $LOCAL_MOUNT_DIR

# Add to /etc/fstab for persistence
echo "$SERVER_IP:$REMOTE_DIR $LOCAL_MOUNT_DIR nfs defaults 0 0" | sudo tee -a /etc/fstab

echo "NFS Client setup complete. Mounted directory: $LOCAL_MOUNT_DIR"