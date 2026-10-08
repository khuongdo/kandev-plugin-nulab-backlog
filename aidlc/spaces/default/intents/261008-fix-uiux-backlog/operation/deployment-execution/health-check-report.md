# Health Check Report - v0.5.0

## Pipeline Health

| Item | Status | Evidence |
|------|--------|----------|
| PR #17 CI (`changes`, `checks`, `packaged-host-contract`, `secret-scan`) | healthy | all pass before merge |
| `main` after merge | healthy | `7682d69`, protected branch, squash merge |
| Release workflow run 37719084025 | healthy | `verify`, `contract`, `publish` success |
| GitHub Release v0.5.0 | healthy | assets present, checksum and provenance verified |

## Runtime Health

The plugin has no hosted service of its own; it runs inside Kandev after a manual install. Runtime health on the self-hosted Kandev (plugin status running, no startup error in the Kandev server log, the smoke checks in smoke-test-results.md) is **pending your install** (release scope B).

The contract job installed and ran the 0.5.0 package on Kandev v0.96.0 with a clean start, which covers the async `initialize` (bounded 3 s ON/OFF check before the Integrations entry) under the host's 10 s initialize limit.

## Verdict

Release healthy up to the GitHub Release. Install and runtime checks on the self-hosted Kandev remain open.
