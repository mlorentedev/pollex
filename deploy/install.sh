#!/bin/bash
set -euo pipefail

echo "=== Pollex first-time setup on Jetson ==="

# 0. Install required tools (JetPack base image lacks curl)
echo "Installing prerequisites..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl zstd

# 1. Add CUDA to PATH (if not already present)
if ! grep -q '/usr/local/cuda/bin' ~/.bashrc; then
  echo "Adding CUDA to PATH..."
  echo 'export PATH=/usr/local/cuda/bin:$PATH' >> ~/.bashrc
fi

# 2. Install Ollama (if not present)
if ! command -v ollama &>/dev/null; then
  echo "Installing Ollama..."
  curl -fsSL https://ollama.com/install.sh | sh
else
  echo "Ollama already installed: $(ollama --version)"
fi

# 3. Start Ollama service
echo "Enabling Ollama service..."
sudo systemctl enable ollama
sudo systemctl start ollama
sleep 3

# 4. Pull default model
echo "Pulling qwen2.5:1.5b..."
ollama pull qwen2.5:1.5b

# 5. Create config directory
echo "Setting up /etc/pollex/..."
sudo mkdir -p /etc/pollex
sudo chown "$(whoami):$(whoami)" /etc/pollex

# 6. Install systemd service
echo "Installing pollex-api service..."
sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service
sudo systemctl daemon-reload
sudo systemctl enable pollex-api

echo "=== Setup complete ==="
echo "Run 'make deploy' from your dev machine to deploy the binary."
