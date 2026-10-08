# CD Configuration: v0.6.0

## Pipeline (existing, unchanged)

No pipeline change is needed; this release reuses the team's existing GitHub Actions setup.

- `ci.yml` on `pull_request`: format, vet, lint, test (`-race`), coverage floor 80%, package, package verification, packaged-host contract test on Kandev 0.96.0. Every check must be green before the squash merge (protected `main`).
- `release.yml` on tag `vX.Y.Z`: runs only from `main`, refuses an existing tag, re-runs all checks plus package verification, creates the GitHub Release with `nulab-backlog-<version>.tar.gz` and `checksums.txt`, and attaches a build provenance attestation (`id-token: write` / `attestations: write` only on the publish job).
- Actions pinned to full commit SHAs; default `permissions: contents: read`.

## Release-Specific Changes (applied in Deployment Execution)

| Item | Change |
|---|---|
| `manifest.yaml` | `version: "0.6.0"` (Q1=A); `min_kandev_version` unchanged (0.96.0) |
| `README.md` | Heading "Unreleased: one source control service per workspace" -> "0.6.0: one source control service per workspace"; add a warning line that task worktrees cannot fetch or push Backlog Git repositories while GitHub, GitLab or Bitbucket is the active service (Q2=A) |
| GitHub Release notes | README 0.6.0 section at the top, generated PR list below it (project Deployment rule) |
| Marketplace registry | PR updating the entry to 0.6.0 after the Release |

## Environment Promotion Matrix

| Environment | Trigger | Gate |
|---|---|---|
| Pull request (CI) | push to PR branch | all CI checks green |
| `main` | squash merge (self-merge) | protected branch, CI green |
| GitHub Release ("production") | manual tag `v0.6.0` on `main` | `release.yml` checks + package verification |
| Self-hosted Kandev | manual install of the released package | admin installs; smoke check below |
| Kandev marketplace | registry PR | Kandev maintainers review |

## Feature Flags

None. The behaviour change is governed by the per-workspace active-service setting itself, and existing multi-service workspaces stay in the pending state until an admin picks a service.
