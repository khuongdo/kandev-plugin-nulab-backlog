## Developer Code Scan Results

Focused scan for intent `261008-ci-path-filter`: run CI and release only for changes in app-related paths, excluding `aidlc/`, `docs/` and other non-app paths. Depth: Minimal.

### Scan Coverage
- **Analyzed deeply**:
  - .github/
  - .github/workflows/ci.yml
  - .github/workflows/release.yml
  - Makefile
- **Skimmed only** (top-level classification, plus the minimum needed to understand what the deep paths call):
  - aidlc/ (AI-DLC records, memory, codekb)
  - docs/ (docs/brand/, docs/manual-checks/)
  - .claude/ (AI-DLC framework shell)
  - internal/ (Go packages; internal/ci/workflows.go, internal/ci/secrets.go and internal/ci/release.go skimmed only for what the Makefile targets enforce)
  - server/
  - cmd/ (cmd/ci, cmd/verifypkg)
  - ui/
  - build/ (local build output, not a CI input)
  - dist/ (gitignored package output)
  - manifest.yaml, go.mod, go.sum, README.md, LICENSE, .golangci.yml, .kandev-sdk-ref, .nvmrc, .gitignore
  - GitHub repository ruleset for `main` (read via `gh api`, evidence for required checks)

### Packages Found
- `.github/workflows/ci.yml` — GitHub Actions workflow `ci` — YAML — PR and main-push checks (jobs `checks`, `packaged-host-contract`)
- `.github/workflows/release.yml` — GitHub Actions workflow `release` — YAML — tag-triggered release (jobs `verify`, `contract`, `publish`)
- `Makefile` — Make — standard targets that both workflows call

Top-level path classification (for the path filter):

| Path | Class | Why |
|------|-------|-----|
| `server/` | app | Go entrypoint, built into the package (`build`) |
| `internal/` | app | Go code; tested, linted, covered (`GO_PKGS := ./internal/... ./server/...`) |
| `cmd/` | app (tooling) | `cmd/ci` runs workflow lint, secret scan, contract test, release preflight; `cmd/verifypkg` verifies the package |
| `ui/` | app | Prettier, tsc, ESLint, Vitest, esbuild bundle into the package |
| `manifest.yaml` | app | Copied into the package; `VERSION` and `MIN_KANDEV_VERSION` are read from it |
| `go.mod`, `go.sum` | app | Build inputs; CI checks `go mod tidy` |
| `.kandev-sdk-ref` | app | Pinned SDK commit used by `check-sdk`, ldflags `SDKRef`, CI checkout |
| `.nvmrc` | app | Node version for CI |
| `.golangci.yml` | app (lint config) | Changes lint results |
| `Makefile` | app (build) | Every CI step goes through it |
| `.github/workflows/` | app (CI) | `make lint` runs actionlint and `cmd/ci workflows` on these files |
| `aidlc/` | non-app | AI-DLC records, memory and codekb; not read by any target except the repo-wide secret scan |
| `.claude/` | non-app | AI-DLC framework shell |
| `docs/brand/` | non-app | Brand note |
| `docs/manual-checks/` | non-app for CI, **read by release** | `release-preflight` reads `docs/manual-checks` (Makefile line 145; internal/ci/release.go) |
| `README.md`, `LICENSE` | non-app | Not read by any target |
| `.gitignore` | non-app | Not read by any target |
| `build/`, `dist/` | generated | Local output, not inputs |

