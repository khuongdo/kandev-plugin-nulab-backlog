# Deployment Pipeline Questions — 261008-ci-path-filter

Context: the team's Deployment practice already settles strategy, gates, approvals and rollback (pull request to `main` with CI green, squash merge, deliberate `vX.Y.Z` tag on `main` → `release.yml` → GitHub Release → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). This change touches only CI (`.github/workflows/`, `internal/ci`, `README.md`); it does not change the plugin package or `manifest.yaml`. So only change-specific questions are asked.

Checked before asking (2026-10-08): the latest GitHub release is `v0.4.2` (2026-10-07T23:52Z); `origin/main` is at `3a983ab` (squash of #13, v0.4.2). This branch sits on `f5a7529`, the pre-squash commit of the same v0.4.2 change, so it must be rebased onto `origin/main` before the pull request. The working tree also holds uncommitted AI-DLC records of the previous intent (`261007-plugin-install-502` Deployment Execution).

Answers recorded 2026-10-08 in guided mode.

## Question 1
Should this change produce a plugin release? It changes no plugin code, manifest or package; only CI workflows, a CI helper under `internal/ci`, and the README CI section.

A. No release: merge the pull request only; the next real plugin change releases as usual
B. Release a patch `0.4.3` anyway
X. Other (please specify)

[Answer]: A

## Question 2
After the pull request has run `secret-scan` once, the `main` ruleset must add `secret-scan` as a required check (requirement FR5.1). This is a change to your GitHub repository settings, outside the code. Who applies it?

A. I run the `gh api` command from `code-summary.md` after `secret-scan` has reported on the pull request, and show you the resulting required-check list
B. You apply it yourself in GitHub settings (or with that command)
X. Other (please specify)

[Answer]: A

## Question 3
The uncommitted AI-DLC records of the previous intent (`261007-plugin-install-502` Deployment Execution) are still in the working tree. Where should they go?

A. A separate records-only pull request after this change merges; it also serves as the first real check that a records-only pull request skips `checks`/`packaged-host-contract` and is still mergeable
B. Include them in this pull request
X. Other (please specify)

[Answer]: A
