# Phase 9: llama.cpp GPU Acceleration on Jetson Nano

## Problem

Ollama on Jetson Nano 4GB uses 100% CPU (41s per polish request). The Maxwell GPU (128 CUDA cores, compute 5.3) sits idle because Ollama dropped CUDA 10.2 support. Switching to llama-server (llama.cpp) compiled with CUDA 10.2 should give ~300-500% speedup (41s → ~8-15s).

## Approach

New `LlamaCppAdapter` talking to llama-server's OpenAI-compatible API. Ollama stays as fallback for dev machines. Idempotent build/deploy scripts. Pinned commit for reproducibility.

---

## 9.1 — LlamaCppAdapter

### `backend/adapter_llamacpp.go`

```go
type LlamaCppAdapter struct {
    BaseURL string        // e.g., "http://localhost:8080"
    Model   string        // display name, e.g., "qwen2.5:1.5b"
    Client  *http.Client
}
```

**Polish()**: POST to `BaseURL + "/v1/chat/completions"` — OpenAI-compatible format:
- Request: `{"messages": [{"role":"system","content":"..."},{"role":"user","content":"..."}], "stream": false}`
- Response: `{"choices": [{"message": {"content": "polished text"}}]}`

**Available()**: GET `BaseURL + "/health"` with 2s timeout → true if HTTP 200

**Name()**: `"llama.cpp (Model)"`

Follow existing patterns from `adapter_ollama.go` (same structure, different API format).

### `backend/adapter_llamacpp_test.go`

Table-driven tests mirroring `adapter_ollama_test.go`:
- Polish happy path (verify request format, response parsing)
- Server error (500 → error)
- Context cancellation
- Available when healthy / when unreachable
- Name format

---

## 9.2 — Config + Registration

### `backend/config.go` — new fields

```go
LlamaCppURL   string `yaml:"llamacpp_url"`
LlamaCppModel string `yaml:"llamacpp_model"`
```

Defaults: empty (disabled). Env vars: `POLLEX_LLAMACPP_URL`, `POLLEX_LLAMACPP_MODEL`.

### `backend/main.go` — `buildAdapters()`

Register when `cfg.LlamaCppURL != ""` (same pattern as Claude adapter):
```go
if cfg.LlamaCppURL != "" {
    model := cfg.LlamaCppModel
    if model == "" { model = "llama.cpp" }
    llamacpp := &LlamaCppAdapter{
        BaseURL: cfg.LlamaCppURL,
        Model:   model,
        Client:  &http.Client{Timeout: 120 * time.Second},
    }
    adapters[model] = llamacpp
    models = append(models, ModelInfo{ID: model, Name: model, Provider: "llamacpp"})
}
```

Note: 120s timeout (vs 60s for Ollama) — GPU inference on Nano can be slow for long texts.

### `backend/handler_health.go` — type switch

```go
case *LlamaCppAdapter:
    return "llama-server unreachable"
```

### `backend/handler_test.go`

Health test with LlamaCpp adapter unavailable (mirrors existing pattern).

---

## 9.3 — Deploy: Compile llama.cpp on Jetson

### `deploy/build-llamacpp.sh` — idempotent build script

Runs ON the Jetson. Variables at top:
```bash
LLAMA_COMMIT="23106f9"
MODEL_URL="https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf"
MODEL_PATH="/opt/llama-models/qwen2.5-1.5b-instruct-q4_k_m.gguf"
```

Steps:
1. **Skip if already built**: check `/usr/local/bin/llama-server` exists
2. **Install build deps**: `apt install gcc-8 g++-8 cmake libcurl4-openssl-dev`
3. **Clone llama.cpp** at pinned commit to `/opt/llama.cpp-build/`
4. **Apply 6 patches** for CUDA 10.2 compatibility:
   - `CMakeLists.txt`: set `CMAKE_CUDA_ARCHITECTURES 53`
   - `ggml/CMakeLists.txt`: add `stdc++fs` linker flag
   - `ggml/src/ggml-cuda/common.cuh`: remove `constexpr` on line 455
   - `ggml/src/ggml-cuda/fattn-common.cuh`: comment `__builtin_assume`
   - `ggml/src/ggml-cuda/fattn-vec-f32.cuh`: comment `__builtin_assume`
   - `ggml/src/ggml-cuda/fattn-vec-f16.cuh`: comment `__builtin_assume`
   - CUDA bf16 compatibility: create stub `cuda_bf16.h`
