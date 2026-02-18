#!/bin/bash
# setup-cloudflared.sh — Install and configure Cloudflare Tunnel on Jetson Nano (ARM64)
# Idempotent: safe to run multiple times.
#
# Usage:
#   bash setup-cloudflared.sh                          # Home Jetson (defaults)
#   TUNNEL_NAME=pollex-office \
#   SSH_HOSTNAME=ssh-pollex.mlorente.dev \
#   TUNNEL_PROTOCOL=http2 \
#     bash setup-cloudflared.sh                        # Office Jetson
#
# Environment variables:
#   TUNNEL_NAME      Tunnel name (default: pollex)
#   API_HOSTNAME     API hostname (default: pollex.mlorente.dev)
#   SSH_HOSTNAME     SSH hostname (optional — enables SSH ingress)
#   TUNNEL_PROTOCOL  Tunnel protocol (optional — set to "http2" for restrictive firewalls)
set -euo pipefail

TUNNEL_NAME="${TUNNEL_NAME:-pollex}"
API_HOSTNAME="${API_HOSTNAME:-pollex.mlorente.dev}"
SSH_HOSTNAME="${SSH_HOSTNAME:-}"
TUNNEL_PROTOCOL="${TUNNEL_PROTOCOL:-}"
LOCAL_PORT=8090

echo "=== Cloudflare Tunnel Setup for Pollex ==="
echo "    Tunnel:   $TUNNEL_NAME"
echo "    API host: $API_HOSTNAME"
[ -n "$SSH_HOSTNAME" ] && echo "    SSH host: $SSH_HOSTNAME"
[ -n "$TUNNEL_PROTOCOL" ] && echo "    Protocol: $TUNNEL_PROTOCOL"

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

{
  echo "tunnel: $TUNNEL_ID"
  echo "credentials-file: $CONFIG_DIR/$TUNNEL_ID.json"
  [ -n "$TUNNEL_PROTOCOL" ] && echo "protocol: $TUNNEL_PROTOCOL"
  echo ""
  echo "ingress:"
  echo "  - hostname: $API_HOSTNAME"
  echo "    service: http://localhost:$LOCAL_PORT"
  [ -n "$SSH_HOSTNAME" ] && echo "  - hostname: $SSH_HOSTNAME" && echo "    service: ssh://localhost:22"
  echo "  - service: http_status:404"
} > "$CONFIG_FILE"

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
echo "Create DNS records via CLI or Cloudflare dashboard:"
echo "  cloudflared tunnel route dns $TUNNEL_NAME $API_HOSTNAME"
[ -n "$SSH_HOSTNAME" ] && echo "  cloudflared tunnel route dns $TUNNEL_NAME $SSH_HOSTNAME"
echo ""
echo "Then start the tunnel:"
echo "  sudo systemctl start cloudflared"
echo "  sudo systemctl status cloudflared"
echo ""
echo "Verify:"
echo "  curl https://$API_HOSTNAME/api/health"
