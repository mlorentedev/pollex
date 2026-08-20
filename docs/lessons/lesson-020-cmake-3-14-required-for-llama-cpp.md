---
id: lesson-020-cmake-3-14-required-for-llama-cpp
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cmake, jetson]
---

# CMake 3.14+ required for llama.cpp

**Context:** Building llama.cpp on Ubuntu 18.04.

**Problem:** System ships CMake 3.10. `pip3 install cmake` fails — needs `skbuild` which is not available on Python 3.6.

**Solution:** Install aarch64 binary from Kitware: `curl | tar` to `/usr/local/`.

**Tags:** `#cmake` `#jetson`

---
