---
id: lesson-071-audit-gate-dependabot-deadlock
type: lesson
status: active
created: "2026-09-27"
owner: manu
tags: [pollex, lesson, ci, npm-audit, dependabot, deadlock, security]
---

# A required npm-audit gate deadlocks per-finding Dependabot security PRs

**Context:** Six open Dependabot PRs, all red on the same CI job (`npm audit (site + extension)`), going back over two weeks (#87–#94). Looked like a flaky job; it was structural.

**Problem:** `/site` had three audit findings (devalue GHSA-9rgm-9g3h-6x36, js-yaml GHSA-2883-xcg3-v3hh, svgo GHSA-w27v-7q3p-w38r/GHSA-4vpr-x523-8j87), all in transitive dependencies. Dependabot opened **one PR per finding**, and each PR fixed exactly one. Since the audit job demands zero findings, every PR failed on the two findings it did not fix — none could ever pass, none could merge, and Dependabot kept the set open forever. A gate that requires *all* findings fixed combined with *one-fix-per-PR* automation is a deadlock by construction, not bad luck. Worse: the red job on every PR made `npm audit` background noise, which is exactly when a *real* high-severity finding would have been missed.

**Solution:** One combined lockfile-only PR (PR #95): `npm audit fix` in `/site`, which moves transitive deps within the ranges already declared in `package.json` (34 lines of `package-lock.json`, no manifest change). It fixes all findings at once, passes the gate, and the superseded Dependabot PRs close (`Closes #96`). Then `@dependabot rebase` the unrelated PRs (#90, #93, #94) so they pick up the fixed `master` and go green.

**Why:** The audit gate is fail-closed, and Dependabot is per-finding — the only path through is a change that fixes everything in one commit. Keep that path cheap so it is the obvious move: the fix must stay **lockfile-only** (no new dependencies, no manifest edit) so it needs no SDD gate and no review overhead. If the pattern recurs, the structural fix is a Dependabot `groups` entry for security updates so one PR carries all fixes from the start.

**Tags:** `#ci` `#npm-audit` `#dependabot` `#deadlock` `#security`
