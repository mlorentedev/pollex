#!/usr/bin/env bash
# Install pollex artifacts on the Jetson Nano after rsync.
# Called by `make deploy` after all files are staged in /tmp.
# Usage: ssh user@host 'bash -s' < scripts/jetson-install.sh
set -euo pipefail

# Binary
sudo mv /tmp/pollex /usr/local/bin/pollex
sudo chmod +x /usr/local/bin/pollex

# Config + prompts
sudo mkdir -p /etc/pollex
sudo mv /tmp/pollex-config.yaml /etc/pollex/config.yaml
sudo mv /tmp/pollex-polish.txt /etc/pollex/polish.txt
sudo mv /tmp/pollex-polish-cloud.txt /etc/pollex/polish-cloud.txt

# Service file
sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service
sudo systemctl daemon-reload
