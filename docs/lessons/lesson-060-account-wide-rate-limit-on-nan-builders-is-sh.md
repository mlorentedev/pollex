---
id: lesson-060-account-wide-rate-limit-on-nan-builders-is-sh
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, nan, rate-limiting, concurrency, cloud]
---

# Account-wide rate limit on nan.builders is shared with all tooling

**Context:** Pollex uses the same `nan.api-key` as Hermes, `qq`, and other TUI tools.

**Problem:** The gateway caps at 100 RPM / 5 concurrent **per account**, not per app. A traffic burst through Pollex (semi-public via Cloudflare Tunnel) starves interactive tooling.

**Solution:** Wrap the chain with a `Throttle` concurrency semaphore (default 3, `POLLEX_NAN_MAX_CONCURRENT`). Cloud path stays API-key gated. Consider per-client backoff for future semi-public deployments.

**Why:** The 5-concurrent cap is shared. If Pollex holds 5 slots, `qq` and Hermes get 429s even for single-turn interactive use.

**Tags:** `#nan` `#rate-limiting` `#concurrency` `#cloud`

---
