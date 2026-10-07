# Rollback Runbook — 261007-uiux-github-style (release v0.1.1)

## When to Roll Back

- The plugin fails to start on the self-hosted Kandev after installing `v0.1.1`.
- Any post-install smoke check step in `deployment-strategy.md` fails.
- A defect in the new issue watch creates unexpected tasks, or the settings screen or `/backlog` is unusable.

## Steps

1. **Pause issue watches if they misbehave** (fast containment, no reinstall): in Settings > Integrations > Backlog, pause each issue watch, or switch the integration off for the workspace with its enable switch (the connection is kept).
2. **Reinstall the previous version**: download `nulab-backlog-0.1.0.tar.gz` and `checksums.txt` from the `v0.1.0` GitHub Release, verify (`sha256sum -c checksums.txt`, `gh attestation verify`), and install it on the self-hosted Kandev in place of `0.1.1`.
3. **Verify the rollback**: the plugin shows `0.1.0`; the three old Integrations entries (Backlog, Watches, Dashboard) are back; existing PR watches, saved queries and issue links are listed as before.
4. **Mark the release**: edit the `v0.1.1` GitHub Release notes to say it is withdrawn and why. Do not delete or move the `v0.1.1` tag (project rule).
5. **Fix forward**: fix on a branch, pass CI, and release a new patch version (`v0.1.2`) through the normal pipeline.

## Data Notes

- Rolling back to `0.1.0` leaves the new issue-watch documents in the plugin state store; `0.1.0` ignores them, and reinstalling `0.1.1`+ picks them up again.
- Tasks already created by issue watches stay in Kandev with their issue links; `0.1.0` shows them as ordinary linked tasks.
- Saved queries saved with a `creator` filter keep that field; `0.1.0` ignores it.

## Contacts and Escalation

Single maintainer project: the maintainer owns rollback and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry repository.
