---
id: "pollex-adr-005-cloudflare-tunnel"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-14"
owner: manu
---

# ADR-005: Cloudflare Tunnel for Public Access

## Status

Accepted (2026-02-13)

## Context

Pollex API runs on a Jetson Nano behind double NAT (192.168.2.x behind Proxmox host). No router access for port forwarding. The Chrome extension needs to reach the API from any network, not just the CubeLab LAN.

The domain `mlorente.dev` already has DNS managed by Cloudflare, with a Hetzner server behind it.

## Decision

Use **Cloudflare Tunnel** (formerly Argo Tunnel) for zero-config ingress, combined with **API key authentication** implemented in the Go middleware layer.

### Why Cloudflare Tunnel

| Alternative | Why Not |
|-------------|---------|
| Port forwarding | No router access (double NAT) |
| WireGuard VPN | Requires relay VPS, more infrastructure |
| ngrok | Paid for custom domain, no systemd integration |
| Cloudflare Access (OIDC) | Overkill for single-user tools, extension can't do OAuth redirects |
| Tailscale | Extension can't join tailnet |

Cloudflare Tunnel is the only option that:
- Works behind double NAT without any router config
- Provides HTTPS with a custom subdomain
- Runs as a systemd service (reliable, auto-restart)
- Free tier is sufficient

### Why API Key (not Cloudflare Access)

The authentication is implemented as Go middleware (`X-API-Key` header) rather than Cloudflare Access because:

1. **Extension compatibility** — browser extensions can set custom headers but can't handle OAuth redirects in popup windows
2. **Portability** — auth works without Cloudflare (local dev, direct LAN access)
3. **Simplicity** — no external dependency for auth, `crypto/subtle.ConstantTimeCompare` is sufficient
4. **Rotation** — managed via dotfiles → `/etc/pollex/secrets.env` → `POLLEX_API_KEY` env var

## Implementation

### Middleware Chain

```
CORS → RequestID → Logging → RateLimit → APIKey → MaxBytes → Timeout → mux
```

- Empty key = pass-through (backward compatible for local dev)
- `/api/health` exempt (monitoring always accessible)
- `crypto/subtle.ConstantTimeCompare` prevents timing attacks

### Secret Management

- Key stored in `/etc/pollex/secrets.env` as `POLLEX_API_KEY=...`
- File managed by dotfiles project (not in pollex repo)
- systemd `EnvironmentFile=-/etc/pollex/secrets.env` (dash prefix = optional, no failure if missing)
- `.gitignore` includes `secrets.env`

### Tunnel Architecture

```
Extension (any network)
  → HTTPS → pollex.mlorente.dev
  → Cloudflare Edge
  → Cloudflare Tunnel (encrypted)
  → cloudflared on Jetson (localhost:8090)
  → pollex-api
  → llama-server (localhost:8080)
```

## Consequences

### Positive

- API accessible from any network without router changes
- HTTPS with valid certificate (Cloudflare-managed)
- Extension works identically on LAN and remote
- Auth is backward compatible (empty key = disabled)
- Minimal infrastructure (no relay VPS, no VPN)

### Negative

- Cloudflare dependency for remote access (local LAN still works without it)
- API key is a shared secret (acceptable for single-user tools)
- Latency increase for remote requests (Cloudflare edge hop)
- `cloudflared` binary needs updating independently (not in apt on JetPack)

### Risks

- Cloudflare Tunnel free tier could change → fallback to WireGuard relay
- API key leak → rotate via dotfiles, restart service
