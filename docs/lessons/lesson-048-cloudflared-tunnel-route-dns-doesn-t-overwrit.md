---
id: lesson-048-cloudflared-tunnel-route-dns-doesn-t-overwrit
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, cloudflare, tunnel, dns]
---

# `cloudflared tunnel route dns` doesn't overwrite

**Context:** Registering DNS routes for the pollex tunnel.

**Problem:** If the CNAME already exists, it fails with `An A, AAAA, or CNAME record with that host already exists`.

**Solution:** Use `--overwrite-dns` for cutover between tunnels.

**Tags:** `#cloudflare` `#tunnel` `#dns`

---
