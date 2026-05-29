---
id: "pollex-runbook-cloudflare-tunnel"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-14"
owner: manu
---

# Setup Cloudflare Tunnel on Jetson

Exposes the Pollex API at `https://pollex.mlorente.dev` via Cloudflare Tunnel. No router config needed (works behind double NAT).

See [005-cloudflare-tunnel-public-access](../adr/005-cloudflare-tunnel-public-access.md) for the decision rationale.

## Prerequisites

- Cloudflare account with `mlorente.dev` DNS managed by CF
- SSH access to Jetson: `ssh jetson-home`
- Pollex API running: `systemctl is-active pollex-api`

## Automated Setup

```sh
make deploy-tunnel
```

This SCPs `setup-cloudflared.sh` and `cloudflared.service` to the Jetson, then runs the setup script. The script is idempotent (safe to re-run).

The script will:
1. Install `cloudflared` binary (ARM64)
2. Authenticate with Cloudflare (interactive — prints a URL to visit)
3. Create tunnel named `pollex`
4. Write `~/.cloudflared/config.yml`
5. Install systemd service

## Manual DNS Step

After the script completes, add a CNAME record in the Cloudflare dashboard:

| Type | Name | Target | Proxy |
|------|------|--------|-------|
| CNAME | pollex | `<TUNNEL_ID>.cfargotunnel.com` | ON |

The tunnel ID is printed by the setup script.

## API Key Configuration

The API key is managed through the dotfiles secret system (age-encrypted). See [deploy-jetson](deploy-jetson.md#API Key Management) for full details.

```sh
# First time: add secret to dotfiles
secrets_add POLLEX_API_KEY pollex.api-key

# Deploy to Jetson
make deploy-secrets

# Rotate later
secrets_rotate POLLEX_API_KEY && make deploy-secrets
```

The `pollex-api.service` reads the key from `/etc/pollex/secrets.env` via `EnvironmentFile`.

## Verify

```sh
# From any machine with internet access
curl -s https://pollex.mlorente.dev/api/health

# Polish with API key
curl -s -X POST https://pollex.mlorente.dev/api/polish \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"text":"this is a test","model_id":"qwen2.5-1.5b-gpu"}'
```

## Service Management

```sh
ssh jetson-home 'sudo systemctl status cloudflared'
ssh jetson-home 'sudo journalctl -u cloudflared -f'
ssh jetson-home 'sudo systemctl restart cloudflared'
```

## Multi-Node Setup

For multi-node deployment (home + office Jetsons), each node has its own tunnel:

### Endpoint Architecture (Blue-Green)

Three DNS endpoints following SRE blue-green deployment pattern:

| Endpoint | Purpose | Points to |
|----------|---------|-----------|
| `pollex.mlorente.dev` | Production (user-facing) | Active tunnel (home or office, DNS cutover) |
| `pollex-home.mlorente.dev` | Direct to home node | Always → tunnel `pollex` |
| `pollex-office.mlorente.dev` | Direct to office node | Always → tunnel `pollex-office` |

The production endpoint serves the extension. Direct endpoints are for monitoring (Prometheus scrapes both independently), pre-cutover verification, and debugging.

### Tunnel Configuration Per Node

| Node | Tunnel Name | Ingress Hostnames | Protocol |
|------|------------|-------------------|----------|
| Home (`jetson-home`) | `pollex` | `pollex.mlorente.dev`, `pollex-home.mlorente.dev` | QUIC (default) |
| Office (`jetson-office`) | `pollex-office` | `pollex.mlorente.dev`, `pollex-office.mlorente.dev`, `ssh-pollex.mlorente.dev` | HTTP/2 (QUIC blocked) |

Only one tunnel can serve `pollex.mlorente.dev` at a time — controlled by the CNAME record in Cloudflare DNS. Both tunnels list it in their ingress so either can serve it after DNS cutover. See [deploy-jetson](deploy-jetson.md#DNS Cutover) for switching procedure.

The office tunnel also serves SSH via a multi-ingress config. See [setup-cloudflare-ssh](setup-cloudflare-ssh.md) for details.

> **Critical: `protocol: http2`** — Office/corporate firewalls commonly block QUIC (UDP 443). The office tunnel must use `protocol: http2` in its config. Without this, cloudflared fails with `failed to dial a quic connection, network is unreachable`.

### Verify Direct Endpoints

```sh
# Both should return health independently of which is serving production
curl -s https://pollex-home.mlorente.dev/api/health
curl -s https://pollex-office.mlorente.dev/api/health
```

## Troubleshooting

### Tunnel not connecting

```sh
# Check cloudflared logs
ssh jetson-home 'journalctl -u cloudflared -n 50 --no-pager'

# Verify config
ssh jetson-home 'cat ~/.cloudflared/config.yml'

# Test manually (reads tunnel name from config.yml)
ssh jetson-home 'cloudflared tunnel --config ~/.cloudflared/config.yml run'
```

### 401 Unauthorized from extension

- Verify API key in extension Settings matches `POLLEX_API_KEY` in secrets.env
- Health endpoint should always return 200 (exempt from auth)
- Rotate key: update dotfiles → deploy secrets.env → restart pollex-api

### DNS not resolving

- Verify CNAME record exists in Cloudflare dashboard
- Check proxy status is ON (orange cloud)
- Allow 1-2 minutes for DNS propagation

## Rollback

To disable remote access without affecting local operation:

```sh
ssh jetson-home 'sudo systemctl stop cloudflared && sudo systemctl disable cloudflared'
```

The Pollex API continues working on the LAN at `http://192.168.2.59:8090`.
