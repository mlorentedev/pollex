---
id: lesson-007-http-maxbyteserror-requires-errors-as
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, error-handling]
---

# `http.MaxBytesError` requires `errors.As()`

**Context:** Handling oversized request bodies.

**Problem:** The error is wrapped by `json.Decoder`, so direct type assertion fails.

**Solution:** Use `errors.As(err, &maxBytesErr)` to unwrap and check.

**Tags:** `#go` `#error-handling`

---
