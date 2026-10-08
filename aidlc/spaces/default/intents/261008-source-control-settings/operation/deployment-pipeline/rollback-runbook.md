# Rollback Runbook: v0.6.0

## Triggers

- Release workflow fails after the tag (no Release published).
- Smoke check fails on the self-hosted server (plugin does not start, Source control section broken, PR list for the active service fails).
- Users report that a needed service or Backlog Git access stopped working in a way the upgrade notes do not explain.

## Procedure

1. Do not delete or move the `v0.6.0` tag (project Forbidden rule).
2. Mark the v0.6.0 GitHub Release as broken in its notes (and as pre-release if it is not yet referenced by the marketplace).
3. Reinstall v0.5.3 on the self-hosted Kandev server from its GitHub Release.
4. Data compatibility: v0.6.0 only adds an optional `active` field to the SCM settings document (schema version stays 1). v0.5.3 ignores the unknown field, so tokens, mappings, queries, watches and links remain usable after downgrade; all connected services work again as before.
5. If the marketplace registry PR was merged, open a PR pointing the entry back to 0.5.3.
6. Fix forward with a patch release (0.6.1) through the normal PR -> tag flow.

## Verification After Rollback

- Plugin reports 0.5.3 and starts.
- Settings > Source control shows the previous multi-service layout; PR lists and Backlog Git access work.

## Notes

Re-upgrading to a fixed 0.6.x restores the stored `active` value if one was saved, because the field survives in the settings document.
