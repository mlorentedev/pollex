---
id: lesson-041-error-truncation-200-chars
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, chrome-extension, errors]
---

# Error truncation (200 chars)

**Context:** Storing error messages in `chrome.storage.local`.

**Problem:** Server errors (stack traces, internal paths) can be very long.

**Solution:** Truncate to 200 chars to prevent storage bloat.

**Tags:** `#chrome-extension` `#errors`

---
