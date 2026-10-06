# Unit Test Instructions — connection (U2)

## Frameworks and Setup

- **Go**: the standard `testing` package with `github.com/stretchr/testify/require`. Tests are table-driven with `t.Run`, and each name describes a behaviour. They always run with `-race`, which needs CGO. Go 1.26 must be on `PATH` (the U1/U5 runs used the toolchain under `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin` with `GOTOOLCHAIN=local`).
- **Virtual time**: `testing/synctest` from the Go standard library, as in U1. The `backlog.Client` takes an injected `Now` and `Wait func(ctx, time.Duration) error`, and `connection.Service` and `Store` take an injected clock. No test uses a real `time.Sleep`. Rate-limit waits, queue spacing, token expiry and the 10-minute OAuth state expiry are all asserted on the virtual clock.
- **Fake Backlog**: `net/http/httptest` TLS servers with fake JSON in `internal/backlog/testdata/` — from U1: `myself_*.json`, `error_401.json`, `error_429.json`; new: `token_ok.json`, `token_refreshed.json`, `token_invalid_grant.json`, `projects_ok.json`, `projects_empty.json`, `error_500_bait.json`. The fake counts requests per path and requests in flight per group; those counts give the "0 requests" and "never in parallel" assertions.
- **Kandev host**: the in-memory `SecretStore` and `StateStore` fakes in `internal/connection/fakes_test.go`, extended with `DeleteState`; the fake `pluginsdk.Host` in `internal/plugin`, extended with `GetConfig`. Both can inject failures.
- **UI**: Vitest with `jsdom` (`ui/vitest.config.ts`), the shared fake `host` in `ui/src/testing/harness.ts`, `axe-core` for accessibility, and `vi.useFakeTimers()` for the rate-limit countdown.

## Commands for This Unit

Run the Go commands from the repository root. `../kandev` must be at the commit in `.kandev-sdk-ref`.

- **Go, U2 tests only.** Every new U2 test name starts with one of the listed prefixes (see the plan's test naming rule):

  ```
  go test -race ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./internal/testutil/... -run '^Test(OAuth|Token|RateLimit|Queue|Projects|Recheck|Disconnect|Replace|SpaceChange|Restore|ConnectionChanged|ConnectionReader|Webhook|APIKeyHeader|ActionFailureLog|U2)'
  ```

  This runs now: before Step 3, each package reports `ok … [no tests to run]`.

- **Go, one layer while working (example):**

  ```
  go test -race ./internal/backlog/... -run '^Test(RateLimit|Queue)'
  ```

- **Go, coverage of the packages U2 changes:**

  ```
  go test -race -coverprofile=coverage.out ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./internal/testutil/...
  ```

  Then run `go tool cover -func=coverage.out`. `make coverage` checks the floor over the whole scope.

- **UI, U2 test files only** (run in `ui/`):

  ```
  npx vitest run src/settings/state.test.ts src/settings/oauth.test.tsx src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/settings/confirm-dialog.test.tsx --passWithNoTests
  ```

  `--passWithNoTests` keeps this command runnable before Step 13 creates the files. From Step 13 on, all five files exist and must run.

- **U1 regression for files U2 edits** (run in `ui/`):

  ```
  npx vitest run src/settings/settings.test.tsx src/switch src/page
  ```

## Coverage Target

At least **80% line coverage** over `./internal/...` and `./server/...`, with only `server/main.go` excluded (team Testing Posture, NFR8). The floor is never lowered and no exclusion is added. U2 aims to keep every package it changes at 90% or higher.

## Expected Test Volume (Standard strategy)

| Component | Test file(s) | Approximate tests |
|-----------|--------------|-------------------|
| `internal/backlog` OAuth and token types | `oauth_types_test.go` | 5–6 |
| `internal/backlog` projects and call groups | `projects_types_test.go`, `group_test.go` | 5 + table rows |
| `internal/backlog` client: headers, 429, queues, wait logs | `client_test.go` (U2 cases), `limiter_test.go` | 8 + a 3-row 429 table |
| `internal/backlog` OAuth and Projects calls | `oauth_test.go`, `projects_test.go` | 7 + 5 |
| `internal/connection` OAuth config and state token | `oauth_config_test.go`, `state_token_test.go` | 6 + 6 |
| `internal/connection` project keys and change reasons | `project_key_test.go`, `change_test.go`, `record_compat_test.go` | 5 + 7-row table + 2 |
| `internal/connection` store (U2 cases) | `store_test.go` | 8 |
| `internal/connection` reader and token refresh | `credentials_test.go` | 7, including the 5-goroutine single refresh |
| `internal/connection` OAuth start and callback | `oauth_test.go` | 8 |
| `internal/connection` test, disconnect, replace, space change, restore | `lifecycle_test.go` | 8 |
| `internal/connection` projects and events | `projects_test.go`, `events_test.go` | 6 + 5 |
| `internal/plugin` actions and logs (U2 cases) | `actions_test.go` | 8 |
| `internal/plugin` webhook and config | `webhook_test.go`, `config_test.go` | 7 + 2 |
| `internal/plugin` manifest (U2 cases) | `manifest_test.go` | 3 |
| `ui/src/settings` reducer | `state.test.ts` | 8 |
| `ui/src/settings` OAuth | `oauth.test.tsx` | 6 |
| `ui/src/settings` connected panel and M12 | `connected-panel.test.tsx` | 8 |
| `ui/src/settings` project picker | `project-picker.test.tsx` | 6 |
| `ui/src/settings` confirm dialog | `confirm-dialog.test.tsx` | 5 |

The integration tests at the key boundaries are: gateway ↔ fake Backlog (client tests); connection ↔ gateway fake ↔ store fakes (`credentials_test.go`, `oauth_test.go`, `lifecycle_test.go`); Kandev adapter ↔ fake Host (`actions_test.go`, `webhook_test.go`).

## Mocking and Stubbing

- Never call real Backlog or a real Kandev in automated tests.
- Use fakes, not mocks with call expectations. Assert "0 requests" and "exactly 1 refresh" by counting requests on the fake server.
- Inject time and waiting (`Now`, `Wait`, store clock) and run timing tests inside a `synctest` bubble. On the UI side, use `vi.useFakeTimers()`.
- The concurrency tests (single refresh, per-group queue, per-workspace write lock) must pass under `-race`.
- Every secret in a test (API key, access token, refresh token, client secret, auth code, state nonce) comes from the `internal/testutil` helpers, which return a fresh random value per run. Each leak test asserts that no 8-character window of the random part appears in logs, errors, responses, events or the webhook `Location`.

## Test Data

- Keep each test's data local. Use `t.TempDir()` and `t.Setenv()` where a test touches files or the environment.
- JSON fixtures contain only fake users (`"name": "Test User"`), fake hosts (`example-space.backlog.com`) and fake projects (`PROJ`, `DEMO`). A token or key that a test asserts on is generated at run time, never committed, so `make check-secrets` stays clean.
- Config maps for OAuth are built in each test with `https://kandev.example.test` as `public_base_url` and helper-generated client credentials.
