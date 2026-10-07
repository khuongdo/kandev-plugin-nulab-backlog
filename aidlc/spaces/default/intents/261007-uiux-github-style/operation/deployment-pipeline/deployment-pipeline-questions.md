# Deployment Pipeline Questions — 261007-uiux-github-style

Context: the release pipeline already exists and the team's Deployment practice settles strategy, gates, approvals and rollback (tag `vX.Y.Z` on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). `v0.1.0` is already released and `manifest.yaml` still says `0.1.0`, so this change needs a new version. Answer each by writing the letter after the answer tag.

## Q1. Which version should this change be released as?

The change adds features (issue watch, PR list action, six new actions) and changes the UI (one Integrations entry, removed `/backlog/watches` and `/backlog/dashboard` pages). Semver for a 0.x plugin treats new features as a minor bump.

A. `0.2.0` — minor: new features and UI changes
B. `0.1.1` — patch
C. `1.0.0` — first stable major
X. Other (please specify)

[Answer]: B

## Q2. How should users be told about the removed pages and the new places?

Old bookmarks to `/backlog/watches` and `/backlog/dashboard` stop working; PR watches and saved queries now live in Settings > Integrations > Backlog, and the PR list is a tab on `/backlog`.

A. An "Upgrade notes" section in the GitHub Release notes and in the README
B. Release notes only
C. No special note
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Version (Q1): release this change as `0.1.1`; `manifest.yaml` moves from `0.1.0` to `0.1.1` and the release tag is `v0.1.1`.
- Upgrade notes (Q2): an "Upgrade notes" section in both the GitHub Release notes and the README, saying the two pages were removed and where watches, saved queries and the PR list now live.
- Everything else follows the team's Deployment practice unchanged: pull request to `main` with CI green, squash merge, deliberate `v0.1.1` tag, `release.yml` re-runs all checks and package verification, GitHub Release with checksums and provenance, marketplace registry pull request, manual install on the self-hosted Kandev, rollback by reinstalling `v0.1.0` and fixing forward; no feature flags.

Does this all look correct before I generate the deployment pipeline artifacts?

Looks correct
Request changes

[Answer]: Looks correct
