---
id: lesson-038-stale-job-detection
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, chrome-extension, stale-jobs]
---

# Stale job detection

**Context:** Handling service worker termination mid-request.

**Solution:** Compare `Date.now() - polishJob.startedAt` against a threshold (150s). If exceeded, mark as failed.

**Tags:** `#chrome-extension` `#stale-jobs`

---
