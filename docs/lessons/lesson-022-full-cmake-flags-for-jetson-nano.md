---
id: lesson-022-full-cmake-flags-for-jetson-nano
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cmake, jetson, arm64]
---

# Full cmake flags for Jetson Nano

**Context:** Building llama.cpp with CUDA on ARM64.

**Solution:** `-DGGML_CUDA=ON -DCMAKE_CUDA_STANDARD=14 -DCMAKE_CUDA_STANDARD_REQUIRED=TRUE -DGGML_CPU_ARM_ARCH=armv8-a -DGGML_NATIVE=OFF`.

**Tags:** `#cmake` `#jetson` `#arm64`

---
