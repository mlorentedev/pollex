# Pollex — Lessons Learned

## Fase 1 — Backend Core

- **Default prompt path relativo**: `../prompts/polish.txt` es relativo al directorio de ejecución (`backend/`), no al binario. Al correr `go run .` desde `backend/`, funciona. Desde otro directorio, falla.

## Phase 6 — Integration Testing

- **Extraer funciones de `main()` habilita testabilidad**: `buildAdapters()` y `setupMux()` permiten crear `httptest.NewServer` con el stack completo de middleware sin levantar el servidor real.
- **`httptest.NewServer` > `httptest.NewRecorder`** para E2E: el recorder solo prueba handlers individuales, el server prueba conexiones TCP reales, middleware chain completo, y headers de transporte.

## Phase 7 — Hardening

- **Orden del middleware importa**: CORS debe ser primero (para que OPTIONS preflight no sea bloqueado por rate limit). requestID antes de logging (para que los logs tengan el ID). Rate limit antes de maxBytes (rechazar antes de leer el body ahorra recursos).
- **`http.MaxBytesError` requiere `errors.As()`**: el error viene envuelto por `json.Decoder`, no se puede hacer type assertion directa. Usar `errors.As(err, &maxBytesErr)`.
- **Rate limiter sliding window con `[]time.Time`**: simple y efectivo para uso LAN. No necesita token bucket ni Redis para un server single-instance.
- **`signal.Notify` necesita buffer 1**: `done := make(chan os.Signal, 1)` — sin buffer, la señal se puede perder si nadie está escuchando en el momento exacto.

## Phase 8 — Documentation & Deploy

- **JetPack 4.6.6** (no 4.6.5) es la última versión soportada por Jetson Nano 4GB. JetPack 5.x+ requiere Orin o superior.
- **No hacer `apt dist-upgrade`** en Jetson — rompe los drivers CUDA y la compatibilidad con JetPack.
- **Primer boot tarda ~45 min** en SD card — no interrumpir. SD card lenta solo afecta boot/instalación, no operación normal (todo corre en RAM).
- **sshd no arranca hasta completar OEM setup** — requiere HDMI + teclado obligatoriamente.
- **JetPack base image no trae `curl`** — hay que instalarlo como prerequisito en `install.sh`.
- **CUDA no está en PATH por defecto** — hay que añadir `/usr/local/cuda/bin` a `~/.bashrc`.
- **Sudo sin password necesario** para scripts de deploy remoto (`/etc/sudoers.d/manu`).
- **SSH a Jetson requiere jump host** — el Jetson está en 192.168.2.x detrás del Proxmox. Configurar `~/.ssh/config` con `ProxyJump pve`.
- **WiFi dongles necesitan drivers** — usar Ethernet para setup inicial.
- **Makefile usa `JETSON_HOST=nvidia`** — resuelve vía SSH config, no por DNS.
- **SCP a `/usr/local/bin` falla por permisos** — SCP a `/tmp/` primero, luego `sudo mv` vía SSH. Mismo patrón para `/etc/pollex/`.
- **`zstd` necesario para Ollama** — el instalador de Ollama usa zstd para descomprimir. Añadir al `install.sh` junto con `curl`.
- **`curl` directo al Jetson no funciona** — está detrás de NAT. `jetson-status` debe hacer `ssh nvidia 'curl -s localhost:8090/...'`.
