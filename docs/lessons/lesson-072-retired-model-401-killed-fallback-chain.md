---
id: lesson-072-retired-model-401-killed-fallback-chain
type: lesson
status: active
created: "2026-10-06"
owner: manu
tags: [pollex, lesson, nan, fallback, error-handling, health-check, observability]
---

# A retired model answering 401 silently killed the whole fallback chain

**Context:** Pre-migration audit, 2026-10-06. NaN had retired `mimo-v2.5` on 2026-09-30, the first model of the `nan-cloud` chain. A real `nan-cloud` polish against production returned 502 in 1.6 s while `/api/health` reported `nan-cloud: available` (#106).

**Problem:** NaN answers a retired model with `401 {"type":"auth_error","param":"model"}`, a per-model 401, not a per-key one. `shouldFallback` classified every 401 as a client error that "recurs identically" (lesson 061, ADR-009 decision 4) and stopped the chain, so the three-model failover never reached `qwen3.6` or `gemma4`, both healthy. Nothing noticed for a week: `Available()` checks configuration, not the upstream, and no user traffic hit the cloud path in that window.

**Solution:** 401 now advances the chain like 404, and the default primary moved to `mimo-v2.6-flash` (ADR-009 amendment). A table test pins that a 401 reaches the next adapter. The fix only counts once deployed: the closing evidence is the same production request returning 200.

**Why:** A failover policy encodes assumptions about the upstream's error semantics, and those drift without notice. When deciding what to fail fast on, weigh the cost of each mistake: failing over on a truly fatal error costs a few fast calls; refusing to fail over on a recoverable one takes the whole engine down. Decide by that asymmetry, not by the status code's textbook meaning. A static health check hides this class of outage, and the only detector is a live request.

**Tags:** `#nan` `#fallback` `#error-handling` `#health-check` `#observability`
