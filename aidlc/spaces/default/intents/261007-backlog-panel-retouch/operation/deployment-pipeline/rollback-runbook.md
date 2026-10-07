# Rollback Runbook — 261007-backlog-panel-retouch (release v0.4.1)

## When to Roll Back

- The plugin fails to start, or the `/backlog` page, a pull-request list or a Kanban card fails to render, on the self-hosted Kandev after installing `v0.4.1`.
- Any post-install smoke check step in `deployment-strategy.md` fails in a way that blocks users (for example, the issue or pull-request list cannot be filtered or searched, the provider selector does not switch lists, or linked tasks cannot be opened).

## Steps

1. **Reinstall the previous version**: download `nulab-backlog-0.4.0.tar.gz` and `checksums.txt` from the `v0.4.0` GitHub Release, verify (`sha256sum -c checksums.txt`, `gh attestation verify`), and install it on the self-hosted Kandev in place of `0.4.1`.
2. **Verify the rollback**: the plugin shows `0.4.0`; the old toolbars (auto-search, labelled filters) and plain task links are back; saved queries (Backlog and provider), issue links, watches and provider connections are listed as before.
3. **Mark the release**: edit the `v0.4.1` GitHub Release notes to say it is withdrawn and why. Do not delete or move the `v0.4.1` tag (project rule).
4. **Fix forward**: fix on a branch, pass CI, and release a new patch version (`v0.4.2`) through the normal pipeline.

## Data Notes

- This release writes no new data and changes no stored format, so `0.4.0` reads everything `0.4.1` saved; rollback has no data step.

## Contacts and Escalation

Single maintainer project: the maintainer owns rollback and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry repository.
