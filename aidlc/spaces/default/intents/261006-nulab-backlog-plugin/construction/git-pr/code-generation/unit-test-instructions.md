# Unit Test Instructions — git-pr (U4)

## Frameworks and Setup

- **Go.** `testing` with `github.com/stretchr/testify/require`; table-driven with `t.Run`; behaviour-describing names; always `-race` (needs CGO). Go 1.26 on `PATH` (the toolchain under `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin` with `GOTOOLCHAIN=local`). `../kandev` at the commit in `.kandev-sdk-ref` (v0.96.0).
- **Virtual time.** `testing/synctest`. `backlog.Client` takes `Now` and `Wait`; the connection store takes `Now`; `git.Watcher` takes `Every` and `Now` and runs its ticker inside the bubble. No real `time.Sleep`. Watch cycles, 429 waits, queue spacing and the 15-minute credential expiry are asserted on the virtual clock.
- **Fake Backlog.** `net/http/httptest` TLS servers with JSON in `internal/backlog/testdata/`: new `repositories_ok.json`, `repositories_empty.json`, `pullrequests_ok.json`, `pullrequest_ok.json`, `pullrequest_created.json`, `issue_ok.json`; from U1/U2 `error_401.json`, `error_429.json`, `error_500_bait.json`. The Git smart-HTTP probe is served on `/git/<PROJ>/<repo>.git/info/refs`. The fake counts requests per path and in flight per group.
- **Kandev host.** In-memory `SecretStore`/`StateStore` fakes (`internal/connection/fakes_test.go`, plus a small local fake in `internal/git`); a fake `git.HostPort` recording created tasks and metadata and able to fail on the Nth create; the fake `pluginsdk.Host` in `internal/plugin` extended with `Tasks()` (`Create`, `List`) and `Repositories()` (`List`).
- **UI.** Vitest with `jsdom` (`ui/vitest.config.ts`); shared fake `host` in `ui/src/testing/harness.ts` extended with `registerRepositoryProvider`, `registerReviewProvider`, `registerTaskAction`, `openTaskLinkDialog`, `openModal`, `context.getTaskCreationContext`; `axe-core`; `vi.useFakeTimers()`.

## Commands for This Unit

Run Go commands from the repository root and UI commands from `ui/`.

- **Go, U4 tests only, runnable now** (before Step 1 creates `internal/git`):

  ```
  go test -race ./internal/backlog/... ./internal/connection/... ./internal/plugin/... -run '^TestU4_'
  ```

- **Go, U4 tests only, from Step 1 on:**

  ```
  go test -race ./internal/backlog/... ./internal/connection/... ./internal/git/... ./internal/plugin/... -run '^TestU4_'
  ```

- **Go, one layer while working (example):**

  ```
  go test -race ./internal/git/... -run '^TestU4_Watch'
  ```

- **Go, coverage of the packages U4 changes:**

  ```
  go test -race -coverprofile=coverage.out ./internal/backlog/... ./internal/connection/... ./internal/git/... ./internal/plugin/...
  ```

  Then `go tool cover -func=coverage.out`. `make coverage` checks the floor over the whole scope.

- **UI, U4 test files only:**

  ```
  npx vitest run src/git/git-state.test.ts src/git/git-access.test.tsx src/git/repository-provider.test.ts src/git/pr-link.test.ts src/git/review-provider.test.tsx src/git/create-pr.test.ts src/git/watches-page.test.tsx src/git/watch-form.test.tsx src/git/dashboard-page.test.tsx --passWithNoTests
  ```

  `--passWithNoTests` keeps this runnable before Step 15; from Step 15 on all nine files exist and must run.

- **U2 regression for files U4 edits:**

  ```
  npx vitest run src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/settings/settings.test.tsx src/switch src/page
  ```

  ```
  go test -race ./internal/backlog/... ./internal/connection/... ./internal/plugin/... -run '^Test(OAuth|Token|RateLimit|Queue|Projects|Recheck|Disconnect|Replace|SpaceChange|Restore|ConnectionChanged|ConnectionReader|Webhook|APIKeyHeader|ActionFailureLog|U2)'
  ```

## Coverage Target

