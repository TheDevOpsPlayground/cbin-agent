#!/bin/bash

# Define variables
INSTALL_DIR="/opt/recycler-cli"
BIN_PATH="/usr/local/bin/recycler-cli"
CONFIG_DIR="/etc/recycler-cli"
LOG_DIR="/var/log/recycler-cli"
CONFIG_FILE="$CONFIG_DIR/config.conf"
SYSTEMD_FILE="/etc/systemd/system/recycler-cli.service"

# Create directories
echo "Creating necessary directories..."
sudo mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"

# Download the binary from GitHub
echo "Downloading recycler-cli binary..."
curl -L -o "$INSTALL_DIR/recycler-cli" "https://github.com/Toymakerftw/recycler/raw/refs/heads/go/recycler-cli"

# Set permissions
echo "Setting permissions..."
sudo chmod +x "$INSTALL_DIR/recycler-cli"
sudo chown -R root:root "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"

# Create symbolic link to make globally accessible
echo "Linking binary to /usr/local/bin..."
sudo ln -sf "$INSTALL_DIR/recycler-cli" "$BIN_PATH"

# Copy configuration file (create a default one if not exists)
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Creating default configuration..."
  echo -e "{\n  \"recycleBinDir\": \"/mnt/recycle-bin\",\n  \"numWorkers\": 4\n}" | sudo tee "$CONFIG_FILE" > /dev/null
fi

# Create systemd service file
echo "Creating systemd service..."
sudo bash -c "cat > $SYSTEMD_FILE" <<EOF
[Unit]
Description=Recycler CLI Service
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/recycler-cli
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start the service
echo "Enabling and starting recycler-cli service..."
sudo systemctl daemon-reload
sudo systemctl enable recycler-cli.service
sudo systemctl start recycler-cli.service

echo "Installation complete. Recycler CLI is ready to use."

# Inform the user about configuration and log file locations
echo "----------------------------------------"
echo "Recycler CLI has been installed."
echo "Configuration file is located at: $CONFIG_FILE"
echo "Log file can be found at: $LOG_DIR/recycler-cli.log"
echo "----------------------------------------"