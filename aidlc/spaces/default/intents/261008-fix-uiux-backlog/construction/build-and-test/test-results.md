# Test Results - Fix UIUX

Run on 2026-10-08 in the task worktree (branch `feature/fix-uiux-i41`), Go 1.26.8, Node v25.2.1, `../kandev` -> `~/repo/kandev` at f099a46 (v0.96.0).

## Build Status

| Command | Result |
|---------|--------|
| `make check-sdk` | pass |
| `make check-format` | pass |
| `make vet` | pass |
| `make lint` (golangci-lint + gosec, prettier, tsc, eslint, actionlint) | pass, 0 issues |
| `make build` | pass |
| `make package` | pass, `dist/nulab-backlog-0.4.2.tar.gz` 23,397,332 bytes |
| `make verify-package` | `verifypkg: OK dist/nulab-backlog-0.4.2.tar.gz (nulab-backlog@0.4.2)` |
| `make check-secrets` | pass |

## Test Results

| Suite | Command | Total | Passed | Failed | Skipped |
|-------|---------|-------|--------|--------|---------|
| Go, all packages (`-race`) | `make test` / `go test -race -count=1 ./...` | 709 top-level tests, 13 packages | 709 | 0 | 0 |
| UI (Vitest) | `make test` / `npx vitest run` | 387 tests, 33 files | 387 | 0 | 0 |
| Unit-scoped Go (unit-test-instructions) | `go test -race ./internal/issues/ ./internal/git/` | ok | ok | 0 | 0 |
| Unit-scoped UI (unit-test-instructions) | `npx vitest run src/index.test.ts src/issues/issue-badge.test.tsx src/issues/issues-state.test.ts src/settings/sections.test.tsx` | 58 | 58 | 0 | 0 |
| Packaged-host contract on Kandev v0.96.0 | `make contract-test KANDEV_MIN_DIR=../kandev`, 10 runs | 10 | 10 | 0 | 0 |

Baseline before the change (recorded at Code Generation): Go 326 top-level tests in `internal/issues`, `internal/git`, `internal/plugin`, now 331; UI 373, now 387. No regressions.

## Failure Details

None.

## Coverage

`make coverage`: **92.8%** line coverage over `./internal/...` and `./server/...` (floor 80%, excluded `server/main.go` only). Profile under `build/`; no `coverage.out` at the repo root.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| T-COV | `code-generation-plan.md` > Testing Contract (team) | >= 80% Go line coverage | 92.8% | `make coverage` above | build-and-test | Met |
| T-RACE | Testing Contract (team) | Go tests green with `-race` | 709/709 pass | `make test` above | build-and-test | Met |
| T-CONTRACT | Testing Contract (team) + project note (10 runs) | contract test passes on `min_kandev_version` | 10/10 pass on v0.96.0 | contract runs above | build-and-test | Met |
| T-REQ-TESTS | Testing Contract obligations (Minimal: one test per requirement) | every FR/NFR has a test | 28/28 covered | [cross-unit-traceability.md](cross-unit-traceability.md) | build-and-test | Met |
| T-SUITE | Testing Contract scope floor (express) | existing suite green | 709 Go + 387 UI pass | `make test` above | build-and-test | Met |
| T-LINT | Team Code Style (CI gates) | format, vet, lint clean | clean, 0 issues | `make check-format vet lint` | build-and-test | Met |
| T-PKG | Project Mandated rule (package verification) | package verifies | verifypkg OK | `make verify-package` | build-and-test | Met |

No `nfr-requirements/` or `nfr-design/` artifacts exist (express scope skips them); requirements NFR1-NFR5 are covered by tests listed in cross-unit-traceability.md.
