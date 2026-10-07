# Rollback Runbook — 261007-opt-in-default (release v0.3.0)

## When to Roll Back

- The plugin fails to start on the self-hosted Kandev after installing `v0.3.0`.
- Any post-install smoke check step in `deployment-strategy.md` fails (for example, turning the switch on does not restore the workspace).
- Users are blocked and an admin cannot turn the switch on.

## Steps

1. **Fast containment (no reinstall)**: if the only problem is that workspaces are off, an admin turns the switch on per workspace in Settings > Integrations > Backlog. This is the expected upgrade step, not a rollback.
2. **Reinstall the previous version**: download `nulab-backlog-0.2.0.tar.gz` and `checksums.txt` from the `v0.2.0` GitHub Release, verify (`sha256sum -c checksums.txt`, `gh attestation verify`), and install it on the self-hosted Kandev in place of `0.3.0`.
3. **Verify the rollback**: the plugin shows `0.2.0`; workspaces with no switch record are on again (old default); saved connections, watches and saved queries are listed as before.
4. **Mark the release**: edit the `v0.3.0` GitHub Release notes to say it is withdrawn and why. Do not delete or move the `v0.3.0` tag (project rule).
5. **Fix forward**: fix on a branch, pass CI, and release a new patch version (`v0.3.1`) through the normal pipeline.

## Data Notes

- The fix writes no new data. Switch records saved while on `0.3.0` (on or off) are read the same way by `0.2.0`, so explicit choices survive a rollback; only workspaces with no record flip back to on.

## Contacts and Escalation

Single maintainer project: the maintainer owns rollback and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry repository.
