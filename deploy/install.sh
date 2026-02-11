#!/bin/bash
set -euo pipefail

echo "=== Pollex first-time setup on Jetson ==="

# 1. Install Ollama (if not present)
if ! command -v ollama &>/dev/null; then
  echo "Installing Ollama..."
  curl -fsSL https://ollama.com/install.sh | sh
else
  echo "Ollama already installed: $(ollama --version)"
fi

# 2. Start Ollama service
echo "Enabling Ollama service..."
sudo systemctl enable ollama
sudo systemctl start ollama
sleep 3

# 3. Pull default model
echo "Pulling qwen2.5:1.5b..."
ollama pull qwen2.5:1.5b

# 4. Create config directory
echo "Setting up /etc/pollex/..."
sudo mkdir -p /etc/pollex
sudo chown manu:manu /etc/pollex

# 5. Install systemd service
echo "Installing pollex-api service..."
sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service
sudo systemctl daemon-reload
sudo systemctl enable pollex-api

echo "=== Setup complete ==="
echo "Run 'make deploy' from your dev machine to deploy the binary."
