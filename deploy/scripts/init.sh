#!/bin/bash
set -euo pipefail

echo "=== Pollex first-time setup on Jetson ==="

# 1. Install required tools
echo "Installing prerequisites..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl zstd

# 2. Add CUDA to PATH (if not already present)
if ! grep -q '/usr/local/cuda/bin' ~/.bashrc; then
  echo "Adding CUDA to PATH..."
  echo 'export PATH=/usr/local/cuda/bin:$PATH' >> ~/.bashrc
  echo "CUDA added to PATH (reload shell or source ~/.bashrc)"
fi

# 3. Create config directory
echo "Setting up /etc/pollex/..."
sudo mkdir -p /etc/pollex
sudo chown "$(whoami):$(whoami)" /etc/pollex

# 4. Install systemd services
echo "Installing systemd services..."
sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service
sudo cp /tmp/llama-server.service /etc/systemd/system/llama-server.service
sudo systemctl daemon-reload
sudo systemctl enable pollex-api
sudo systemctl enable llama-server

echo ""
echo "=== Setup complete ==="
echo "Next steps:"
echo "  1. make deploy-llamacpp   # Build llama.cpp with CUDA (~85 min)"
echo "  2. make deploy            # Deploy binary + config"
echo "  3. make deploy-secrets    # Deploy API key"
echo "  4. make deploy-tunnel     # Setup Cloudflare Tunnel"
