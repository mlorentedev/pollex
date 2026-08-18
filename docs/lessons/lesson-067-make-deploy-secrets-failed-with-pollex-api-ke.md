---
id: lesson-067-make-deploy-secrets-failed-with-pollex-api-ke
type: lesson
status: active
created: "2026-08-06"
owner: manu
tags: [pollex, lesson, makefile, deploy, secrets, dotf, jetson]
---

# `make deploy-secrets` failed with "POLLEX_API_KEY not set" in a plain shell — auto-resolve via `dotf secrets run`

**Context:** Deployment to the Jetson from a fresh terminal: `make deploy && make deploy-secrets`. The second target aborted with `POLLEX_API_KEY not set` because the keys only exist inside the dotfiles environment (`dotf secrets run`), not in a plain shell.

**Problem:** The Makefile assumed the keys were exported in the ambient shell. On a machine/shell without dotfiles env loaded, the target fails with a confusing message — and the user cannot know how to proceed.

**Solution:** `deploy-secrets` now checks if either key is missing and, if so, re-invokes an internal `_deploy-secrets` target under `dotf secrets run` (injects keys into the child process only, never the ambient shell). SSOT stays the dotfiles registry (`secrets/registry.yaml`).

**Why:** Deployment targets should never assume ambient env that only exists in one setup. Auto-resolving from the canonical source (dotf) makes the command work identically everywhere and documents the fallback path in the Makefile itself.

**Tags:** `#makefile` `#deploy` `#secrets` `#dotf` `#jetson`

---
