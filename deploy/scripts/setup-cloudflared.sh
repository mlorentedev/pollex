#!/bin/bash
# setup-cloudflared.sh — Install and configure Cloudflare Tunnel on Jetson Nano (ARM64)
# Idempotent: safe to run multiple times.
# Usage: bash setup-cloudflared.sh
set -euo pipefail

TUNNEL_NAME="pollex"
LOCAL_PORT=8090

echo "=== Cloudflare Tunnel Setup for Pollex ==="

# 1. Install cloudflared (ARM64)
if command -v cloudflared &>/dev/null; then
  echo "[ok] cloudflared already installed: $(cloudflared --version)"
else
  echo "[*] Installing cloudflared (ARM64)..."
  curl -L -o /tmp/cloudflared https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
  sudo install -m 755 /tmp/cloudflared /usr/local/bin/cloudflared
  rm /tmp/cloudflared
  echo "[ok] cloudflared installed: $(cloudflared --version)"
fi

# 2. Authenticate (interactive — opens browser or prints URL)
if [ -f "$HOME/.cloudflared/cert.pem" ]; then
  echo "[ok] Already authenticated (cert.pem exists)"
else
  echo "[*] Authenticating with Cloudflare..."
  echo "    A browser window will open (or copy the URL if headless)."
  cloudflared tunnel login
fi

# 3. Create tunnel (idempotent)
if cloudflared tunnel list | grep -q "$TUNNEL_NAME"; then
  echo "[ok] Tunnel '$TUNNEL_NAME' already exists"
else
  echo "[*] Creating tunnel '$TUNNEL_NAME'..."
  cloudflared tunnel create "$TUNNEL_NAME"
fi

# 4. Write tunnel config
TUNNEL_ID=$(cloudflared tunnel list -o json | python3 -c "
import json, sys
tunnels = json.load(sys.stdin)
for t in tunnels:
    if t['name'] == '$TUNNEL_NAME':
        print(t['id'])
        break
")

CONFIG_DIR="$HOME/.cloudflared"
CONFIG_FILE="$CONFIG_DIR/config.yml"

cat > "$CONFIG_FILE" <<EOF
tunnel: $TUNNEL_ID
credentials-file: $CONFIG_DIR/$TUNNEL_ID.json

ingress:
  - hostname: pollex.mlorente.dev
    service: http://localhost:$LOCAL_PORT
  - service: http_status:404
EOF

echo "[ok] Config written to $CONFIG_FILE"

# 5. Install systemd service
if [ -f /etc/systemd/system/cloudflared.service ]; then
  echo "[ok] systemd service already installed"
else
  echo "[*] Installing systemd service..."
  sudo cp /tmp/cloudflared.service /etc/systemd/system/cloudflared.service
  sudo systemctl daemon-reload
  sudo systemctl enable cloudflared
  echo "[ok] cloudflared.service enabled"
fi

echo ""
echo "=== MANUAL STEP REQUIRED ==="
echo "Add a CNAME DNS record in Cloudflare dashboard:"
echo "  Name:    pollex"
echo "  Target:  $TUNNEL_ID.cfargotunnel.com"
echo "  Proxy:   ON (orange cloud)"
echo ""
echo "Then start the tunnel:"
echo "  sudo systemctl start cloudflared"
echo "  sudo systemctl status cloudflared"
echo ""
echo "Verify:"
echo "  curl https://pollex.mlorente.dev/api/health"
