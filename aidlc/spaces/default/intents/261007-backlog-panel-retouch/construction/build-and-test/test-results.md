# Test Results — 261007-backlog-panel-retouch

Run date: 2026-10-08 (UTC), re-run after Loop-back 2 (rebased onto v0.4.0 `5724d88`; SCM pull-request list brought to the GitHub-style toolbar and task display; version 0.4.1). Environment: WSL2 Linux; Node/npm from `ui/`; Go 1.26.0 complete toolchain with `GOTOOLCHAIN=local`; `../kandev` at `v0.96.0`. Logs kept in the session scratchpad (`bt3/*.log`; earlier runs in `bt2/`, `bt/`).

## Build Status

| Command | Result | Evidence |
|---|---|---|
| `make check-format vet lint` | **Pass** (exit 0) | `ci workflows: OK` |
| `npm --prefix ui run typecheck && lint && format:check` | **Pass** (exit 0) | no findings |
| `make package` | **Pass** (exit 0) | `dist/nulab-backlog-0.4.1.tar.gz` |
| `make verify-package` | **Pass** (exit 0) | `verifypkg: OK dist/nulab-backlog-0.4.1.tar.gz (nulab-backlog@0.4.1)` |

## Test Results

| Suite | Total | Passed | Failed | Skipped | Evidence |
|---|---|---|---|---|---|
| Unit — scoped command from `unit-test-instructions.md` (8 files) | 147 | 147 | 0 | 0 | `Test Files 8 passed (8) / Tests 147 passed (147)` |
| Unit — full Vitest suite (`npm --prefix ui test`) | 373 | 373 | 0 | 0 | `Test Files 33 passed (33) / Tests 373 passed (373)`; v0.4.0 baseline included its SCM tests |
| Go — `go test -race ./internal/... ./server/...` (via `make coverage`) | all packages | all ok | 0 | 0 | no Go source changed |
| Packaged-host contract test on Kandev v0.96.0, 10 runs | 10 | 10 | 0 | 0 | each run `ci contract: OK nulab-backlog on Kandev v0.96.0` |

No failures; no failure details to report.

## Coverage

- Go: **92.8%** (floor 80%, excluded `server/main.go`) — `make coverage`; profile under `build/`, no `coverage.out` at the repo root.
- UI: no coverage floor is defined; every new helper and branch is exercised by the scoped tests (see code-summary.md).

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-SUITE-GREEN | code-generation-plan.md § Testing Contract (scope_floor) | Existing suite stays green | Vitest 373/373; Go all ok | vitest-full.log, coverage.log | build-and-test | Met |
| TC-MIN-PER-REQ | Testing Contract (strategy_volume) | ≥1 verifiable test per requirement; happy path per component | FR1–FR5, NFR1/3/4/5 each mapped to a test or gate | cross-unit-traceability.md | build-and-test | Met |
| TC-GO-COVERAGE-80 | Testing Contract (team note) | Go line coverage ≥ 80% (`./internal/...`, `./server/...`) | 92.8% | coverage.log | build-and-test | Met |
| TC-GO-RACE | Testing Contract (team note) | Go tests run with `-race`, pass | pass | coverage.log | build-and-test | Met |
| TC-CONTRACT-10 | Testing Contract (team + project notes) | Contract test on min Kandev version passes 10/10 | 10/10 | contract-1..10.log | build-and-test | Met |
| TC-NO-ROOT-COVERAGE | Testing Contract (project note) | No `coverage.out` at repo root | absent | `ls coverage.out` → not found | build-and-test | Met |
| NFR1 | requirements.md | Host components; no plugin CSS | controls test pass; package verified | vitest-scoped.log, verify.log | build-and-test | Met |
| NFR2-MIN-VERSION | requirements.md NFR2 | Package installs and runs on Kandev v0.96.0 | 10/10 contract runs | contract logs | build-and-test | Met |
| NFR2-UI-IN-HOST | requirements.md NFR2 | Host components render inside the plugin route on a real host | Not observed — no browser in this environment; contract test does not render UI | integration-test-instructions.md § Manual Real-Host UI Check | deployment-execution | Unverified |
| NFR3 | requirements.md | Accessible names, keyboard reach, safe external links | badge/toolbar/filter a11y tests + axe pass | vitest-scoped.log | build-and-test | Met |
| NFR4 | requirements.md | Go/UI suites green, coverage ≥ 80%, tests updated not deleted | as above | logs above | build-and-test | Met |
| NFR5 | requirements.md | No secret leakage; locked behaviour unchanged | redaction tests pass; no Go/manifest change | coverage.log, git status | build-and-test | Met |

