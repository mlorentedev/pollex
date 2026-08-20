---
id: lesson-033-mlock-prevents-model-paging
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, jetson, llama, performance]
---

# `--mlock` prevents model paging

**Context:** Optimizing llama.cpp server on Jetson Nano.

**Problem:** Without mlock, the kernel can swap the model to disk during inactivity, causing cold-start latency.

**Solution:** Always use `--mlock` on the Jetson.

**Tags:** `#jetson` `#llama.cpp` `#performance`

---
