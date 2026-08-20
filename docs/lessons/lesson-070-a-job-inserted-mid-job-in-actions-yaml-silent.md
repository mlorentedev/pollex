---
id: lesson-070-a-job-inserted-mid-job-in-actions-yaml-silent
type: lesson
status: active
created: "2026-08-08"
owner: manu
tags: [pollex, lesson, ci, github-actions, yaml, false-green, branch-protection]
---

# A job inserted mid-job in Actions YAML silently steals the trailing steps

**Context:** `ci.yml` grew a new `audit` job (#47) while `test-extension` (#45) already existed. Routine review; both PRs green; nobody noticed anything for three days.

**Problem:** The `audit:` block was inserted **between** `test-extension`'s `setup-node` step and its four run steps. GitHub Actions job boundaries are pure indentation — there is no schema error when a sibling job appears mid-job — so those four steps were silently reparented onto `audit`. From then on `test-extension` ran checkout + setup-go + setup-node and stopped: a **permanently green check that asserted nothing**, while `audit` quietly ran the extension suite under a misleading name. No lint, no warning, no failing build. `actionlint` passes on both versions, because both are structurally valid YAML.

**Solution:** Move the steps back under their job. To catch the class rather than the instance, assert the shape instead of reading it:

```bash
python3 -c "
import yaml
d=yaml.safe_load(open('.github/workflows/ci.yml'))
for j,v in d['jobs'].items():
    print(j, len(v.get('steps',[])), [s.get('name') or s.get('uses') for s in v.get('steps',[])])
"
```

**Why:** A green check is trusted in proportion to how boring it is, and this one had no symptom — nothing failed, so nothing drew attention. **Job duration is the cheap tell:** `test-extension` ran 10–17s while broken and 58s once fixed. Treat a suspiciously fast job as suspect. The danger compounds with branch protection: requiring a check that tests nothing converts a silent gap into a documented guarantee. Verify a required check actually executes *before* requiring it (relevant to SEC-002).

**Tags:** `#ci` `#github-actions` `#yaml` `#false-green` `#branch-protection`
