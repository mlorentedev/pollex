---
id: lesson-052-build-llamacpp-sh-downloaded-wrong-model
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, deployment, llama]
---

# `build-llamacpp.sh` downloaded wrong model

**Context:** Deploying llama.cpp build script.

**Problem:** Script had `q4_k_m.gguf` hardcoded but production switched to `q4_0.gguf` (23% faster). The bug went unnoticed because the home Jetson was already running q4_0 (manually fixed).

**Solution:** Always verify model filename matches between script, service file, and actual file on disk.

**Tags:** `#deployment` `#llama.cpp`

---
