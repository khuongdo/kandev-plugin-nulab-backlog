# Code Structure — kandev-plugin-nulab-backlog

## Layout

| Path | Classification | Notes |
|---|---|---|
| `server/main.go` | entrypoint | `pluginsdk.Serve(plugin.NewRuntime())` only |
| `internal/plugin/` | adapter | handlers map, webhook, events, host port, credentials; `Version`/`SDKRef` set by ldflags |
| `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/` | domain services | see [component-inventory.md](component-inventory.md) |
| `internal/backlog/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` | HTTP clients | stdlib `net/http`, `testdata/` JSON fixtures |
| `internal/redact/`, `internal/testutil/` | utility / test helper | |
| `internal/pkgverify/`, `cmd/verifypkg/` | build tooling | offline package verifier |
| `internal/ci/`, `cmd/ci/` | build tooling | `secrets`, `workflows`, `contract`, `preflight`, `marketplace` |
| `ui/src/` | UI bundle source | `index.ts` registration; areas `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, `testing/` |
| `manifest.yaml` | plugin contract | see [api-documentation.md](api-documentation.md) |
| `Makefile` | build | every CI step goes through it |
| `.github/workflows/` | CI/CD | `ci.yml`, `release.yml`; see [architecture.md](architecture.md#ci-and-release-pipeline) |
| `docs/` | docs | `brand/`, `manual-checks/` |
| `aidlc/`, `.claude/` | process records / framework | AI-DLC records, memory, codekb; AI-DLC shell |

## Top-Level Path Classification

What each top-level path means for CI and release (developer scan, 2026-10-08):

| Path | Class | Why |
|---|---|---|
| `server/`, `internal/` | app | Go code; `GO_PKGS := ./internal/... ./server/...` is tested, linted, covered; `./server` is built |
| `cmd/` | app (tooling) | `cmd/ci` (workflow lint, secret scan, contract test, release preflight), `cmd/verifypkg` |
| `ui/` | app | Prettier, tsc, ESLint, Vitest, esbuild bundle into the package |
| `manifest.yaml` | app | copied into the package; `VERSION` and `MIN_KANDEV_VERSION` are read from it |
| `go.mod`, `go.sum` | app | build inputs; CI checks `go mod tidy` |
| `.kandev-sdk-ref` | app | SDK pin for `check-sdk`, ldflags `SDKRef`, CI checkout |
| `.nvmrc` | app | Node version for CI |
| `.golangci.yml` | app (lint config) | changes lint results |
| `Makefile` | app (build) | every CI step |
| `.github/workflows/` | app (CI) | `make lint` runs actionlint and `cmd/ci workflows` on these files |
| `aidlc/`, `.claude/` | non-app | read by no target except the repo-wide `check-secrets` scan |
| `docs/brand/` | non-app | brand note |
| `docs/manual-checks/` | non-app for CI, **release input** | read by `release-preflight` (Makefile line 145, `internal/ci/release.go`); allow-listed in the secret scan |
| `README.md`, `LICENSE`, `.gitignore` | non-app | read by no target |
| `build/`, `dist/` | generated | local output, not inputs |

## Build and Packaging

`Makefile` (GNU Make, `SHELL := /bin/bash`, `-eu -o pipefail`). Every Go target first runs `check-sdk`: `../kandev` HEAD must equal `.kandev-sdk-ref`.

| Target | Does |
|---|---|
| `check-format` | `gofmt -l server internal`; `prettier --check` in `ui/` |
| `vet` | `go vet ./...` |
| `lint` | golangci-lint v2.14.0 (`./...`), `tsc --noEmit`, ESLint, actionlint v1.7.12, `go run ./cmd/ci workflows -dir .github/workflows` |
| `test` | `go test -race ./internal/... ./server/...`; `vitest run` |
| `coverage` | 80% floor, profile under `build/`, excludes only `server/main.go` |
| `check-secrets` | `go run ./cmd/ci secrets -root .` over the whole repo (including `aidlc/`, `docs/`) |
| `build` | `CGO_ENABLED=0` cross-build of `./server` for `PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64` |
| `ui-build` | esbuild `ui/src/index.ts` -> `build/ui/bundle.js`; fails if React is bundled |
| `package` | build + ui-build -> stage -> Kandev `cmd/plugin-pack` -> `dist/nulab-backlog-<version>.tar.gz` + `dist/checksums.txt` |
| `verify-package` | `go run ./cmd/verifypkg` |
| `contract-test` | builds Kandev at `v$(MIN_KANDEV_VERSION)` from `KANDEV_MIN_DIR`, installs the package, runs `cmd/ci contract` |
| `release-preflight` | needs `TAG`; tag format, tag == manifest version, tag on `origin/main`, no existing Release, first-release record in `docs/manual-checks/` |
| `marketplace-entry`, `clean`, `help` | not used by CI |

The platform list lives in four places: `manifest.yaml` `runtime.executables`, `Makefile` `PLATFORMS`, `internal/pkgverify` executables, `internal/plugin/manifest_test.go`.

## Code Patterns

- Ports-and-adapters: only `internal/plugin` maps domain errors to `pluginsdk` codes.
- Errors wrapped with `%w`; `context.Context` first on I/O; responses bounded by `io.LimitReader`.
- Tests co-located (`_test.go`, `*.test.ts[x]`), fake servers via `httptest`.
- Makefile targets carry comments with traceability IDs (US7.4, AC7.5.2, R-01..R-03).
