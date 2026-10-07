# Test Results — github-parity-actions

Run on 2026-10-07 against the working tree on base commit `2b4325f` (uncommitted Code Generation changes), with go1.26.8, Node v25.2.1 and Kandev `v0.96.0` (`f099a46`).

## Build

| Command | Result |
|---|---|
| `make check-format vet lint` | Pass: gofmt clean, go vet clean, golangci-lint `0 issues.`, `ci workflows: OK` |
| `make package verify-package` | Pass: `verifypkg: OK dist/nulab-backlog-0.2.0.tar.gz (nulab-backlog@0.2.0)` |
| `(cd ui && npm run typecheck && npm run lint && npm run format:check)` | Pass: `All matched files use Prettier code style!` |
| `git diff go.mod go.sum` | Empty (no new dependency) |

## Tests

| Suite | Total | Passed | Failed | Skipped |
|---|---|---|---|---|
| Go `go test -race ./internal/... ./server/...` (9 packages, via `make test coverage`) | 9 packages | 9 | 0 | 0 |
| Vitest `npx vitest run` (31 files) | 317 | 317 | 0 | 0 |
| Stage-level scoped commands from `unit-test-instructions.md` | Covered by the full runs above; each was run once | Pass | 0 | 0 |
| Packaged-host contract test `make contract-test KANDEV_MIN_DIR=../kandev` (10 runs, project rule) | 10 | 10 | 0 | 0 |

Every contract run ended with `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Coverage

`make coverage`: **92.8%** (floor 80%, excluded: `server/main.go`).

| Package | Coverage |
|---|---|
| internal/backlog | 96.1% |
| internal/ci | 91.0% |
| internal/connection | 94.6% |
| internal/git | 90.9% |
| internal/issues | 91.5% |
| internal/pkgverify | 93.2% |
| internal/plugin | 93.8% |
| internal/redact | 97.4% |
| internal/testutil | 88.0% |

The profile is written to `build/coverage.out` (gitignored). There is no `coverage.out` in the repo root.

## Failure Details

None.

## Target Verification Matrix

Final verdicts are in [build-and-test-summary.md](build-and-test-summary.md#target-verification-matrix). All targets are Met.
