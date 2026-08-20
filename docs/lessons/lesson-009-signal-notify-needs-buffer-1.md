---
id: lesson-009-signal-notify-needs-buffer-1
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, signals]
---

# `signal.Notify` needs buffer 1

**Context:** Graceful shutdown with OS signals.

**Problem:** Without a buffer, the signal can be lost if nobody is listening at the exact moment.

**Solution:** `done := make(chan os.Signal, 1)`.

**Tags:** `#go` `#signals`

---
