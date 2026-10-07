# Unit Test Instructions — github-parity-actions

## Strategy and Scope

- Strategy: **Minimal**, run test-first (TDD) per the Testing Contract. Each requirement group in `requirements.md` gets one verifiable test, and every new component gets at least one happy-path test. Per the construction phase rule, each new test file covers the happy path plus at least two error or edge cases.
- Expected new tests: about 14 (Go ~8, Vitest ~6). The existing suite must stay green: 9 Go packages and 286 Vitest tests (baseline from `developer-scan.md`).
- Coverage target: at least 80% Go line coverage over `./internal/...` and `./server/...` (enforced by `make coverage`). No threshold may be lowered.

## Framework Setup

- Go: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, always run with `-race`. Requirements: Go 1.26.x on `PATH` (`export PATH=$HOME/.local/go/bin:$PATH`) and `../kandev` pointing at the Kandev `v0.96.0` checkout (the symlink already exists next to this worktree).
- UI: Vitest + jsdom (existing `ui/vitest.config.ts`). Host components are faked in `ui/src/testing/harness.ts`. The new fakes are added in plan Step 12. Run `npm ci` in `ui/` once if `node_modules` is missing.

## How to Run This Change's Tests

Each command is limited to the files and tests of this change. Run them from the repository root.

```bash
export PATH=$HOME/.local/go/bin:$PATH

# Repository + business logic: quick actions, issue saved queries, assignee "me"
go test -race ./internal/issues/ -run 'QuickAction|IssueQuer|SaveQuery|SetQueryDefault|List'

# PR saved query default flag
go test -race ./internal/git/ -run 'Quer'

# Actions, manifest, redaction
go test -race ./internal/plugin/ -run 'QuickAction|IssueQuer|QueryDefault|Manifest|Leak|Redact'

# UI (from ui/)
(cd ui && npx vitest run src/page/quick-actions.test.ts src/page/start-task.test.tsx src/settings/sections.test.tsx src/page/backlog-lists.test.tsx src/page/backlog-page.test.tsx src/issues/issues-page.test.tsx)
```

Readiness (plan Step 2): before any change, these same commands must pass. The new test files do not exist yet at that point, so for the UI command run the existing files only (`src/settings/sections.test.tsx src/page/backlog-lists.test.tsx src/page/backlog-page.test.tsx src/issues/issues-page.test.tsx`). For each Red step, write the failing output into the progress notes.

## Mocking and Test Data

- Backlog is never called for real. Go tests use the existing `net/http/httptest` fake Backlog harness (`internal/issues/harness_test.go`, `internal/git/harness_test.go`) and the fake state store (`fakes_test.go`). The `me` resolution test asserts that the fake receives one `/api/v2/users/myself` call and an `assigneeId[]` equal to that user's id.
- Plugin state: use the in-memory `StateStore` fakes already in each package. Old-document compatibility (no `isDefault`) is tested by seeding a raw map without the field.
- UI: the harness records every `invokeAction(name, opts)` call. Tests assert the action name and body. The `TaskCreateDialog` fake exposes `initialValues` and a confirm button that calls `onSuccess({ id: "t-1" })`. A failing `issues.link` is simulated by having the harness reject that action.
- Test data contains no real credentials. Use the existing placeholder keys from `testdata/`.

## Full-Suite Verification (Build and Test, not per step)

`make check-format vet lint test coverage` (delete `coverage.out` afterwards), then in `ui/`: `npm run typecheck && npm run lint && npm run format:check && npm test`.
