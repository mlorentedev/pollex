#!/bin/bash
set -euo pipefail
export PATH="$HOME/.local/bin:/usr/local/cuda/bin:$PATH"

# === Configuration ===
LLAMA_COMMIT="23106f9"
BUILD_DIR="/opt/llama.cpp-build"
MODEL_URL="https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_0.gguf"
MODEL_DIR="/opt/llama-models"
MODEL_PATH="${MODEL_DIR}/qwen2.5-1.5b-instruct-q4_0.gguf"

echo "=== llama.cpp GPU build for Jetson Nano (CUDA 10.2) ==="

# --- Step 1: Skip if already built ---
if [ -f /usr/local/bin/llama-server ]; then
  echo "llama-server already installed at /usr/local/bin/llama-server"
  echo "Delete it first if you want to rebuild."
  /usr/local/bin/llama-server --version 2>/dev/null || true
else
  # --- Step 2: Install build dependencies ---
  echo "Installing build dependencies..."
  sudo apt-get update -qq
  sudo apt-get install -y -qq gcc-8 g++-8 git libcurl4-openssl-dev

  # CMake 3.14+ required (Ubuntu 18.04 ships 3.10).
  # Download official aarch64 binary from Kitware.
  CMAKE_VER="3.28.6"
  if cmake --version 2>/dev/null | grep -qE '3\.(1[4-9]|[2-9][0-9])'; then
    echo "CMake already meets minimum version."
  else
    echo "Installing CMake ${CMAKE_VER} (aarch64 binary)..."
    curl -fsSL "https://github.com/Kitware/CMake/releases/download/v${CMAKE_VER}/cmake-${CMAKE_VER}-linux-aarch64.tar.gz" \
      | sudo tar -xz -C /usr/local --strip-components=1
  fi
  echo "CMake version: $(cmake --version | head -1)"

  # --- Step 3: Clone llama.cpp at pinned commit ---
  if [ -d "${BUILD_DIR}" ]; then
    echo "Removing previous build directory..."
    sudo rm -rf "${BUILD_DIR}"
  fi

  echo "Cloning llama.cpp at commit ${LLAMA_COMMIT}..."
  sudo git clone https://github.com/ggml-org/llama.cpp.git "${BUILD_DIR}"
  sudo chown -R "$(whoami):$(whoami)" "${BUILD_DIR}"
  cd "${BUILD_DIR}"
  git checkout "${LLAMA_COMMIT}"

  # --- Step 4: Apply CUDA 10.2 + gcc-8.4 compatibility patches ---
  echo "Applying patches..."

  # 4a. Force CUDA architecture 53 (Maxwell, Jetson Nano)
  sed -i 's/set(CMAKE_CUDA_ARCHITECTURES.*/set(CMAKE_CUDA_ARCHITECTURES "53")/' CMakeLists.txt

  # 4b. Add stdc++fs linker flag and --copy-dt-needed-entries for gcc-8
  if ! grep -q 'stdc++fs' ggml/CMakeLists.txt; then
    cat >> ggml/CMakeLists.txt << 'CMAKEOF'
target_link_libraries(ggml-base PRIVATE stdc++fs)
add_link_options(-Wl,--copy-dt-needed-entries)
CMAKEOF
  fi

  # 4c. Replace 'static constexpr' with 'static const' ONLY on variable declarations.
  # Lines with '(' are function declarations — keep constexpr on those (C++14 supports it).
  # Without this, constexpr functions lose their constexpr-ness and can't be used in
  # template arguments / constant expressions (breaks mmvq.cu, warp_reduce_sum, etc.).
  sed -i '/(/ !s/static constexpr/static const/' ggml/src/ggml-cuda/common.cuh

  # 4d. Comment out __builtin_assume (not supported by nvcc 10.2)
  for f in ggml/src/ggml-cuda/fattn-common.cuh \
           ggml/src/ggml-cuda/fattn-vec-f32.cuh \
           ggml/src/ggml-cuda/fattn-vec-f16.cuh; do
    if [ -f "$f" ]; then
      sed -i 's/__builtin_assume/\/\/ __builtin_assume/' "$f"
    fi
  done

  # 4e. Create proper cuda_bf16.h stub (map nv_bfloat16 to half)
  # CUDA 10.2 has no bf16 support — redirect to fp16 types.
  CUDA_INCLUDE="/usr/local/cuda/include"
  echo "Creating cuda_bf16.h stub..."
  sudo tee "${CUDA_INCLUDE}/cuda_bf16.h" > /dev/null << 'STUBEOF'
