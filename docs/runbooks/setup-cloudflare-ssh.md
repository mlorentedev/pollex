---
id: "pollex-runbook-cloudflare-ssh"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-17"
owner: manu
---

# Setup Cloudflare Tunnel for SSH Access

## Overview

> **Status (Feb 2026)**: The office Jetson (`jetson-office`) was retired. SSH via Cloudflare
> Tunnel is **no longer in active use**. The home Jetson is now `kubelab-jet1` in the
> KubeLab headscale mesh — SSH uses `ssh jet1` (headscale) or `ssh jet1-lan` (LAN).
> This runbook is kept as reference in case office Jetson is re-deployed.

Remote SSH access to a Jetson Nano through Cloudflare Tunnel. Useful when the Jetson is behind a restrictive firewall, corporate WiFi with client isolation, or any network where direct SSH is not possible.

Requires `cloudflared` installed on **both** the Jetson (server) and the dev machine (client).

See [setup-cloudflare-tunnel](setup-cloudflare-tunnel.md) for the API tunnel setup. See [005-cloudflare-tunnel-public-access](../adr/005-cloudflare-tunnel-public-access.md) for the decision rationale.

## Architecture

```
Dev Machine                  Cloudflare Edge              Jetson Nano
┌──────────┐                ┌──────────────┐            ┌──────────────┐
│ ssh cmd  │──ProxyCommand──│ ssh-pollex.  │──tunnel──▶│ sshd :22     │
│ cloudflared│               │ mlorente.dev │            │ cloudflared  │
└──────────┘                └──────────────┘            └──────────────┘
```

All traffic flows outbound over HTTPS (port 443). No inbound ports needed on the Jetson's network.

## Server Setup (Jetson)

### 1. Install cloudflared

```sh
curl -L -o /tmp/cloudflared https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
sudo install -m 755 /tmp/cloudflared /usr/local/bin/cloudflared
cloudflared --version
```

> **Note:** JetPack base image lacks `curl`. Install first: `sudo apt update && sudo apt install -y curl`. Alternatively use `wget -O`.

### 2. Authenticate with Cloudflare

```sh
cloudflared tunnel login
```

This prints a URL — open it on your phone/laptop browser, select the `mlorente.dev` zone, and authorize. A `cert.pem` file is saved to `~/.cloudflared/`.

> **Headless Jetson:** The Jetson has no browser. Copy the URL and open it on another device. Authorization completes remotely.

### 3. Create the Tunnel

```sh
cloudflared tunnel create pollex-office
# Note the tunnel UUID in the output
```

### 4. Write Tunnel Config

```sh
cat > ~/.cloudflared/config.yml << EOF
tunnel: <TUNNEL_UUID>
credentials-file: /home/manu/.cloudflared/<TUNNEL_UUID>.json
protocol: http2

ingress:
  - hostname: ssh-pollex.mlorente.dev
    service: ssh://localhost:22
  - service: http_status:404
EOF
```

> **Critical: `protocol: http2`** — Corporate/office firewalls often block QUIC (UDP 443). Without this setting, cloudflared defaults to QUIC and fails with `failed to dial a quic connection, network is unreachable`. Forcing HTTP/2 (TCP 443) resolves this.

### 5. Create DNS Record

```sh
cloudflared tunnel route dns pollex-office ssh-pollex.mlorente.dev
# Creates CNAME: ssh-pollex → <UUID>.cfargotunnel.com
```

### 6. Install systemd Service

```sh
sudo tee /etc/systemd/system/cloudflared.service << 'EOF'
[Unit]
Description=Cloudflare Tunnel (pollex-office)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudflared tunnel --config /home/manu/.cloudflared/config.yml run
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now cloudflared
sudo systemctl status cloudflared
```

> **Note: No `User=manu` directive.** On JetPack 4.6 (Ubuntu 18.04), systemd sometimes fails with `failed to determine user credentials: No such process` when using `User=` in custom service files. Running as root with the config pointing to manu's home directory works reliably. The `ProtectSystem=strict` and `ProtectHome=read-only` hardening directives were also removed for the same compatibility reason.

### Verify

```sh
sudo systemctl status cloudflared
# Should show: active (running)

journalctl -u cloudflared -n 10 --no-pager
# Should show: "Registered tunnel connection" (4 connections)
```

## Client Setup

### Windows

```powershell
winget install Cloudflare.cloudflared
# Restart terminal for PATH to update
cloudflared --version
```

Add to `%USERPROFILE%\.ssh\config`:

```
Host jetson-office
    HostName ssh-pollex.mlorente.dev
    User manu
    ProxyCommand cloudflared access ssh --hostname %h
    IdentityFile ~/.ssh/id_ed25519
    ControlMaster auto
    ControlPath /tmp/ssh-%r@%h:%p
    ControlPersist 10m
```

