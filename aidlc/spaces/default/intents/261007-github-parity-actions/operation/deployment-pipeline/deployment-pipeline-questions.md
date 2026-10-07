# Deployment Pipeline Questions — 261007-github-parity-actions

Context: the release pipeline already exists, and the team's Deployment practice in `team.md` already settles strategy, gates, approvals, rollback and feature flags. The flow is: tag `vX.Y.Z` on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request → manual install on the self-hosted Kandev. Rollback means reinstalling the previous version and fixing forward. There are no feature flags. `v0.1.1` is the latest release. Code Generation already set `manifest.yaml` to `0.2.0` and added a README "Upgrade notes → 0.2.0" section. Only release-specific questions are asked here.

## Q1. Which version should this change be released as?

The change adds features: quick actions, default queries, saved issue queries, and 7 new actions. It also changes the UI: a scope bar replaces the tabs, and the issue row's "Create task" item is removed.

A. `0.2.0` — minor: new features and UI changes (already set in `manifest.yaml`)
B. `0.1.2` — patch
X. Other (please specify)

[Answer]: A

## Q2. How should users be told about the changed UI?

The issue row's "Create task" item is gone, replaced by the "+ Task" menu. The PR toolbar's saved-query dropdown moved to the scope bar's Saved menu, and the lists now open on a default query.

A. The "Upgrade notes" section in the README (already written) and the same text in the GitHub Release notes
B. README only
X. Other (please specify)

[Answer]: A
