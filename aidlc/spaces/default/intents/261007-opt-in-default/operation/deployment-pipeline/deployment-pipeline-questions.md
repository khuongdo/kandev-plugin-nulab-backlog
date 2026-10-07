# Deployment Pipeline Questions — 261007-opt-in-default

Context: the release pipeline already exists and the team's Deployment practice settles strategy, gates, approvals and rollback (tag `vX.Y.Z` on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). How users learn about the upgrade effect is already decided (requirements FR5.1/FR5.2: README note plus release notes). `v0.1.1` is released and this branch's `manifest.yaml` says `0.1.1`, so this fix needs a new version. (Correction found during confirmation: `v0.2.0` was released on `main` while this work ran, so the next version must be above `0.2.0`.) Answer by writing the letter after the answer tag.

## Q1. Which version should this fix be released as?

The change is a bug fix, but it changes default behaviour: after upgrading, workspaces that never touched the switch turn off until an admin turns them on. For a 0.x plugin, semver allows a patch for a fix, while a behaviour change that existing users notice is often shipped as a minor bump.

A. `0.1.2` — patch: it is a bug fix
B. `0.2.0` — minor: the default behaviour changes for existing users
X. Other (please specify)

[Answer]: X. Other: 0.3.0 (v0.2.0 is already released on main; changed via Requested Changes Feedback)

## Consolidated Summary Confirmation

- Version (Q1): release this fix as `0.3.0`; `manifest.yaml` moves to `0.3.0` and the release tag is `v0.3.0`. `v0.2.0` is already the latest release on `main`.
- Upgrade notes (already decided, FR5.1/FR5.2): the README note becomes "Upgrading from v0.2.0 or earlier" under a `0.3.0` heading, and the same text goes into the GitHub Release notes for `v0.3.0`.
- Before release, this branch is rebased onto `origin/main` (v0.2.0) and the opt-in change is re-verified against the v0.2.0 code (new quick actions must go through the switch guard).
- Everything else follows the team's Deployment practice unchanged: pull request to `main` with CI green, squash merge, deliberate `v0.3.0` tag, `release.yml` re-runs all checks and package verification, GitHub Release with checksums and provenance, marketplace registry pull request, manual install on the self-hosted Kandev, rollback by reinstalling `v0.2.0` and fixing forward; no feature flags.

Does this all look correct before I generate the deployment pipeline artifacts?

- Looks correct
- Request changes

[Answer]: Looks correct

## Requested Changes Feedback

What should change?

[Answer]: 0.3.0
