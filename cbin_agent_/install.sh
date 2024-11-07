#!/bin/bash

# Variables
INSTALL_DIR="/etc/cbin"
MOUNT_POINT="/mnt/recyclebin"
CURRENT_DIR="$(pwd)"   
ENV_FILE_SRC="$CURRENT_DIR/env" 
ENV_FILE="$INSTALL_DIR/env"      
BINARY_SRC="$CURRENT_DIR/health"
BINARY="$INSTALL_DIR/health"     
SERVICE_FILE="/etc/systemd/system/cbin.service"
MAIN_SRC="$CURRENT_DIR/recycle"
MAIN_BINARY="$INSTALL_DIR/recycle" 


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

# Check if source binary file exists
if [[ ! -f "$BINARY_SRC" ]]; then
    echo "Error: Binary file $BINARY_SRC not found."
    exit 1
fi

if [[ ! -f "$MAIN_SRC" ]]; then
    echo "Error: Binary file $BINARY_SRC not found."
    exit 1
fi







mkdir -p "$INSTALL_DIR"
mkdir -p "$MOUNT_POINT"

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

cp "$BINARY_SRC" "$BINARY"
chmod +x "$BINARY"

cp "$MAIN_SRC" "$MAIN_BINARY"
chmod +x "$MAIN_BINARY"




cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=Recycle Bin Service
After=network.target

[Service]
EnvironmentFile=$ENV_FILE
ExecStart=$BINARY
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable cbin.service
systemctl start cbin.service

# Add alias to global bashrc for 'rm' command replacement
if ! grep -Fxq "alias rm='$MAIN_BINARY'" /etc/bash.bashrc; then
   echo "alias rm='$MAIN_BINARY'" | sudo tee -a /etc/bash.bashrc > /dev/null
fi

source /etc/bash.bashrc

echo "Installation completed successfully."
