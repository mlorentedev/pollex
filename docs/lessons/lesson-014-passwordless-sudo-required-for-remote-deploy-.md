---
id: lesson-014-passwordless-sudo-required-for-remote-deploy-
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, deployment, sudo, ssh]
---

# Passwordless sudo required for remote deploy scripts

**Context:** Running deploy scripts via SSH.

**Problem:** SSH remote commands fail with "password required" for `sudo`.

**Solution:** Configure `/etc/sudoers.d/manu` for passwordless sudo.

**Tags:** `#deployment` `#sudo` `#ssh`

---
