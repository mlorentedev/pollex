---
id: lesson-036-chrome-popup-lifecycle-destroyed-on-focus-los
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, chrome-extension, service-worker]
---

# Chrome popup lifecycle — destroyed on focus loss

**Context:** Building the Chrome extension.

**Problem:** Any `fetch()` in popup.js is aborted when the popup loses focus.

**Solution:** Move fetch to the service worker (background.js) which persists independently.

**Tags:** `#chrome-extension` `#service-worker`

---
