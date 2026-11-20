#!/bin/bash

# Target configuration
PI_IP="192.168.1.217"

# Ask for the SSH username
echo "Enter SSH username for the Raspberry Pi 5 at $PI_IP (e.g. 'pi' or 'sabrodigan'):"
read -r PI_USER

if [ -z "$PI_USER" ]; then
    echo "Error: Username cannot be empty."
    exit 1
fi

echo "--------------------------------------------------"
echo "Step 1: Testing SSH connection to $PI_USER@$PI_IP..."
echo "--------------------------------------------------"
ssh -o StrictHostKeyChecking=no "$PI_USER@$PI_IP" "echo 'SSH connection successful!'" || {
    echo "Error: Could not connect to $PI_USER@$PI_IP. Check the username and password."
    exit 1
}

echo "--------------------------------------------------"
echo "Step 2: Building piMonitor locally for ARM64..."
echo "--------------------------------------------------"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.build=$(git rev-parse --short HEAD)" -o piMonitor . || {
    echo "Error: Local build failed."
    exit 1
}
echo "Build complete."

echo "--------------------------------------------------"
echo "Step 3: Creating directory and copying files to Pi..."
echo "--------------------------------------------------"
ssh "$PI_USER@$PI_IP" "mkdir -p ~/piMonitor"
scp -r piMonitor templates "$PI_USER@$PI_IP:~/piMonitor/" || {
    echo "Error: Failed to copy files to the Pi."
    exit 1
}

echo "--------------------------------------------------"
echo "Step 4: Setting up systemd background service..."
echo "--------------------------------------------------"

# Generate the systemd service file dynamically with the correct user and paths
cat <<EOF > /tmp/pimonitor.service
[Unit]
Description=Pi 5 Hardware Monitor Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/$PI_USER/piMonitor
ExecStart=/home/$PI_USER/piMonitor/piMonitor
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Copy service file to the Pi and activate it
scp /tmp/pimonitor.service "$PI_USER@$PI_IP:/tmp/"
rm /tmp/pimonitor.service

ssh -t "$PI_USER@$PI_IP" "sudo mv /tmp/pimonitor.service /etc/systemd/system/ && \
                        sudo systemctl daemon-reload && \
                        sudo systemctl enable pimonitor && \
                        sudo systemctl restart pimonitor" || {
    echo "Error: Failed to register and start systemd service on the Pi."
    exit 1
}

echo "--------------------------------------------------"
echo "Deployment successful!"
echo "--------------------------------------------------"
echo "piMonitor is now running in the background on your Pi 5."
echo "You can check the running service logs anytime with:"
echo "  ssh $PI_USER@$PI_IP 'sudo journalctl -u pimonitor -f'"
echo "--------------------------------------------------"
