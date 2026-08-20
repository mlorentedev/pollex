---
id: lesson-056-cloudflare-dns-route-conflict-requires-overwr
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, cloudflare, dns, tunnel]
---

# Cloudflare DNS route conflict requires --overwrite-dns

**Context:** Registering `pollex.mlorente.dev` to the pollex tunnel.

**Problem:** An existing A/CNAME record for `pollex.mlorente.dev` prevented creating the tunnel CNAME. The tunnel and API were running fine on the Jetson — the error was purely at the Cloudflare DNS layer.

**Solution:** `cloudflared tunnel route dns --overwrite-dns pollex pollex.mlorente.dev` to force the CNAME.

**Tags:** `#cloudflare` `#dns` `#tunnel`

---
