---
id: lesson-050-restarting-cloudflared-kills-your-ssh
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, cloudflare, tunnel, ssh]
---

# Restarting cloudflared kills your SSH

**Context:** Restarting the Cloudflare Tunnel service.

**Problem:** If you access the Jetson via the same tunnel you restart, the connection drops (`Broken pipe`).

**Solution:** Wait ~15s and reconnect.

**Tags:** `#cloudflare` `#tunnel` `#ssh`

---