#ifndef CUDA_BF16_H
#define CUDA_BF16_H

#include <cuda_fp16.h>

// Stub for CUDA 10.2: map bfloat16 types to half (fp16).
// Precision differs but allows compilation. Inference quality unaffected
// for Q4-quantized models (bf16 only used in intermediate compute).
typedef half nv_bfloat16;
typedef half2 nv_bfloat162;
typedef half  __nv_bfloat16;
typedef half2 __nv_bfloat162;

#endif // CUDA_BF16_H
STUBEOF

  echo "Creating cuda_bf16.hpp stub..."
  sudo tee "${CUDA_INCLUDE}/cuda_bf16.hpp" > /dev/null << 'HPPEOF'
#ifndef CUDA_BF16_HPP
#define CUDA_BF16_HPP
#include "cuda_bf16.h"
#endif // CUDA_BF16_HPP
HPPEOF

  # 4f. Fix ARM NEON intrinsic conflicts for gcc-8 on aarch64
  # gcc-8 on aarch64 already provides vld1q_*_x2 in arm_neon.h, but
  # llama.cpp defines its own polyfills that clash (different inline attrs).
  # Comment out the upstream _x2 polyfills, and add _x4 stubs only (gcc-8
  # does NOT provide _x4 variants).
  IMPL_FILE="ggml/src/ggml-cpu/ggml-cpu-impl.h"
  if [ -f "$IMPL_FILE" ]; then
    echo "Patching ARM NEON intrinsics in ggml-cpu-impl.h for gcc-8..."

    # Comment out upstream vld1q_*_x2 polyfills (already in gcc-8 arm_neon.h)
    sed -i '/^static inline int8x16x2_t vld1q_s8_x2/,/^}/s/^/\/\//' "$IMPL_FILE"
    sed -i '/^static inline uint8x16x2_t vld1q_u8_x2/,/^}/s/^/\/\//' "$IMPL_FILE"
    sed -i '/^static inline int16x8x2_t vld1q_s16_x2/,/^}/s/^/\/\//' "$IMPL_FILE"

    # Add _x4 stubs only (not provided by gcc-8 arm_neon.h)
    NEON_X4_STUBS='
/* --- gcc-8 ARM NEON _x4 stubs (not in arm_neon.h until gcc-10+) --- */
#if defined(__ARM_NEON) && defined(__GNUC__) && __GNUC__ < 10
static inline int8x16x4_t vld1q_s8_x4(const int8_t *ptr) {
    int8x16x4_t res;
    res.val[0] = vld1q_s8(ptr);
    res.val[1] = vld1q_s8(ptr + 16);
    res.val[2] = vld1q_s8(ptr + 32);
    res.val[3] = vld1q_s8(ptr + 48);
    return res;
}
static inline uint8x16x4_t vld1q_u8_x4(const uint8_t *ptr) {
    uint8x16x4_t res;
    res.val[0] = vld1q_u8(ptr);
    res.val[1] = vld1q_u8(ptr + 16);
    res.val[2] = vld1q_u8(ptr + 32);
    res.val[3] = vld1q_u8(ptr + 48);
    return res;
}
#endif
/* --- end gcc-8 _x4 stubs --- */
'
    # Insert _x4 stubs after the first #include <arm_neon.h> in the file
    TMPFILE=$(mktemp)
    awk -v stubs="$NEON_X4_STUBS" '
      /#include <arm_neon.h>/ && !done { print; print stubs; done=1; next }
      { print }
    ' "$IMPL_FILE" > "$TMPFILE"
    mv "$TMPFILE" "$IMPL_FILE"
  fi

  # 4g. Create <charconv> shim (C++17 header, not available with nvcc C++14)
  # gcc-8 only provides <charconv> in -std=c++17 mode, but nvcc 10.2 is limited to C++14.
  # Provide a minimal implementation using strtol/strtof, placed in a shim dir
  # that's injected into the include path via CMAKE_CUDA_FLAGS.
  SHIM_DIR="${BUILD_DIR}/shims"
  mkdir -p "$SHIM_DIR"
  echo "Creating <charconv> C++14 shim..."
  cat > "${SHIM_DIR}/charconv" << 'CHARCONVEOF'
// Minimal <charconv> shim for C++14 (CUDA 10.2 / nvcc / gcc-8).
// Provides std::from_chars for int/long/float using C stdlib functions.
#pragma once
#include <cstdlib>
#include <cerrno>
#include <climits>
#include <system_error>

