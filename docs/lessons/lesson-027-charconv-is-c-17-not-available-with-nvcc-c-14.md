---
id: lesson-027-charconv-is-c-17-not-available-with-nvcc-c-14
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cuda, c, jetson]
---

# `<charconv>` is C++17, not available with nvcc C++14

**Context:** Cross-compiling llama.cpp for Jetson Nano.

**Problem:** gcc-8 only provides `<charconv>` in `-std=c++17` mode, but nvcc 10.2 is forced to C++14.

**Solution:** Create a `charconv` shim with `std::from_chars` implemented over `strtol`/`strtof`, inject via `-isystem` in `CMAKE_CUDA_FLAGS`.

**Tags:** `#cuda` `#c++17` `#jetson`

---