### Build System
- **Type**: GNU Make (`SHELL := /bin/bash`, `-eu -o pipefail`) over Go modules and npm.
- **Config Files**: `Makefile`, `go.mod`/`go.sum`, `ui/package.json`/`ui/package-lock.json`, `.golangci.yml`, `.nvmrc`, `.kandev-sdk-ref`, `manifest.yaml`.
- **Build Dependencies** (target → prerequisites; what it touches):
  - `check-sdk` → `../kandev` must be at `.kandev-sdk-ref`.
  - `check-format` → `gofmt -l server internal`; `prettier --check` in `ui/`.
  - `vet` → check-sdk; `go vet ./...`.
  - `lint` → check-sdk; golangci-lint v2.14.0 (`./...`), `tsc --noEmit`, ESLint, actionlint v1.7.12, `go run ./cmd/ci workflows -dir .github/workflows`.
  - `test` → check-sdk; `go test -race ./internal/... ./server/...`; `vitest run`.
  - `coverage` → check-sdk; 80% floor, profile under `build/`, excludes only `server/main.go`.
  - `check-secrets` → check-sdk; `go run ./cmd/ci secrets -root .` walks the whole repo (including `aidlc/` and `docs/`; `docs/manual-checks/` is allow-listed in internal/ci/secrets.go line 98).
  - `build` → check-sdk; `CGO_ENABLED=0` cross-builds `./server` for `linux-amd64 linux-arm64 darwin-amd64 darwin-arm64`.
  - `ui-build` → esbuild `ui/src/index.ts` into `build/ui/bundle.js`; fails if React is bundled.
  - `package` → build, ui-build; stages `manifest.yaml`, executables, bundle; runs Kandev `plugin-pack`; writes `dist/checksums.txt`.
  - `verify-package` → check-sdk; `go run ./cmd/verifypkg`.
  - `contract-test` → builds Kandev at `v$(MIN_KANDEV_VERSION)` from `KANDEV_MIN_DIR`, installs the package, runs `cmd/ci contract`.
  - `release-preflight` → check-sdk; needs `TAG`; checks tag format, tag == manifest version, tag on `origin/main`, no existing Release, first-release record in `docs/manual-checks`.
  - `marketplace-entry`, `clean`, `help` — not used by CI.

### APIs Discovered
- GitHub Actions triggers (the surface this intent changes):
  - `ci.yml` (lines 3-7): `on: pull_request: branches: [main]` and `push: branches: [main]`. **No `paths` or `paths-ignore` filter today.** No `concurrency` block.
  - `release.yml` (lines 6-8): `on: push: tags: ['v*']`. **No `paths` filter.** `concurrency: group: release, cancel-in-progress: false`.
- Jobs:
  - `ci` / `checks`: checkout plugin + Kandev at SDK ref, setup-go, setup-node, `npm ci`, `go mod tidy` diff check, `make check-format vet lint test coverage check-secrets build package verify-package`, upload `plugin-package`.
  - `ci` / `packaged-host-contract` (`needs: checks`, 30 min): checks out Kandev at SDK ref and at `v<min_kandev_version>`, downloads the uploaded package, `sha256sum -c`, `make verify-package contract-test KANDEV_MIN_DIR=../kandev-min`.
  - `release` / `verify`: same as `checks` with `fetch-depth: 0`, plus `release-preflight TAG=$TAG` (`GH_TOKEN`), uploads `release-package`.
  - `release` / `contract` (`needs: verify`): same as `packaged-host-contract` against `release-package`.
  - `release` / `publish` (`needs: [verify, contract]`): only job with `contents: write`, `id-token: write`, `attestations: write`; attests `dist/nulab-backlog-*.tar.gz`; `gh release create --verify-tag --generate-notes` (prerelease when the tag has `-`).
- Required status checks on `main` (repository ruleset 24580280, `gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280`, read 2026-10-08): `required_status_checks` = contexts **`checks`** and **`packaged-host-contract`** (integration 15368 = GitHub Actions), `strict_required_status_checks_policy: false`. Other rules: `deletion`, `non_fast_forward`, `pull_request` with `allowed_merge_methods: [squash]`, 0 required approvals. Classic branch protection is not used (`branches/main/protection` returns 404).
- Evidence the required checks run on records-only PRs today: PR #12 `records/backlog-panel-retouch-v0.4.1` (AI-DLC Operation records only) shows `checks pass 2m17s` and `packaged-host-contract pass 53s`. Merged PRs #3, #5, #7, #10, #12 are records-only branches that would be skipped by the intended filter.

