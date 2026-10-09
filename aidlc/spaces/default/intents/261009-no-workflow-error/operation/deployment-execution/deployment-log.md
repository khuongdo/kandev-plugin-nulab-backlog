# Deployment Log - 261009-no-workflow-error (release v0.6.1)

Scope (Q1 = A): up to the GitHub Release. The self-hosted Kandev install and the marketplace step are done by the user.

## Timeline (UTC, 2026-10-09)

| Step | Result |
|---|---|
| Version | `manifest.yaml` `0.6.0` -> `0.6.1`; README heading "Next release: ..." -> `### 0.6.1: "+ Task" works when the Backlog page is opened directly` (Deployment Pipeline Q2=A). `make test` and `make package verify-package` re-run green with 0.6.1 |
| Commit | `480fbdc` on `feature/fix-kandev-has-no-wo-ik5` (code, tests, README, AI-DLC records); base `origin/main` `3803248`, no rebase needed |
| Push + pull request | as `khuongdo` (active `gh` account): PR #28 https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/28 |
| CI | `changes` pass (29s), `checks` pass (2m18s), `packaged-host-contract` pass (58s), `secret-scan` pass (37s) |
| Merge | squash-merged -> `main` at `6f152d8` |
| Pre-tag re-check | latest release still `v0.6.0`; no `v0.6.1` tag on the remote; `main` carries `version: "0.6.1"` |
| Tag | annotated `v0.6.1` on `6f152d8`, pushed ~11:20 |
| `release.yml` run 37923098232 | `verify` success, `contract` success, `publish` success |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.6.1 - `nulab-backlog-0.6.1.tar.gz` (23,588,696 bytes), `checksums.txt`; notes start with the README 0.6.1 upgrade note, generated PR list below (`gh release edit`) |
| Release package check | downloaded assets: `sha256sum -c` OK, `verifypkg` OK (`nulab-backlog@0.6.1`), `gh attestation verify` exit 0 |

## Post-Release: `main` CI Failure and Fix

| Step | Result |
|---|---|
| `ci` on `main` after the merge (run 37923081764) | failed in `checks`: `TestMyselfReturnsCancellationUnchanged` (`internal/backlog`) got a `body` error instead of `context.Canceled`. Pre-existing flaky test, unrelated to the 0.6.1 change; the v0.6.1 release run passed, so released artefacts are unaffected |
| Root cause | an empty 200 from the fake server won the race against the caller's cancellation and `attempt()` treated it as success |
| Fix | `attempt()` returns `context.Canceled` when the caller cancelled, even if a response arrived; deterministic regression test with a cancelling transport (red before, green 200/200 after) |
| PR #29 | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/29 — CI green; squash-merged to `main` at `c14628c` (user chose merge without a new release) |
| `ci` on `main` (run 37925058796) | success — `changes`, `checks`, `packaged-host-contract` all green |

## Database Migrations

None. No stored data, settings or secret format changed.

## Not Done (by choice, Q1 = A)

- Install of the released 0.6.1 package on the self-hosted Kandev (the user installs and approves the new `workflows` read permission). The same code was already installed and checked manually on real Kandev v0.96.0 during Build and Test (pre-release 0.6.0-numbered build).
- Marketplace registry step (the user handles it).

## Rollback

Per `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.6.0`; never move or delete the tag; fix forward with 0.6.2).
