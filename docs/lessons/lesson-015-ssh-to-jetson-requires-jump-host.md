---
id: lesson-015-ssh-to-jetson-requires-jump-host
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, ssh, proxmox]
---

# SSH to Jetson requires jump host

**Context:** Connecting to Jetson from development machine.

**Problem:** The Jetson is on 192.168.2.x behind Proxmox.

**Solution:** Configure `~/.ssh/config` with `ProxyJump pve`.

**Tags:** `#ssh` `#proxmox`

---
