---
id: lesson-031-host-permissions-all-urls-required-in-manifes
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, chrome-extension, manifest-v3]
---

# `host_permissions: ["<all_urls>"]` required in Manifest V3

**Context:** Building the Chrome extension for Cloudflare Tunnel.

**Problem:** Extension cannot fetch external URLs without host permissions.

**Solution:** Add `"<all_urls>"` to `host_permissions` in `manifest.json`.

**Tags:** `#chrome-extension` `#manifest-v3`

---
