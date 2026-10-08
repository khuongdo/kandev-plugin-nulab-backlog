# Test Results — Opt-in by default

Run date: 2026-10-07. Environment: Go 1.26.x (`~/.local/go/bin`), `../kandev` at v0.96.0 (`f099a46dc`), Node per `.nvmrc`.

## Build Status

- `make check-format vet lint check-secrets`: success — golangci-lint `0 issues.`, `tsc`/ESLint/Prettier clean, actionlint OK, `ci workflows: OK`, `ci secrets: OK`.
- `make package`: success — `dist/nulab-backlog-0.1.1.tar.gz` written.
- `make verify-package`: success — `verifypkg: OK dist/nulab-backlog-0.1.1.tar.gz (nulab-backlog@0.1.1)`.

## Test Results

| Command | Result |
|---|---|
| `go test -race -count=1 ./internal/connection/... ./internal/plugin/... ./internal/ci/... ./internal/issues/... ./internal/git/...` (unit-test-instructions, run once) | all 5 packages `ok` |
| `make coverage` (`go test -race -coverprofile=build/coverage.out ./internal/... ./server/...`) | all packages `ok`; 1169 Go tests pass, 0 fail (count from Code Generation run; this run had no failures) |
| `cd ui && npx vitest run` (full UI suite, covers `src/switch/switch.test.tsx` and `src/index.test.ts`) | 29 files, 291/291 passed |
| `make contract-test KANDEV_MIN_DIR=../kandev` | `ci contract: OK nulab-backlog on Kandev v0.96.0` (1 run here; 10/10 in Code Generation) |

Failures: none. Skipped: none.

## Coverage Report

`coverage: 92.8% (floor 80%, excluded: server/main.go)`. Per package: backlog 96.1%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.7%, redact 97.4%, testutil 88.0%. Profile kept under `build/`; no `coverage.out` at the repo root.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-COV-80 | code-generation-plan.md § Testing Contract (team Testing Posture) | Go line coverage ≥ 80% over `./internal/...`, `./server/...` | 92.8% | `make coverage` output above | build-and-test | Met |
| TC-RACE | code-generation-plan.md § Testing Contract | Go tests run with `-race`, all pass | all packages ok with `-race` | `make coverage`, scoped `go test -race` | build-and-test | Met |
| TC-REGRESSION | code-generation-plan.md § Testing Contract (bugfix scope floor) | Targeted regression for the bug | `TestSwitchWithNoRecord…` (store) + 6 tests in `internal/plugin/opt_in_test.go` pass | Go runs above | build-and-test | Met |
| TC-SUITE-GREEN | code-generation-plan.md § Testing Contract (scope floor) | Existing suite stays green | Go all ok; Vitest 291/291 | runs above | build-and-test | Met |
| TC-CONTRACT | code-generation-plan.md § Testing Contract (team Testing Posture) | Packaged-host contract test passes on min Kandev version | `ci contract: OK` on v0.96.0 | `make contract-test` | build-and-test | Met |
| TC-LINT | team Code Style (via Testing Contract notes) | gofmt, vet, golangci-lint+gosec, tsc, ESLint, Prettier clean | all clean | `make check-format vet lint` | build-and-test | Met |
| TC-SECRETS | Testing Contract (no secrets in tests) / NFR3 | No credential-shaped strings | `ci secrets: OK` | `make check-secrets` | build-and-test | Met |

No `nfr-requirements/` or `nfr-design/` artifacts exist (bugfix scope skips those stages).
