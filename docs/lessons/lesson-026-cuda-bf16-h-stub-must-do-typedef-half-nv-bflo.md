---
id: lesson-026-cuda-bf16-h-stub-must-do-typedef-half-nv-bflo
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cuda, jetson]
---

# `cuda_bf16.h` stub must do `typedef half nv_bfloat16`

**Context:** Cross-compiling llama.cpp for Jetson Nano.

**Problem:** Defining `__nv_bfloat16` as a struct is not enough — the code uses both names (`nv_bfloat16` and `__nv_bfloat16`).

**Solution:** Include `cuda_fp16.h` and typedef both to `half`.

**Tags:** `#cuda` `#jetson`

---
