# Deployment Strategy — CI path filter

## Strategy

Trunk-based merge to `main`, no plugin release (team Way of Working and Deployment; Q1 = A). The "deployment" is the CI configuration taking effect on `main`: GitHub Actions reads workflow files from the commit under test, so the new behaviour applies to the pull request itself and to every later pull request and `main` push.

## Why No Release

The plugin package is unchanged: no plugin code, `manifest.yaml`, UI bundle or executables changed. `release.yml` and its checks are unchanged. Releasing would publish a byte-equivalent package under a new version.

## Gates

1. Pull request CI green: `checks`, `packaged-host-contract`, `secret-scan`.
2. Ruleset update (Q2 = A), applied by the agent once `secret-scan` has reported on the pull request, before merge, so `main` immediately requires it. Shown to you as the read-back required-check list.
3. Squash merge.

## Success Criteria (Smoke Checks)

- Pull request run: `changes` job log shows `app=true`; all three checks pass.
- `main` push run after merge: `changes` shows `app=true`; full CI green.
- First records-only pull request (Q3 = A): `changes` shows `app=false`; `checks` and `packaged-host-contract` are skipped; `secret-scan` passes; GitHub shows the pull request as mergeable without bypass.

## Risk Notes

- If the `changes` job ever fails, `checks` still runs (fail-safe), so a classifier failure costs CI time, never coverage.
- Known limitation (accepted at Code Generation): the credential scanner does not read `aidlc/` or most of `docs/`.
