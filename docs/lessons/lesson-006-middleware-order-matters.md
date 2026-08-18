---
id: lesson-006-middleware-order-matters
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, http, middleware]
---

# Middleware order matters

**Context:** Building the HTTP middleware stack.

**Problem:** Wrong order causes preflight failures, missing request IDs, or wasted resources.

**Solution:** CORS → RequestID → Logging → Metrics → APIKey → RateLimit → MaxBytes → Timeout → mux.

**Tags:** `#go` `#http` `#middleware`

---
