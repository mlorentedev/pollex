# Pollex — TODO

## Fase 1 — Backend Core
- [x] `go.mod` + `config.go` + tests
- [x] `adapter.go` (interfaz) + `adapter_mock.go` (respuestas simuladas con delay configurable)
- [x] `adapter_ollama.go` + tests (con `httptest`)
- [x] `handler_health.go` + `handler_models.go` + `handler_polish.go` + tests
- [x] `middleware.go` (CORS + logging + timeout)
- [x] `main.go` (wiring, flag `--mock` para modo desarrollo sin Ollama)
- [x] `prompts/polish.txt`
- [x] Verificar: `make test` pasa (26 tests), `curl` contra servidor funciona (mock)

## Fase 2 — Extension MVP
- [x] `manifest.json` (Manifest V3, permisos: storage)
- [x] `popup.html` + `popup.css` (layout completo)
- [x] `api.js` (cliente HTTP con AbortController, 70s timeout)
- [x] `popup.js` (UI wiring, Ctrl+Enter, timer en vivo, cancel, copy)
- [x] `settings.html` + `settings.js` (API URL + Test Connection)
- [x] Icons (16/48/128px PNG)
- [x] Verificar: extensión carga en Chrome, flujo polish funciona end-to-end

## Fase 3 — Claude Adapter
- [x] `adapter_claude.go` + tests (7 tests, Messages API)
- [x] Actualizar config y `main.go` (auto-register when `claude_api_key` set)
- [x] Verificar: dropdown muestra modelos locales + Claude (33 tests total)

## Fase 4 — UX Polish
- [x] Timer en vivo durante inference (built into Fase 2)
- [x] Botón cancelar (built into Fase 2)
- [x] Agrupación en dropdown (Local / Cloud) — optgroup by provider
- [x] Estilos de error + cancelled state

## Fase 5 — Deploy
- [x] Archivos de deploy: `pollex-api.service`, `config.yaml`, `install.sh`, `ollama-models.sh`
- [x] `Makefile` completo con todos los targets (deploy-setup SCPs service file)
- [x] Binarios: local 10MB, ARM64 9.5MB
- [x] 33 tests passing, `go vet` clean
- [x] `make deploy-setup` en Jetson (primera vez)
- [x] `make deploy` + `make jetson-status` → servicio activo en Jetson

## Phase 6 — E2E / Integration Testing
- [x] Refactorizar `main.go`: extraer `buildAdapters()` + `setupMux()` para testabilidad
- [x] `integration_test.go` con 8 tests E2E via `httptest.NewServer`
- [x] Tests: PolishFullFlow, HealthFullFlow, ModelsFullFlow, OptionsPreflightCORS, UnknownRoute, ConcurrentPolish, ContextCancellation, AdapterErrorPropagation
- [x] Vault: ADR-002 Testing Strategy

## Phase 7 — Hardening
- [x] `requestid.go` + `requestid_test.go` (crypto/rand, 32 hex, context helpers)
- [x] `requestIDMiddleware` + logging con request ID (`[req-id] METHOD PATH STATUS DURATION`)
- [x] `maxBytesMiddleware(64KB)` + detección de `MaxBytesError` → 413 en handler_polish
- [x] `maxTextLength = 10000` validación en handler_polish → 400
- [x] `ratelimit.go` + `ratelimit_test.go` (sliding window, 10 req/min/IP, 429)
- [x] Rich health check: `/api/health` reporta status por-adaptador
- [x] Graceful shutdown: `http.Server` + `signal.Notify(SIGINT, SIGTERM)` + 10s drain
- [x] Middleware chain: CORS → requestID → logging → rateLimit → maxBytes → timeout → mux
- [x] Integration tests actualizados: RateLimit, OversizedBody, TextTooLong, X-Request-ID
- [x] Vault: ADR-003 Hardening Decisions

## Phase 8 — Documentation
- [x] Reescribir `flash-jetson.md` (runbook completo desde cero)
- [x] Actualizar `_index.md` (Go 1.26, status completo, links a ADR-002/003)
- [x] Actualizar `deploy-jetson.md` (pre-flight checklist + procedimiento de rollback)
- [x] Actualizar `jetson-memory.md` (sección de swap: file + ZRAM)

## Phase 9 — llama.cpp GPU Acceleration on Jetson Nano

