# Test Results — Link Task modal, GitHub-style

Run date: 2026-10-08. Commit base: `d3d17e5` plus the uncommitted Code Generation changes.

## Build Status

| Command | Result |
|---|---|
| `cd ui && npm run build` | Success: `../build/ui/bundle.js 226.9kb` |
| `make vet` | Success, no findings |
| `make check-format` | Success (gofmt plus Prettier: "All matched files use Prettier code style!") |

## Test Results

| Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|
| Unit-scoped: `cd ui && npx vitest run src/issues/issue-link.test.ts src/issues/link-task-dialog.test.tsx src/issues/task-menu.test.ts src/index.test.ts` | 59 (4 files) | 59 | 0 | 0 |
| Full UI: `cd ui && npx vitest run` | 421 (34 files) | 421 | 0 | 0 |
| Go: `make coverage` (`go test -race` over `./internal/... ./server/...`) | 13 packages with tests | 13 ok | 0 | 0 |
| `cd ui && npx tsc --noEmit` | — | pass | — | — |
| `cd ui && npx eslint .` | — | pass | — | — |

UI baseline before the change: 387 tests in 33 files, all passing. 34 tests were added and no existing test failed.

## Failure Details

None.

## Coverage Report

`make coverage`: **92.9%** total (floor 80%, excluded: `server/main.go`). Per package, from 88.0% (`internal/testutil`) to 97.4% (`internal/redact`). The profile is written under `build/` (git-ignored), not at the repository root.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-GO-COV | Testing Contract (`team` Testing Posture, coverage floor) | ≥ 80% Go line coverage over `./internal/...` and `./server/...` | 92.9% | `make coverage` output | build-and-test | Met |
| TC-GO-RACE | Testing Contract (`team`, `go test -race`) | All Go tests pass with `-race` | 13/13 packages ok | `make coverage` (runs `go test -race`) | build-and-test | Met |
| TC-SCOPE-REGRESSION | Testing Contract (`org`, bugfix scope floor) | A targeted regression for the defect | Dialog test: success → toast + links-store refresh + `onLinked` + `onClose`; task action refresh test | `ui/src/issues/link-task-dialog.test.tsx`, `ui/src/issues/issue-link.test.ts` | build-and-test | Met |
| TC-SUITE-GREEN | Testing Contract (`org`, bugfix scope floor) | Existing suite stays green | UI 421/421, Go 13/13 | Full vitest and `make coverage` | build-and-test | Met |
| TC-MIN-PER-REQ | Testing Contract (Minimal strategy) | ≥ 1 test per requirement, happy path per component | Every FR/NFR mapped to a test file (see `cross-unit-traceability.md`) | `cross-unit-traceability.md` | build-and-test | Met |
| TC-TS-STRICT | Team Code Style (TypeScript) | `tsc --noEmit` strict, ESLint, Prettier clean | All pass | Commands above | build-and-test | Met |
| TC-PKG-CONTRACT | Testing Contract (`team`, packaged-host contract test on min Kandev); project rule: run 10 times | Package installs and runs on Kandev v0.96.0, 10/10 runs | 10/10 runs `ci contract: OK nulab-backlog on Kandev v0.96.0` | `make package` + `make verify-package` (`verifypkg: OK`), then `make contract-test KANDEV_MIN_DIR=../kandev` x10 | build-and-test | Met |

No NFR Requirements or NFR Design artifacts exist for this bugfix scope, so the inventory comes only from the Testing Contract and the team Code Style.

## Packaging

| Command | Result |
|---|---|
| `make package` | `dist/nulab-backlog-0.5.0.tar.gz` (`dist/` is git-ignored) |
| `make verify-package` | `verifypkg: OK dist/nulab-backlog-0.5.0.tar.gz (nulab-backlog@0.5.0)` |
| `make contract-test KANDEV_MIN_DIR=../kandev` (10 runs, `../kandev` at tag `v0.96.0`) | 10/10 passed |
