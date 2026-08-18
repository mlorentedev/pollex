---
id: lesson-023-neon-stubs-go-in-ggml-cpu-impl-h-not-in-ggml-
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, arm, neon, llama]
---

# NEON stubs go in `ggml-cpu-impl.h`, NOT in `ggml-cpu-quants.c`

**Context:** Cross-compiling llama.cpp for ARM64 with gcc-8.

**Problem:** `ggml_vld1q_s8_x4` macros are defined in `impl.h`. Injecting stubs in `quants.c` doesn't work because it doesn't include `arm_neon.h` directly.

**Solution:** Put stubs in `ggml-cpu-impl.h`.

**Tags:** `#arm` `#neon` `#llama.cpp`

---
