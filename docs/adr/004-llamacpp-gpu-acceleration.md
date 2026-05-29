---
id: "pollex-adr-004-llamacpp-gpu"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-12"
owner: manu
---

# ADR-004: llama.cpp GPU Acceleration on Jetson Nano

**Status:** Accepted
**Date:** 2026-02-12
**Supersedes:** Partial update to [001-local-llm-on-jetson-nano](001-local-llm-on-jetson-nano.md)

## Context

After deploying Pollex to the Jetson Nano 4GB (Phase 5), we discovered that **Ollama runs at 100% CPU** — taking **~41 seconds per request** while the 128 Maxwell CUDA cores sit idle. Investigation revealed that Ollama dropped CUDA 10.2 support in late 2024. The Jetson Nano is stuck on CUDA 10.2 (JetPack 4.6.6, Ubuntu 18.04), and NVIDIA has end-of-lifed software support for this hardware.

This makes Ollama viable only as a CPU fallback, not as the primary inference engine on the Jetson.

## Decision

Replace Ollama with **llama.cpp** (`llama-server`) as the primary inference engine on the Jetson Nano. Compile llama.cpp from source with CUDA 10.2 support using targeted compatibility patches. Keep Ollama registered for development machines and as a CPU fallback.

## Alternatives Considered

| Option | Verdict | Reason |
|--------|---------|--------|
| **Ollama (status quo)** | Rejected | No CUDA 10.2 support. 41s/request on CPU is unacceptable UX. |
| **vLLM** | Rejected | Requires CUDA 11.8+, Python runtime. Too heavy for 4GB. |
| **TGI (Text Generation Inference)** | Rejected | Requires CUDA 11.8+, Rust toolchain, high memory overhead. |
| **llama.cpp (chosen)** | Accepted | C/C++, minimal dependencies, compiles with CUDA 10.2 via patches, OpenAI-compatible server mode. |
| **ONNX Runtime** | Rejected | Model conversion complexity, limited chat template support. |

llama.cpp is the only viable option that (a) compiles on CUDA 10.2 with patches, (b) fits in 4GB RAM, and (c) provides an HTTP API compatible with existing adapter patterns.

## Implementation

### Adapter

New `LlamaCppAdapter` in the backend, targeting llama-server's OpenAI-compatible API:
- **Polish:** `POST /v1/chat/completions` (standard OpenAI chat format)
- **Health:** `GET /health` (2s timeout)
- **Timeout:** 120s (vs 60s for Ollama) — GPU inference on Nano can be slow for long texts

Registered conditionally when `POLLEX_LLAMACPP_URL` is configured. Coexists with Ollama and Claude adapters.

### Pinned Commit Strategy

llama.cpp is pinned to commit `23106f9` rather than a release tag because:
1. **CUDA 10.2 patches are commit-specific** — patches target specific file lines that change between versions
2. **No release guarantees CUDA 10.2 compat** — the project officially dropped it
3. **Reproducibility** — exact same binary on every build, no surprise breakage
4. **Update path** — when upgrading, test new commit locally, update patches, then pin new commit

**Risk:** Pinning means no automatic security fixes. Mitigation: llama-server listens on `127.0.0.1` only, not exposed to network.

### CUDA 10.2 Patches

Eight patches required to compile on CUDA 10.2 / nvcc from JetPack 4.6.6:
1. `CMAKE_CUDA_ARCHITECTURES=53` (Maxwell, Jetson Nano compute capability)
2. `stdc++fs` linker flag + `--copy-dt-needed-entries` for gcc-8 filesystem support
3. `static constexpr` → `static const` in common.cuh (nvcc 10.2 limitation)
4. Comment `__builtin_assume` in flash attention files (gcc-8 / nvcc 10.2 incompatibility)
5. Stub `cuda_bf16.h` + `cuda_bf16.hpp` (CUDA 10.2 has no bf16 — typedefs to `half`/fp16)
6. ARM NEON intrinsic stubs in `ggml-cpu-impl.h` (gcc-8.4 missing `vld1q_*_x2/x4`, added in gcc-8.5)
7. `<charconv>` shim (C++17 header unavailable under nvcc C++14 — `std::from_chars` via `strtol`/`strtof`)
8. Stub out `fattn-wmma-f16.cu` (WMMA requires Volta+ compute 7.0, Maxwell is 5.3)

Key cmake flags beyond the patches:
- `-DCMAKE_CUDA_STANDARD=14` — nvcc 10.2 does not support C++17
- `-DCMAKE_CUDA_STANDARD_REQUIRED=TRUE` — prevent silent fallback
- `-DGGML_CPU_ARM_ARCH=armv8-a` — correct ARM architecture for Cortex-A57
- `-DGGML_NATIVE=OFF` — avoid auto-detected CPU flags that may fail

Additional build prerequisites discovered during deployment:
- CMake 3.14+ required (Ubuntu 18.04 ships 3.10) — install aarch64 binary from Kitware (`pip3 install cmake` fails on Python 3.6)
- `/opt/llama.cpp-build/` needs `sudo git clone` + `chown` under `/opt/`

All patches are applied by `deploy/build-llamacpp.sh` via `sed`/`awk`/`cat`. Build takes ~85 minutes on the Nano (to be validated).

### llama-server Flags

```
-ngl 999    # Offload ALL layers to GPU
-c 2048     # Context size (conservative for 4GB shared RAM)
-t 4        # All 4 ARM A57 cores for CPU fallback layers
--host 127.0.0.1  # Local only
--port 8080
```

## Consequences

### Positive

- **~300-500% speedup** expected (41s → ~8-15s per request)
- GPU utilization finally enabled on the Nano
- OpenAI-compatible API simplifies adapter code
- Same model (Qwen2.5-1.5B Q4_K_M) — no quality regression
- Ollama remains available on dev machines (no workflow change)

### Negative

- Custom build script with patches — maintenance burden on llama.cpp upgrades
- ~85 minute build time on Nano (one-time, idempotent)
- Two inference services on Jetson (Ollama for fallback + llama-server primary)
- Pinned commit means manual security review on updates

### Rollback

If llama-server fails on the Jetson:
1. `sudo systemctl stop llama-server`
2. Remove `llamacpp_url` from `/etc/pollex/config.yaml`
3. `sudo systemctl restart pollex-api` — falls back to Ollama (CPU, slower but working)

## Memory Budget (Updated)

| Component | RAM |
|-----------|-----|
| JetPack OS (headless) | ~500MB |
| llama-server + model (GPU layers) | ~1.2GB |
| Pollex Go API | ~15MB |
| **Total** | **~1.7GB** |
| **Free** | **~2.3GB** |

Ollama can be stopped on the Jetson once llama-server is validated, freeing ~200MB.

## References

- [001-local-llm-on-jetson-nano](001-local-llm-on-jetson-nano.md) — Original LLM decision
- `tasks/plan-phase9.md` — Detailed implementation plan
- `deploy/build-llamacpp.sh` — Build script with patches
