# Integration Test Instructions — 261009-no-workflow-error

## Scope

The Minimal test strategy generates no separate integration suite. Two integration-level checks still apply to this bugfix, because the unit tests use a test double for Kandev's `TaskCreateDialog`:

## Packaged-Host Contract Test (automated)

```bash
make contract-test KANDEV_MIN_DIR=../kandev   # run 10 times (project rule, catches host startup races)
```

Expected: 10/10 pass on Kandev v0.96.0 with the new `api_read: workflows` capability.

## Real-Kandev Manual Check (bugfix regression, required before release)

Environment: self-hosted Kandev v0.96.0 with a workspace that has at least one workflow and a connected Backlog space.

1. Install `dist/nulab-backlog-<version>.tar.gz`; approve the new `workflows` permission if asked.
2. Open `/backlog` directly (reload; do not visit the Kanban board first). Pick a "+ Task" quick action on an issue → Kandev's task dialog opens with a selectable workflow and step; no "no workflow" error.
3. Create the task → the issue row shows the linked task badge.
4. At about 360 px and at desktop width, trigger an error (e.g. disable the Backlog connection) → errors appear as a toast or a full-width wrapping alert; no row overflow or horizontal scroll.

Record the result (pass/fail per step, date, Kandev version) in `test-results.md`.
