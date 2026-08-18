---
id: lesson-065-astro-build-artifacts-site-astro-must-not-be-
type: lesson
status: active
created: "2026-08-05"
owner: manu
tags: [pollex, lesson, astro, docs-site, gitignore, generated-artifacts]
---

# Astro build artifacts (`site/.astro/`) must not be committed

**Context:** The docs site is a Starlight/Astro project. `site/.astro/` (content types, schema, modules) is regenerated on every `astro build`.

**Problem:** The directory was accidentally committed since the initial docs-site PR. Every local build changes it, producing noisy working-tree diffs and stale committed files.

**Solution:** `git rm -r --cached site/.astro/` + add `site/.astro/` to `.gitignore`. Builds regenerate it locally; CI builds don't need it tracked.

**Why:** Generated artifacts pollute diffs and drift from reality. The deploy flow (GitHub Pages via `pages.yml`) builds from source anyway.

**Tags:** `#astro` `#docs-site` `#gitignore` `#generated-artifacts`

---