### Frameworks & Libraries
- actions/checkout — `3d3c42e5aac5ba805825da76410c181273ba90b1` (v7.0.1) — checkout plugin and Kandev
- actions/setup-go — `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0) — Go from `go.mod`
- actions/setup-node — `820762786026740c76f36085b0efc47a31fe5020` (v7.0.0) — Node from `.nvmrc`
- actions/upload-artifact — `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a` (v7.0.1)
- actions/download-artifact — `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c` (v8.0.1)
- actions/attest-build-provenance — `4d101475d8b20a2381f78447822ac1eab6504dd8` (v4.2.2) — release only
- golangci-lint v2.14.0 and actionlint v1.7.12 via `go run` (pinned in Makefile)
- Every `uses:` is pinned to a full 40-character SHA with a version comment.

### Test Coverage
- **Test Directories**: `internal/**` (Go, `-race`), `ui/` (Vitest); workflow policy tested in `internal/ci/workflows_test.go` (skimmed).
- **Test Frameworks**: Go `testing` + testify; Vitest.
- **Coverage Config**: present (`coverage` target, 80% floor, `COVERAGE_EXCLUDE := server/main.go`).

### Code Quality Indicators
- **Linting**: `.golangci.yml`, ESLint/Prettier/tsc in `ui/`, actionlint, and a repo-specific workflow policy `go run ./cmd/ci workflows -dir .github/workflows` (internal/ci/workflows.go lines 36-40): every action pinned to a full SHA; top-level `permissions` exactly `contents: read`; no `pull_request_target`; write permission only in `release.yml` job `publish`. It does not inspect `paths`/`paths-ignore`, so a filter passes this policy as long as the rest is unchanged.
- **CI/CD**: `.github/workflows/ci.yml`, `.github/workflows/release.yml`. Top-level `permissions: contents: read` in both; `publish` is the only job with write.
- **Documentation**: Makefile targets carry comments with traceability IDs (US7.4, AC7.5.2, R-01..R-03).

### Technical Debt Signals
- `ci.yml` and `release.yml` duplicate the setup and contract steps (lines 13-58 vs 145-191, 68-124 vs 200-256). Out of scope for this intent.
- `ci.yml` has no `concurrency` group, so superseded PR pushes keep running. Out of scope; note only.
- `release-preflight` uses `gh release list --limit 1000` (marked `ponytail:` in the Makefile).

## Handoff Summary
- **Intent-relevant finding**: `ci.yml` (lines 3-7) triggers on every `pull_request` and `push` to `main` with no path filter, and `release.yml` (lines 6-8) triggers on `push` of `v*` tags with no path filter. The `main` ruleset requires the status checks `checks` and `packaged-host-contract`. If `ci.yml` simply gains `paths`/`paths-ignore` (for `aidlc/**`, `docs/**`, `.claude/**`, `README.md`, `LICENSE`), a records-only PR (like #3, #5, #7, #10, #12) never starts the workflow, the two required checks stay "Expected - waiting" and the PR cannot be merged under the squash-only, required-checks ruleset. The design needs a pattern that still reports `checks` and `packaged-host-contract` on skipped PRs, for example: keep the workflow triggering always and add a cheap change-detection job (`git diff --name-only` against the base, no third-party action, or `dorny/paths-filter` pinned by SHA) with the two required jobs gated by `if:` (a job skipped by `if:` reports as success for required checks), or a separate same-named no-op workflow on `paths-ignore` (fragile; GitHub documents it but names must match exactly).
- **Risks / follow-up**:
  - Release: GitHub ignores `paths` filters for tag pushes (filters apply only to branch pushes), so a path filter on `release.yml` would have no effect; and a release is a deliberate manual tag, so "skip release for non-app changes" needs a different meaning (for example, `release-preflight` refusing a tag whose diff from the previous release touches no app path). Clarify with the user what "release" filtering should do.
  - `release-preflight` reads `docs/manual-checks/` (Makefile line 145, internal/ci/release.go), so `docs/manual-checks/**` is a release input even if it is a non-app path for CI.
  - `check-secrets` scans the whole repo, including `aidlc/` and `docs/`; skipping CI on records-only PRs drops the credential scan for those files (project Forbidden: no real credentials in the repo). Decide whether to keep a light secret-scan job for skipped PRs.
  - `.github/workflows/**` and `Makefile` must stay in the "app" set, because `make lint` checks the workflows and every job runs through the Makefile.
  - Any new action must be pinned to a full commit SHA and keep top-level `permissions: contents: read`, or `make lint` (internal/ci/workflows.go) fails.
  - `push` to `main` after a squash merge would also be filtered; for app changes the main-push run still matters (it builds the package that the release later rebuilds), so the filter must apply consistently to both events.
