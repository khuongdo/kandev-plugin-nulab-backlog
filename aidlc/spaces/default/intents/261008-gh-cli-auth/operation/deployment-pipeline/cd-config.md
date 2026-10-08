# CD Configuration — release 0.5.1

## Pipeline (unchanged, from the team Deployment practice)

No pipeline file changes in this intent. The existing GitHub Actions workflows are used as they are:

| Workflow | Trigger | What it does |
|---|---|---|
| `.github/workflows/ci.yml` | `pull_request`, push to `main` | format, vet, lint (gosec), test (`-race`), coverage ≥ 80%, package, package verification, packaged-host contract test on Kandev `v0.96.0`; required checks on the protected `main` |
| `.github/workflows/secrets.yml` | `pull_request`, push | credential scan |
| `.github/workflows/release.yml` | push of tag `vX.Y.Z` | runs only for a tag on `main` that equals `manifest.yaml` `version` and has no Release; re-runs every check, package verification and the contract test; publishes the GitHub Release with `nulab-backlog-X.Y.Z.tar.gz`, `checksums.txt` and a build provenance attestation |

All actions are pinned to full commit SHAs; default `permissions: contents: read`; only the publish job has `id-token: write` / `attestations: write`.

## Release-specific changes for 0.5.1

1. `manifest.yaml`: `version: "0.5.1"` (in the feature pull request).
2. Every other place that carries the version, if any (checked in Deployment Execution with `git grep -n '0\.5\.0'` outside `aidlc/` and `internal/plugin/testdata/`).
3. `README.md` → `## Upgrade notes`: add at the top
   `### 0.5.1: connect GitHub or GitLab with the gh / glab CLI login` with the approved text (deployment-pipeline-questions.md Q2).
4. Pull request from `feature/th-m-auth-method-cho-0pe` to `main`, self-merged with squash when CI is green.
5. On `main`: `make package verify-package`, then `git tag v0.5.1 && git push origin v0.5.1`.
6. After the Release: put the README upgrade note at the top of the auto-generated Release notes (`gh release edit v0.5.1`), keeping the generated PR list below (project rule).
7. Marketplace registry pull request for 0.5.1 (team practice), reviewed by the Kandev maintainers.

## Promotion

| Environment | How it gets the version | Gate |
|---|---|---|
| CI (throwaway Kandev v0.96.0) | contract test in `ci.yml` and `release.yml` | all required checks green |
| GitHub Release | tag `v0.5.1` on `main` | creating the tag is the human approval |
| Self-hosted Kandev | manual install of the Release package | admin installs; smoke check below |

## Smoke check after install

- Settings > Integrations > Nulab Backlog shows version 0.5.1.
- Settings > Source control: GitHub and GitLab cards show the "Use gh CLI login" / "Use glab CLI login" button; existing token connections still show "connected".
- If the Kandev server user is logged in with `gh`: press "Use gh CLI login" → card shows "Connected via gh CLI as <account>", "Test" succeeds.
