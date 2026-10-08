# Code Summary — CI path filter

Zero-Unit `express` work, one iteration, TDD per the Testing Contract (`sha256:6db87aafc6be218a4f8f762cfd33d3fec2e691661f16cdc25e6dcc4236038b2d`). Applicable layers: business logic (classification) and CLI/endpoint (`cmd/ci changes`, workflow policy). Data model, repository and frontend layers are not applicable.

## Files

| File | Change | Requirements |
|------|--------|--------------|
| `internal/ci/changes.go` | New. `nonAppDirs` / `nonAppFiles` (the single app vs non-app list), `ClassifyChanges`, the `changes` subcommand, and the injectable `diffNames` git runner | FR1.1–FR1.3, FR2.2, FR2.4 |
| `internal/ci/changes_test.go` | New. `TestClassifyChanges`, `TestChangesCommand`, `TestDiffNamesReadsGit` | FR1.1, FR1.2, FR2.2, FR2.4, NFR4, NFR5 |
| `internal/ci/run.go` | Registers the `changes` subcommand; usage line lists it | FR2.2 |
| `internal/ci/workflows_test.go` | New tests over the real workflow files: `TestRepositoryWorkflowsSkipAppJobsOnlyForNonAppChanges`, `TestRepositoryWorkflowsScanEveryChangeForSecrets` | FR2.1, FR2.3, FR3.1–FR3.3, NFR1, NFR2 |
| `.github/workflows/ci.yml` | New first job `changes`; `checks` and `packaged-host-contract` gated; job names unchanged | FR2.1–FR2.4, NFR1–NFR3 |
| `.github/workflows/secrets.yml` | New workflow `secrets`, job `secret-scan` | FR3.1–FR3.3, NFR1–NFR3 |
| `README.md` | "CI checks" section describes the `changes` job, the non-app list, `secret-scan` and the three required checks | FR2, FR3, FR5.1 |

`cmd/ci/main.go` and `.github/workflows/release.yml` are unchanged (dispatch lives in `internal/ci/run.go`; FR4.1).

## Key Decisions

