---
id: lesson-008-rate-limiter-sliding-window-with-time-time
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, rate-limiting]
---

# Rate limiter sliding window with `[]time.Time`

**Context:** Building a rate limiter for a single-instance server.

**Problem:** Token bucket or Redis are overkill for LAN use.

**Solution:** Simple sliding window using a `[]time.Time` slice — effective and dependency-free.

**Tags:** `#go` `#rate-limiting`

---