Ollama uses 100% CPU on Jetson (41s/request) because it dropped CUDA 10.2 support.
Switch to llama-server (llama.cpp compiled with CUDA 10.2) for ~300-500% speedup (→ ~8-15s).

### 9.1 — LlamaCppAdapter
- [x] `backend/adapter_llamacpp.go` — OpenAI-compatible API (`/v1/chat/completions`, `/health`)
- [x] `backend/adapter_llamacpp_test.go` — 7 tests (Polish, ServerError, EmptyChoices, ContextCancel, Available, NotAvailable, Name)

### 9.2 — Config + Registration
- [x] `backend/config.go` — add `LlamaCppURL`, `LlamaCppModel` fields + env vars (`POLLEX_LLAMACPP_*`)
- [x] `backend/main.go` — register in `buildAdapters()` when URL configured (120s timeout)
- [x] `backend/handler_health.go` — add type switch case for `*LlamaCppAdapter`
- [x] Config tests updated: YAML load + env override for new fields

### 9.3 — Deploy: Compile llama.cpp on Jetson
- [x] `deploy/build-llamacpp.sh` — idempotent CUDA build script (pinned commit `23106f9`, 6 patches for CUDA 10.2)
- [x] `deploy/llama-server.service` — systemd unit (`-ngl 999 -c 2048 -t 4`, hardened)
- [x] `deploy/config.yaml` — add `llamacpp_url`, `llamacpp_model`
- [x] `Makefile` — `deploy-llamacpp` target

### 9.4 — Documentation
- [x] Vault: ADR-004 llama.cpp GPU Acceleration (decision rationale, alternatives, patches, rollback)
- [x] Vault: `runbooks/build-llamacpp-jetson.md` — full build runbook with troubleshooting
- [x] Vault: update `runbooks/deploy-jetson.md` — llama-server section + rollback
- [x] Vault: update `_index.md` — GPU acceleration status, updated stack table
- [x] Vault: update `troubleshooting/jetson-memory.md` — memory budget for llama-server

### Verification
- [x] `make test` — 62 top-level tests (97 with subtests), -race clean, go vet clean
- [x] Local smoke test: Docker `llama-server` (CPU) + pollex API → polish end-to-end OK (2.1s)
- [x] `make deploy-llamacpp` — llama-server compiled and running on Jetson
- [x] `ssh nvidia 'systemctl is-active llama-server'` — service active
- [x] `make deploy` — new binary with LlamaCpp adapter deployed
- [x] Benchmark: llama-server/GPU ~7.9s (short) vs Ollama/CPU ~41s = **~5x speedup**

## Phase 10 — Remote Access via Cloudflare Tunnel + API Key Auth

Jetson behind double NAT (no router access). Cloudflare Tunnel for zero-config ingress + API key middleware in Go.

### 10.1 — API Key Auth (Backend)
- [x] `internal/config/config.go` — add `APIKey` field + `POLLEX_API_KEY` env override
- [x] `internal/middleware/apikey.go` — `X-API-Key` header, `crypto/subtle.ConstantTimeCompare`, health exempt
- [x] `internal/middleware/apikey_test.go` — 6 subtests (disabled, valid, missing, wrong, health exempt, models requires auth)
- [x] `internal/middleware/chain.go` — APIKey before RateLimit (reordered in Phase 12.4)
- [x] `internal/middleware/cors.go` — `Access-Control-Allow-Headers: "Content-Type, X-API-Key"`
- [x] `internal/server/server.go` — `SetupMux` accepts `apiKey` parameter
- [x] `cmd/pollex/main.go` — pass `cfg.APIKey`, log auth mode

### 10.2 — Integration Tests
- [x] `internal/server/integration_test.go` — `TestIntegration_APIKeyRequired` with 5 subtests
- [x] Updated `SetupMux` calls + CORS header assertions

### 10.3 — Extension API Key Support
- [x] `extension/api.js` — `getApiKey()`, `buildHeaders()`, inject `X-API-Key` in fetchModels + fetchPolish
- [x] `extension/popup.html` — API Key password input in Settings
- [x] `extension/popup.js` — load/save `apiKey` in `chrome.storage.local`
- [x] `extension/popup.css` — `.form-hint` style

