# Health Check Report - v0.5.2

## Release Pipeline Health

| Item | Status | Evidence |
|------|--------|----------|
| PR #21 required checks | Healthy | `checks`, `packaged-host-contract`, `secret-scan` and `changes` all pass |
| `main` after the merge | Healthy | `abee1f4`, `manifest.yaml` version `0.5.2` |
| `release.yml` on `v0.5.2` | Healthy | Run 37731187049: `verify`, `contract` and `publish` all succeeded |
| GitHub Release | Healthy | v0.5.2 published with package, checksums and provenance attestation (`publish` job) |

## Runtime Health

The plugin's runtime health on the self-hosted Kandev is not checked here, because release scope B leaves the install to you. After installing, the health check is: Settings > Plugins shows nulab-backlog 0.5.2 with status running, and the Kandev server log shows no plugin start errors. Then run the smoke checks in `smoke-test-results.md`.

## Known Advisory Items (not blockers)

- **R-02:** the issue-side dialog can show an error after a link that actually succeeded, if the toast or the row update throws. A fix-forward candidate.
- **R-01:** on a cold start, the first Link menu open can omit "Link Backlog issue" until the links store has loaded.
