---
id: lesson-024-gcc-8-on-aarch64-provides-vld1q-x2-but-not-x4
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, arm, gcc, llama]
---

# gcc-8 on aarch64 provides `vld1q_*_x2` but NOT `_x4`

**Context:** Cross-compiling llama.cpp for Jetson Nano.

**Problem:** Initial assumption that gcc-8.4 lacked all `_x2/_x4` was wrong. Only the `_x4` variants need stubs.

**Solution:** gcc-8's `arm_neon.h` includes `vld1q_s8_x2`, `vld1q_u8_x2`, `vld1q_s16_x2`. Only `_x4` variants need stubs. Comment out llama.cpp's own polyfills in `ggml-cpu-impl.h` to avoid "redeclared inline without 'gnu_inline' attribute" errors.

**Tags:** `#arm` `#gcc` `#llama.cpp`

---
