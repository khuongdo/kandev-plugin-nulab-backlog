# Unit Test Instructions — Opt-in by default

## Framework and Setup

- Go: standard `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run`, always `-race`. Prerequisites: Go 1.26.x and `../kandev` linked to the pinned v0.96.0 SDK checkout (see `.kandev-sdk-ref`).
- UI: Vitest (config `ui/vitest.config.ts`), run from `ui/` after `npm ci`.
- No new test framework or configuration is needed; the existing runners are verified in plan Step 1 before the first Red step.

## Commands (scoped to this change)

Go — the packages this change touches:

```bash
go test -race ./internal/connection/... ./internal/plugin/... ./internal/issues/... ./internal/git/... ./internal/ci/...
```

Single regression (repository layer):

```bash
go test -race ./internal/connection/ -run 'TestSwitchWithNoRecord'
```

UI — the files this change touches:

```bash
cd ui && npx vitest run src/switch/switch.test.tsx src/index.test.ts
```

Contract test (packaged host, minimum Kandev version; run 10 times locally when the environment allows):

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

## Expected Tests (Minimal strategy + bugfix regression floor)

- Repository: no switch record → `LoadSwitch` returns `false` (inverted `TestSwitchWithNoRecordIsEnabled`); undecodable record still an error; `connection.get` view reports `enabled: false`.
- Business/API: fresh workspace → guarded action refused with `integration_disabled`, no Backlog call; workers and Git credential RPC idle/refused; admin turns on → connect succeeds; non-admin cannot turn on; explicit ON record stays ON.
- UI: view without `enabled` → switch and page show the Off state.
- Contract: fresh install reports `enabled: false` and refuses connect with 409; after `set_enabled` the connect validation (400 `spaceUrl`) still runs.

## Coverage

- Go line coverage stays at or above 80% (`make coverage`, measured over `./internal/...` and `./server/...`). Do not lower the floor or add exclusions. Delete `coverage.out` after a local run (or write it under `build/`).

## Mocking and Test Data

- Use the existing fakes: `newFakeSecrets`/`newFakeState`/`newTestStore`, `newHarness` (connection), `newRig`/`newFakeHost` (plugin), the issues/git rigs, and the `httptest` fake Backlog server. No real Backlog calls, no real credentials.
- Tests that exercise ON behaviour turn the switch on explicitly through one shared helper; off-state tests keep their explicit setup.
