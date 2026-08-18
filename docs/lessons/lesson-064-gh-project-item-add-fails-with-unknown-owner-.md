---
id: lesson-064-gh-project-item-add-fails-with-unknown-owner-
type: lesson
status: active
created: "2026-08-05"
owner: manu
tags: [pollex, lesson, ci, github-projects, fine-grained-pat, graphql, bitacora]
---

# `gh project item-add` fails with "unknown owner type" under a fine-grained PAT — use GraphQL in workflows

**Context:** The bitácora `add-to-project` workflow used `gh project item-add 1 --owner mlorentedev` to add PRs to the board (the `actions/add-to-project` action only handles issues, not PRs).

**Problem:** The step failed in CI with `unknown owner type` even though the PAT was valid and the project existed. Classic PATs resolve `--owner`; fine-grained PATs don't expose the owner type the CLI needs.

**Solution:** Use `actions/github-script` with the Projects v2 GraphQL mutation instead — the same pattern `bitacora-status.yml` already used: `addProjectV2ItemById(input: { projectId, contentId })`. It's idempotent, so re-runs are safe.

**Why:** The CLI path works interactively (classic PAT) but breaks in CI with a fine-grained PAT. GraphQL via github-script is the portable, proven path. Keep project ID (`PVT_kwHOAM7xJs4BZ6GY`) and content ID (`context.payload.pull_request.node_id`) in the script — never shell-expand URLs (template injection, zizmor flags it).

**Tags:** `#ci` `#github-projects` `#fine-grained-pat` `#graphql` `#bitacora`

---
