---
id: lesson-039-timer-ticks-best-effort
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, chrome-extension, timer]
---

# Timer ticks best-effort

**Context:** Progress bar in the extension.

**Problem:** `chrome.runtime.sendMessage` from background to popup fails silently if the popup is closed.

**Solution:** Wrap in try/catch.

**Tags:** `#chrome-extension` `#timer`

---
