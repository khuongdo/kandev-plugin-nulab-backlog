# Test Results — 261009-no-workflow-error

Date: 2026-10-09. Commit base: `3803248` + uncommitted bugfix changes. Kandev SDK `../kandev` at `v0.96.0`.

## Build Status

- `make check-format vet lint`: success (gofmt, go vet, golangci-lint incl. gosec 0 issues, tsc/ESLint/Prettier, actionlint, `ci workflows: OK`).
- `make package verify-package`: success — `verifypkg: OK dist/nulab-backlog-0.6.0.tar.gz (nulab-backlog@0.6.0)`.

## Test Results

| Suite | Command | Result |
|---|---|---|
| Go (all, race) | `go test -race ./...` | 13/13 packages ok, 0 failed |
| Go coverage | `make coverage` | 92.9% (floor 80%, excluded: `server/main.go`); profile under `build/`, none in repo root |
| UI (all) | `cd ui && npx vitest run` | 36 files, 462 passed, 0 failed, 0 skipped |
| UI typecheck | `cd ui && npx tsc --noEmit` | OK |
| Unit-scoped commands (unit-test-instructions.md) | `go test -race ./internal/plugin/ -run 'HasWorkflow|WorkflowsStatus|Manifest'`; Vitest `start-task`, `error-alert` and surface files | included in the full runs above, all passing (deduplicated, run once) |
| Packaged-host contract test | `make contract-test KANDEV_MIN_DIR=../kandev` x10 | 10/10 passed on Kandev v0.96.0 |
| Real-Kandev manual check | integration-test-instructions.md steps 1-4 | Passed (reported by the user, 2026-10-09): dialog opens from a direct `/backlog` load, task links to the issue, errors display well on mobile and desktop |

## Failure Details

None.

## Target Verification Matrix (final)

See `build-and-test-summary.md` § Target Verification Matrix — every applicable target is `Met`.
