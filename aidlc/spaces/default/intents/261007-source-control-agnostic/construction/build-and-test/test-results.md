# Test Results — Multi-provider source control

Run on 2026-10-07 in the worktree `feature/source-control-agnos-2jr` (uncommitted changes from Code Generation), Go 1.26.8, Node 25.2.1, `../kandev` at v0.96.0.

## Build

| Command | Result |
|---|---|
| `make check-format` (gofmt, prettier) | Success |
| `make vet` | Success |
| `make lint` (golangci-lint v2.14.0 + gosec: `0 issues.`; `tsc --noEmit`; eslint; actionlint; CI workflow check) | Success |
| `make check-secrets` | Success (`ci secrets: OK`) |
| `make build` (UI bundle + all platform binaries) | Success |
| `make package` | Success (`dist/nulab-backlog-0.3.0.tar.gz`) |
| `make verify-package` | Success (`verifypkg: OK ... (nulab-backlog@0.3.0)`) |

`build/` and `dist/` were removed afterwards with `make clean`; no `coverage.out` remains in the repository root.

## Tests

| Suite | Command | Total | Passed | Failed | Skipped |
|---|---|---|---|---|---|
| Go, whole module (`-race`) | `go test -race -count=1 -json ./...` | 1324 | 1324 | 0 | 0 |
| Go, this work's scoped command 1 | `go test -race -count=1 ./internal/scm/... ./internal/github/... ./internal/gitlab/... ./internal/bitbucket/...` | 83 top-level | 83 | 0 | 0 |
| Go, this work's scoped command 2 | `go test -race -count=1 -run 'SCM\|Scm\|Manifest\|V030' ./internal/plugin/...` | 20 top-level | 20 | 0 | 0 |
| UI, whole suite | `npx vitest run` (in `ui/`) | 364 (33 files) | 364 | 0 | 0 |
| UI, this work's scoped command | `npx vitest run src/settings/source-control-section.test.tsx src/git/pr-list.test.tsx src/git/git-state.test.ts src/git/watch-form.test.tsx src/issues/issue-panel.test.tsx` | 53 (5 files) | 53 | 0 | 0 |
| Packaged-host contract test | `make contract-test KANDEV_MIN_DIR=../kandev`, 10 runs | 10 | 10 | 0 | 0 |
| NFR8 check | `go test -race -count=1 -run 'NFR8\|FirstPage' -v ./internal/scm/...` | 1 | 1 | 0 | 0 |
| Security-focused selection | `go test -race -count=1 -run 'Redact\|Token\|Secret\|Host\|Admin\|Guard' ./internal/scm/... ./internal/plugin/...` | 44 | 44 | 0 | 0 |

Baseline before Code Generation (from `code-summary.md`): Go 1201 passed, Vitest 322 passed (31 files), coverage 92.8%. No existing test was removed; the existing suite stays green.

## Failure Details

None.

## Coverage

`make coverage`: **92.8%** total line coverage over `./internal/...` and `./server/...` (floor 80%, only `server/main.go` excluded, as before). New packages (from Code Generation's run): `scm` 93.1%, `github` 91.7%, `gitlab` 88.7%, `bitbucket` 90.2%, `plugin` 93.8%.

## Target Verification Matrix

See `build-and-test-summary.md` § Target Verification Matrix (all targets `Met`).
