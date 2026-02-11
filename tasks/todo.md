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
- [ ] `make deploy-setup` en Jetson (primera vez) — requiere hardware
- [ ] `make deploy` + `make jetson-status` → test E2E en LAN — requiere hardware
