# Code Generation Plan — CI path filter

Zero-Unit `express` work: one implementation iteration, scoped from `inception/requirements-analysis/requirements.md` (FR1–FR5, NFR1–NFR5) and the CodeKB (`aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`). Brownfield: modify files in place.

## Design Summary

- **Classification lives in Go** (`internal/ci/changes.go`), exposed as a new `cmd/ci changes` subcommand, following the existing `cmd/ci <subcommand>` + `internal/ci` pattern (workflows, secrets, preflight, contract). This is the single definition of app vs non-app paths (FR1.3) and is unit-tested (NFR4).
- **`ci.yml`** keeps its triggers (FR2.1). A new first job `changes` checks out the plugin with full history (plus the pinned Kandev SDK checkout, because the module's `replace ../kandev` makes any `go run` need it), runs `go run ./cmd/ci changes -base <sha> -head <sha>`, and exposes the output `app=true|false`. `checks` gets `needs: changes` and `if: needs.changes.outputs.app == 'true'`; `packaged-host-contract` keeps `needs: checks` (a skipped `checks` skips it too, reported as passing). Job names `checks` and `packaged-host-contract` are unchanged (FR2.3).
  - Base/head: `pull_request` → `github.event.pull_request.base.sha` and `github.event.pull_request.head.sha` (three-dot merge-base diff); `push` → `github.event.before` and `github.sha`.
  - Fail-safe (FR2.4): an empty/all-zero base, an unknown commit, a `git diff` error, or an empty file list prints `app=true` (with a warning on stderr), never a failure that blocks the pipeline.
- **New `.github/workflows/secrets.yml`** (workflow `secrets`, job `secret-scan`): runs on every `pull_request` to `main` and `push` to `main`, no path filter; checks out the plugin and the pinned Kandev SDK commit (git clone only, no Kandev build), sets up Go, runs `make check-secrets` (FR3, NFR2). Top-level `permissions: contents: read`, every action pinned by the same full SHAs already used in `ci.yml` (NFR1, NFR3).
- **`check-secrets` in `checks`**: removed from the `make` list in `ci.yml` per the approved assumption, UNLESS `internal/ci/secrets.go` scans generated `build/` or `dist/` files that only exist inside `checks` (for example `build/coverage.out`); in that case keep it in `checks` as well and record the deviation in `code-summary.md`.
- **`release.yml` unchanged** (FR4). Ruleset update is manual (FR5); the exact `gh api` command is written into `code-summary.md` for the user.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "express",
  "test_strategy": "minimal",
  "project_type": "brownfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    },
    {
      "layer": "project",
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07)"
    }
  ],
  "obligations": {
    "strategy": "minimal",
    "strategy_volume": [
      "One verifiable test per requirement at the narrowest effective level.",
      "At least one happy-path unit test per component.",
      "Unit tests are the default; a bugfix/security scope floor may require an integration or E2E regression when that is the narrowest level that reproduces the defect."
    ],
    "scope_floor": [
      "Keep the existing test suite green.",
      "This scope adds no extra new-test floor beyond the selected test strategy."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Verify the existing test runner/configuration and record the exact unit-scoped command.",
    "runner_ready_before_first_test": true,
    "testable_layers": [
      "Data model / database behavior",
      "Repository / data access",
      "Business logic",
      "API / endpoint",
      "Frontend behavior"
    ],
    "steps": [
      "Project structure and production configuration skeleton.",
      "Verify the existing test runner/configuration and record the exact unit-scoped command.",
      "Data model / database behavior - Red: write the failing tests and record the failing command output.",
      "Data model / database behavior - Green: implement only enough behavior to pass.",
      "Data model / database behavior - Refactor: improve the implementation while tests stay green.",
      "Repository / data access - Red: write the failing tests and record the failing command output.",
      "Repository / data access - Green: implement only enough behavior to pass.",
      "Repository / data access - Refactor: improve the implementation while tests stay green.",
      "Business logic - Red: write the failing tests and record the failing command output.",
      "Business logic - Green: implement only enough behavior to pass.",
      "Business logic - Refactor: improve the implementation while tests stay green.",
      "API / endpoint - Red: write the failing tests and record the failing command output.",
      "API / endpoint - Green: implement only enough behavior to pass.",
      "API / endpoint - Refactor: improve the implementation while tests stay green.",
      "Frontend behavior - Red: write the failing tests and record the failing command output.",
      "Frontend behavior - Green: implement only enough behavior to pass.",
      "Frontend behavior - Refactor: improve the implementation while tests stay green.",
      "Environment/build configuration.",
      "Documentation and traceability."
    ]
  },
  "input_sha256": "sha256:4442c5fad82de74b31fc5e573b07e7444027e7437773b5002fb422bdfb2749f7",
  "contract_sha256": "sha256:6db87aafc6be218a4f8f762cfd33d3fec2e691661f16cdc25e6dcc4236038b2d"
}
```

Inapplicable layers (no change): data model / database, repository / data access, frontend behavior. Applicable testable layers: business logic (path classification, fail-safe) and the CLI/endpoint layer (`cmd/ci changes` output, workflow policy over the new and changed workflow files).

## Steps

- [x] **Step 1 — Structure check.** Confirm no file other than those listed under "Files" below needs changing; read `internal/ci/run.go` (subcommand dispatch), `internal/ci/workflows.go`/`workflows_test.go` (does any test pin the content of `ci.yml`?), and `internal/ci/secrets.go` (which paths the scanner walks or skips, to settle the `check-secrets` decision in the Design Summary). Traces: FR2, FR3.
- [x] **Step 2 — Runner readiness.** Verify `../kandev` is at `.kandev-sdk-ref` (`make check-sdk`) and that `go test -race ./internal/ci/ -run 'TestClassifyChanges|TestChangesCommand|TestWorkflows'` runs (the existing workflow tests pass; the new tests do not exist yet). Record the command in `unit-test-instructions.md`.
- [x] **Step 3 — Business logic, Red.** Add `internal/ci/changes_test.go`: a table-driven `TestClassifyChanges` with at least: only `aidlc/` + `docs/` files → non-app; `docs/` + `internal/` → app; only `.github/workflows/ci.yml` → app; only `README.md`, `LICENSE`, `.gitignore`, `.claude/x` → non-app; `docs` prefix look-alikes such as `docsite/a.go` or `aidlc.go` → app; empty list → app (fail-safe). Run it and record the failing output. Traces: FR1.1, FR1.2, FR2.4, NFR4.
- [x] **Step 4 — Business logic, Green.** Add `internal/ci/changes.go` with the single non-app list (`aidlc/`, `.claude/`, `docs/` as directory prefixes; `README.md`, `LICENSE`, `.gitignore` as exact paths) and `ClassifyChanges(paths []string) bool` (true = app). Minimum code to pass. Traces: FR1.3.
- [x] **Step 5 — Business logic, Refactor** while green (doc comments, naming per team Code Style).
- [x] **Step 6 — CLI layer, Red.** Add tests for the `changes` subcommand (in `internal/ci/run_test.go` or `changes_test.go`, matching how other subcommands are tested): with an injected git runner, (a) a non-app diff prints `app=false`; (b) an app diff prints `app=true`; (c) an empty base, an all-zero base (`0000…`), and a git error each print `app=true` and a warning on stderr, exit 0; (d) missing `-head` is a usage error. Record the failing output. Traces: FR2.2, FR2.4.
- [x] **Step 7 — CLI layer, Green.** Wire `changes -base <sha> -head <sha>` into `internal/ci/run.go`/`cmd/ci/main.go`, running `git diff --name-only <base>...<head>` through an injectable runner (`context.Context` first, errors wrapped with `%w`), printing `app=<bool>` on stdout for `$GITHUB_OUTPUT`. Traces: FR2.2.
- [x] **Step 8 — CLI layer, Refactor** while green.
- [x] **Step 9 — Workflow policy, Red.** If `internal/ci/workflows_test.go` checks real repo workflow files, add a case asserting that `ci.yml` keeps jobs named `checks` and `packaged-host-contract`, that `checks` is gated on `needs.changes.outputs.app`, and that `secrets.yml` exists with job `secret-scan` triggered on `pull_request` and `push` to `main` without `paths`/`paths-ignore`; otherwise add the smallest such test next to it. Record the failing output. Traces: FR2.1, FR2.3, FR3.1, FR3.2, FR4.1.
- [x] **Step 10 — Environment/build configuration (Green for Step 9).**
  - Edit `.github/workflows/ci.yml`: add the `changes` job (checkout plugin with `fetch-depth: 0`, read SDK ref, checkout Kandev at the SDK ref, setup-go, run `go run ./cmd/ci changes` with the event's base/head passed via `env:`, not inline `${{ }}` in `run:`), add `needs: changes` + `if:` to `checks`, and apply the `check-secrets` decision from Step 1.
  - Add `.github/workflows/secrets.yml` as in the Design Summary.
  - Leave `.github/workflows/release.yml` untouched.
  - Run `make lint` (actionlint + `cmd/ci workflows`) and the Step 9 tests green.
- [x] **Step 11 — Regression.** Run `make check-format vet lint test coverage check-secrets` locally; the existing suite stays green and coverage stays at or above 80% (never lower the floor or add exclusions). Traces: NFR1, NFR4, NFR5.
- [x] **Step 12 — Documentation and traceability.** Update the README's CI/development section if it describes when CI runs; write `code-summary.md` (including the exact `gh api` command to add `secret-scan` to ruleset 24580280 for FR5, to run after the check has reported once), `source-manifest.json` and `traceability.json` (FR1.1–FR5.1, NFR1–NFR5).

## Files

- `internal/ci/changes.go` (new), `internal/ci/changes_test.go` (new)
- `internal/ci/run.go`, `internal/ci/run_test.go`, `cmd/ci/main.go` (if dispatch lives there), `internal/ci/workflows_test.go` (only if Step 9 needs it)
- `.github/workflows/ci.yml` (modified), `.github/workflows/secrets.yml` (new)
- `README.md` (only if it documents CI triggers)

## Story / Requirement Traceability

| Step | Requirements |
|------|--------------|
| 3–5 | FR1.1, FR1.2, FR1.3, FR2.4, NFR4 |
| 6–8 | FR2.2, FR2.4 |
| 9–10 | FR2.1, FR2.3, FR3.1, FR3.2, FR3.3, FR4.1, NFR1, NFR2, NFR3 |
| 11 | NFR1, NFR4, NFR5 |
| 12 | FR5.1 |
