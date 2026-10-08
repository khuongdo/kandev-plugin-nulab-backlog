# Rollback Runbook — 261007-source-control-agnostic (release v0.4.0)

## When to Roll Back

- The plugin fails to start on the self-hosted Kandev after installing `v0.4.0`.
- Smoke check step 2 fails: an existing Backlog feature (issues, Backlog Git PRs, saved queries, watches) regresses.
- A provider token or provider response content appears in logs, errors or the UI (security incident: roll back first, then rotate the affected token at the provider).

A problem limited to one new provider (for example a GitHub API difference) does not need a rollback: an admin removes that provider's token in Settings > Source control, which stops all calls to it, and the fix ships as a patch.

## Steps

1. **Fast containment (no reinstall)**: remove the affected provider's token in Settings > Source control; or turn Backlog off for the workspace (all `scm.*` actions are refused while it is off).
2. **Reinstall the previous version**: download `nulab-backlog-0.3.0.tar.gz` and `checksums.txt` from the `v0.3.0` GitHub Release, verify (`sha256sum -c checksums.txt`, `gh attestation verify`), and install it in place of `0.4.0`.
3. **Verify the rollback**: the plugin shows `0.3.0`; Backlog issues, Backlog Git PR links, watches and saved queries work as before; "Git access" is back in its own section.
4. **Mark the release**: edit the `v0.4.0` GitHub Release notes to say it is withdrawn and why. Do not delete or move the `v0.4.0` tag (project rule).
5. **Fix forward**: fix on a branch, pass CI, release `v0.4.1` through the normal pipeline.

## Data Notes

- `v0.4.0` never modifies v0.3.0 documents or secrets, so `0.3.0` reads them unchanged after a rollback.
- The new `scm.*` documents and `backlog.scm.<provider>.<ws>` secrets stay in the Kandev store, unused by `0.3.0`, and are picked up again when `0.4.x` is reinstalled. To remove the tokens completely, remove each provider in Settings > Source control **before** rolling back, and revoke the tokens at the providers.

## Contacts and Escalation

Single-maintainer project: the maintainer owns rollback, token revocation and the fix-forward release. Kandev marketplace issues go to the Kandev maintainers through the registry repository.
