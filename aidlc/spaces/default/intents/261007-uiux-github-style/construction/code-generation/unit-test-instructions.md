# Unit Test Instructions — 261007-uiux-github-style

Zero-Unit refactor intent: "this unit" is the whole change, limited to the packages and UI folders it touches.

## Test Framework Setup

- Go: standard `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run`, always with `-race` (CGO on). Requires Go 1.26.x on `PATH` (`export PATH="$HOME/.local/go/bin:$PATH"`) and `../kandev` linked to the v0.96.0 checkout (`f099a46`, the commit in `.kandev-sdk-ref`).
- UI: Vitest with jsdom (`ui/vitest.config.ts`), axe-core for accessibility checks, the fake host in `ui/src/testing/harness.ts`. Node 22 (`.nvmrc`); `npm ci` in `ui/` if `node_modules` is missing.
- No new test framework or configuration file is needed.

## How to Run This Change's Tests

Run from the repository root. Both commands are runnable on the unchanged tree before the first Red step.

Go (packages this change touches):

```bash
PATH="$HOME/.local/go/bin:$PATH" go test -race ./internal/backlog/... ./internal/issues/... ./internal/git/... ./internal/plugin/...
```

A single behaviour during a Red/Green cycle (example):

```bash
PATH="$HOME/.local/go/bin:$PATH" go test -race ./internal/issues/... -run 'TestIssueWatch'
```

UI (folders this change touches):

```bash
cd ui && npx vitest run src/index.test.ts src/brand src/page src/settings src/issues src/git src/switch
```

Static checks for the UI after each Green step:

```bash
cd ui && npx tsc --noEmit && npx eslint . && npx prettier --check .
```

## Expected Coverage Targets

- Go: the team floor of 80% line coverage over `./internal/...` and `./server/...` stays in force (baseline 92.9%); measure with the profile under `build/`, never at the repository root:

```bash
mkdir -p build && PATH="$HOME/.local/go/bin:$PATH" go test -race -coverprofile=build/coverage.out ./internal/... ./server/... && go tool cover -func=build/coverage.out | tail -1 && rm build/coverage.out
```

- Minimal strategy volume: one test per requirement (FR1–FR6, NFR2–NFR4, NFR6) at the narrowest level, plus at least one happy-path test per new component (IssueWatch validation, store, gateway additions, watcher, PR list service, each new action, each new UI section/dialog, PR list, icon). The plan's tests for rules BR3.x exceed this floor where a rule guards against duplicate tasks or lost issues.
- The existing suites stay green (scope floor).

## Mocking and Stubbing Guidance

- Backlog: the `net/http/httptest` fake server with JSON fixtures under `internal/backlog/testdata/` (issue pages sorted by created, PR pages, PR count, 401, 429 with `Retry-After`). Never call real Backlog.
- Kandev host: the existing fake host ports in `internal/issues` and `internal/git` tests (task creation, state store); add a "workflow not found" failure mode for BR3.14.
- Time: inject the clock and wait functions (no `time.Sleep`); the watcher tick and per-watch interval are driven by the fake clock.
- UI: stub every host component in `harness.ts` as plain DOM with `data-testid`, and pass `variant`/`size` through as `data-variant`/`data-size` so button style rules are assertable.

## Test Data Management

- Fixtures are synthetic: fake space `example.backlog.com`, fake keys from `internal/testutil`; no real credentials anywhere (project rule).
- Tests are independent: `t.TempDir()`, `t.Setenv()`, fresh fake stores per test; UI tests create a fresh harness per test.
- Deleted UI pages (`git/watches-page.tsx`, `git/dashboard-page.tsx`) take their tests with them; behaviour they covered is re-asserted in the Settings and PR list tests.