### 10.4 — Deploy Artifacts
- [x] `deploy/config.yaml` — comment for api_key via env
- [x] `deploy/pollex-api.service` — `EnvironmentFile=-/etc/pollex/secrets.env`, After=llama-server
- [x] `deploy/cloudflared.service` — systemd unit for Cloudflare Tunnel
- [x] `deploy/setup-cloudflared.sh` — idempotent setup script (install, auth, create tunnel, config, DNS hint)
- [x] `Makefile` — `deploy-cloudflared` target
- [x] `.gitignore` — `secrets.env`

### 10.5 — Documentation
- [x] Vault: ADR-005 Cloudflare Tunnel for Public Access
- [x] Vault: `runbooks/setup-cloudflare-tunnel.md`
- [x] Vault: update `_index.md` — Phase 10 status, package layout, ADR-005 link
- [x] Vault: update `runbooks/deploy-jetson.md` — cloudflared section + API key rotation

### 10.6 — Deploy & Verification
- [x] `make test` — all tests pass + 11 new subtests (6 unit + 5 integration)
- [x] `make dev` (no api_key) — backward compatible, auth disabled
- [x] `POLLEX_API_KEY=test make dev` — auth enforcement (401 without, 200 with key)
- [x] Extension: API key input in Settings, polish works with key
- [x] `make deploy` + `make deploy-secrets` — binary, service, and secrets on Jetson
- [x] `make deploy-cloudflared` — tunnel created, DNS CNAME configured
- [x] `curl https://pollex.mlorente.dev/api/health` — 200 from internet
- [x] `curl -H "X-API-Key: ..." https://pollex.mlorente.dev/api/polish` — 200, ~3s GPU inference
- [x] Extension with remote URL + API key — end-to-end OK
- [x] `extension/manifest.json` — added `host_permissions: ["<all_urls>"]` for remote access
- [x] `Makefile` — `deploy-secrets`, `tunnel-start`, `tunnel-status`, `tunnel-logs` targets
- [x] `Makefile` — `deploy` now includes service file + daemon-reload
- [x] Vault + CLAUDE.md — secrets flow documented (dotfiles → age → deploy-secrets → Jetson)

### Known Limitations (resolved)
- [x] Rate limiter: reads `Cf-Connecting-Ip` header for real client IP behind Cloudflare Tunnel

### 10.7 — Chrome Web Store Publishing — WON'T DO
CWS publishing adds cost ($5), maintenance burden (privacy policy, review process, responding to reviews), and ongoing compliance for zero practical benefit. The target audience capable of deploying a Jetson Nano with llama.cpp + Cloudflare Tunnel can load an unpacked extension. Chrome blocks .crx installs outside CWS since Chrome 75, so the only alternative is load unpacked (current approach), which works fine for a self-hosted tool.

## Phase 12 — Performance Optimization + Extension UX

### 12.1 — Extension UX
- [x] Persist textarea draft in `chrome.storage.local` (restore on popup reopen — popup closes on focus loss)
- [x] Hard character limit: reduce MAX_CHARS from 10000 to 1500 (120s timeout ÷ 68ms/char ≈ 1764 max)
- [x] Adjust estimated time warning threshold accordingly

### 12.2 — Jetson Inference Tuning
- [x] Q4_0 quantization: `qwen2.5-1.5b-instruct-q4_0.gguf` — 23% faster (3.0s vs 3.9s short text)
- [x] `--mlock` + `LimitMEMLOCK=infinity`: model locked in RAM, no paging
- [x] Headless mode: `systemctl set-default multi-user.target` (frees ~400MB RAM, effective after reboot)
- [x] A/B test `-t 2` vs `-t 4`: no difference with full GPU offload, keeping `-t 4`
- [~] Zram tuning: WON'T DO — only 29MB/2GB used, negligible overhead, not worth the complexity

### 12.3 — Model Upgrade (deferred)
- [x] **Skipped:** 3B model descartado — latencia ~2x haría textos >750 chars inutilizables (timeout 120s). 1.5B Q4_0 ya pasa los 5 quality samples. Reconsiderar solo si la calidad resulta insuficiente en uso real.

### 12.4 — Benchmark Improvements
- [x] Rate limiter: authenticated requests (X-API-Key) bypass rate limiting; APIKey moved before RateLimit in chain
- [x] Output results to JSON file (`--json results.json`)
- [x] Add warmup run (`--warmup`, discards first result per sample)

## Phase 13 — Observability & SRE Foundations

