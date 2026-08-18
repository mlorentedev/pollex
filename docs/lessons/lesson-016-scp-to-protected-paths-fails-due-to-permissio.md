---
id: lesson-016-scp-to-protected-paths-fails-due-to-permissio
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, deployment, scp, permissions]
---

# SCP to protected paths fails due to permissions

**Context:** Deploying files to the Jetson.

**Problem:** Can't SCP directly to `/usr/local/bin` or `/etc/pollex/`.

**Solution:** SCP to `/tmp/` first, then `ssh ... 'sudo mv /tmp/file /target/'`.

**Tags:** `#deployment` `#scp` `#permissions`

---
