# Pollex — TODO

## Fase 1 — Backend Core
- [ ] `go.mod` + `config.go` + tests
- [ ] `adapter.go` (interfaz) + `adapter_mock.go` (respuestas simuladas con delay configurable)
- [ ] `adapter_ollama.go` + tests (con `httptest`)
- [ ] `handler_health.go` + `handler_models.go` + `handler_polish.go` + tests
- [ ] `middleware.go` (CORS)
- [ ] `main.go` (wiring, flag `--mock` para modo desarrollo sin Ollama)
- [ ] `prompts/polish.txt`
- [ ] Verificar: `make test` pasa, `curl` contra servidor funciona (con mock y/o Ollama local)

## Fase 2 — Extension MVP
- [ ] `manifest.json`
- [ ] `popup.html` + `popup.css` (layout estático)
- [ ] `api.js` (cliente HTTP)
- [ ] `popup.js` (conectar UI al API)
- [ ] `settings.html` + `settings.js`
- [ ] Verificar: extensión carga en Chrome, flujo polish funciona end-to-end

## Fase 3 — Claude Adapter
- [ ] `adapter_claude.go` + tests
- [ ] Actualizar config y `main.go`
- [ ] Verificar: dropdown muestra modelos locales + Claude

## Fase 4 — UX Polish
- [ ] Timer en vivo durante inference
- [ ] Botón cancelar
- [ ] Agrupación en dropdown (Local / Cloud)
- [ ] Estilos de error

## Fase 5 — Deploy
- [ ] Archivos de deploy (systemd, scripts)
- [ ] `Makefile` completo con todos los targets
- [ ] `make deploy-setup` en Jetson (primera vez)
- [ ] `make deploy` + `make jetson-status` → test E2E en LAN
