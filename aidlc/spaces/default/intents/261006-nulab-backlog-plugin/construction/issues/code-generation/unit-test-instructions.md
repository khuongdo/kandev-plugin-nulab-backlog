# Unit Test Instructions — issues (U3)

## Frameworks and Setup

- **Go.**
  - `testing` with `github.com/stretchr/testify/require`; table-driven tests with `t.Run`; test names that describe behaviour; always `-race` (needs CGO).
  - Go 1.26 on `PATH` (the toolchain under `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin` with `GOTOOLCHAIN=local`).
  - `../kandev` at the commit in `.kandev-sdk-ref` (v0.96.0).
  - Every U3 test function starts with `TestU3_`, and each underscore part is under 32 characters (U5 `check-secrets`).
- **Virtual time.** `testing/synctest`.
  - `backlog.Client` takes `Now` and `Wait`; `issues.Store` takes `Now`; `issues.Syncer` takes `Tick` and `Now` and runs its ticker inside the bubble; `Suggest` takes the injected wait for its 250 ms coalescing.
  - No real `time.Sleep`.
  - Cycles, the 2-minute interval, 429 waits, Search-queue spacing, the 300 ms fake delay (AC8.1.1) and the 10 s timeout (AC8.1.3) are asserted on the virtual clock. Timing tests use the in-memory `RoundTripper` from `internal/backlog/limiter_test.go`, so no socket is opened inside the bubble.
- **Fake Backlog.** `net/http/httptest` TLS servers shaped exactly like Backlog payloads.
  - Fixtures in `internal/backlog/testdata/`: new `issues_ok.json`, `issues_count_ok.json`, `issue_detail_ok.json`, `comments_ok.json`, `attachments_ok.json`, `statuses_ok.json`, `project_users_ok.json`; existing `issue_ok.json`, `projects_ok.json`, `error_401.json`, `error_429.json`, `error_500_bait.json`.
  - The fake counts requests per path, in flight per group, and **non-GET requests** (must stay 0, AC3.2.3).
- **Kandev host.**
  - In-memory `StateStore` fake in `internal/issues`.
  - A fake `issues.HostPort` that records created tasks (with `Identifier` `T-<n>`), lists tasks, and can fail on the Nth create.
  - The fake `pluginsdk.Host` in `internal/plugin`, extended for `Tasks().Create` returning `Identifier`, `Tasks().List` with deletion, and `OnEvent` delivery. It nests task metadata under `"plugin:nulab-backlog"` exactly as Kandev v0.96.0 does.
  - Fake `Connection` with `Current`, `Credentials`, `Subscribe` and `RequireEnabled`. A real `connection.Service` is used in `runtime_u3_test.go`.
- **UI.**
  - Vitest with `jsdom` (`ui/vitest.config.ts`); `axe-core`; `vi.useFakeTimers()` for the 400 ms search wait and the countdown.
  - The shared fake `host` in `ui/src/testing/harness.ts` is extended with `registerTaskPanel`, `registerTaskMenuAction`, `registerComponent`, `registerTranslations`, `i18n {locale, t}`, `useResponsiveBreakpoint`, `toast` and `ui.Skeleton`.

## Commands for This Unit

Run Go commands from the repository root and UI commands from `ui/`.

- **Go, U3 tests only, runnable now** (before Step 1 creates `internal/issues`):

  ```
  go test -race ./internal/backlog/... ./internal/plugin/... -run '^TestU3_'
  ```

- **Go, U3 tests only, from Step 1 on:**

  ```
  go test -race ./internal/backlog/... ./internal/issues/... ./internal/plugin/... -run '^TestU3_'
  ```

- **Go, one area while working (example):**

  ```
  go test -race ./internal/issues/... -run '^TestU3_Sync'
  ```

- **Go, coverage of the packages U3 changes:**

  ```
  go test -race -coverprofile=coverage.out ./internal/backlog/... ./internal/issues/... ./internal/plugin/...
  ```

  Then run `go tool cover -func=coverage.out`. `make coverage` checks the floor over the whole scope.

- **UI, U3 test files only:**

  ```
  npx vitest run src/issues/issues-state.test.ts src/issues/issues-page.test.tsx src/issues/link-task-dialog.test.tsx src/issues/issue-badge.test.tsx src/issues/task-menu.test.ts src/issues/issue-panel.test.tsx src/issues/poll-interval.test.tsx src/issues/i18n.test.ts --passWithNoTests
  ```

  `--passWithNoTests` keeps this runnable before Step 10. From Step 10 on, all eight files exist and must run.

- **Regression for shared files U3 edits:**

  ```
  npx vitest run src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/page/backlog-page.test.tsx src/index.test.ts
  ```

  ```
  go test -race ./internal/backlog/... ./internal/git/... ./internal/plugin/... -run '^Test(U4_|U2_|Projects|Queue|RateLimit)'
  ```

