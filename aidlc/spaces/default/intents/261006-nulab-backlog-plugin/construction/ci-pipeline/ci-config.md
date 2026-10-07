# CI Configuration

## Summary

The pipeline already exists. It was built by the ci-release Unit (U5) and has been passing on `main`: the two `ci` runs for commit `4b0a090` on 2026-10-06 (37540712015, 37541196810) both succeeded. This stage keeps both workflows unchanged ([Q1] A, [Q3] A, [Q4] A) and fixes one gap: protection of `main` on GitHub ([Q2] A).

| Item | Choice | Source |
|------|--------|--------|
| CI tool | GitHub Actions | [Q1] A; team Way of Working |
| Branch strategy | Trunk-based. Short-lived branches into `main` through a pull request, merged with squash by the author once CI is green | [Q2] A; team Way of Working |
| Release trigger | A `vX.Y.Z` tag on `main` | team Deployment |
| Artifact storage | GitHub Actions artifacts for CI; GitHub Releases for published packages | [Q4] A |

## Workflows

### `.github/workflows/ci.yml` (name `ci`)

Triggers: `pull_request` to `main`, and `push` to `main`. It never uses `pull_request_target`. Default permissions: `contents: read`.

| Job | Needs | What it does |
|-----|-------|--------------|
| `checks` | - | Checks out the plugin and Kandev at the commit in `.kandev-sdk-ref`, sets up Go (from `go.mod`) and Node (from `.nvmrc`), runs `npm ci`, checks that `go mod tidy` does not change `go.mod`/`go.sum`, then runs `make check-format vet lint test coverage check-secrets build package verify-package`. Uploads `dist/` as the `plugin-package` artifact |
| `packaged-host-contract` | `checks` | Checks out Kandev at the SDK commit and at `v<min_kandev_version>` from `manifest.yaml`, downloads `plugin-package`, checks `sha256sum -c checksums.txt`, then runs `make verify-package contract-test KANDEV_MIN_DIR=../kandev-min`. Timeout 30 minutes |

The contract job tests the exact bytes that `checks` built and uploaded; it never rebuilds them.

### `.github/workflows/release.yml` (name `release`)

Trigger: `push` of a tag matching `v*`. Default permissions: `contents: read`. Concurrency group `release`, so a second release waits instead of running in parallel.

| Job | Needs | Permissions | What it does |
|-----|-------|-------------|--------------|
| `verify` | - | read | Full history checkout. Same setup and tidy check as `ci`, then `make check-format vet lint test coverage check-secrets build package verify-package release-preflight TAG=<tag>`. Package verification runs before the preflight (project Mandated rule). Uploads `release-package` |
| `contract` | `verify` | read | Same as `packaged-host-contract`, on the `release-package` artifact |
| `publish` | `verify`, `contract` | `contents: write`, `id-token: write`, `attestations: write` | Downloads `release-package`, checks the checksum, creates a build provenance attestation with `actions/attest-build-provenance`, then runs `gh release create <tag> --verify-tag --generate-notes` with the package and `checksums.txt`. Tags with a `-` become prereleases |

`release-preflight` refuses: a malformed tag, a tag that differs from `manifest.yaml`, a tag not on `origin/main`, a tag that already has a Release, and a first release without a passing record in `docs/manual-checks`. A released tag is never deleted or overwritten (project Forbidden rule); a broken release is fixed forward with a new patch tag.

## Supply-chain settings

- Every action is pinned to a full commit SHA, with the version in a trailing comment. `make lint` enforces this with `go run ./cmd/ci workflows -dir .github/workflows` and also runs `actionlint` (`v1.7.12`).
- Only `publish` has write or OIDC permissions.
- Tool versions are pinned in the `Makefile` (`golangci-lint v2.14.0`, `actionlint v1.7.12`) and run through `go run`, so CI and local runs use the same versions.
- No dependency vulnerability scan (team decision).

## Branch protection on GitHub

On 2026-10-07 the ruleset `main` (id 24580280, target `refs/heads/main`, enforcement `active`) contained only the `deletion` rule. Force-push protection had been turned off earlier to rewrite commit authors. Per [Q2] A it was updated on 2026-10-07 to:

| Rule | Setting |
|------|---------|
| `deletion` | branch deletion blocked |
| `non_fast_forward` | force-push blocked |
| `pull_request` | pull request required; 0 approvals (self-merge); allowed merge method `squash` only |
| `required_status_checks` | `checks` and `packaged-host-contract` from GitHub Actions (integration 15368); branches do not have to be up to date before merge |

There are no bypass actors, so the owner also goes through a pull request. Tags are not covered by this ruleset; a tag ruleset is not added (YAGNI: the release preflight and the "never delete a released tag" rule cover it for a single maintainer).

## Local equivalence

CI calls only `Makefile` targets, so `make check-format vet lint test coverage check-secrets build package verify-package` gives the same result locally (with `../kandev` at `.kandev-sdk-ref`). `make contract-test` needs `../kandev-min` at the minimum-version tag and gcc. After a local `make coverage`, delete `coverage.out` from the repository root during AI-DLC stages (project Testing Posture learning).

## Changes made in this stage

- GitHub ruleset `main` updated as above. No workflow file or `Makefile` changed.

## Open items

- The two `ci` jobs have not yet run on commit `74edd48` (Build and Test loop-backs), which is local only. It must go to `main` through a pull request, which now has to pass both required checks.
- GitHub set `require_extra_approval_for_unattributed_changes: true` as a default on the `pull_request` rule. If it blocks a self-merge, turn it off in the ruleset.
