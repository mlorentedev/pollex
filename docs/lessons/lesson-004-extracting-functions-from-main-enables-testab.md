---
id: lesson-004-extracting-functions-from-main-enables-testab
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, testing, architecture]
---

# Extracting functions from `main()` enables testability

**Context:** Writing tests for the backend.

**Problem:** `main()` is hard to test because it starts the server and blocks.

**Solution:** Extract `buildAdapters()` and `setupMux()` so they can be called independently with `httptest.NewServer` and the full middleware stack.

**Tags:** `#go` `#testing` `#architecture`

---
