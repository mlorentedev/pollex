---
id: lesson-053-user-manu-in-systemd-fails-only-for-cloudflar
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, systemd, jetson, cloudflare]
---

# `User=manu` in systemd fails only for cloudflared

**Context:** Hardening systemd service files on JetPack 4.6.

**Problem:** `pollex-api.service` and `llama-server.service` work fine with `User=manu` and hardening directives. The `cloudflared.service` specifically fails with `failed to determine user credentials`.

**Solution:** Run cloudflared as root with explicit `--config` path.

**Tags:** `#systemd` `#jetson` `#cloudflare`

---
