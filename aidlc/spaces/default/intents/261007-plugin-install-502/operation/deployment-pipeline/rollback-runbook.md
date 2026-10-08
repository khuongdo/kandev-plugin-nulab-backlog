# Rollback Runbook — 261007-plugin-install-502 (release v0.4.2)

## Triggers

- `v0.4.2` fails to install on the self-hosted Kandev, or the smoke check in `deployment-strategy.md` fails.
- A user reports that 0.4.2 breaks their (non-Windows) Kandev host.

## Procedure

1. **Mark the release.** Edit the GitHub Release `v0.4.2` notes: add "Known problem: <summary>. Use v0.4.1." Do not delete or move the tag (project rule: a released tag is never deleted or overwritten).
2. **Reinstall the previous version** on the self-hosted Kandev: **Settings > Plugins > Install plugin > From URL** with `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v0.4.1/nulab-backlog-0.4.1.tar.gz`. Use From URL, not upload: the 29.5 MB 0.4.1 package is the one that hit the 30-second upload limit. If Kandev refuses a downgrade because the version already exists, uninstall `nulab-backlog` first, then install 0.4.1; Backlog stays off until an admin turns it on again.
3. **Verify**: Settings > Plugins shows `nulab-backlog` `0.4.1` active; the Backlog issue list loads.
4. **Fix forward** with a new patch release (`0.4.3`) through the normal pipeline.

## Windows Hosts

A Windows-hosted Kandev cannot install 0.4.2 by design (no Windows executable). Such hosts stay on 0.4.1; this is a documented support change, not a rollback trigger.

## Data

The release changes no stored data, settings or permissions, so a rollback needs no data migration.
