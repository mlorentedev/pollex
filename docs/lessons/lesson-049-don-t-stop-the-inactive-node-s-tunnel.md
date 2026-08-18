---
id: lesson-049-don-t-stop-the-inactive-node-s-tunnel
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, cloudflare, tunnel, multi-node]
---

# Don't stop the inactive node's tunnel

**Context:** Multi-node deployment with Cloudflare Tunnels.

**Problem:** Both tunnels serve independent endpoints (`pollex-home.mlorente.dev`, `pollex-office.mlorente.dev`).

**Solution:** Both tunnels must stay active for independent monitoring. Only the production CNAME is redirected.

**Tags:** `#cloudflare` `#tunnel` `#multi-node`

---
