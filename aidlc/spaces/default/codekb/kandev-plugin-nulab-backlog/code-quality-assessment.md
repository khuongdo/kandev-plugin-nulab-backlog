# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- ~108 Go test files co-located in every `internal/*` package (`httptest` fakes, `testdata/` JSON), run with `-race`; ~40 Vitest files under `ui/src/`.
- `make coverage`: 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile at `build/coverage.out`.
- Packaging: `pkgverify_test.go`, `contract_test.go` (driver units); the real install path runs only in `make contract-test` (CI `packaged-host-contract`, release `contract`).
- Workflow policy is tested in `internal/ci/workflows_test.go`.
- Baseline: not recorded by the 2026-10-08 runs (the active intent changes only CI configuration; the Go suite is verified by CI itself).

## Linting

gofmt, go vet, golangci-lint (`.golangci.yml`, +gosec), `tsc --noEmit` strict, ESLint, Prettier, actionlint, `cmd/ci workflows` policy (`internal/ci/workflows.go` lines 36-40: SHA-pinned actions, top-level `permissions` exactly `contents: read`, no `pull_request_target`, write permission only in `release.yml` job `publish`). The policy does not inspect `paths` / `paths-ignore`.

## CI/CD

- `ci.yml`: `checks` (format, vet, lint, test, coverage, check-secrets, build, package, verify-package) and `packaged-host-contract` (installs the package on Kandev built at `v0.96.0`). Runs on every PR to `main` and every push to `main`; no path filter, no concurrency group.
- `release.yml`: `verify` -> `contract` -> `publish` (build provenance attestation, `gh release create`) on `v*` tag push.
- Required checks `checks` and `packaged-host-contract` on `main` via ruleset `24580280`; squash-only. Details: [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).
- Records-only PRs run the full CI today: PR #12 (`records/backlog-panel-retouch-v0.4.1`) shows `checks` 2m17s and `packaged-host-contract` 53s. PRs #3, #5, #7, #10, #12 are records-only.

## CI Path Filter Constraints

Facts any design for intent `261008-ci-path-filter` must respect:

1. **Required checks must still report.** A plain `paths` / `paths-ignore` on `ci.yml` stops the workflow from starting on a records-only PR, so `checks` and `packaged-host-contract` stay "Expected - waiting" and the PR cannot merge. A job skipped by `if:` does report success. Candidate patterns: an always-triggered workflow with a cheap change-detection job gating the two required jobs by `if:`, or a same-named no-op workflow on the inverse paths (fragile: names must match exactly).
2. **Release filtering has no trigger-level meaning.** `paths` filters do not apply to tag pushes, and a release is a deliberate manual tag. "Skip release for non-app changes" needs a different definition (for example `release-preflight` refusing a tag whose diff from the previous release touches no app path). Needs a user decision.
3. **`docs/manual-checks/` is a release input** (`release-preflight`, Makefile line 145, `internal/ci/release.go`).
4. **Secret scan coverage.** `check-secrets` scans the whole repo, including `aidlc/` and `docs/`. Skipping CI on records-only PRs drops that scan for those files (project Forbidden: no real credentials in the repo). Decide whether a light secret scan stays on skipped PRs.
5. **App set includes CI and build files.** `.github/workflows/**` and `Makefile` must count as app paths, because `make lint` checks the workflows and every job runs through the Makefile. Full classification: [code-structure.md](code-structure.md#top-level-path-classification).
6. **Workflow policy.** Any new action must be SHA-pinned and top-level `permissions: contents: read` kept, or `make lint` fails.
7. **Both events.** The filter must apply consistently to `pull_request` and `push` to `main`; the main-push run builds the package that a later release rebuilds.

## Documentation

README covers build, install (upload via Settings > Plugins), connection, CI, release, marketplace. Go packages have `doc.go`. `docs/manual-checks/` holds the first-release record. Makefile targets carry traceability IDs.

## Known Issue: Plugin install 502

Status: addressed in v0.4.2 (commit `f5a7529`, "Smaller package without Windows") by dropping `windows-amd64` from the package. Original analysis (intent `261007-plugin-install-502`), kept for history:

- Runtime: Kandev v0.97.0 (`kandev --headless`, `:38429`) behind `tailscale serve` (`https://webfrontier.tail152aaa.ts.net`).
- Kandev `server.readTimeout` default 30 s (`KANDEV_SERVER_READTIMEOUT`, `catalog.go` line 61). Multipart upload install parses the whole body before `Install`; a body slower than 30 s is cut.
- Backend logs: `POST /api/plugins/install` 400, `duration_ms` ~30000, 47-byte body `{"error":"missing multipart field \"package\""}`. `tailscale serve` turns the dropped upstream into 502.
- Reproduced: 29.5 MB v0.4.1 upload throttled to 800 KB/s -> 502 after 33.5 s; unthrottled uploads succeed.
- Classification: host timeout plus plugin package size. Remaining option: install by URL (see [architecture.md](architecture.md#improvement-opportunities)).

## Technical Debt

- `ci.yml` and `release.yml` duplicate setup and contract steps (lines 13-58 vs 145-191, 68-124 vs 200-256).
- `ci.yml` has no `concurrency` group; superseded PR pushes keep running.
- `release-preflight` uses `gh release list --limit 1000` (marked `ponytail:` in the Makefile).
- Platform list declared in four places (`manifest.yaml`, `Makefile` `PLATFORMS`, `pkgverify`, `internal/plugin/manifest_test.go`).
- `pkgverify` duplicates Kandev `pkgtar` rules; no size limit check.
- Contract test installs over loopback (30 s client timeout), so it never sees proxy/slow-upload failures.
- Runtime drift: Kandev 0.97.0 running vs 0.96.0 pin.
