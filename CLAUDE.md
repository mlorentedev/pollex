# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Pollex is a text polishing tool: Go API backend + Chrome browser extension + llama.cpp GPU inference on Jetson Nano 4GB. Users paste text in the extension popup, select an LLM model, and get polished English back. Remote access via Cloudflare Tunnel with API key auth.

Vault: `~/Projects/knowledge/10_projects/pollex/`

## Commands

```bash
make dev                # Start API with mock adapter on :8090 (no LLM needed)
make test               # Run all tests with race detector: go test -v -race ./...
make lint               # go vet + gofmt check
make build              # Build for current platform → dist/pollex
make build-arm64        # Cross-compile for Jetson → dist/pollex-arm64
make bench              # Run benchmark against local API
make bench-jetson       # Run benchmark against Jetson via Cloudflare Tunnel
make docker-build       # Build Docker image (alpine:3.21, 24.7MB)
make docker-dev         # Start pollex in Docker (mock mode) on :8090
make docker-down        # Stop pollex Docker container
make monitoring-up      # Start Prometheus + Alertmanager + Grafana (needs make dev running)
make monitoring-down    # Stop monitoring stack
make monitoring-validate # Validate Prometheus rules and config syntax
make deploy             # Build ARM64 + SCP to Jetson + restart service (JETSON_HOST override)
make deploy-secrets     # Deploy API key from dotfiles env to Jetson
make deploy-tunnel      # Setup Cloudflare Tunnel on Jetson (interactive)
make deploy-init        # First-time Jetson setup (packages, CUDA, dirs, systemd)
```

Run a single test: `go test -v -race -run TestHandlePolish ./internal/handler/...`

**Important:** All `go run`/`go test` in Bash tool need `source ~/.zshrc` first (Go 1.26 path).

## Architecture

Module: `github.com/mlorentedev/pollex`. Dependencies: stdlib `net/http` + `gopkg.in/yaml.v3` + `prometheus/client_golang`.

- `LLMAdapter` interface: `Name()`, `Polish(ctx, text, systemPrompt)`, `Available()`
- Adapters: MockAdapter (dev), OllamaAdapter (legacy), ClaudeAdapter (API), LlamaCppAdapter (primary GPU)
- Middleware chain (order matters): CORS → RequestID → Logging → Metrics → APIKey → RateLimit → MaxBytes(64KB) → Timeout(120s) → mux
- `server.SetupMux(adapters, models, systemPrompt, apiKey, version)` — extracted for `httptest.NewServer` testability
- Config: YAML file + `POLLEX_*` env var overrides. `POLLEX_API_KEY` → `/etc/pollex/secrets.env` on Jetson

### Package Layout

```
cmd/pollex/         # Entry point — thin composition root (flags, config, wiring, shutdown)
cmd/benchmark/      # Benchmark CLI tool (samples + runner)
internal/
  adapter/          # LLMAdapter interface + concrete implementations
  config/           # YAML config + env var overrides (POLLEX_* prefix)
  handler/          # HTTP handlers (health, models, polish) + response helpers
  metrics/          # Prometheus metric declarations (promauto, default registry)
  middleware/       # CORS, RequestID, Logging, Metrics, RateLimit, APIKey, MaxBytes, Chain
  server/           # SetupMux wires handlers + middleware (testable via httptest)
deploy/
  systemd/          # pollex-api, llama-server, cloudflared, jetson-clocks services
  scripts/          # init.sh, build-llamacpp.sh, setup-cloudflared.sh
  prometheus/       # alerts.yml, prometheus.yml, prometheus-local.yml, alertmanager.yml
  grafana/          # pollex-dashboard.json, provisioning configs
  loadtest/         # k6 load test scripts (normal, burst, jetson, soak)
  config.yaml       # Production config (deployed to Jetson)
```

## Infrastructure

Single Jetson Nano 4GB (`kubelab-jet1`, headscale `100.64.0.8`, LAN `172.16.1.4`).

| Access | Command | When |
|--------|---------|------|
| SSH primary | `ssh jet1` | Always (headscale mesh) |
| SSH fallback | `ssh jet1-lan` | When headscale is down (local network) |
| Public API | `https://pollex.mlorente.dev` | Cloudflare Tunnel (extension, benchmarks) |

`make deploy` defaults to `JETSON_HOST=jet1`, auto-falls back to `jet1-lan` with warning.

## Critical Gotchas

- **Go path**: `source ~/.zshrc` required in Bash tool — Go 1.26 not in default PATH
- **JetPack systemd**: No `User=` directive in custom service files (fails with `No such process`). No `ProtectSystem=strict` either. Run as root.
- **Cloudflare Tunnel protocol**: Corporate/office networks block QUIC (UDP 443). Use `protocol: http2` in `~/.cloudflared/config.yml` if tunnel fails to connect.
- **jet1 DNS on boot**: tailscale uses systemd-resolved; if it logs out, DNS circular dependency prevents reconnect. Fixed via `/etc/systemd/resolved.conf.d/fallback-dns.conf` (`FallbackDNS=1.1.1.1`). Already applied.
- **Rate limit bypass**: Authenticated requests (`X-API-Key`) bypass rate limiting — by design. APIKey middleware is before RateLimit in the chain.
- **SCP to protected paths**: SCP to `/tmp/` first, then `sudo mv` via SSH. Can't SCP directly to `/usr/local/bin` or `/etc/pollex/`.

## Vault

`~/Projects/knowledge/10_projects/pollex/` — architecture, ADRs, runbooks, benchmarks.

Key files: `_index.md` (overview + status), `architecture.md` (diagrams), `extension.md`,
`02-runbooks/deploy-jetson.md`, `02-runbooks/cicd.md`, `01-adrs/` (ADR-001–008).
