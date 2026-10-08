# Cross-Unit Traceability — CI path filter

Zero-Unit `express` work: one stage-level traceability file, `construction/code-generation/traceability.json`. No user stories (User Stories skipped), so only `FR`/`NFR` IDs from `inception/requirements-analysis/requirements.md` are enumerated.

**Verdict: PASS with one deferred item** — 16 of 17 IDs covered `OK` with existing target files; FR5.1 is `Deferred` (manual ruleset change outside the repo).

| ID | Status | Owning stage / Unit | Target file | Exists |
|----|--------|---------------------|-------------|--------|
| FR1.1 | OK | code-generation / stage-level | `internal/ci/changes.go` | yes |
| FR1.2 | OK | code-generation / stage-level | `internal/ci/changes.go` | yes |
| FR1.3 | OK | code-generation / stage-level | `internal/ci/changes.go` | yes |
| FR2.1 | OK | code-generation / stage-level | `.github/workflows/ci.yml` | yes |
| FR2.2 | OK | code-generation / stage-level | `.github/workflows/ci.yml` | yes |
| FR2.3 | OK | code-generation / stage-level | `.github/workflows/ci.yml` | yes |
| FR2.4 | OK | code-generation / stage-level | `internal/ci/changes_test.go` | yes |
| FR3.1 | OK (with known limitation) | code-generation / stage-level | `.github/workflows/secrets.yml` | yes |
| FR3.2 | OK | code-generation / stage-level | `.github/workflows/secrets.yml` | yes |
| FR3.3 | OK | code-generation / stage-level | `internal/ci/workflows_test.go` | yes |
| FR4.1 | OK | code-generation / stage-level | `.github/workflows/release.yml` (unchanged) | yes |
| FR5.1 | Deferred | manual, after merge | `gh api` command in `code-generation/code-summary.md` | n/a |
| NFR1 | OK | code-generation / stage-level | `internal/ci/workflows_test.go` | yes |
| NFR2 | OK | code-generation / stage-level | `.github/workflows/ci.yml` | yes |
| NFR3 | OK | code-generation / stage-level | `code-generation/code-summary.md` | yes |
| NFR4 | OK | code-generation / stage-level | `internal/ci/changes_test.go` | yes |
| NFR5 | OK | code-generation / stage-level | `internal/ci/changes_test.go` | yes |

## Findings for the Gate

- **FR5.1 deferred**: add `secret-scan` to ruleset 24580280 after it has reported once on the pull request.
- **FR3.1 limitation**: `secret-scan` runs on every change, but the scanner does not read `aidlc/` or most of `docs/`, so the FR3 acceptance case for a credential under `aidlc/` is not met. Accepted at the Code Generation gate.
