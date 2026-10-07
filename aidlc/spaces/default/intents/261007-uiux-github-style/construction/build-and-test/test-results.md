# Test Results — 261007-uiux-github-style

Run date: 2026-10-07 (UTC). Environment: Go 1.26.8 with CGO, `../kandev` at `v0.96.0` (`f099a46`), Node v25.2.1 locally (`.nvmrc` says 22; CI uses 22).

## Build

| Command | Result |
|---|---|
| `make check-format vet lint test coverage check-secrets build package verify-package` | Exit 0 |
| Prettier (`check-format`) | All matched files use Prettier code style |
| `gofmt -l internal server cmd` | No files listed |
| `go vet ./...` | Clean |
| `golangci-lint` v2.14.0 (default + gosec) | 0 issues |
| `tsc --noEmit`, `eslint .` | Clean |
| `actionlint` v1.7.12, `go run ./cmd/ci workflows` | OK |
| `go mod tidy -diff` | No diff |
| `check-secrets` | Pass |
| `build`, `package` | `dist/nulab-backlog-0.1.0.tar.gz`, `dist/checksums.txt` |
| `verify-package` | `verifypkg: OK dist/nulab-backlog-0.1.0.tar.gz (nulab-backlog@0.1.0)` |

## Tests

| Suite | Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| Go, all packages (`make test`) | `go test -race ./internal/... ./server/...` | 9 packages | 9 | 0 | 0 |
| Go, this change (unit-test-instructions) | `go test -race -count=1 ./internal/backlog/... ./internal/issues/... ./internal/git/... ./internal/plugin/...` | 4 packages | 4 | 0 | 0 |
| UI, all (`make test`) | `npx vitest run` | 286 tests / 29 files | 286 | 0 | 0 |
| UI, this change (unit-test-instructions) | `npx vitest run src/index.test.ts src/brand src/page src/settings src/issues src/git src/switch` | 237 tests / 28 files | 237 | 0 | 0 |
| Packaged-host contract on Kandev v0.96.0 | `make contract-test KANDEV_MIN_DIR=../kandev`, 10 consecutive runs | 10 | 10 | 0 | 0 |

The unit-test-instructions UI command omits `ui/src/controls.test.ts` (it sits directly in `ui/src/`); the full `npx vitest run` above includes it and passes.

## Coverage

`make coverage` (profile under `build/`, never the repository root): **92.8%** total (floor 80%; only `server/main.go` excluded). Per package: backlog 96.1%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.7%, redact 97.4%, testutil 88.0%. Baseline before the change: 92.9%.

## Failures

None.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TP-COV | team.md Testing Posture; code-generation-plan.md Testing Contract | Go line coverage ≥ 80% over `./internal/...` and `./server/...`, only `main` excluded | 92.8% | `make coverage` output above | build-and-test | Met |
| TP-RACE | team.md Testing Posture | All Go tests pass with `-race` | 9/9 packages pass | `make test` | build-and-test | Met |
| TP-UI | team.md Testing Posture / Code Style | UI Vitest suite green; `tsc` strict, ESLint, Prettier clean | 286/286; all clean | `make test`, `check-format`, `lint` | build-and-test | Met |
| TP-CONTRACT | team.md Testing Posture; project.md Correction (repeat runs) | Packaged plugin installs and runs on Kandev `min_kandev_version` 0.96.0, repeatedly | 10/10 runs OK | `contract1..10` runs above | build-and-test | Met |
| CS-GO | team.md Code Style | gofmt, go vet, golangci-lint + gosec clean; `go mod tidy` no diff | All clean | `make check-format vet lint`, `go mod tidy -diff` | build-and-test | Met |
| SEC-LEAK | requirements NFR2; project.md Mandated | No key/token in logs, errors or UI responses | Leak tests pass | `internal/plugin/actions_watch_test.go` in `make test` | build-and-test | Met |
| SEC-CREDS | project.md Forbidden | No real credentials in repo/test data | Pass | `make check-secrets` | build-and-test | Met |
| PKG | project.md Mandated; team.md Deployment | Package verification passes | OK | `make verify-package` | build-and-test | Met |
| A11Y | requirements NFR4 | Icon-only buttons named; axe checks pass on changed screens | Pass | `ui/src/page/backlog-lists.test.tsx`, `ui/src/settings/sections.test.tsx` in `make test` | build-and-test | Met |
