---
id: lesson-021-dcmake-cuda-standard-14-is-mandatory-for-jets
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cmake, cuda, jetson]
---

# `-DCMAKE_CUDA_STANDARD=14` is mandatory for Jetson Nano

**Context:** Building llama.cpp with CUDA on Jetson Nano (CUDA 10.2).

**Problem:** CUDA 10.2 nvcc doesn't support C++17. Without this flag, cmake fails with "CUDA17 dialect not supported".

**Solution:** Pass `-DCMAKE_CUDA_STANDARD=14 -DCMAKE_CUDA_STANDARD_REQUIRED=TRUE`.

**Tags:** `#cmake` `#cuda` `#jetson`

---
