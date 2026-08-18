---
id: lesson-045-mount-type-cache-in-docker-build
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, docker, performance]
---

# `--mount=type=cache` in Docker build

**Context:** Optimizing multi-stage Docker builds.

**Problem:** Full rebuilds take ~30s when only code changes.

**Solution:** Cache `GOMODCACHE` and `GOCACHE` between builds. Reduces rebuild time to ~5s.

**Tags:** `#docker` `#performance`

---
