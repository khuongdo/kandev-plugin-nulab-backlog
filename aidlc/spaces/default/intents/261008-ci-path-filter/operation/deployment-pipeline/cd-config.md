# CD Configuration — CI path filter

## Delivery Path for This Change

No plugin release (Q1 = A). The change is delivered by merging one pull request to `main`; `release.yml` is unchanged and not triggered (no tag).

| Step | Action | Gate |
|------|--------|------|
| 1 | Commit the change on this branch and rebase onto `origin/main` (`3a983ab`); the branch's `f5a7529` is the pre-squash copy of #13 and drops out | Clean rebase, no conflicts |
| 2 | Push the branch, open a pull request to `main` | — |
| 3 | CI on the pull request: `changes` prints `app=true` (the change touches `.github/` and `internal/`), so `checks` and `packaged-host-contract` run in full; `secret-scan` runs | All three green |
| 4 | Add `secret-scan` to ruleset 24580280 with the `gh api` command from `construction/code-generation/code-summary.md` (Q2 = A), then read back the required-check list | List is `checks`, `packaged-host-contract`, `secret-scan` |
| 5 | Squash-merge (self-merge on green CI, team Way of Working) | Required checks green |
| 6 | The `push` to `main` run: `changes` diffs `event.before..github.sha` → `app=true`; full CI runs on `main` | Green |
| 7 | Separate records-only pull request for the `261007-plugin-install-502` Deployment Execution records (Q3 = A) | `checks` and `packaged-host-contract` skipped (reported as passing), `secret-scan` green, pull request mergeable |

## Pipeline Files

- `.github/workflows/ci.yml` — triggers unchanged (`pull_request` and `push` to `main`); new first job `changes`; `checks` gated on `needs.changes.outputs.app != 'false'`; `packaged-host-contract` gated on `needs.checks.result == 'success'`.
- `.github/workflows/secrets.yml` — new; job `secret-scan` on every `pull_request` and `push` to `main`, no path filter.
- `.github/workflows/release.yml` — unchanged (tag-triggered release).

## Required Status Checks (ruleset 24580280)

Before: `checks`, `packaged-host-contract`. After step 4: `checks`, `packaged-host-contract`, `secret-scan` (integration 15368, GitHub Actions). All other ruleset rules are kept as they are.

## Environment Promotion Matrix

| Environment | How it receives a change | Applies to this change |
|-------------|--------------------------|------------------------|
| `main` (trunk) | Squash merge after green required checks | Yes |
| GitHub Release / marketplace | Deliberate `vX.Y.Z` tag → `release.yml` | No (no release) |
| Self-hosted Kandev | Manual install of a release | No (package unchanged) |

## Feature Flags

None (team Deployment practice).