### 13.1 — Prometheus Metrics
- [x] Add `prometheus/client_golang` dependency
- [x] `GET /metrics` endpoint (exempt from API key auth + rate limit)
- [x] Metrics: `pollex_polish_duration_seconds` histogram (by model)
- [x] Metrics: `pollex_requests_total` counter (by method, path, status)
- [x] Metrics: `pollex_adapter_available` gauge (per adapter)
- [x] Metrics: `pollex_input_chars` histogram (text size distribution)
- [x] Metrics middleware in chain: CORS → RequestID → Logging → **Metrics** → APIKey → RateLimit → MaxBytes → Timeout
- [x] Integration tests: MetricsEndpoint, MetricsExemptFromAuth, MetricsAfterPolish
- [x] Unit tests: metrics middleware counter, apikey /metrics exempt

### 13.2 — Structured Logging
- [x] JSON log format via `slog.NewJSONHandler` (timestamp, level, msg, request_id, method, path, status, duration_ms)
- [x] Replace `log.Printf` with `log/slog` in middleware/logging.go
- [x] Replace `log.Printf/Fatalf/Println` with `slog.Info/Error` in cmd/pollex/main.go
- [x] Log adapter name + model in buildAdapters registration

### 13.3 — SLOs & SLIs
- [x] Define SLIs: availability (composite: API up + llamacpp available), latency (p50/p95 polish), error rate (5xx on /api/polish)
- [x] Define SLOs (7d rolling): 99% availability (100.8 min budget), p50 < 20s, p95 < 60s, error rate < 1%
- [x] Error budget policy: healthy → warning → freeze → post-mortem
- [x] Document in vault as ADR-007

### 13.4 — Alerting & Dashboard
- [x] Prometheus alerting rules (`deploy/prometheus/alerts.yml`): PollexDown, LlamaCppDown, HighLatency p50/p95, HighErrorRate, ErrorBudgetBurn
- [x] Alertmanager config template (`deploy/prometheus/alertmanager.yml`): Slack webhook routing, severity-based repeat intervals
- [x] Prometheus scrape config (`deploy/prometheus/prometheus.yml`): pollex.mlorente.dev target, 30s interval
- [x] Grafana dashboard (`deploy/grafana/pollex-dashboard.json`): SLO status row, traffic/errors, latency percentiles, adapter availability
- [x] Background adapter probe goroutine (30s interval) in `cmd/pollex/main.go` — keeps `pollex_adapter_available` gauge fresh for Prometheus
- [x] Deploy: configure scrape target in Docker Prometheus on monitoring host

## Phase 14 — Containerization

### 14.1 — Dockerfile
- [x] Multi-stage build (Go 1.26 builder → `alpine:3.21` runtime, 24.7MB image)
- [x] Non-root user (`pollex:pollex` UID/GID 1001)
- [x] Health check via `curl` in compose (alpine has curl, no separate binary needed)
- [x] Build args: `VERSION` (git describe), `VCS_REF` (commit SHA), OCI labels
- [x] Build cache: `--mount=type=cache` for go mod + go build
- [x] `.dockerignore` (extension/, deploy/, tasks/, .github/, .git/)

### 14.2 — Docker Compose (local dev)
- [x] `docker-compose.yml`: pollex-api (mock mode, port 8090)
- [x] `make docker-build`, `make docker-dev`, `make docker-down` targets
- [x] Verified: health, polish, metrics endpoints + Docker HEALTHCHECK `healthy`

## Phase 15 — IaC & Load Testing

### 15.1 — Ansible Playbook (Removed)
- [x] ~~Ansible playbook~~ — removed. Jetson Nano runs Ubuntu 18.04 (Python 3.6), incompatible with ansible-core 2.20 (requires 3.9+). deadsnakes PPA only provides up to 3.8 for bionic arm64. Deploy via `make deploy-*` targets (SSH + SCP), which are simpler and proven for single-node.

### 15.2 — Load Testing
- [x] k6 script (`deploy/loadtest/pollex.js`) with custom metrics (polish_duration_ms, errors)
- [x] Test scenarios: normal load (12 req/min, 2 min), burst (5 VUs, 25 iterations), soak (12 req/min, 30 min)
- [x] Thresholds aligned to SLOs: p50 < 20s, p95 < 60s, error rate < 1%
- [x] `make loadtest`, `make loadtest-jetson`, `make loadtest-soak` targets
- [x] Run against Jetson and document results in vault benchmarks

