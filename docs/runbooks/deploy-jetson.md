---
id: "pollex-runbook-deploy-jetson"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-10"
owner: manu
---

# Deploy Pollex to Jetson Nano

## SSH Access

The Jetson (`kubelab-jet1`) is in the KubeLab headscale mesh. Two access paths:

| Path | Alias | When to use |
|------|-------|-------------|
| Headscale VPN | `ssh jet1` | Primary — works from anywhere |
| LAN ProxyJump | `ssh jet1-lan` | Fallback — when headscale is down, local network only |

```
# ~/.ssh/config
Host jet1
  Hostname 100.64.0.8      # headscale mesh IP
  User manu
  IdentityFile ~/.ssh/id_ed25519

Host jet1-lan
  Hostname 172.16.1.4      # KubeLab LAN IP
  User manu
  IdentityFile ~/.ssh/id_ed25519
  ProxyJump rpi4-lan
```

> **If `ssh jet1` fails after reboot**: tailscale DNS may have lost its upstream.
> See [headscale-setup](headscale-setup.md#incident-jet1-dns-boot)
> for the fix (`FallbackDNS` in systemd-resolved — already applied on jet1).

### Makefile SSH target resolution

`make deploy` defaults to `JETSON_HOST=jet1` and automatically falls back to `jet1-lan`
with a warning if jet1 is unreachable (3s probe). Override manually if needed:

```sh
make deploy                          # jet1 (auto-fallback to jet1-lan)
make deploy JETSON_HOST=jet1-lan     # force LAN route
```

## Pre-Flight Checklist

Before deploying, verify:

- [ ] All tests pass: `make test` (expect 75+ tests, 0 failures)
- [ ] SSH works: `ssh jetson-home`
- [ ] llama-server active: `ssh jetson-home 'systemctl is-active llama-server'`
- [ ] Sufficient disk space: `ssh jetson-home 'df -h /usr/local/bin'` (need ~10MB)

## First-Time Setup

```sh
make deploy-init
```

This SCPs systemd service files and runs `deploy/init.sh` on the Jetson, which:
1. Installs prerequisites (`curl`, `zstd`)
2. Adds CUDA to PATH
3. Creates `/etc/pollex/` config directory
4. Installs and enables systemd services (`pollex-api`, `llama-server`)

Then continue with:

```sh
make deploy-llamacpp   # Build llama.cpp with CUDA on Jetson (~85 min)
make deploy            # Binary + config + prompt
make deploy-secrets    # API key
make deploy-tunnel     # Cloudflare Tunnel
```

See [build-llamacpp-jetson](build-llamacpp-jetson.md) for llama.cpp build details. See [setup-cloudflare-tunnel](setup-cloudflare-tunnel.md) for tunnel setup.

## Deployment Workflow

### Deploy to Production (office Jetson)

```sh
make test && make deploy JETSON_HOST=jetson-office
```

Run tests first, then deploy. All deploy targets accept `JETSON_HOST` override.

### Blue-Green Deploy (major changes)

For risky changes, deploy to the backup node first:

```sh
# 1. Deploy to staging (home = backup node)
make deploy JETSON_HOST=jetson-home

# 2. Verify via direct endpoint
curl -s https://pollex-home.mlorente.dev/api/health

# 3. If OK, deploy to production
make deploy JETSON_HOST=jetson-office

# 4. If staging failed, production was never touched
```

### Deploy to Specific Node

```sh
make deploy                                  # Default node (jetson-home)
make deploy JETSON_HOST=jetson-office        # Office node (production)
```

Internally, `make deploy`:
1. Cross-compiles for ARM64 (`GOOS=linux GOARCH=arm64`)
2. SCPs binary, config, prompt, and service file to `/tmp/` on the Jetson
3. Moves files to final locations via `sudo mv`
4. Reloads systemd and restarts `pollex-api`

## Verify

```sh
# Via direct endpoints (from any machine)
curl -s https://pollex-home.mlorente.dev/api/health
curl -s https://pollex-office.mlorente.dev/api/health

# Via SSH (on the Jetson itself)
make jetson-status    # Health check
make jetson-logs      # Tail service logs
make jetson-test      # End-to-end polish test
```

## Rollback Procedure

If a deploy causes issues:

### 1. Backup Before Deploy (recommended)

```sh
ssh jetson-home 'sudo cp /usr/local/bin/pollex /usr/local/bin/pollex.bak'
```

### 2. Restore Previous Binary

```sh
ssh jetson-home 'sudo systemctl stop pollex-api && sudo cp /usr/local/bin/pollex.bak /usr/local/bin/pollex && sudo systemctl start pollex-api'
```

### 3. Verify Rollback

```sh
make jetson-status    # Should show healthy
make jetson-logs      # Check for errors
```

### 4. Emergency: Disable Service

If the service keeps crashing:
```sh
ssh jetson-home 'sudo systemctl stop pollex-api && sudo systemctl disable pollex-api'
```

Then investigate logs: `ssh jetson-home 'journalctl -u pollex-api -n 100 --no-pager'`

## llama-server (GPU Inference)

The Jetson uses `llama-server` (llama.cpp compiled with CUDA 10.2) for GPU-accelerated inference. See [004-llamacpp-gpu-acceleration](../adr/004-llamacpp-gpu-acceleration.md) for the decision rationale.

### First-Time llama-server Setup

```sh
make deploy-llamacpp    # ~85 min build on Jetson
```

This compiles llama.cpp with CUDA 10.2, downloads the model, and installs the systemd service. See [build-llamacpp-jetson](build-llamacpp-jetson.md) for details and troubleshooting.

### Verify llama-server

```sh
ssh jetson-home 'systemctl is-active llama-server'
ssh jetson-home 'curl -s http://localhost:8080/health'
```

## Cloudflare Tunnel (Remote Access)

The API is exposed publicly via Cloudflare Tunnel at `https://pollex.mlorente.dev`. See [setup-cloudflare-tunnel](setup-cloudflare-tunnel.md) for first-time setup.

### Verify Remote Access

```sh
# Health check (no auth required)
curl -s https://pollex.mlorente.dev/api/health

# Polish (requires API key)
curl -s -X POST https://pollex.mlorente.dev/api/polish \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"text":"test","model_id":"qwen2.5-1.5b-gpu"}'
```

### Service Management

```sh
make jetson-tunnel-status   # Check tunnel status
make jetson-tunnel-logs     # Tail tunnel logs
make jetson-tunnel-start    # Start tunnel
```

### API Key Management

The API key is managed through the dotfiles secret system (age-encrypted):

```
dotfiles/sensitive/pollex.api-key.secret.age  → encrypted (committed)
dotfiles/sensitive/env-mapping.conf           → POLLEX_API_KEY=pollex.api-key
```

**Deploy key to Jetson:**

```sh
make deploy-secrets
```

This reads `$POLLEX_API_KEY` from the shell environment (loaded by dotfiles on login) and writes `/etc/pollex/secrets.env` on the Jetson with permissions `600`.

**Rotate key:**

```sh
# 1. Generate new key
openssl rand -hex 32

# 2. Update in dotfiles
secrets_rotate POLLEX_API_KEY

# 3. Deploy to Jetson
make deploy-secrets

# 4. Update the key in the browser extension Settings
```

## Benchmarking

After deploy, run benchmarks to verify performance:

```sh
make bench-jetson    # Via Cloudflare Tunnel (needs $POLLEX_API_KEY)
```

This sends 5 sample texts (tiny to max) through the API and prints a results table with elapsed times.

## Troubleshooting

- See [Jetson Memory Issues](../troubleshooting/jetson-memory.md)
- See [build-llamacpp-jetson](build-llamacpp-jetson.md) for build and llama-server issues
- See [setup-cloudflare-tunnel](setup-cloudflare-tunnel.md) for tunnel and remote access issues
- API logs: `make jetson-logs` or `ssh jetson-home 'journalctl -u pollex-api -n 50'`
- llama-server logs: `ssh jetson-home 'journalctl -u llama-server -n 50'`
- cloudflared logs: `make jetson-tunnel-logs` or `ssh jetson-home 'journalctl -u cloudflared -n 50'`
- See [setup-cloudflare-ssh](setup-cloudflare-ssh.md) for SSH tunnel issues (office Jetson)
- See [setup-wifi-jetson](setup-wifi-jetson.md) for WiFi dongle issues

## Endpoint Architecture (Blue-Green)

Three DNS endpoints:

| Endpoint | Purpose | Use |
|----------|---------|-----|
| `pollex.mlorente.dev` | Production | Extension, end users |
| `pollex-home.mlorente.dev` | Direct to home node | Monitoring, debugging, pre-cutover verification |
| `pollex-office.mlorente.dev` | Direct to office node | Monitoring, debugging, pre-cutover verification |

Verify both nodes independently:

```sh
curl -s https://pollex-home.mlorente.dev/api/health
curl -s https://pollex-office.mlorente.dev/api/health
```

## DNS Cutover (Switching Production Node)

To switch `pollex.mlorente.dev` between Jetsons:

### Move Production to Office Jetson

1. Verify office Jetson is healthy via direct endpoint: `curl -s https://pollex-office.mlorente.dev/api/health`
2. Cutover DNS (from any machine with tunnel credentials):
   ```sh
   cloudflared tunnel route dns --overwrite-dns pollex-office pollex.mlorente.dev
   ```
   > **`--overwrite-dns` required** — without it, the command fails because the CNAME already exists pointing to the other tunnel.
3. Wait ~30s for DNS propagation
4. Verify production now serves office: `curl -s https://pollex.mlorente.dev/api/health` (check `version` field)
5. Do NOT stop home tunnel — it serves `pollex-home.mlorente.dev` for monitoring

### Rollback to Home Jetson (< 1 min)

1. Cutover DNS back:
   ```sh
   cloudflared tunnel route dns --overwrite-dns pollex pollex.mlorente.dev
   ```
2. Wait ~30s
3. Verify: `curl -s https://pollex.mlorente.dev/api/health` (version should show home build)

## Client Prerequisites (Office Jetson)

To deploy or SSH to the office Jetson, `cloudflared` must be installed on the dev machine:

| Platform | Install |
|----------|---------|
| Windows | `winget install Cloudflare.cloudflared` |
| Linux (deb) | `sudo dpkg -i cloudflared-linux-amd64.deb` ([releases](https://github.com/cloudflare/cloudflared/releases)) |
| macOS | `brew install cloudflared` |

See [setup-cloudflare-ssh](setup-cloudflare-ssh.md) for full client setup including SSH config.

### Dev Environment Prerequisites (WSL/Linux)

To run `make deploy` from WSL or Linux, you need:

1. **Go 1.26+** (for cross-compilation):
   ```sh
   wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
   echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.zshrc   # or ~/.bashrc
   source ~/.zshrc
   ```
   > **WSL gotcha:** `sudo apt install golang-go` installs Go 1.22 on Ubuntu — too old. Must install manually from go.dev. Also ensure `/usr/local/go/bin` is **before** `/usr/bin` in PATH, or the system Go 1.22 takes precedence.

2. **cloudflared** (for SSH to office Jetson):
   ```sh
   curl -L -o /tmp/cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
   sudo dpkg -i /tmp/cloudflared.deb
   ```

3. **SSH config** for `jetson-office` — see [setup-cloudflare-ssh](setup-cloudflare-ssh.md#WSL)

4. **SSH key** copied from Windows if needed:
   ```sh
   cp /mnt/c/Users/Manu/.ssh/id_ed25519* ~/.ssh/
   chmod 600 ~/.ssh/id_ed25519
   eval $(ssh-agent) && ssh-add ~/.ssh/id_ed25519
   ```

> **CRLF warning:** Files on `/mnt/c/` (Windows filesystem) may have CRLF line endings. The repo includes `.gitattributes` to force LF for `.sh` and `.service` files. If deploy scripts fail with `invalid option namepefail`, run: `sed -i 's/\r$//' deploy/scripts/*.sh deploy/systemd/*.service`
