---
id: lesson-062-gremlins-v0-6-0-appends-internally-pass-bare-
type: lesson
status: active
created: "2026-06-06"
owner: manu
tags: [pollex, lesson, gremlins, mutation-testing, ci, go]
---

# gremlins v0.6.0 appends `...` internally — pass bare package paths in CI

**Context:** Mutation-testing CI step (`gremlins unleash --timeout-coefficient 5 ...`) over `./internal/adapter` and `./internal/config` in `.github/workflows/ci.yml`.

**Problem:** The job failed on master with `impossible to executeCoverage: exit status 1`. The command used Go-style wildcards (`./internal/adapter/...`), which resolved to no packages and aborted the run before any mutation ran.

**Solution:** Pass the bare package path with **no** trailing `...`: `gremlins unleash --timeout-coefficient 5 ./internal/adapter`. A repo-root `.gremlins.toml` (`threshold = 80`, `excluding = ["_test.go"]`) drives both local and CI runs. Fixed in PR #33, released in v1.8.1.

**Why:** gremlins v0.6.0 performs the recursive `...` expansion *internally*, so a caller-supplied `./internal/adapter/...` becomes `./internal/adapter/.../...`, which matches zero packages. Earlier versions passed `...` through verbatim, so the same invocation used to work — this is a silent behavioural change across versions. The CLI surfaces it only as a generic coverage failure with no hint that the path is the cause.

**Tags:** `#gremlins` `#mutation-testing` `#ci` `#go`

---
