# Deployment Log — 261007-backlog-panel-retouch (release v0.4.1)

Plan: `operation/deployment-pipeline/cd-config.md`, `deployment-strategy.md`. Decisions: `deployment-execution-questions.md` (Q1 = B: PR → CI → merge and tag only after the manual UI check passes, no marketplace registry pull request; Q2 = B: the human runs the UI check later).

## Pre-deployment Checks

| Check | Result | Evidence |
|---|---|---|
| Build and Test | All commands green; NFR2-UI-IN-HOST accepted as Unverified (closed by the manual UI check below) | `construction/build-and-test/test-results.md` |
| Latest release / main before PR | `v0.4.0` latest; `origin/main` = `ad4adcf` (PR #10, records only) | `gh release list`, `git log origin/main` (2026-10-08) |
| Data migration | None | — |
| Dependent services / window | None | — |

## Execution

| Time (UTC) | Step | Result |
|---|---|---|
| 2026-10-08 | Commit code and AI-DLC records on `feature/refactor-backlog-pan-03d` | `d816740` "GitHub-style issue and pull request lists, version 0.4.1" |
| 2026-10-08 | Rebase onto `origin/main` `ad4adcf` | One conflict in `aidlc/spaces/default/intents/intents.json` (both intents' rows kept; `source-control-agnostic` = complete) |
| 2026-10-08 | Push and open pull request | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/11 |
| 2026-10-08 | CI on the pull request | `checks` pass (2m9s), `packaged-host-contract` pass (56s) — run 37694739794 |
| 2026-10-08 | Manual real-host UI check (NFR2-UI-IN-HOST) | **Passed** — reported by the human |
| 2026-10-08 | Re-check before merge/tag | `v0.4.0` still latest; `origin/main` still `ad4adcf`; PR #11 `MERGEABLE`/`CLEAN` |
| 2026-10-08 | Squash merge PR #11 | `main` = `1819cc3` "GitHub-style issue and pull request lists, version 0.4.1 (#11)"; `manifest.yaml` 0.4.1 |
| 2026-10-08 | Tag `v0.4.1` on `1819cc3`, `release.yml` | Run 37696276123 `success` (verify, contract, publish) |
| 2026-10-08 | GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.4.1 — `nulab-backlog-0.4.1.tar.gz`, `checksums.txt`, provenance; notes = README 0.4.1 text + R-07 line |
| — | Marketplace registry pull request | Not planned (Q1 = B) |
| — | Install `v0.4.1` on the self-hosted Kandev + smoke check | **Pending** (human) |

## Status

Released: `v0.4.1` is merged, tagged and published. Remaining: the human installs `v0.4.1` on the self-hosted Kandev and runs the post-install smoke check. AI-DLC Operation records for this release (this directory) go to `main` in a follow-up records pull request, as for v0.4.0 (#10).