> **SSH multiplexing (`ControlMaster`)** is critical for Cloudflare Tunnel performance. Each new SSH/SCP connection through the tunnel takes 2-5 seconds to negotiate. With multiplexing, the first connection pays this cost but all subsequent connections reuse it instantly. A `make deploy` with 5 SCP calls goes from ~25s to ~8s.

### Linux (Debian/Ubuntu)

```sh
curl -L -o /tmp/cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i /tmp/cloudflared.deb
cloudflared --version
```

Add to `~/.ssh/config`:

```
Host jetson-office
    HostName ssh-pollex.mlorente.dev
    User manu
    ProxyCommand cloudflared access ssh --hostname %h
    IdentityFile ~/.ssh/id_ed25519
    ControlMaster auto
    ControlPath /tmp/ssh-%r@%h:%p
    ControlPersist 10m
```

### WSL (Windows Subsystem for Linux)

WSL has its own `~/.ssh/config` separate from Windows. You must install cloudflared and configure SSH inside WSL independently.

```sh
# Install cloudflared inside WSL
curl -L -o /tmp/cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i /tmp/cloudflared.deb

# Copy SSH key from Windows if needed
cp /mnt/c/Users/Manu/.ssh/id_ed25519* ~/.ssh/
chmod 600 ~/.ssh/id_ed25519
```

Same SSH config as Linux (add `Host jetson-office` to `~/.ssh/config` inside WSL).

> **Common mistake:** Having SSH config only in Windows but running `make deploy` from WSL. Both environments need their own cloudflared + SSH config.

### macOS

```sh
brew install cloudflared
```

Same SSH config as Linux.

### Test

```sh
ssh jetson-office
# First connection may open a browser for Cloudflare Access auth (if configured)

ssh jetson-office 'hostname'
# Expected: jetson-office
```

### Copy SSH Key (passwordless)

On Linux/macOS:

```sh
ssh-copy-id jetson-office
```

On Windows (`ssh-copy-id` not available):

```powershell
type %USERPROFILE%\.ssh\id_ed25519.pub | ssh jetson-office "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

## Adding API Ingress (Later)

Once Pollex API is deployed, update `~/.cloudflared/config.yml` on the Jetson to serve both SSH and API:

```yaml
tunnel: <TUNNEL_UUID>
credentials-file: /home/manu/.cloudflared/<TUNNEL_UUID>.json
protocol: http2

ingress:
  - hostname: pollex.mlorente.dev
    service: http://localhost:8090
  - hostname: ssh-pollex.mlorente.dev
    service: ssh://localhost:22
  - service: http_status:404
```

Then restart: `sudo systemctl restart cloudflared`

See [setup-cloudflare-tunnel](setup-cloudflare-tunnel.md) for the full API tunnel verification steps.

## Troubleshooting

### `websocket: bad handshake` from SSH client

The tunnel is not running on the Jetson. Check:

```sh
ssh jetson-office    # if you have another way in, or physically
sudo systemctl status cloudflared
journalctl -u cloudflared -n 30 --no-pager
```

### `failed to dial a quic connection, network is unreachable`

The network blocks QUIC (UDP 443). Add `protocol: http2` to `config.yml` and restart cloudflared.

### `failed to determine user credentials: No such process`

systemd `User=` directive issue on JetPack 4.6. Remove `User=manu` from the service file and run as root. See note in Step 6.

### SSH works but is slow to connect

First connection through Cloudflare Tunnel takes 2-5 seconds (tunnel negotiation). Subsequent commands in the same session are fast. This is normal.

### SSH fails after Jetson reboot

The Jetson needs time to: connect WiFi → start cloudflared → register tunnel. Wait 1-2 minutes after power-on before attempting SSH.

## Lessons Learned

- **`protocol: http2` is essential** on corporate/office networks — QUIC (UDP) is commonly blocked
- **`User=manu` in systemd fails** on JetPack 4.6 for custom services — run as root instead
- **Boot sequence takes 1-2 min** — WiFi must connect before cloudflared can register
- **`cloudflared` needed on client too** — no way to get clean SSH without it through Cloudflare Tunnel
- **WSL and Windows have separate SSH configs** — installing cloudflared on Windows does NOT make it available in WSL. Must install and configure independently in each environment
- **Go 1.26 in WSL** — `sudo apt install golang-go` installs Go 1.22 (too old for pollex). Must install manually from go.dev and ensure `/usr/local/go/bin` is before `/usr/bin` in PATH
- **CRLF on Windows filesystem** — files on `/mnt/c/` have CRLF line endings. Repo `.gitattributes` forces LF for deploy files, but may need `sed -i 's/\r$//'` if checkout predates the fix
- **ssh-agent needed in WSL** — SSH key with passphrase requires `eval $(ssh-agent) && ssh-add` per terminal session to avoid repeated prompts
- **`ssh-copy-id` does not exist on Windows** — use `type ... | ssh ... cat >>` workaround
- **Client isolation on office WiFi** is common — direct SSH between devices on the same network is blocked, but Cloudflare Tunnel works because it only needs outbound HTTPS
