# Test Results — Plugin install failed: 502

Run date: 2026-10-07 (UTC), worktree branch `feature/plugin-install-faile-9qm`, Go 1.26.8, `../kandev` at `v0.96.0` (`f099a46dc`).

## Build Status

| Command | Result |
|---|---|
| `make check-format` | pass (exit 0) |
| `make vet` | pass (exit 0) |
| `make lint` (golangci-lint + gosec, tsc, eslint, actionlint, workflow policy) | pass (exit 0) |
| `make build` | pass — 4 executables built |
| `make package` | pass — `dist/nulab-backlog-0.4.1.tar.gz`, 23,393,498 bytes |
| `make verify-package` | pass — `verifypkg: OK dist/nulab-backlog-0.4.1.tar.gz (nulab-backlog@0.4.1)` |

Package contents (`tar tzf`): `manifest.yaml`, `server/plugin-darwin-amd64`, `server/plugin-darwin-arm64`, `server/plugin-linux-amd64`, `server/plugin-linux-arm64`, `ui/bundle.js`, `checksums.txt`. No Windows executable.

## Test Results

| Suite | Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| Fix's unit tests (stage-level, run once) | `go test -race ./internal/pkgverify/...` | 29 | 29 | 0 | 0 |
| Go, whole module | `go test -race -json ./...` (also `make test`) | 1325 | 1325 | 0 | 0 |
| UI (Vitest) | `make test` (`npx vitest run`) | 373 (33 files) | 373 | 0 | 0 |
| Secret scan | `make check-secrets` | — | pass | — | — |
| Packaged-host contract on Kandev v0.96.0 | `make contract-test KANDEV_MIN_DIR=../kandev`, run 10 times | 10 | 10 | 0 | 0 |

Baseline before the change (from Code Generation): 1324 Go tests passed; the fix adds one regression subtest.

## Coverage

`make coverage`: **92.8%** line coverage over `./internal/...` and `./server/...` (floor 80%, excluded `server/main.go`). Profile written under `build/`; no `coverage.out` at the repository root.

## Failure Details

None.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| T-COV | `code-generation-plan.md` > Testing Contract (team floor) | ≥ 80% line coverage | 92.8% | `make coverage` output | build-and-test | Met |
| T-RACE | Testing Contract (team) | Go tests green with `-race` | 1325/1325 pass | `go test -race -json ./...` | build-and-test | Met |
| T-CONTRACT | Testing Contract (team, project note: 10 runs) | contract test passes on `min_kandev_version` | 10/10 pass on v0.96.0 | `make contract-test` ×10 | build-and-test | Met |
| T-REGRESSION | Testing Contract scope floor (bugfix) / requirements NFR3 | targeted regression present and green | Windows-listed manifest and stray `.exe` rejected; tests pass | `internal/pkgverify/pkgverify_test.go` | build-and-test | Met |
| T-SUITE | Testing Contract scope floor / NFR2 | existing suite green | all Go, UI, lint, secret checks pass | commands above | build-and-test | Met |
| T-SIZE | requirements NFR1 | package ≤ 25 MB | 23,393,498 bytes | `ls -l dist/` | build-and-test | Met |