At least **80% line coverage** over `./internal/...` and `./server/...`, excluding only `server/main.go` (team Testing Posture, NFR8). The floor is never lowered and no exclusion is added. U4 aims for ≥ 85% in `internal/git` and keeps `internal/backlog`, `internal/connection`, `internal/plugin` at or above their U2 levels.

## Expected Test Volume (Standard strategy)

| Component | Test file(s) | Approximate tests |
|-----------|--------------|-------------------|
| `internal/backlog` repository and PR types | `repositories_types_test.go`, `pullrequests_types_test.go` | 5 + 7 (+ table rows) |
| `internal/backlog` git calls and Git probe | `git_client_test.go` | 8 |
| `internal/connection` Git credential types | `git_credential_types_test.go` | 5 |
| `internal/connection` Git secret store (U4 cases) | `store_test.go` | 6 |
| `internal/connection` set/read/test/delete | `git_credential_test.go` | 8 |
| `internal/git` entities and parsers | `types_test.go` | 8 tables |
| `internal/git` state store | `store_test.go` | 6 |
| `internal/git` repository source | `repos_test.go` | 6 |
| `internal/git` link, unlink, status | `links_test.go`, `status_test.go` | 6 + 5 |
| `internal/git` create PR | `create_test.go` | 7 |
| `internal/git` credential scope | `resolver_test.go` | 6 |
| `internal/git` watches and watcher | `watches_test.go`, `watcher_test.go` | 6 + 8 |
| `internal/git` events, impact, queries, leak | `events_test.go`, `impact_test.go`, `queries_test.go`, `leak_test.go` | 7 + 2 + 5 + 2 |
| `internal/plugin` U4 actions | `actions_u4_test.go` | 8 |
| `internal/plugin` credential RPCs, manifest, wiring | `credential_test.go`, `manifest_test.go` (U4), `runtime_u4_test.go` | 5 + 3 + 4 |
| `ui/src/git` helpers | `git-state.test.ts` | 6 |
| `ui/src/git` Git access block | `git-access.test.tsx` | 5 |
| `ui/src/git` repository provider | `repository-provider.test.ts` | 6 |
| `ui/src/git` link dialog | `pr-link.test.ts` | 5 |
| `ui/src/git` review provider | `review-provider.test.tsx` | 6 |
| `ui/src/git` create PR transport | `create-pr.test.ts` | 6 |
| `ui/src/git` watches page and form | `watches-page.test.tsx`, `watch-form.test.tsx` | 7 + 6 |
| `ui/src/git` dashboard | `dashboard-page.test.tsx` | 6 |

Integration tests at key boundaries: gateway ↔ fake Backlog (`git_client_test.go`); connection ↔ gateway fake ↔ store fakes (`git_credential_test.go`); GitIntegration ↔ connection fake ↔ host-port fake ↔ store (`watcher_test.go`, `events_test.go`, `create_test.go`); Kandev adapter ↔ fake Host (`actions_u4_test.go`, `credential_test.go`, `runtime_u4_test.go`).

## Mocking and Stubbing

- Never call real Backlog, real Git, or a real Kandev in automated tests.
- Use fakes, not mocks with call expectations. Assert "0 requests", "exactly 3 tasks" and "no second create" by counting on the fake server and the fake host port.
- Inject time and waiting (`Now`, `Wait`, `Every`), run timer tests inside `synctest`; use `vi.useFakeTimers()` in the UI.
- Concurrency tests pass under `-race`: `Run` racing a tick, a concurrent create PR and token refresh on the Update queue, parallel link writes in one workspace.
- Test a restart by building a new `Watcher` and `Service` over the same in-memory store.
- Every secret in a test (API key, access token, Git password) comes from `internal/testutil` helpers (fresh random per run). Leak tests assert no 8-character window of the random part appears in logs, errors, action replies, task descriptions, credential RPC errors or events.

## Test Data

- Keep each test's data local; use `t.TempDir()`/`t.Setenv()` where files or the environment are touched.
- JSON fixtures contain only fake data: users (`"Test User"`, `"Lan"`), hosts (`example-space.backlog.com`), projects (`PROJ`, `DEMO`), repositories (`web-app`, `api`). Lists of 25 PRs are generated in the test.
- No password, key or token is committed, so `make check-secrets` stays clean.
- Kandev identifiers are fixed fake strings (`ws-1`, `task-17`, `repo-k1`, `session-1`).
