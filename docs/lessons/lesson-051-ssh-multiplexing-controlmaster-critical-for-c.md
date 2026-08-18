---
id: lesson-051-ssh-multiplexing-controlmaster-critical-for-c
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, ssh, cloudflare, performance]
---

# SSH multiplexing (`ControlMaster`) critical for Cloudflare Tunnel

**Context:** Deploying via `make deploy` (multiple SCP calls through the tunnel).

**Problem:** Each SCP through the tunnel takes 2-5s to negotiate. `make deploy` with 5 SCP calls takes ~25s without multiplexing.

**Solution:** Add `ControlMaster auto`, `ControlPath /tmp/ssh-%r@%h:%p`, `ControlPersist 10m` to SSH config. Reduces deploy time to ~8s.

**Tags:** `#ssh` `#cloudflare` `#performance`

---
