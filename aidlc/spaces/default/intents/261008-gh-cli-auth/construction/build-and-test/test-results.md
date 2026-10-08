# Test Results — CLI login for GitHub and GitLab

Run on 2026-10-08, branch `feature/th-m-auth-method-cho-0pe` (base commit `d3d17e5`, change uncommitted), Go 1.26.8, `../kandev` at `v0.96.0` (`f099a46`).

## Build Status

| Command | Exit | Result |
|---|---|---|
| `make check-format` | 0 | gofmt and Prettier clean |
| `make vet` | 0 | clean |
| `make lint` | 0 | golangci-lint (+gosec), tsc, ESLint, actionlint, `ci workflows: OK` |
| `make check-secrets` | 0 | `ci secrets: OK` |
| `make package` | 0 | `dist/nulab-backlog-0.5.0.tar.gz` + `checksums.txt` |
| `make verify-package` | 0 | `verifypkg: OK dist/nulab-backlog-0.5.0.tar.gz (nulab-backlog@0.5.0)` |

## Test Results

| Command | Exit | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| `make test` — Go (`go test -race` on all 13 packages) | 0 | 13 packages | 13 ok | 0 | 0 |
| `make test` — UI (`npx vitest run`) | 0 | 391 tests / 33 files | 391 | 0 | 0 |
| Unit commands of this change: `go test -race ./internal/scm/... ./internal/gitlab/... ./internal/plugin/...` (uncached, `-v`) | 0 | 214 top-level tests | 214 | 0 | 0 |
| Unit commands of this change: `cd ui && npx vitest run src/settings/source-control-section.test.tsx src/git/git-state.test.ts` | 0 | 25 | 25 | 0 | 0 |
| Security: `go test -race -count=1 ./internal/scm/ -run 'CLI\|Leak\|Secret\|Redaction'` | 0 | ok | ok | 0 | 0 |
| Parity: `go test -race -count=1 ./internal/plugin/ -run 'Manifest'` | 0 | ok | ok | 0 | 0 |
| Contract: `make contract-test KANDEV_MIN_DIR=../kandev` × 10 | 0 | 10 runs | 10 | 0 | 0 |

Baseline before the change (from code-summary.md): 194 top-level Go tests in the three touched packages and 387 UI tests, all passing. After: 214 and 391, all passing. No regression.

Contract test output (each run): `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Failure Details

None.

## Coverage Report

`make coverage`: **92.9%** total (floor 80%, excluded: `server/main.go`). Touched packages: `internal/scm` 93.7%, `internal/gitlab` 88.7%, `internal/plugin` 93.8%. The profile is written under `build/`; no `coverage.out` at the repo root.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-COV | code-generation-plan.md § Testing Contract (team Testing Posture) | ≥ 80% line coverage, `./internal/...` + `./server/...`, only `server/main.go` excluded | 92.9% | `make coverage` output | build-and-test | Met |
| TC-RACE | Testing Contract (team) | all Go tests run with `-race` | `go test -race` in `make test` and in the unit commands | Makefile `test` target; command logs | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team) + project rule (10 runs) | packaged-host contract test passes on Kandev `v0.96.0`, 10/10 | 10/10 | `make contract-test` × 10 | build-and-test | Met |
| TC-LEAK | Testing Contract (team) + NFR1 | a test asserts no token in logs, errors or UI replies | `TestRedaction_NoCLITokenInLogsErrorsOrReplies` and `TestCLIToken_EveryFailureIsCLIUnavailable` pass | security command | build-and-test | Met |
| TC-SUITE | Testing Contract obligations (scope floor) | existing suite stays green | all Go packages ok, 391/391 UI | `make test` | build-and-test | Met |
| TC-MINIMAL | Testing Contract obligations (strategy) | ≥ 1 test per requirement, happy path per component | every FR/NFR traced to a test or code file covered by tests (see cross-unit-traceability.md) | traceability.json | build-and-test | Met |
| TC-STYLE | team Code Style | format, vet, lint (gosec), tsc, ESLint, Prettier pass | all pass | `make check-format vet lint` | build-and-test | Met |
| TC-PKG | project Mandated (package verification) | package builds and verifies | `verifypkg: OK` | `make package verify-package` | build-and-test | Met |
| NFR2 | requirements.md NFR2 | fixed args, no shell, timeout ≤ 10 s, capped output | 10 s timeout, 4 KiB cap, fixed args | `TestCLIToken_RunsTheFixedCommandAndTrims`, `TestRunCLI_LimitsOutputAndReportsMissingBinary` | build-and-test | Met |
| NFR4 | requirements.md NFR4 | ≤ 1 CLI process per provider per 5 minutes in steady state | cache reused for 5 minutes on injected clock | `TestCLIToken_CachesForFiveMinutes` | build-and-test | Met |
