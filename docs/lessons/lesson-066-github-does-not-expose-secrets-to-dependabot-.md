---
id: lesson-066-github-does-not-expose-secrets-to-dependabot-
type: lesson
status: active
created: "2026-08-06"
owner: manu
tags: [pollex, lesson, ci, dependabot, secrets, github-actions, bitacora]
---

# GitHub does not expose secrets to Dependabot PRs — skip them in workflows that need secrets

**Context:** The `add-to-project` workflow (bitácora board) failed on Dependabot PRs (#48/#49) with "Bad credentials"-style errors even after the PAT was fixed.

**Problem:** GitHub deliberately does not provide repository secrets to workflows triggered by Dependabot PRs (same rule as fork PRs). `secrets.BITACORA_PAT` arrives empty, so any step using it fails. Re-running the workflow or fixing the PAT does not help — the PR type itself is the blocker.

**Solution:** Add `&& github.actor != 'dependabot[bot]'` to the job condition (skip Dependabot PRs; issues and normal PRs unaffected). Dependabot PRs reach the board via the rollout backfill or a manual item-add.

**Why:** Secrets are intentionally withheld from Dependabot PRs to prevent a compromised dependency file from exfiltrating credentials. Detect the pattern early: a workflow that needs secrets should gate on `github.actor`.

**Tags:** `#ci` `#dependabot` `#secrets` `#github-actions` `#bitacora`

---
