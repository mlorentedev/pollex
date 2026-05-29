---
id: "pollex-architecture"
type: architecture
status: stable
created: "2026-02-22"
owner: manu
---

# Pollex: System Design

## Architecture Decisions (ADRs)
*   [ADR-001](../adr/001-local-llm-on-jetson-nano.md): Local LLM on Jetson Nano
*   [ADR-004](../adr/004-llamacpp-gpu-acceleration.md): llama.cpp GPU Acceleration
*   [ADR-005](../adr/005-cloudflare-tunnel-public-access.md): Cloudflare Tunnel
*   [ADR-006](../adr/006-q4-0-quantization.md): Q4_0 Quantization

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/polish` | `X-API-Key` | Polish text via selected model |
| GET | `/api/models` | `X-API-Key` | List available models |
| GET | `/api/health` | None | Rich health check |
| GET | `/metrics` | None | Prometheus metrics |

## Repo Structure
Clean package layout: `cmd/pollex` + `internal/{adapter,config,handler,middleware,server}`.