namespace std {

struct from_chars_result {
    const char* ptr;
    errc ec;
};

inline from_chars_result from_chars(const char* first, const char* last, int& value, int base = 10) {
    (void)last;
    char* end = nullptr;
    errno = 0;
    long r = strtol(first, &end, base);
    if (errno == ERANGE || r > INT_MAX || r < INT_MIN)
        return {first, errc::result_out_of_range};
    if (end == first)
        return {first, errc::invalid_argument};
    value = static_cast<int>(r);
    return {end, errc{}};
}

inline from_chars_result from_chars(const char* first, const char* last, long& value, int base = 10) {
    (void)last;
    char* end = nullptr;
    errno = 0;
    long r = strtol(first, &end, base);
    if (errno == ERANGE)
        return {first, errc::result_out_of_range};
    if (end == first)
        return {first, errc::invalid_argument};
    value = r;
    return {end, errc{}};
}

inline from_chars_result from_chars(const char* first, const char* last, float& value) {
    (void)last;
    char* end = nullptr;
    errno = 0;
    float r = strtof(first, &end);
    if (errno == ERANGE)
        return {first, errc::result_out_of_range};
    if (end == first)
        return {first, errc::invalid_argument};
    value = r;
    return {end, errc{}};
}

} // namespace std
CHARCONVEOF

  # 4h. Disable WMMA flash attention (requires Volta+ compute 7.0, Maxwell is 5.3)
  # Must provide a no-op function body (not empty file) because the symbol
  # is declared in a header and called from the flash-attention dispatcher.
  WMMA_FILE="ggml/src/ggml-cuda/fattn-wmma-f16.cu"
  if [ -f "$WMMA_FILE" ]; then
    echo "Stubbing out fattn-wmma-f16.cu (WMMA not supported on Maxwell)..."
    cat > "$WMMA_FILE" << 'WMMAEOF'
// Stubbed: WMMA requires Volta+ (compute 7.0). Jetson Nano Maxwell is 5.3.
#include "common.cuh"
#include "fattn-wmma-f16.cuh"

void ggml_cuda_flash_attn_ext_wmma_f16(ggml_backend_cuda_context & ctx, ggml_tensor * dst) {
    (void)ctx; (void)dst;
    GGML_ABORT("WMMA flash attention not supported on compute capability < 7.0");
}
WMMAEOF
  fi

  # --- Step 5: Build with CUDA ---
  echo "Building llama.cpp (this takes ~85 minutes on Jetson Nano)..."
  mkdir -p build && cd build
  CC=gcc-8 CXX=g++-8 cmake .. \
    -DGGML_CUDA=ON \
    -DCMAKE_CUDA_COMPILER=/usr/local/cuda/bin/nvcc \
    -DCMAKE_CUDA_ARCHITECTURES=53 \
    -DCMAKE_CUDA_STANDARD=14 \
    -DCMAKE_CUDA_STANDARD_REQUIRED=TRUE \
    -DCMAKE_CUDA_FLAGS="-isystem ${BUILD_DIR}/shims" \
    -DGGML_CPU_ARM_ARCH=armv8-a \
    -DGGML_NATIVE=OFF \
    -DLLAMA_CURL=ON \
    -DCMAKE_BUILD_TYPE=Release
  make -j4 2>&1 | tee /tmp/llama-build.log

  # --- Step 6: Install binary ---
  echo "Installing llama-server..."
  sudo cp bin/llama-server /usr/local/bin/llama-server
  sudo chmod +x /usr/local/bin/llama-server
  echo "llama-server installed."
fi

# --- Step 7: Download model ---
sudo mkdir -p "${MODEL_DIR}"
if [ -f "${MODEL_PATH}" ]; then
  echo "Model already downloaded at ${MODEL_PATH}"
else
  echo "Downloading model (this takes a few minutes)..."
  sudo curl -L -o "${MODEL_PATH}" "${MODEL_URL}"
  echo "Model downloaded."
fi

# --- Step 8: Install systemd service ---
if [ -f /tmp/llama-server.service ]; then
  echo "Installing llama-server systemd service..."
  sudo cp /tmp/llama-server.service /etc/systemd/system/llama-server.service
  sudo systemctl daemon-reload
  sudo systemctl enable llama-server
  sudo systemctl restart llama-server
  sleep 2
  echo "llama-server service status:"
  systemctl is-active llama-server || true
fi

echo "=== llama.cpp build complete ==="
