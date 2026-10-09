# Rollback Runbook: v0.6.1

## When to Roll Back

- After installing 0.6.1, "+ Task" fails to open Kandev's task dialog, created tasks do not link, or the Backlog page fails to render.
- The plugin fails to start or install on Kandev 0.96.0 (e.g. the new permission is refused).

## Steps

1. On the self-hosted Kandev, uninstall 0.6.1 and reinstall the previous release `v0.6.0` package from its GitHub Release (verify with `checksums.txt` and `gh attestation verify`).
2. Smoke check 0.6.0: Backlog page renders; issue list loads (the original "no workflow" bug returns only when `/backlog` is opened directly — acceptable while rolled back).
3. Mark the 0.6.1 GitHub Release as broken in its notes (never delete or overwrite the tag — project Forbidden rule).
4. Fix forward with a new patch release (0.6.2) through the normal PR → tag flow.

## Data Considerations

No stored data or settings change between 0.6.0 and 0.6.1, so a downgrade needs no migration. The extra `workflows` read permission granted for 0.6.1 is harmless if left approved.

## Escalation

Single-maintainer project: the maintainer (khuongdo) owns the rollback decision; Kandev marketplace maintainers are contacted only if the registry PR was already merged.
