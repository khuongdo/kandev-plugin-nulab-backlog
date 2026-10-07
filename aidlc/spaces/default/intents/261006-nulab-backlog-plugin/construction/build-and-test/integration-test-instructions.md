# Integration Test Instructions

Test strategy: **Standard** (unit tests plus integration tests at key boundaries). All integration tests are Go or Vitest tests that run inside `make test`; none need a network, a real Backlog space or a real Kandev, except the packaged-host contract test.

## Framework and setup

- Go `testing` + `testify/require`, `-race`, `testing/synctest` for virtual time, `net/http/httptest` TLS fakes shaped like Backlog API v2 payloads (`internal/*/testdata/`), an in-memory `pluginsdk.Host` fake shaped like Kandev v0.96.0 (task metadata nested under `"plugin:nulab-backlog"`, provider repository names in `ProviderName`).
- Vitest + jsdom + axe-core with the shared fake `host` in `ui/src/testing/harness.ts`.
- Secrets in tests come from `internal/testutil` (fresh random per run).

## Key boundaries and their tests

| Boundary | Units | Tests | Command |
|----------|-------|-------|---------|
| Plugin ↔ Backlog HTTP (client, rate limiting, queues, errors, body limit, https-only) | U1, U2, U3, U4 | `internal/backlog/*_test.go` (`client_test`, `limiter_test`, `projects_test`, `git_client_test`, `issues_client_test`) | `go test -race ./internal/backlog/...` |
| Connection service ↔ Backlog fake ↔ secret/state store fakes | U1, U2, U4 | `internal/connection/*_test.go` | `go test -race ./internal/connection/...` |
| Git integration ↔ connection fake ↔ host port fake ↔ store | U4 | `internal/git/watcher_test.go`, `events_test.go`, `create_test.go` | `go test -race ./internal/git/...` |
| Issue integration ↔ connection fake ↔ host port fake ↔ store (incl. 300 ms latency budget) | U3 | `internal/issues/sync_test.go`, `events_test.go`, `create_test.go`, `list_test.go` | `go test -race ./internal/issues/...` |
| Kandev adapter ↔ fake Host (actions, RPCs, events, real `connection.Service` switch) | U1–U4 | `internal/plugin/actions_*_test.go`, `references_test.go`, `events_test.go`, `runtime_u3_test.go`, `runtime_u4_test.go` | `go test -race ./internal/plugin/...` |
| UI ↔ plugin actions (fake host) | U1–U4 | `ui/src/**/*.test.ts(x)` | `cd ui && npx vitest run` |
| Packaged plugin ↔ real Kandev v0.96.0 (install, manifest validation, start, actions) | U5 (covers all) | `internal/ci/contract.go` driver | `make contract-test KANDEV_MIN_DIR=../kandev-min` |

## Cross-unit interaction checks (run once, all units together)

```
go test -race -count=1 ./internal/... ./server/...
cd ui && npx vitest run
make package verify-package
make contract-test KANDEV_MIN_DIR=../kandev-min
```

## Coverage expectations

80% line coverage floor over `./internal/...` and `./server/...` (only `server/main.go` excluded), enforced by `make coverage`; `internal/git` and `internal/issues` aim for ≥ 85%.

## Test data and environment

- Fixtures contain fake data only (`example-space.backlog.com`, `PROJ`, `DEMO`, `Test User`, `Lan`). Larger lists are generated in tests.
- Each test owns its data (`t.TempDir()`, `t.Setenv()`); no shared state.
- The contract test uses ports 38529/39529 and a throwaway data directory; it never uses the developer's real Kandev (38429/39429).
