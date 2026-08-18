---
id: lesson-025-wmma-fattn-wmma-f16-cu-requires-volta-compute
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cuda, wmma, jetson]
---

# WMMA (fattn-wmma-f16.cu) requires Volta+ (compute 7.0)

**Context:** Building llama.cpp with CUDA on Jetson Nano (Maxwell, compute 5.3).

**Problem:** Maxwell doesn't support WMMA.

**Solution:** Empty the file leaving only `#include "common.cuh"` for it to compile.

**Tags:** `#cuda` `#wmma` `#jetson`

---
