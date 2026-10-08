# Smoke Test Results — CI path filter

Success criteria from `operation/deployment-pipeline/deployment-strategy.md`.

| Check | Expected | Observed | Evidence | Result |
|-------|----------|----------|----------|--------|
| App pull request (#14) | `changes` runs; `checks`, `packaged-host-contract`, `secret-scan` all run and pass | All four jobs pass | Actions runs 37714280452 (`ci`), 37714280624 (`secrets`) | PASS |
| App push to `main` (`dd89cc2`) | Full CI runs and passes | `changes`, `checks`, `packaged-host-contract` success; `secret-scan` success | Runs 37714637744 (`ci`), 37714637558 (`secrets`) | PASS |
| Records-only pull request (#15) | `checks` and `packaged-host-contract` skipped and reported as passing; `secret-scan` passes; mergeable without bypass | `checks` skipping, `packaged-host-contract` skipping, `changes` pass, `secret-scan` pass; `mergeStateStatus: CLEAN`, merged | Run 37714993471 (`ci`), 37714993483 (`secrets`) | PASS |
| Records-only push to `main` (`a23845f`) | Heavy jobs skipped; `secret-scan` runs | `checks` skipped, `packaged-host-contract` skipped, `secret-scan` success | `gh run list --commit a23845f…` | PASS |
| Ruleset requires `secret-scan` | Required checks are `checks`, `packaged-host-contract`, `secret-scan` | Exactly those three | `gh api …/rulesets/24580280` read-back | PASS |
| Release workflow untouched | `release.yml` unchanged; no run | No `release` run triggered; file unchanged | `git diff 3a983ab dd89cc2 -- .github/workflows/release.yml` empty | PASS |

All smoke checks pass.
