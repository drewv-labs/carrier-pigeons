#!/bin/bash
set -e

VERSION="v1.0.0" # Update this or fetch dynamically via GitHub API
REPO="drewv-labs/carrier-pigeons" # Swap if your repo is private/named differently
NODE_ID=$(hostname)

echo "[*] Detecting architecture for $NODE_ID..."
ARCH=$(uname -m)
case $ARCH in
    aarch64) BIN_NAME="pigeoneer-linux-arm64" ;;
    riscv64) BIN_NAME="pigeoneer-linux-riscv64" ;;
    x86_64)  BIN_NAME="pigeoneer-linux-amd64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BIN_NAME"

echo "[*] Downloading Pigeoneer $VERSION from GitHub..."
# The -L flag is mandatory to follow GitHub's S3 redirects
curl -sL -o /usr/local/bin/pigeoneer "$DOWNLOAD_URL"
chmod +x /usr/local/bin/pigeoneer

echo "[*] Configuring Systemd Service..."
cat <<EOF > /etc/systemd/system/pigeoneer.service
[Unit]
Description=Carrier Pigeons Edge Telemetry Daemon
After=network.target

[Service]
ExecStart=/usr/local/bin/pigeoneer
Restart=always
RestartSec=5
Environment="PIGEONEER_NODE_ID=$NODE_ID"
Environment="PIGEONEER_BROKER_URL=tcp://pigeoncoop.local:1883"

[Install]
WantedBy=multi-user.target
EOF

echo "[*] Enabling and starting Pigeoneer..."
systemctl daemon-reload
systemctl enable --now pigeoneer

echo "[✔] Node $NODE_ID is online and reporting."
