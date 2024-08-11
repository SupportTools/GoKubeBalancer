#!/bin/bash

# Define variables
DOWNLOAD_URL="https://github.com/SupportTools/GoKubeBalancer/releases/download/v1.0.0/GoKubeBalancer_Linux_x86_64.tar.gz"
INSTALL_DIR="/opt/GoKubeBalancer"
SERVICE_FILE="/etc/systemd/system/gokubebalancer.service"

# Prompt user for values
read -p "Enter RANCHER_API: " RANCHER_API
read -p "Enter RANCHER_KEY: " RANCHER_KEY
read -p "Enter RANCHER_CLUSTER: " RANCHER_CLUSTER

# Create install directory
sudo mkdir -p "$INSTALL_DIR"

# Download and extract tar.gz
curl -L "$DOWNLOAD_URL" | sudo tar xz -C "$INSTALL_DIR"

# Create .env file
cat <<EOL | sudo tee "$INSTALL_DIR/.env"
export RANCHER_API=$RANCHER_API
export RANCHER_KEY=$RANCHER_KEY
export RANCHER_CLUSTER=$RANCHER_CLUSTER
EOL

# Create systemd service file
cat <<EOL | sudo tee "$SERVICE_FILE"
[Unit]
Description=GoKubeBalancer Service
After=network.target

[Service]
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/GoKubeBalancer
WorkingDirectory=$INSTALL_DIR
Restart=always
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=gokubebalancer

[Install]
WantedBy=multi-user.target
EOL

# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable gokubebalancer
sudo systemctl start gokubebalancer

echo "Installation complete. GoKubeBalancer service is now running."
