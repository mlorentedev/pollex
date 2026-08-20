---
id: lesson-028-don-t-blanket-replace-static-constexpr-in-fun
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, sed, llama]
---

# Don't blanket-replace `static constexpr` in functions

**Context:** Patching llama.cpp for Jetson Nano.

**Problem:** `sed 's/static constexpr/static const/'` blanket breaks constexpr functions used as template args (mmvq.cu, warp_reduce_sum).

**Solution:** Only replace on lines without `(`: `sed '/(/ !s/static constexpr/static const/'`.

**Tags:** `#sed` `#llama.cpp`

---
