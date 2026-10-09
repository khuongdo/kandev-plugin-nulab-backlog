# Build and Test Summary — 261009-no-workflow-error

## Overall Status

Build: success. Tests: all passing. Prerequisites: Go 1.26.x (`~/.local/go`), Node/npm, `../kandev` at `v0.96.0`.

## Test Type Inventory

- Unit tests (Go + Vitest) from Code Generation, including the bugfix regression (`start-task.test.tsx`: null context + workspace with a workflow → dialog opens with `workflowId: null`).
- Packaged-host contract test (automated integration on Kandev v0.96.0, x10).
- Real-Kandev manual check (bugfix regression on the real host, mobile + desktop widths).
- Security checks via gosec and the host-error leak test (security-test-instructions.md).
- Minimal strategy: no separate load/performance suite (performance-test-instructions.md explains how NFR2 is checked).

## Coverage Expectations

Zero-Unit bugfix: Go line coverage floor 80% for `./internal/...` and `./server/...` (team rule); new Go code fully covered.

## Target Verification Matrix

Inventory sources: `construction/code-generation/code-generation-plan.md` § Testing Contract (no `nfr-requirements/` or `nfr-design/` artifacts exist in this bugfix scope) and `inception/requirements-analysis/requirements.md` NFR1-NFR6.

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-COV | Testing Contract (team) — coverage floor | ≥ 80% Go line coverage | 92.9% | `make coverage` output, test-results.md | build-and-test | Met |
| TC-RACE | Testing Contract (team) — `go test -race` | all packages pass with `-race` | 13/13 ok | `go test -race ./...` | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team/project) — contract test on min Kandev, 10 runs | 10/10 pass on v0.96.0 | 10/10 | `build/contract-*.log`, test-results.md | build-and-test | Met |
| TC-REGRESSION | Testing Contract (bugfix scope floor) — targeted regression | regression test present and passing | present, passing | `ui/src/page/start-task.test.tsx` | build-and-test | Met |
| TC-SUITE | Testing Contract (bugfix scope floor) — existing suite green | 0 failures | Go 13/13, Vitest 462/462 | test-results.md | build-and-test | Met |
| NFR1 | requirements.md NFR1 | inline alerts wrap, no horizontal scroll at 320–1280 px | wrapping full-width alerts; user confirmed good display on mobile and desktop | `ui/src/error-alert.tsx` + tests; manual check 2026-10-09 | build-and-test | Met |
| NFR2 | requirements.md NFR2 | ≤ 1 s, at most one host call | one `Limit: 1` host call; dialog opened promptly on real Kandev | `workflow_actions_test.go`; manual check | build-and-test | Met |
| NFR3 | requirements.md NFR3 | `role="alert"`, text always present | asserted | `ui/src/error-alert.test.tsx` | build-and-test | Met |
| NFR4 | requirements.md NFR4 | no host/Backlog text or secrets in errors | generic `internal` code only; gosec 0 issues | `TestWorkflowsStatus_HostErrorLeaksNoText`, `make lint` | build-and-test | Met |
| NFR5 | requirements.md NFR5 | TDD with regression, `-race`, 80% floor, suites green | as above | code-summary.md Red evidence; test-results.md | build-and-test | Met |
| NFR6 | requirements.md NFR6 | works on Kandev v0.96.0 with new capability | contract 10/10; real host check passed | test-results.md | build-and-test | Met |

## Readiness Assessment

- Build-ready: yes.
- Test-ready: yes — all commands passed and every target is `Met`.
- Deployment-ready: yes, pending release-version choice; the README upgrade note covers the new `api_read: workflows` permission.

## Known Limitations / Outstanding Items

- Review findings carried forward (minor): no pending indicator on "+ Task" while the workflow check runs (R-02); host errors return the generic `internal` code rather than an "unavailable" code (R-01, recorded deviation).
- The three watch dialogs' workflow lookup remains a follow-up (out of scope, Q2=A).
- The new save-query dialog alert also appears on the Settings page (shared component).
