# CD Configuration — 261007-plugin-install-502 (release v0.4.2)

The delivery pipeline already exists and does not change for this intent. This document records how the fix flows through it and the release inputs it adds (Q1, Q2 in `deployment-pipeline-questions.md`).

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package`, `go mod tidy` check, packaged-host contract job on Kandev `min_kandev_version` (0.96.0) | All required checks green; `main` is protected |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.4.2` on `main` → `.github/workflows/release.yml` | `verify` (full make chain + `release-preflight TAG=v0.4.2`), `contract` (contract test on the exact package bytes), `publish` (provenance attestation + `gh release create` with package and `checksums.txt`) | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install **From URL** of the released package | Manual |

The workflows build with `make build`, which now uses the 4-platform `PLATFORMS` list, so CI and release produce the same 4-executable package. No workflow file changes are needed: no workflow names the Windows executable (checked during Code Generation review).

## Release Inputs Added by This Intent

1. **Base.** This branch is based on `1819cc3` (v0.4.1); `origin/main` is at `bf20039` (PR #12, AI-DLC records for v0.4.1, no code). Rebase onto `origin/main` before opening the pull request (expected conflicts only under `aidlc/`, e.g. `intents.json` and the shared code KB) and let CI re-run. Re-check `gh release list` and `origin/main` immediately before tagging (project rule).
2. **Version (Q1 = `0.4.2`).** Change `manifest.yaml` `version` from `0.4.1` to `0.4.2` in this change set; `release-preflight` refuses `v0.4.2` until `main` carries that version.
3. **Release notes (Q2 = README + Release).** Add under README `## Upgrade notes`, and use the same text as the GitHub Release body for `v0.4.2`:

   ```markdown
   ### 0.4.2: smaller package, install From URL

   - The package is smaller (about 23 MB instead of 29.5 MB): it now carries server
     executables for Linux and macOS on amd64 and arm64 only.
   - **Windows servers are no longer supported.** Kandev running on Windows cannot install
     or upgrade to 0.4.2; stay on 0.4.1 there.
   - Install **From URL** (Settings > Plugins > Install plugin) with the GitHub Release
     package URL. Uploading the file still works, but Kandev stops reading an upload after
     30 seconds, which shows as `Plugin install failed: 502` on slow connections (see
     "Troubleshooting" in the README).
   - Nothing else changes: no data, setting or permission changes.
   ```

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks + contract test on Kandev 0.96.0 |
| GitHub Release | `nulab-backlog-0.4.2.tar.gz` + `checksums.txt` + provenance | Tag `v0.4.2` on `main` | `release.yml` verify/contract/publish; `gh attestation verify` |
| Self-hosted Kandev (v0.97.0, "production") | The released package | Manual install From URL by the maintainer | Smoke check in `deployment-strategy.md` |
| Marketplace | Registry entry for `0.4.2` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The change is packaging and documentation only; there is no runtime behaviour to toggle.
