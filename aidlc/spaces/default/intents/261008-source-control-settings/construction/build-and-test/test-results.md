# Test Results

Run on 2026-10-08 in this worktree, Go 1.26.8, `../kandev` at v0.96.0 (`f099a46d`, matches `.kandev-sdk-ref`).

## Build Status

SUCCESS.

- `make check-format vet lint`: exit 0; golangci-lint (with gosec) `0 issues.`
- `cd ui && npm run typecheck && npm run lint && npm run format:check`: exit 0; "All matched files use Prettier code style!"
- `make package verify-package`: exit 0; `verifypkg: OK dist/nulab-backlog-0.5.3.tar.gz (nulab-backlog@0.5.3)` (version bump is decided at Deployment Pipeline).

## Test Results

| Suite | Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| UI (Vitest, all files) | `make test` | 446 tests / 34 files | 446 | 0 | 0 |
| Go (`-race`, all packages) | `make test` / `make coverage` | 14 packages | 13 with tests ok, `server` has no tests | 0 | 0 |
| Packaged-host contract (Kandev v0.96.0) | `make contract-test KANDEV_MIN_DIR=../kandev` x10 | 10 runs | 10 | 0 | 0 |

The stage-level unit-test commands from `unit-test-instructions.md` (Go `./internal/scm/ ./internal/git/ ./internal/plugin/` and the five Vitest files) are subsets of the full runs above and were not run a second time.

## Failure Details

None.

## Coverage Report

`make coverage`: `coverage: 92.9% (floor 80%, excluded: server/main.go)`. Changed packages: `internal/scm` 93.7%, `internal/git` 91.0%, `internal/plugin` 93.7%. Coverage profile deleted after the run.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-COV | code-generation-plan.md Testing Contract (team Testing Posture) | Go line coverage >= 80% over ./internal/... and ./server/... | 92.9% | `make coverage` output | build-and-test | Met |
| TC-RACE | Testing Contract (team) | All Go tests pass with `-race` | All packages ok | `make coverage` / `make test` output | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team) | Packaged plugin installs and runs on min Kandev (0.96.0) | 10/10 runs pass | `make contract-test` x10 | build-and-test | Met |
| TC-UI | Testing Contract (team Code Style) | Vitest, tsc strict, ESLint, Prettier pass | 446/446, all clean | `make test`, `npm run typecheck/lint/format:check` | build-and-test | Met |
| TC-LINT | team Code Style | gofmt, go vet, golangci-lint + gosec clean | 0 issues | `make check-format vet lint` | build-and-test | Met |
| TC-PKG | project Mandated (package verification) | Package verification passes | verifypkg OK | `make package verify-package` | build-and-test | Met |
| TC-SUITE | Testing Contract scope floor | Existing suite stays green | All pass | full runs above | build-and-test | Met |