## Coverage Target

At least **80% line coverage** over `./internal/...` and `./server/...`, excluding only `server/main.go` (team Testing Posture, NFR8). The floor is never lowered and no exclusion is added. U3 aims for **≥ 85% in `internal/issues`** and keeps `internal/backlog` and `internal/plugin` at or above their U4 levels (95.8% and 93.3%).

## Expected Test Volume (Standard strategy)

| Component | Test file(s) | Approximate tests |
|-----------|--------------|-------------------|
| `internal/backlog` issue types and query encoding | `issues_types_test.go` | 7 (+ table rows) |
| `internal/backlog` issue calls (groups, errors, timeout, 0 writes) | `issues_client_test.go` | 8 |
| `internal/issues` entities, validation, mapping | `types_test.go` | 8 tables |
| `internal/issues` state store | `store_test.go` | 6 |
| `internal/issues` list and filters | `list_test.go`, `filters_test.go` | 8 + 4 |
| `internal/issues` create task | `create_test.go` | 7 |
| `internal/issues` link, unlink, task search | `links_test.go` | 7 |
| `internal/issues` detail and comments | `detail_test.go` | 7 |
| `internal/issues` suggestions and authorize | `suggest_test.go` | 7 |
| `internal/issues` sync cycle | `sync_test.go` | 8 |
| `internal/issues` events, impact, settings, leak | `events_test.go`, `impact_test.go`, `settings_test.go`, `leak_test.go` | 7 + 2 + 3 + 2 |
| `internal/plugin` U3 actions | `actions_u3_test.go` | 8 |
| `internal/plugin` reference RPCs, events, host port, manifest, wiring | `references_test.go`, `events_test.go`, `host_port_u3_test.go`, `manifest_test.go` (U3), `runtime_u3_test.go` | 5 + 3 + 3 + 3 + 3 |
| `ui/src/issues` state helpers | `issues-state.test.ts` | 7 |
| `ui/src/issues` issue list (M2, M2m) | `issues-page.test.tsx` | 8 |
| `ui/src/issues` link task dialog (M3) | `link-task-dialog.test.tsx` | 6 |
| `ui/src/issues` card badge (M6) | `issue-badge.test.tsx` | 6 |
| `ui/src/issues` task menu unlink | `task-menu.test.ts` | 5 |
| `ui/src/issues` task panel (M8) | `issue-panel.test.tsx` | 8 |
| `ui/src/issues` poll interval (M1) | `poll-interval.test.tsx` | 5 |
| `ui/src/issues` language | `i18n.test.ts` | 5 |

Integration tests at key boundaries:

- gateway ↔ fake Backlog: `issues_client_test.go`;
- IssueIntegration ↔ connection fake ↔ host-port fake ↔ store: `sync_test.go`, `events_test.go`, `create_test.go`, `list_test.go` (including the 300 ms latency budget);
- Kandev adapter ↔ fake Host: `actions_u3_test.go`, `references_test.go`, `events_test.go`, `runtime_u3_test.go` with a real `connection.Service`.

## Mocking and Stubbing

- Never call real Backlog or a real Kandev in automated tests.
- Use fakes, not mocks with call expectations. Assert "exactly 1 search", "0 non-GET requests", "0 Update calls", "one task" and "≤ 1 in flight" by counting on the fake server and the fake host.
- Fakes copy the real payload shapes (Backlog JSON, Kandev task metadata nesting and repository names), so a test cannot pass against a shape the real host never produces.
- Inject time and waiting (`Now`, `Wait`, `Tick`) and run timer tests inside `synctest`. Use `vi.useFakeTimers()` in the UI.
- Concurrency tests pass under `-race`: a refresh racing a tick; a double create-task; parallel link writes in one workspace; a list's Search calls racing a background cycle.
- Test a restart by building a new `Syncer` and `Service` over the same in-memory store.
- Every secret in a test (API key, access token) comes from the `internal/testutil` helpers, fresh and random per run. Leak tests use `testutil.AssertNoLeak` to check that no 8-character window appears in logs, errors, action replies, task descriptions, reference candidates or authorization reasons. Bait bodies come from `error_500_bait.json`.

## Test Data

- Keep each test's data local; use `t.TempDir()`/`t.Setenv()` where files or the environment are touched.
- JSON fixtures contain only fake data:
  - users `"Test User"` and `"Lan"`;
  - host `example-space.backlog.com`;
  - projects `PROJ` and `DEMO`;
  - issues `PROJ-118`, `PROJ-120`, `PROJ-123`;
  - one 200-character Japanese summary.

  The 57-issue lists and 150-comment lists are generated inside the test.
- No password, key or token is committed, so `make check-secrets` stays clean.
- Kandev identifiers are fixed fake strings: `ws-1`, `task-17` with key `T-17`, `wf-1`, `step-1`.
