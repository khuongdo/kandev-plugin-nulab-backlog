# Deployment Pipeline Questions — 261009-no-workflow-error

The team's Deployment practice (`team.md` § Deployment) already settles strategy, gates, approvals, rollback and feature flags: tag `vX.Y.Z` on `main` → `release.yml` (checks, package verification, GitHub Release, provenance attestation) → marketplace registry PR → manual install on the self-hosted Kandev. Only release-specific questions are asked here.

Context checked 2026-10-09: latest GitHub release `v0.6.0` (2026-10-08); `origin/main` at `3803248` with `manifest.yaml` version `0.6.0`; no newer release.

## Question 1
Which version should this release use? The change is a bugfix plus an error-display retouch, but it adds a new manifest capability (`api_read: workflows`) and a new backend action.

A. 0.6.1 — patch: it is a bugfix; the new read permission only serves the fix (matches how 0.5.2 treated a small new action as a patch)
B. 0.7.0 — minor: a new capability/permission changes what admins approve on upgrade
X. Other (please specify)

[Answer]: A

## Question 2
The README "Next release" note (new `workflows` read permission, "+ Task" fix, error-display retouch) will be renamed to the chosen version and put at the top of the GitHub Release notes. Is that note enough?

A. Yes, use the README note as is
B. Add more (please specify)
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Q1=A: release version 0.6.1 (patch).
- Q2=A: the README "Next release" note is renamed to 0.6.1 and used as-is at the top of the GitHub Release notes.

Does this all look correct before I generate the deployment pipeline artifacts?

Looks correct / Request changes

[Answer]: Looks correct
