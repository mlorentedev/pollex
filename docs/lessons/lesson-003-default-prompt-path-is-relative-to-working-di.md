---
id: lesson-003-default-prompt-path-is-relative-to-working-di
type: lesson
status: active
created: "2026-03-28"
owner: manu
tags: [pollex, lesson, go, paths, deployment]
---

# Default prompt path is relative to working directory, not binary

**Context:** Setting up the Go backend with a relative prompt path.

**Problem:** `../prompts/polish.txt` is relative to the working directory, not the binary location. Running `go run .` from the package directory works. From another directory, it fails.

**Solution:** Use absolute paths or `embed` for resource files. For the deploy target, `/etc/pollex/polish.txt` is the canonical path.

**Tags:** `#go` `#paths` `#deployment`

---
