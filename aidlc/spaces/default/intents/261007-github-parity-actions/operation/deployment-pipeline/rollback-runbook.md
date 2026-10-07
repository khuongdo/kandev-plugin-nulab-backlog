# Rollback Runbook — 261007-github-parity-actions (release v0.2.0)

## When to Roll Back

- The plugin fails to start on the self-hosted Kandev after installing `v0.2.0`.
- Any post-install smoke check step in `deployment-strategy.md` fails.
- `/backlog` or the settings screen is unusable, or quick actions create tasks that are not linked.

## Steps

1. **Contain**, if needed: switch the integration off for the affected workspace with its enable switch. The connection is kept.
2. **Reinstall the previous version**: download `nulab-backlog-0.1.1.tar.gz` and `checksums.txt` from the `v0.1.1` GitHub Release and verify them (`sha256sum -c checksums.txt`, `gh attestation verify`). Then install the package on the self-hosted Kandev in place of `0.2.0`.
3. **Verify the rollback**: the plugin shows `0.1.1`, `/backlog` has the Issues / Pull requests tabs again, and existing PR watches, saved PR queries and issue links are listed as before.
4. **Mark the release**: edit the `v0.2.0` GitHub Release notes to say it is withdrawn and why. Do not delete or move the `v0.2.0` tag (project rule).
5. **Fix forward**: fix on a branch, pass CI, and release a new patch version (`v0.2.1`) through the normal pipeline.

## Data Notes

- After a rollback to `0.1.1`, the new documents `issues.quick_actions` and `issues.queries` stay in the plugin state store. `0.1.1` ignores them, and a later `0.2.x` reinstall picks them up again.
- The `isDefault` field on saved PR queries is ignored by `0.1.1`. Saving a query in `0.1.1` drops the star, and the user can star it again after upgrading.
- Tasks created through quick actions stay in Kandev with their issue or PR links.

## Contacts and Escalation

This is a single-maintainer project: the maintainer owns the rollback and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry repository.