- **Classification**: a path is non-app when it starts with `aidlc/`, `.claude/` or `docs/`, or is exactly `README.md`, `LICENSE` or `.gitignore`. Everything else is app, including look-alikes (`docsite/a.go`, `aidlc.go`) and nested files such as `ui/README.md`. An empty list is app.
- **`changes` subcommand**: `go run ./cmd/ci changes -base <sha> -head <sha>` runs `git diff --name-only <base>...<head>` and prints `app=true|false` for `$GITHUB_OUTPUT`. Base and head must be full 40-hex SHAs, and base must not be all zeros; anything else (empty or all-zero push `before`, a branch name, an option-like value) never reaches git. That case, a git error, or an empty diff prints `app=true` plus a `ci changes: warning: …` line on stderr, and exits 0. A missing `-head` is a usage error (exit 2).
- **Event values** reach the shell only through `env:` (`BASE_SHA`, `HEAD_SHA`). `pull_request` uses `pull_request.base.sha`/`head.sha`; `push` uses `event.before`/`github.sha`. The checkout uses `fetch-depth: 0` so the merge base is available.
- **Fail-safe gating (stronger than the plan's literal `if:`)**: `checks` uses `if: ${{ !cancelled() && needs.changes.outputs.app != 'false' }}`, so it is skipped only on an explicit `app=false`; if the `changes` job itself fails (output empty), `checks` still runs (FR2.4). `packaged-host-contract` keeps `needs: checks` and adds `if: ${{ !cancelled() && needs.checks.result == 'success' }}`, because GitHub's implicit `success()` also looks at the failed transitive `changes` job and would otherwise skip the contract test, which GitHub reports as passing.
- **`check-secrets` stays in `checks` (deviation from the default in the plan's Design Summary, as that summary allows)**: `internal/ci/secrets.go` `inScanScope` scans `coverage.out` and `build/coverage.filtered.out`, which exist only after `make coverage` inside `checks`. `secret-scan` covers the committed tree on every change; `checks` keeps covering the generated coverage profile on app changes.
- **No Kandev checkout in `changes` or `secret-scan` (deviation from the plan, required by NFR2)**: the plan assumed every `go run` needs the `../kandev` replace target. A local check showed `go run ./cmd/ci changes` and `go run ./cmd/ci secrets -root .` both work with no `../kandev` directory and an empty module cache (`internal/ci` never imports the SDK; module graph pruning). So neither light job clones Kandev, which meets NFR2 as written ("no job checks out … Kandev"). `secret-scan` runs `make -o check-sdk check-secrets`: still the standard Makefile target, with only the `check-sdk` prerequisite marked as up to date. The workflow tests assert that neither job checks out `kdlbs/kandev`.
- **Workflow and job names**: workflow `secrets`, job `secret-scan` (unique; no clash with `checks` or `packaged-host-contract`).
- **No new dependencies** (NFR3): only standard library (`os/exec`, `regexp`, `slices`) and the already-used actions `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1` and `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`. `go mod tidy` leaves `go.mod`/`go.sum` unchanged.

## TDD Record

Scoped command (recorded in `unit-test-instructions.md`; the existing workflow tests are named `TestCheckWorkflows…`/`TestRepositoryWorkflows…`):

```bash
go test -race ./internal/ci/ -run 'TestClassifyChanges|TestChangesCommand|TestCheckWorkflows|TestRepositoryWorkflows'
```

Baseline before any change: `make check-sdk` OK; the 4 existing workflow tests pass; `go test -race ./internal/... ./server/...` green.

### Business logic — Red (Step 3)

```
$ go test -race ./internal/ci/ -run 'TestClassifyChanges'
internal/ci/changes_test.go:26:29: undefined: ClassifyChanges
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
```

Green (Step 4): `ok … internal/ci`. Refactor (Step 5): loop replaced by `slices.ContainsFunc`; still green.

### CLI layer — Red (Step 6)

```
$ go test -race ./internal/ci/ -run 'TestChangesCommand|TestDiffNames'
internal/ci/changes_test.go:51:21: undefined: diffNames
internal/ci/changes_test.go:124:16: undefined: diffNames
internal/ci/changes_test.go:128:11: undefined: diffNames
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
```

Green (Step 7): `ok … internal/ci`. Refactor (Step 8): single `if` instead of a one-case `switch`; git output split per line, not per whitespace (paths with spaces stay whole). Smoke run on this repo: the records-only commit `bf20039` prints `app=false`; `f5a7529` (v0.4.2 code change) prints `app=true`; an all-zero base prints the warning and `app=true`.

### Workflow policy — Red (Step 9)

```
$ go test -race ./internal/ci/ -run 'TestRepositoryWorkflows'
--- FAIL: TestRepositoryWorkflowsSkipAppJobsOnlyForNonAppChanges (0.00s)
        	Error:      	Should be true
        	Messages:   	ci.yml needs a changes job
--- FAIL: TestRepositoryWorkflowsScanEveryChangeForSecrets (0.00s)
        	Error:      	Received unexpected error:
        	Messages:   	secrets.yml must exist
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci	0.041s
```

Green (Step 10): workflow edits; tests pass; `make lint` clean (two gosec findings in the new tests were fixed with `//nolint:gosec // <reason>` per team Code Style).

## Regression Results (Step 11)

- `make check-format vet lint test coverage check-secrets`: all pass. golangci-lint `0 issues`; actionlint clean; `ci workflows: OK`; Vitest 33 files / 373 tests pass; Go suite green with `-race`.
- Coverage: **92.8%** total (floor 80%, exclusions unchanged: `server/main.go`); `internal/ci` 91.2%; `ClassifyChanges`, `isApp`, `changesCommand`, `changedPaths` 100%.
- `ci secrets: OK`. No `coverage.out` at the repository root (profile under `build/`).
- `.github/workflows/release.yml`: `git diff` empty (FR4.1).

## FR5.1 — Ruleset update (manual, after `secret-scan` has reported once)

The current ruleset 24580280 requires `checks` and `packaged-host-contract` (integration 15368, GitHub Actions). After the first pull request with this change has run `secret-scan`, add it while keeping every existing rule:

```bash
gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280 \
  --jq '{name, target, enforcement, conditions, bypass_actors, rules: [.rules[] | if .type == "required_status_checks" then .parameters.required_status_checks += [{"context": "secret-scan", "integration_id": 15368}] else . end]} | with_entries(select(.value != null))' \
  > ruleset-24580280.json
gh api -X PUT repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280 --input ruleset-24580280.json
gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280 --jq '[.rules[] | select(.type == "required_status_checks") | .parameters.required_status_checks[].context]'
rm ruleset-24580280.json
```

The last read should print `["checks","packaged-host-contract","secret-scan"]`.

## Deviations and Open Issues

- `check-secrets` kept in `checks` (see Key Decisions).
- No Kandev checkout in `changes`/`secret-scan`; `make -o check-sdk check-secrets` (see Key Decisions).
- Gating expressions are stricter than the plan's `needs.changes.outputs.app == 'true'` (see Key Decisions).
- **FR3 acceptance gap (not fixed, outside the approved plan):** the scanner's scope (`inScanScope` in `internal/ci/secrets.go`) covers only `_test.go` files, `testdata/`, `docs/manual-checks/`, UI test files and coverage profiles. It does **not** read `aidlc/` or other `docs/` files, so the FR3 acceptance case "a records-only pull request that adds a file under `aidlc/` containing a test credential pattern … fails" does not hold today. `secret-scan` runs on every change as required, but widening the scan scope is a separate decision (it may flag existing records).
