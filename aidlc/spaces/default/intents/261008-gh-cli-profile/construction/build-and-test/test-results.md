# Test Results — 261008-gh-cli-profile

Run date: 2026-10-08, branch `feature/gh-cli-profile-scope-q1o` (uncommitted working tree), Go 1.26.8, `../kandev` → v0.96.0.

## Build Status

| Command | Result |
|---|---|
| `make check-sdk` | pass |
| `make check-format` | pass |
| `make vet` | pass |
| `make lint` (golangci-lint v2.14.0 + gosec) | pass, 0 issues |
| `make check-secrets` | pass |
| `make package` | pass — `dist/nulab-backlog-0.5.3.tar.gz` |
| `make verify-package` | pass — `verifypkg: OK dist/nulab-backlog-0.5.3.tar.gz (nulab-backlog@0.5.3)` |
| `npx tsc --noEmit` (ui) | pass |
| `npx eslint .` (ui) | pass |
| `npx prettier --check .` (ui) | pass |

## Test Results

| Suite | Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| Go (all packages, `-race`) | `make coverage` | 1433 test results (tests + subtests, counted with `go test -json ./...`) | 1433 | 0 | 0 |
| UI (Vitest, 34 files) | `npx vitest run` | 435 | 435 | 0 | 0 |
| Packaged-host contract | `make contract-test KANDEV_MIN_DIR=../kandev` ×10 | 10 | 10 | 0 | 0 |

The unit-scoped commands from `unit-test-instructions.md` (`go test -race -count=1 ./internal/scm/ ./internal/plugin/`, `npx vitest run src/settings/source-control-section.test.tsx`) are subsets of the full runs above and were not run separately (no double counting).

Baseline before the change: 1383 Go test results → 1433 (+50).

## Failure Details

None.

## Coverage

`make coverage`: total **92.9%** (floor 80%, excluded: `server/main.go`); `internal/scm` 93.7%. Profile written under `build/`; no `coverage.out` in the repo root.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-coverage-go | code-generation-plan.md Testing Contract (team Testing Posture) | ≥ 80% line coverage over `./internal/...`, `./server/...` | 92.9% | `make coverage` output | build-and-test | Met |
| TC-race | Testing Contract (team) | Go tests run with `-race`, all pass | 1433/1433 pass with `-race` | `make coverage` (`go test -race`) | build-and-test | Met |
| TC-existing-green | Testing Contract (express scope floor) | existing suite stays green | 0 failures (baseline 1383 → 1433) | `go test -json ./...` count | build-and-test | Met |
| TC-contract | Testing Contract (team, project rule 10×) | packaged-host contract test passes on min Kandev 10/10 | 10/10 | `ci contract: OK nulab-backlog on Kandev v0.96.0` | build-and-test | Met |
| TC-package-verify | project Mandated rule | package verification passes | OK | `make verify-package` | build-and-test | Met |
| TC-secrets-redaction | Testing Contract (team) | tokens never in logs/errors/UI responses | redaction tests pass, secret scan clean | AC3.1.3 tests, `make check-secrets` | build-and-test | Met |
| TC-lint | team Code Style | gofmt, vet, golangci-lint+gosec, tsc, ESLint, Prettier clean | all clean | commands above | build-and-test | Met |
| TC-ui-tests | Testing Contract (team) | Vitest suite passes | 435/435 | `npx vitest run` | build-and-test | Met |
| NFR2-timeout-ttl | requirements.md NFR2 | 10 s CLI timeout; 5 min per-login cache | unit tests with injected clock pass | `internal/scm/cli_token_test.go` | build-and-test | Met |
