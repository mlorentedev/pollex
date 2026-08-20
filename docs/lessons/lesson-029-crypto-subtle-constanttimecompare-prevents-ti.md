---
id: lesson-029-crypto-subtle-constanttimecompare-prevents-ti
type: lesson
status: active
created: "2026-04-15"
owner: manu
tags: [pollex, lesson, security, go]
---

# `crypto/subtle.ConstantTimeCompare` prevents timing attacks

**Context:** Implementing API key authentication.

**Problem:** Comparing API keys with `==` short-circuits, enabling timing attacks.

**Solution:** Always use `crypto/subtle.ConstantTimeCompare` for secret comparison.

**Tags:** `#security` `#go`

---