5. **cmake + build** (~85 min on Nano)
6. **Install**: copy `build/bin/llama-server` to `/usr/local/bin/`
7. **Download model**: `qwen2.5-1.5b-instruct-q4_k_m.gguf` to `/opt/llama-models/`

### `deploy/llama-server.service`

```ini
[Unit]
Description=llama.cpp inference server (GPU)
After=network.target

[Service]
Type=simple
User=manu
ExecStart=/usr/local/bin/llama-server \
    -m /opt/llama-models/qwen2.5-1.5b-instruct-q4_k_m.gguf \
    --port 8080 --host 127.0.0.1 \
    -ngl 999 -c 2048 -t 4
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Key flags:
- `-ngl 999`: offload ALL layers to GPU
- `-c 2048`: context size (conservative for 4GB RAM)
- `-t 4`: all 4 ARM A57 cores for CPU fallback layers
- `--host 127.0.0.1`: local only (no external access)

### `deploy/config.yaml` — additions

```yaml
llamacpp_url: "http://localhost:8080"
llamacpp_model: "qwen2.5:1.5b-gpu"
```

Keep existing `ollama_url` as fallback.

### `Makefile` — new target

```makefile
deploy-llamacpp: ## Build llama.cpp with CUDA on Jetson (~85 min)
	scp deploy/build-llamacpp.sh $(JETSON_USER)@$(JETSON_HOST):/tmp/build-llamacpp.sh
	scp deploy/llama-server.service $(JETSON_USER)@$(JETSON_HOST):/tmp/llama-server.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash /tmp/build-llamacpp.sh'
```

---

## 9.4 — Documentation

- Vault: `runbooks/build-llamacpp-jetson.md` — full build runbook with troubleshooting
- Vault: update `runbooks/deploy-jetson.md` — add llama-server deployment section
- Vault: update `_index.md` — GPU acceleration status

---

## File Summary

| File | Action | Purpose |
|------|--------|---------|
| `backend/adapter_llamacpp.go` | CREATE | LlamaCpp adapter (OpenAI-compat API) |
| `backend/adapter_llamacpp_test.go` | CREATE | Adapter tests (~6) |
| `backend/config.go` | MODIFY | Add LlamaCppURL, LlamaCppModel fields |
| `backend/main.go` | MODIFY | Register adapter in buildAdapters() |
| `backend/handler_health.go` | MODIFY | Add type switch case |
| `backend/handler_test.go` | MODIFY | Health test with LlamaCpp |
| `deploy/build-llamacpp.sh` | CREATE | Idempotent CUDA build script |
| `deploy/llama-server.service` | CREATE | systemd service |
| `deploy/config.yaml` | MODIFY | Add llamacpp_url/model |
| `Makefile` | MODIFY | deploy-llamacpp target |

## Constraints

- Go stdlib only, zero new dependencies
- gcc-8 from apt (Ubuntu 18.04 has it in repos)
- Pinned llama.cpp commit `23106f9` for reproducibility
- All scripts idempotent (safe to re-run)
- Ollama stays as fallback, not removed

## Verification

1. `make test` — all existing + new tests pass (~65+ top-level)
2. `make deploy-llamacpp` — builds llama-server on Jetson (~85 min)
3. `ssh nvidia 'systemctl is-active llama-server'` — service running
4. `make deploy` — deploy new binary with LlamaCpp adapter
5. Compare: `qwen2.5:1.5b` (Ollama/CPU) vs `qwen2.5:1.5b-gpu` (llama-server/GPU)
