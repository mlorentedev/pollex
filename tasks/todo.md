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
- [ ] Verificar: extensión carga en Chrome, flujo polish funciona end-to-end

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
- [ ] `make deploy-llamacpp` — builds llama-server on Jetson (~85 min)
- [ ] `ssh nvidia 'systemctl is-active llama-server'` — service running
- [ ] `make deploy` — deploy new binary with LlamaCpp adapter
- [ ] Compare: Ollama/CPU vs llama-server/GPU tokens/s

## Phase 10 — Remote Access & Chrome Web Store

Expose Pollex API remotely and publish extension for easy install on any PC.

### 10.1 — Remote Tunnel (Jetson → Internet)
- [ ] Set up Cloudflare Tunnel or WireGuard (Jetson → VPS relay) — no router access needed
- [ ] HTTPS termination + auth (API key or basic auth to prevent open access)
- [ ] Test latency from outside LAN

### 10.2 — Chrome Web Store Publishing
- [ ] Chrome Developer account ($5 one-time)
- [ ] Default API URL: empty (force user to configure in settings)
- [ ] Privacy policy (required by CWS)
- [ ] Screenshots + description for store listing
- [ ] Submit for review + publish