## Outcome

Every command passed. One applicable target, **NFR2-UI-IN-HOST**, is `Unverified`: it needs a browser on a real Kandev host and is owned by the manual check in Deployment Execution. Under the stage's failure predicate an `Unverified` target counts as a failure, so the human decides at the halt-and-ask question whether to accept it and proceed.

No identifiable code fix exists for this target (it is an environment/observation gap, not a defect), so no loop-back to Code Generation applies for it.

## Accepted Failure

2026-10-08 (re-run after Loop-back 2): all commands passed again (Vitest 373/373, scoped 147/147, coverage 92.8%, `verify-package` OK for 0.4.1, contract 10/10); NFR2-UI-IN-HOST is still `Unverified` and now also covers the provider (GitHub/GitLab/Bitbucket) pull-request list. The human chose **Accept failure** again; the target stays `Unverified` and is carried to the manual real-host UI check in Deployment Execution.

2026-10-07 (re-run after Loop-back 1): all commands passed; NFR2-UI-IN-HOST remains `Unverified`. The human chose **Accept failure**: the target stays `Unverified` and is carried to the manual real-host UI check in Deployment Execution (checklist in integration-test-instructions.md).

## Loop-Back Log

### Loop-back 1 — 2026-10-07T13:30:00Z

- **Diagnosis:** At the halt-and-ask the human chose to fix the code review's open Minor findings before proceeding (code-generation review-01): R-01 the "one reload" test passes only because the fake dropdown does not blur the query input, while the real host commits on blur first (possible double reload); R-04 the page passes `loading` on every page change, so the host toolbar shows "…" for the count and disables refresh during pagination; R-05 the last checked PR status checkbox uses `disabled`, removing it from keyboard focus instead of `aria-disabled`.
- **Root-cause stage:** code-generation (generated source and test harness fidelity).
- **Planned fix:** R-01 make the `IntegrationRepositoryFilter`/toolbar fakes reproduce the host blur-before-select order and make the page coalesce a blur commit followed by a filter change into one reload (or correct the test title if the host order makes one reload impossible), with a test; R-04 pass the toolbar `loading` only for refresh/initial load, not page changes, with a test; R-05 keep the last checked status focusable with `aria-disabled="true"` and ignore its toggle, with a test.
- **Estimated impact:** effort ~1 hour; financial cost none; risk low (UI-only, covered by existing and new tests; NFR2-UI-IN-HOST stays with the manual check).

### Loop-back 2 — 2026-10-07T14:10:00Z

- **Diagnosis:** At the Deployment Pipeline gate the human reported that `v0.4.0` is already released (PR #9, GitHub/GitLab/Bitbucket pull requests). The branch was rebased onto `origin/main` (`5724d88`); `ui/src/git/pr-list.tsx` conflicts were merged, but the build no longer type-checks: `scm-pr-list.tsx` (new in v0.4.0) imports the removed `taskHref`, uses the removed `messages.taskLink` and passes the removed `idPrefix` to `createPrToolbar`; the provider selector in `pr-list.tsx` uses `Label`/`Select*` no longer destructured. The human chose to bring the SCM PR list to the same GitHub-style toolbar and linked-task display now, and to release as `0.4.1`.
- **Root-cause stage:** code-generation (upstream code merged after this intent's code was written).
- **Planned fix:** make the merged code compile and pass; apply FR5-equivalent behaviour to `scm-pr-list.tsx` (host `TaskRowIndicator` for linked tasks, `IntegrationRepositoryFilter`-style label-less dropdowns, lookalike toolbar layout, provider selector restyled); update PR #9's tests (`pr-list.test.tsx`, SCM tests) to the new test ids; bump `manifest.yaml` to `0.4.1`; re-run all gates and the contract test.
- **Estimated impact:** effort ~2 hours; financial cost none; risk medium-low (UI-only; touches v0.4.0's new SCM list, covered by its existing tests plus new ones).
