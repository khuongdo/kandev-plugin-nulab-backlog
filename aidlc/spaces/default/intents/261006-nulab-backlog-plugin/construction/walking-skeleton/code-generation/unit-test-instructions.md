# Unit Test Instructions — walking-skeleton (U1)

## Frameworks and Setup

- **Go**: the standard `testing` package with `github.com/stretchr/testify/require`. Tests are table-driven with `t.Run`, and every test name describes a behaviour. Tests always run with `-race`, which needs CGO (the default on Linux and macOS with a C toolchain).
- **Fake Backlog**: `net/http/httptest` servers. JSON samples live in `internal/backlog/testdata/` (`myself_ok.json`, `myself_missing_id.json`, `error_401.json`, `error_429.json` and others).
- **Kandev stores**: in-memory fakes of `connection.SecretStore` and `connection.StateStore`, plus a fake `pluginsdk.Host` in `internal/plugin`. These fakes can inject failures (a write error, a slow call, a failing rollback).
- **Clock and deadlines**: a clock and the per-step limits are injected into `connection.Service` and `backlog.Client`. Tests set short limits instead of sleeping, and never use real `time.Sleep`.
- **UI**: Vitest with `jsdom`, configured in `ui/vitest.config.ts`. A fake `host` object provides `host.jsx`, `host.React`, `host.ui` stubs and `host.api.invokeAction`. Accessibility is checked with `axe-core`.

## Commands for This Unit

Run these from the repository root. `../kandev` must exist at the commit in `.kandev-sdk-ref`.

- Go, all U1 packages:

  ```
  go test -race ./internal/redact/... ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./server/...
  ```

- Go, one layer while working (example):

  ```
  go test -race ./internal/connection/... -run 'TestParseSpaceAddress|TestValidateAPIKey'
  ```

- UI, the U1 screens and registrations only (run in `ui/`):

  ```
  npx vitest run src/settings src/brand src/switch src/page src/index.test.ts
  ```

- Go, the in-repo package verifier (used by `verify-package`):

  ```
  go test -race ./internal/pkgverify/...
  ```

- Coverage for the U1 Go packages:

  ```
  go test -race -coverprofile=coverage.out ./internal/redact/... ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./server/...
  ```

  Then run `go tool cover -func=coverage.out`. `make coverage` runs this same command, excludes `server/main.go` and checks the floor.

Before Step 2 creates the first tests, these commands report "no test files". Step 2 adds smoke tests so each command runs before the first Red step.

## Coverage Target

At least 80% line coverage over `./internal/...` and `./server/...`, excluding only `server/main.go` (team-practices, NFR8.1). The floor is never lowered, and no exclusion is added to make it pass.

## Expected Test Volume (Standard strategy)

| Component | Test file(s) | Approximate tests |
|-----------|--------------|-------------------|
| `internal/redact` | `redact_test.go` | 5–6 |
| `internal/backlog` | `types_test.go`, `client_test.go` | 8 + table rows for status mapping |
| `internal/connection` | `address_test.go`, `apikey_test.go`, `store_test.go`, `service_test.go` | 8 per file, plus table rows for address cases |
| `internal/plugin` | `actions_test.go`, `manifest_test.go` | 8 + 5 |
| `internal/pkgverify` | `pkgverify_test.go` | 5–6, including the Nulab/Backlog asset-URL check |
| `ui/src/settings` | `settings.test.tsx` | 8 + one per M1 state, plus the Off state and spacing |
| `ui/src` registrations | `index.test.ts` | 5–6 (card icon and switch, nav item, route, workspace sync) |
| `ui/src/brand` | `backlog-logo.test.tsx` | 3–4 |
| `ui/src/switch` | `switch.test.tsx` | 5–6 |
| `ui/src/page` | `backlog-page.test.tsx` | 8, one per BacklogPage state |

Revision 2 also adds switch cases to `store_test.go`, `service_test.go`, `actions_test.go` and `manifest_test.go` (plan Steps 15–19).

The store and service tests with fakes are the integration tests for the key boundaries (gateway ↔ connection ↔ Kandev stores).

## Mocking and Stubbing

- Never call real Backlog in automated tests.
- Use fakes, not mocks with call expectations, except to assert "0 requests" (count requests on the fake server).
- Secrets in tests: build every test key with the shared helper. It returns `test-api-key-` plus 32 random hex characters per run. Never commit a fixed realistic key (NFR4.1).

## Test Data

- Keep each test's data local. Use `t.TempDir()` and `t.Setenv()` where files or the environment are touched.
- JSON fixtures in `testdata/` contain only fake users (for example `"name": "Test User"`) and fake hosts (`example-space.backlog.com`).
