---
id: lesson-035-promauto-registers-metrics-automatically
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, prometheus, metrics]
---

# `promauto` registers metrics automatically

**Context:** Setting up Prometheus metrics.

**Problem:** Manual `prometheus.MustRegister()` is error-prone.

**Solution:** Use `promauto` — no manual registration needed. Beware: don't use in tests that create multiple registries.

**Tags:** `#prometheus` `#metrics`

---
