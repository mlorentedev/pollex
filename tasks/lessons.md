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

## Phase 9 — llama.cpp GPU Acceleration

- **llama.cpp repo migró de `ggerganov` a `ggml-org`** — la imagen Docker es `ghcr.io/ggml-org/llama.cpp:server`, no `ghcr.io/ggerganov/llama.cpp:server`.
- **Probar con Docker real antes de desplegar** — un fake/mock server no valida el contrato real de la API (edge cases, headers, latencia). Usar la imagen oficial de llama-server en CPU para smoke test local.
- **Docker image es `ghcr.io/ggml-org/llama.cpp:server`** — no `ggerganov`, el repo migró a `ggml-org`.
- **CMake 3.14+ necesario** — Ubuntu 18.04 trae 3.10. Instalar binario aarch64 de Kitware: `curl | tar` a `/usr/local/`.
- **`pip3 install cmake` falla en Python 3.6** — necesita `skbuild` que no está disponible. Usar binario de Kitware.
- **`-DCMAKE_CUDA_STANDARD=14` es obligatorio** — CUDA 10.2 nvcc no soporta C++17. Sin este flag, cmake falla con "CUDA17 dialect not supported".
- **Flags completos de cmake para Jetson Nano**: `-DGGML_CUDA=ON -DCMAKE_CUDA_STANDARD=14 -DCMAKE_CUDA_STANDARD_REQUIRED=TRUE -DGGML_CPU_ARM_ARCH=armv8-a -DGGML_NATIVE=OFF`.
- **NEON stubs van en `ggml-cpu-impl.h`, NO en `ggml-cpu-quants.c`** — los macros `ggml_vld1q_s8_x4` etc. están definidos en impl.h. Inyectar stubs en quants.c no funciona porque no incluye arm_neon.h directamente y los macros resuelven antes.
- **gcc-8 on aarch64 DOES provide `vld1q_*_x2` but NOT `_x4`** — initial assumption that gcc-8.4 lacked all `_x2/_x4` was wrong. gcc-8's `arm_neon.h` includes `vld1q_s8_x2`, `vld1q_u8_x2`, `vld1q_s16_x2`. Only the `_x4` variants need stubs. llama.cpp's own polyfills in `ggml-cpu-impl.h` must be commented out to avoid "redeclared inline without 'gnu_inline' attribute" errors.
- **WMMA (fattn-wmma-f16.cu) requiere Volta+ (compute 7.0)** — Maxwell (Jetson Nano, compute 5.3) no lo soporta. Hay que vaciar el archivo dejando solo `#include "common.cuh"` para que compile.
- **`cuda_bf16.h` stub debe hacer `typedef half nv_bfloat16`** — no basta con definir `__nv_bfloat16` como struct, el código usa ambos nombres (`nv_bfloat16` y `__nv_bfloat16`). Incluir `cuda_fp16.h` y hacer typedef de ambos a `half`.
- **`<charconv>` es C++17, no disponible con nvcc C++14** — gcc-8 solo provee `<charconv>` en modo `-std=c++17`, pero nvcc 10.2 está forzado a C++14. Solución: crear un shim `charconv` con `std::from_chars` implementado sobre `strtol`/`strtof`, e inyectarlo via `-isystem` en `CMAKE_CUDA_FLAGS`.
- **No reemplazar `static constexpr` en funciones** — `sed 's/static constexpr/static const/'` blanket rompe funciones constexpr que se usan como template args (mmvq.cu, warp_reduce_sum). Solo reemplazar en líneas sin `(` (variables): `sed '/(/ !s/static constexpr/static const/'`.
