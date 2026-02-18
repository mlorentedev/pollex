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

# 4. Set headless mode (frees ~400MB RAM for inference)
CURRENT_TARGET=$(systemctl get-default)
if [ "$CURRENT_TARGET" != "multi-user.target" ]; then
  echo "Setting headless mode (multi-user.target)..."
  sudo systemctl set-default multi-user.target
  echo "Headless mode set (effective after reboot)"
else
  echo "Already in headless mode"
fi

# 5. Install systemd services
echo "Installing systemd services..."
sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service
sudo cp /tmp/llama-server.service /etc/systemd/system/llama-server.service
sudo cp /tmp/jetson-clocks.service /etc/systemd/system/jetson-clocks.service
sudo systemctl daemon-reload
sudo systemctl enable pollex-api
sudo systemctl enable llama-server
sudo systemctl enable --now jetson-clocks

echo ""
echo "=== Setup complete ==="
echo "Next steps:"
echo "  1. make deploy-llamacpp   # Build llama.cpp with CUDA (~85 min)"
echo "  2. make deploy            # Deploy binary + config"
echo "  3. make deploy-secrets    # Deploy API key"
echo "  4. make deploy-tunnel     # Setup Cloudflare Tunnel"
