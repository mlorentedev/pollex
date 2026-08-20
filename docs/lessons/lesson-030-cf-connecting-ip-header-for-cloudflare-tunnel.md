---
id: lesson-030-cf-connecting-ip-header-for-cloudflare-tunnel
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, cloudflare, tunnel, rate-limiting]
---

# `Cf-Connecting-Ip` header for Cloudflare Tunnel

**Context:** Rate limiting requests through Cloudflare Tunnel.

**Problem:** Without reading the real client IP, the rate limiter would see `127.0.0.1` for everyone.

**Solution:** Read `Cf-Connecting-Ip` header.

**Tags:** `#cloudflare` `#tunnel` `#rate-limiting`

---
