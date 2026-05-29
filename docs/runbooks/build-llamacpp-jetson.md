---
id: "pollex-runbook-build-llamacpp"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-12"
owner: manu
---

# Build llama.cpp on Jetson Nano (CUDA 10.2)

Compiles `llama-server` from source with GPU support on the Jetson Nano 4GB. Required because Ollama dropped CUDA 10.2 support.

See [004-llamacpp-gpu-acceleration](../adr/004-llamacpp-gpu-acceleration.md) for the decision rationale.

## Prerequisites

- Jetson Nano running JetPack 4.6.6 (Ubuntu 18.04, CUDA 10.2)
- SSH access configured: `ssh jetson-home` (see [deploy-jetson](deploy-jetson.md))
- ~2GB free disk space for build artifacts
- ~85 minutes of build time

## Automated Build

From the dev machine:

```sh
make deploy-llamacpp
```

This SCPs `deploy/build-llamacpp.sh` and `deploy/llama-server.service` to the Jetson, then runs the build script. The script is idempotent — safe to re-run.

## What the Build Script Does

1. **Skips if already built** — checks for `/usr/local/bin/llama-server`
2. **Installs gcc-8, g++-8, cmake** — required for CUDA 10.2 compatible compilation
3. **Clones llama.cpp** at pinned commit `23106f9` to `/opt/llama.cpp-build/`
4. **Applies 8 CUDA 10.2 / gcc-8.4 patches**:
   - Force `CMAKE_CUDA_ARCHITECTURES=53` (Maxwell)
   - Add `stdc++fs` linker flag + `--copy-dt-needed-entries` (gcc-8)
   - Replace `static constexpr` with `static const` in common.cuh
   - Comment out `__builtin_assume` in flash attention files
   - Create stub `cuda_bf16.h` / `cuda_bf16.hpp` (no bf16 in CUDA 10.2, typedef to half)
   - Inject ARM NEON intrinsic stubs in `ggml-cpu-impl.h` (gcc-8.4 missing `vld1q_*_x2/x4`)
   - Create `<charconv>` C++14 shim (`std::from_chars` via strtol/strtof, injected via `-isystem`)
   - Stub out `fattn-wmma-f16.cu` (WMMA requires Volta+ compute 7.0)
5. **Builds with cmake** (~85 min, uses all 4 cores)
6. **Installs** `llama-server` to `/usr/local/bin/`
7. **Downloads model** (`qwen2.5-1.5b-instruct-q4_0.gguf`) to `/opt/llama-models/`
8. **Installs systemd service** from `/tmp/llama-server.service`

## Manual Build (If Script Fails)

SSH into the Jetson and run steps manually:

```sh
ssh jetson-home

# Install deps
sudo apt-get install -y gcc-8 g++-8 git libcurl4-openssl-dev

# Install CMake 3.28+ (Ubuntu 18.04 ships 3.10, too old)
curl -fsSL "https://github.com/Kitware/CMake/releases/download/v3.28.6/cmake-3.28.6-linux-aarch64.tar.gz" \
  | sudo tar -xz -C /usr/local --strip-components=1

# Clone at pinned commit
sudo git clone https://github.com/ggml-org/llama.cpp.git /opt/llama.cpp-build
sudo chown -R "$(whoami):$(whoami)" /opt/llama.cpp-build
cd /opt/llama.cpp-build
git checkout 23106f9

# Apply patches (see deploy/build-llamacpp.sh for all 7 patches)
# ...

# Build
mkdir build && cd build
CC=gcc-8 CXX=g++-8 cmake .. \
  -DGGML_CUDA=ON \
  -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc \
  -DCMAKE_CUDA_ARCHITECTURES=53 \
  -DCMAKE_CUDA_STANDARD=14 \
  -DCMAKE_CUDA_STANDARD_REQUIRED=TRUE \
  -DGGML_CPU_ARM_ARCH=armv8-a \
  -DGGML_NATIVE=OFF \
  -DLLAMA_CURL=ON \
  -DCMAKE_BUILD_TYPE=Release
make -j4
```

## Upgrading llama.cpp

To upgrade to a newer commit:

1. Update `LLAMA_COMMIT` in `deploy/build-llamacpp.sh`
2. Remove the old binary: `ssh jetson-home 'sudo rm /usr/local/bin/llama-server'`
3. Review patch compatibility — newer commits may require different patches
4. Test locally first (see [deploy-jetson](deploy-jetson.md))
5. Run `make deploy-llamacpp`

## Verifying the Build

```sh
# Check service is running
ssh jetson-home 'systemctl is-active llama-server'

# Check health endpoint
ssh jetson-home 'curl -s http://localhost:8080/health'

# Test inference
ssh jetson-home 'curl -s http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}]}"'
```

## Troubleshooting

### Build fails with nvcc errors

The patches may not apply cleanly on a different llama.cpp commit. Check:
- `deploy/build-llamacpp.sh` for the exact sed commands
- Compare the target files against the pinned commit to verify line numbers

### Out of memory during build

```sh
# Add swap if not already present
sudo fallocate -l 2G /buildswap
sudo chmod 600 /buildswap
sudo mkswap /buildswap
sudo swapon /buildswap
```

See [jetson-memory](../troubleshooting/jetson-memory.md) for permanent swap setup.

### llama-server crashes on startup

Check logs:
```sh
journalctl -u llama-server -n 50 --no-pager
```

Common causes:
- Model file missing or corrupt → re-download
- Not enough GPU memory → reduce `-c` (context size) from 2048 to 1024
- CUDA driver mismatch → verify `nvcc --version` shows 10.2
