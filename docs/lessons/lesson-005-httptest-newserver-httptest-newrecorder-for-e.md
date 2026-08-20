---
id: lesson-005-httptest-newserver-httptest-newrecorder-for-e
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, testing]
---

# `httptest.NewServer` > `httptest.NewRecorder` for E2E

**Context:** Choosing between handler-level and integration tests.

**Problem:** `NewRecorder` only tests individual handlers in isolation.

**Solution:** `NewServer` tests real TCP connections, the full middleware chain, and transport headers.

**Tags:** `#go` `#testing`

---
