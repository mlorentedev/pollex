---
id: lesson-017-zstd-required-for-ollama
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, ollama, dependencies]
---

# `zstd` required for Ollama

**Context:** Installing Ollama on Jetson.

**Problem:** Ollama installer uses zstd for decompression.

**Solution:** Add `zstd` to prerequisites in `install.sh` along with `curl`.

**Tags:** `#ollama` `#dependencies`

---
