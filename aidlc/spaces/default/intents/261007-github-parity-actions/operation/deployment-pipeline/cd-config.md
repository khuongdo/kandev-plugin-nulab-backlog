# CD Configuration — 261007-github-parity-actions (release v0.2.0)

The delivery pipeline already exists and does not change for this intent. Express scope skipped CI Pipeline and Infrastructure Design, so this document is based on the workspace's existing `.github/workflows/ci.yml` and `release.yml`, the team Deployment practice, and the Build and Test results in `construction/build-and-test/`. It records how this change goes through that pipeline, plus the two release inputs answered in Q1 and Q2.

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package`, the `go mod tidy` check, and the packaged-host contract job on Kandev `min_kandev_version` (0.96.0) | All required checks green. `main` is protected, so no direct push |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.2.0` on `main` → `.github/workflows/release.yml` | `verify` job: the full make chain plus `release-preflight TAG=v0.2.0`. `contract` job: `verify-package contract-test` on Kandev 0.96.0. `publish` job: provenance attestation and `gh release create` with the package and `checksums.txt` | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Updated entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install of the released package | Manual |

Every action is pinned to a full commit SHA. Workflows default to `permissions: contents: read`, and only `publish` gets `id-token: write` and `attestations: write`. No pipeline change is needed for the new code: the 7 new actions are covered by the existing manifest and contract checks, which passed locally 10/10.

## Release Inputs Added by This Intent

1. **Version (Q1 = `0.2.0`)**: `manifest.yaml` already says `version: "0.2.0"` in the change set. `release-preflight` accepts `v0.2.0` only once `main` carries that version.
2. **Upgrade notes (Q2 = README + Release notes)**: `README.md` § "Upgrade notes → 0.2.0" is already in the change set. The GitHub Release body for `v0.2.0` carries the same text:
   - Rows on `/backlog` get a **+ Task** quick action menu (Implement, Investigate, Reproduce for issues; Review, Address feedback, Fix CI for PRs). The actions can be edited in Settings > Integrations > Backlog > Quick actions. The issue row's old **Create task** item is gone, and **Link to task** stays in the row menu.
   - A scope bar replaced the Issues / Pull requests tabs. Saved queries are picked from its **Saved** menu (the saved-query dropdown in the PR toolbar was removed), and issue filters can now be saved too.
   - The lists open on a default query ("Assigned to me, open" / "Open, assigned to me") instead of an empty filter. Star a saved query to make it the default.
   - Existing saved PR queries, links and watches keep working. No data migration is needed.

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks and the contract test on Kandev 0.96.0 |
| GitHub Release | `nulab-backlog-0.2.0.tar.gz`, `checksums.txt`, provenance | Tag `v0.2.0` on `main` | `release.yml` verify/contract/publish; users run `gh attestation verify` |
| Self-hosted Kandev ("production") | The released package | Manual install by the maintainer | Smoke check in `deployment-strategy.md` |
| Marketplace | Registry entry for `0.2.0` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The team does not use feature flags. The per-workspace enable switch is product behaviour, not a release flag.

## Security Implications

No IAM, network, secret or encryption setting changes. The new actions store user text (prompts, saved queries) in plugin state with input limits and `authenticated` access. They are covered by the leak/redaction test and gosec ([security-test-instructions.md](../../construction/build-and-test/security-test-instructions.md)).
