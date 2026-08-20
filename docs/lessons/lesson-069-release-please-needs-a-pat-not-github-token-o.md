---
id: lesson-069-release-please-needs-a-pat-not-github-token-o
type: lesson
status: active
created: "2026-08-05"
owner: manu
tags: [pollex, lesson, ci, release-please, fine-grained-pat, branch-protection, github-actions]
---

# `release-please` needs a PAT, not `GITHUB_TOKEN`, once CI is a required check

**Context:** Hardening the release pipeline (`release-please.yml`) ahead of enabling branch protection with required status checks.

**Problem:** By GitHub design, events produced with the default `GITHUB_TOKEN` do not trigger other workflows. release-please pushes its release branch with that token, so CI never runs on the release PR — with required checks enabled, the PR is blocked forever on "Waiting for status to be reported".

**Solution:** Authenticate the release-please action with a repo-scoped **fine-grained PAT** (`secrets.RELEASE_TOKEN`, Contents + Pull requests: write). Its pushes are real user events, so `ci.yml` fires normally. Workflow-level `permissions:` dropped to `{}` with each job declaring its own (goreleaser keeps `contents: write` for asset upload).

**Why:** The PAT fixes the root cause instead of working around it — the alternative is running the test matrix inside the release workflow and reporting results through the Statuses API, which is far more plumbing. The cost is a stored secret with a rotation obligation. Relevant to SEC-002 (branch protection): required checks only work if the release PR can actually get them.

**Tags:** `#ci` `#release-please` `#fine-grained-pat` `#branch-protection` `#github-actions`

---