## Phase 11 — Performance Benchmarking + System Prompt + CI/CD

### 11.1 — System Prompt Improvement
- [x] `prompts/polish.txt` — expanded to three dimensions: grammar, coherence, conciseness
- [x] Added constraints: no AI-isms, preserve formatting/intent, output only polished text

### 11.2 — Benchmark CLI Tool
- [x] `cmd/benchmark/samples.go` — 5 realistic email samples (tiny/short/medium/long/max)
- [x] `cmd/benchmark/main.go` — standalone CLI: auto-discover models, N runs, markdown table output
- [x] Makefile: `bench`, `bench-jetson` targets

### 11.3 — CI/CD (GitHub Actions)
- [x] `.github/workflows/ci.yml` — lint + test + build (push to master, PRs)
- [x] `.github/workflows/release.yml` — goreleaser + extension zip on `v*` tags
- [x] `.goreleaser.yml` — linux/amd64 + linux/arm64 binaries, changelog from commits
- [x] Extension version synced from git tag in release workflow

### 11.4 — Extension Improvements
- [x] `extension/manifest.json` — professional fields (name, short_name, description, homepage_url, minimum_chrome_version)
- [x] `extension/popup.js` + `popup.html` + `popup.css` — single-model mode: static label instead of dropdown when only one model available

### 11.5 — Makefile Refactoring
- [x] Removed 6 obsolete Ollama targets (deploy-setup, deploy-models, dev-ollama, ollama-up/down/pull)
- [x] Removed obsolete scripts: `deploy/install.sh`, `deploy/ollama-models.sh`
- [x] Renamed for coherence: `tunnel-*` → `jetson-tunnel-*`, `deploy-cloudflared` → `deploy-tunnel`, `ext-pack` → `ext-zip`
- [x] Reorganized into 6 sections: Development, Build, Benchmark, Deploy, Jetson Remote, Utilities

### Verification
- [x] `make test` — 75+ tests pass, -race clean, go vet clean
- [x] Push to GitHub → CI workflow runs (lint + test + build)
- [x] `make bench-jetson` — baseline: ~4 tok/s on Qwen 2.5 1.5B Q4_K_M
- [x] release-please + goreleaser + extension zip automation verified

## Phase 16 — Extension: Service Worker + History

### 16.1 — Service Worker for Persistent Requests
- [x] `extension/manifest.json` — add `background.service_worker: "background.js"`
- [x] `extension/background.js` — new file: message listener (POLISH_START/CANCEL), fetchPolish via api.js, polishJob state in storage, tick interval, appendHistory
- [x] `extension/popup.js` — refactor: doPolish sends message to background, cancel sends message, storage.onChanged listener, recoverJobState on popup open, clearStaleJob on input

### 16.2 — Rolling History (7 entries)
- [x] `extension/popup.html` — history section + detail overlay
- [x] `extension/popup.css` — history list/item/detail styles
- [x] `extension/popup.js` — loadHistory, renderHistory, showHistoryDetail, formatRelativeTime, back button

### 16.3 — Color Scheme Update
- [x] Accent color changed from indigo (#6366f1) to cyan-700 (#0e7490) family
- [x] Brand SVG + PNG icons (16/48/128) regenerated to match

### 16.4 — UX Polish
- [x] Progress bar with ETA percentage + elapsed time (replaces text status)
- [x] ETA padded +15% for conservative estimate, capped at 99%
- [x] Copy button in history detail view
- [x] Clean interface on popup reopen (no stale results shown, history below for recovery)
- [x] Empty input shake animation on Polish click (no text status noise)

### 16.5 — Hardening
- [x] Input validation in background.js (type check, empty, max length 1500)
- [x] Error message truncation (200 chars max in storage)
- [x] Prompt injection defense in `prompts/polish.txt`

### 16.6 — Backend Version in Extension
- [x] `cmd/pollex/main.go` — `var version = "dev"`, injected via `-ldflags -X main.version`
- [x] `internal/handler/health.go` — `version` field in health response JSON
- [x] `internal/server/server.go` — pass version through SetupMux
- [x] `Makefile` — VERSION from `git describe --tags`, LDFLAGS for build/build-arm64
- [x] `.goreleaser.yml` — `-X main.version={{.Version}}` in ldflags
- [x] Extension Settings — shows `API vX.Y.Z` fetched from `/api/health`
