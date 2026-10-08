# Rollback Runbook — release 0.5.1

## Triggers

- The plugin fails to install or start on the self-hosted Kandev.
- Existing GitHub / GitLab / Bitbucket token connections stop working after the upgrade (card shows "error", "Test" fails).
- The CLI method leaks a token anywhere (log, error, UI) — treat as a security incident.

## Steps (admin of the self-hosted Kandev)

1. In the GitHub Release `v0.5.1`, mark it as broken at the top of the notes (do not delete the tag or the Release — a released tag is never deleted or overwritten).
2. Reinstall `v0.5.0` on the Kandev server (Settings > Plugins, From URL of the `v0.5.0` Release package, or upload it).
3. Providers connected with the CLI in 0.5.1: 0.5.0 ignores the field `source` and still sees the provider as connected, but there is no stored token, so its calls fail (token required). Save a typed token for them (or Remove, then save). Mappings, links, queries and watches are kept.
4. Token connections need no action: 0.5.1 kept the stored secret format unchanged.
5. If a token leaked: revoke it on GitHub/GitLab (`gh auth refresh` / regenerate the PAT), then continue with step 1.
6. Fix forward with a new patch release (`v0.5.2`) through the normal pull request and tag flow.

## Verification after rollback

- Settings > Integrations shows Nulab Backlog 0.5.0.
- Source control cards show "connected" for token connections; "Test" succeeds.

## Escalation

Single-maintainer project: the repository owner (`khuongdo`) handles the rollback and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry pull request.
