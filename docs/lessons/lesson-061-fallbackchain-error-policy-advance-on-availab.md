---
id: lesson-061-fallbackchain-error-policy-advance-on-availab
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, go, fallback, error-handling, nan]
---

# FallbackChain error policy: advance on availability, fail fast on client errors

**Context:** Implementing a 3-model fallback chain (mimo-v2.5 → qwen3.6 → gemma4) over nan.builders.

**Problem:** "Advance on any error" is simple but re-tries a 400 (malformed prompt) against all models — same error, wasted calls, full latency.

**Solution:** Advance only on availability/quota errors (HTTP 429, 404, 5xx, network/timeout). Fail fast on client errors (400, 401) and context cancellation. Carry HTTP status via a typed `*StatusError` sentinel so the classifier is a simple switch, not string matching.

**Why:** 400/401 are deterministic — retrying against a different model produces the same result. Availability/quota errors are transient and model-specific — the next model may succeed.

**Tags:** `#go` `#fallback` `#error-handling` `#nan`

---
