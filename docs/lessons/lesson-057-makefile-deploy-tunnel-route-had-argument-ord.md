---
id: lesson-057-makefile-deploy-tunnel-route-had-argument-ord
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, makefile, cloudflare]
---

# Makefile `deploy-tunnel-route` had argument order wrong

**Context:** The Makefile target for registering DNS routes.

**Problem:** `cloudflared tunnel route dns` expects `<tunnel> <hostname>`, but the Makefile had `<hostname> <tunnel>`.

**Solution:** Fixed argument order in the Makefile.

**Tags:** `#makefile` `#cloudflare`

---
