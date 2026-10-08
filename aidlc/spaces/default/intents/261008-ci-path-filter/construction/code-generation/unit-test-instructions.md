# Unit Test Instructions — CI path filter

## Framework and Setup

- Go `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, behaviour-describing names (team Testing Posture).
- Prerequisite: `../kandev` checked out at the commit in `.kandev-sdk-ref` (`make check-sdk` passes). The module's `replace ../kandev` means every `go test`/`go run` needs it.
- Go tests always run with `-race`.
- No new test configuration or dependency is needed; the existing `internal/ci` test setup is reused.

## How to Run This Work's Tests

Exact, scoped command (runnable before the first Red step; it already runs the existing workflow tests):

```bash
go test -race ./internal/ci/ -run 'TestClassifyChanges|TestChangesCommand|TestCheckWorkflows|TestRepositoryWorkflows'
```

Recorded at Step 2 (2026-10-08): the existing workflow tests are named `TestCheckWorkflows…` and `TestRepositoryWorkflows…`, not `TestWorkflows…`, so the `-run` pattern above uses those prefixes. Go 1.26.8; `make check-sdk` passes (`../kandev` at `.kandev-sdk-ref`). Before the change, the command passes the 4 existing workflow tests and the full suite `go test -race ./internal/... ./server/...` is green.

Workflow lint for the changed and new workflow files:

```bash
make lint
```

## Planned Tests (Minimal strategy: one per requirement, happy-path floor per component)

| Test | Covers |
|------|--------|
| `TestClassifyChanges` — non-app-only (`aidlc/`, `docs/`, `.claude/`, `README.md`, `LICENSE`, `.gitignore`) → non-app | FR1.1 |
| `TestClassifyChanges` — mixed `docs/` + `internal/` → app; `.github/workflows/ci.yml` → app; look-alikes `docsite/a.go`, `aidlc.go` → app | FR1.2 |
| `TestClassifyChanges` — empty list → app | FR2.4 |
| `TestChangesCommand` — injected git runner, non-app diff prints `app=false`, app diff prints `app=true` | FR2.2 |
| `TestChangesCommand` — empty base, all-zero base, git error → `app=true` + stderr warning, exit 0; missing `-head` → usage error | FR2.4 |
| `TestWorkflows…` (existing file, new case) — `ci.yml` keeps `checks` and `packaged-host-contract`, `checks` gated on `needs.changes.outputs.app`; `secrets.yml` has job `secret-scan` on `pull_request`/`push` to `main`, no path filter | FR2.1, FR2.3, FR3.1, FR3.2 |

The test names above are exact enough for the `-run` filter; if the existing workflow tests use a different prefix than `TestWorkflows`, update the `-run` pattern in this file before the first Red step (it is still scoped to `./internal/ci/`).

## Coverage Targets

- Project floor stays 80% line coverage over `./internal/...` and `./server/...` (`make coverage`); new code in `internal/ci` is covered by the tests above. Never lower the floor or add exclusions.
- The coverage profile stays under `build/`; never leave `coverage.out` at the repository root.

## Mocking / Stubbing

- `git diff` is reached through an injected runner function so tests never shell out to git; no network.
- Workflow-file tests read the real `.github/workflows/*.yml` from the repo root, or fixtures in `internal/ci/testdata/` if the existing tests use that pattern.

## Test Data

- File-path lists are inline in the table tests.
- Any credential-shaped test input uses obviously fake values; no real credentials (project Forbidden).
