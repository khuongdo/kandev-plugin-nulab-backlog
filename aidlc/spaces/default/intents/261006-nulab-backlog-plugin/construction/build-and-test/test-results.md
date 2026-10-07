
# Test Results — Build and Test

Run date: 2026-10-07. Code under test: `main` at `4b0a090` (same tree as `be4d545`). Toolchain: Go 1.26.8 (`GOTOOLCHAIN=local`), Node v25, Kandev SDK v0.96.0.

## Build status: SUCCESS

`make clean check-format vet lint test coverage check-secrets build package verify-package` exited 0 in 35 s:

- gofmt, go vet, golangci-lint (gosec) 0 issues, tsc, ESLint, Prettier, actionlint: clean.
- `ci secrets: OK`.
- `build:` linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64; UI bundle 79.9 kB.
- `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`.
- `go mod tidy`: no change to `go.mod`/`go.sum`.

Packaged-host contract test: `make contract-test KANDEV_MIN_DIR=../kandev-min` exited 0 — `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Unit and integration tests

Fresh run (`go test -race -count=1 -v ./internal/... ./server/...`): **529 top-level tests passed, 509 subtests passed, 0 failed, 0 skipped.**

UI (`npx vitest run`): **28 files, 224 tests passed, 0 failed.**

Per-unit commands from every `construction/*/code-generation/unit-test-instructions.md`, deduplicated (coverage-profile variants are covered by `make coverage`; `--passWithNoTests` variants are covered by the full Vitest run), each run once:

| # | Command | Units | Result |
|---|---------|-------|--------|
| 1 | `go test -race -count=1 ./internal/redact/... ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./server/...` | U1 | PASS |
| 2 | `go test -race -count=1 ./internal/connection/... -run 'TestParseSpaceAddress\|TestValidateAPIKey'` | U1 | PASS |
| 3 | `go test -race -count=1 ./internal/pkgverify/...` | U1, U5 | PASS |
| 4 | `go test -race -count=1 ./internal/ci/...` | U5 | PASS |
| 5 | `go run ./cmd/ci workflows -dir .github/workflows` | U5 | PASS |
| 6 | `go run ./cmd/ci secrets -root .` | U5 | PASS |
| 7 | `go test -race -count=1 ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./internal/testutil/... -run '^Test(OAuth\|Token\|RateLimit\|Queue\|Projects\|Recheck\|Disconnect\|Replace\|SpaceChange\|Restore\|ConnectionChanged\|ConnectionReader\|Webhook\|APIKeyHeader\|ActionFailureLog\|U2)'` | U2, U4 | PASS |
| 8 | `go test -race -count=1 ./internal/backlog/... ./internal/connection/... ./internal/git/... ./internal/plugin/... -run '^TestU4_'` | U4 | PASS |
| 9 | `go test -race -count=1 ./internal/backlog/... ./internal/issues/... ./internal/plugin/... -run '^TestU3_'` | U3 | PASS |
| 10 | `go test -race -count=1 ./internal/backlog/... ./internal/git/... ./internal/plugin/... -run '^Test(U4_\|U2_\|Projects\|Queue\|RateLimit)'` | U3 | PASS |
| 11 | `cd ui && npx vitest run src/settings src/brand src/switch src/page src/index.test.ts` | U1, U2, U3, U4 | PASS |
| 12 | `cd ui && npx vitest run src/git src/issues` | U3, U4 | PASS |
| 13 | `make package contract-test KANDEV_MIN_DIR=../kandev-min` | U5 | PASS |

Failures: none.

## Coverage

`make coverage`: **92.9%** (floor 80%, excluded: `server/main.go` only).

| Package | Coverage |
|---------|----------|
| internal/backlog | 96.0% |
| internal/ci | 91.0% |
| internal/connection | 94.5% |
| internal/git | 90.8% |
| internal/issues | 91.3% |
| internal/pkgverify | 93.2% |
| internal/plugin | 93.6% |
| internal/redact | 97.4% |
| internal/testutil | 87.0% |
| server (wiring only, excluded) | 0.0% |

## Target Verification Matrix (final)

See `build-and-test-summary.md` → `## Target Verification Matrix` for the full matrix. Run 1 totals (46 targets): **35 Met, 2 Not Met, 9 Unverified**; see Run 2 below for the current totals.

Not Met:
- T-SEC-02: NFR3.2 asks for a 4-character substring scan; tests use 8-character windows.
- T-RATE-01: NFR2.1 asks Connect to return `rate_limited` at once (< 1 s, no retry) on 429; since U2, Connect retries waits ≤ 3 s up to 3 times.

Unverified (need a real Backlog space, a deployed Kandev, GitHub, or a missing test): T-PERF-02, T-PERF-03, T-SEC-06 (live 403), T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01.

## Result

**FAILED by the stage's failure predicate**: every build and test command passed, but 2 targets are Not Met and 9 are Unverified. Failure handling: gated mode → halt-and-ask (no loop-back has been taken yet).

## Run 2 — after Loop-back 1 (2026-10-07)

- `make clean check-format vet lint test coverage check-secrets build package verify-package`: exit 0; golangci-lint 0 issues; Vitest 28 files / 229 tests pass; coverage **92.9%**; `ci secrets: OK`; `verifypkg: OK`; `go mod tidy` no diff.
- `go test -race -count=1 -v ./internal/... ./server/...`: **546 top-level + 534 subtests pass, 0 fail**. New tests pass: `TestNoRetryReturns429AtOnce` (4 cases), `TestAssertNoLeak_Catches4CharWindow`, `TestConnectionGetP95Under500ms`, `TestSetEnabledP95Under500ms`, `TestU2_EmptyVerifierNeverMatches` (4 cases).
- Loop-back 1 targets: T-RATE-01 **Met**, T-SEC-02 **Met**, T-PERF-03 **Met** (virtual-time p95: get 400 ms, set_enabled 200 ms at 100 ms per store operation).
- **Packaged-host contract test is flaky: FAILED 2 of 7 runs** (`make contract-test` once, then 5 manual runs with the same binary and package: OK, OK, OK, OK, FAIL; plus 1 manual OK). Failure: `ci contract: contract: connection.get answered 500` a few milliseconds after install.
  - Root cause: `internal/plugin/runtime.go` resolves the Kandev Host on every store call and returns `errNoHost` (→ 500 `internal`) when it is not injected yet. The SDK injects the Host from a background goroutine that retries the broker dial (`pluginsdk/serve.go:19–21,119–126`, up to 30 s), so an action that arrives right after the plugin starts can race the injection. This is the U1 review Minor R-04 ("workers start before Host injection") showing up on the action path, not only in background workers.
  - T-COMPAT-01 is therefore **Not Met** (the package does not reliably run on v0.96.0).
  - Side observation: the plugin's own stderr log lines do not appear in Kandev's `backend-logs.log` (T-OBS-03 stays Unverified; may need Kandev's plugin log location).

Totals after Run 2 (46 targets): **37 Met, 1 Not Met (T-COMPAT-01), 8 Unverified** (T-PERF-02, T-SEC-06, T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01). Result: **FAILED** by the failure predicate.

## Run 3 — after Loop-back 2 (2026-10-07)

- `make clean check-format vet lint test coverage check-secrets build package verify-package`: exit 0; golangci-lint 0 issues; Vitest 229 tests pass; coverage **92.9%**; `ci secrets: OK`; `verifypkg: OK`; `go mod tidy` no diff.
- `go test -race -count=1 -v ./internal/... ./server/...`: **553 top-level + 534 subtests pass, 0 fail** (includes the new `TestHost_*` Host-readiness tests).
- `make contract-test KANDEV_MIN_DIR=../kandev-min` run **10 times in a row: 10/10 OK** (`ci contract: OK nulab-backlog on Kandev v0.96.0`). T-COMPAT-01 is **Met**.

Totals after Run 3 (46 targets): **38 Met, 0 Not Met, 8 Unverified** (T-PERF-02, T-SEC-06, T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01). Every build and test command passed. Result: still **FAILED** by the failure predicate only because the 8 Unverified targets need a real Backlog space, a deployed Kandev, GitHub or a manual check; no code fix exists for them.

**Accepted failure (2026-10-07):** at the halt-and-ask the human chose "Accept failure". The 8 Unverified targets (T-PERF-02, T-SEC-06, T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01) and the 9 Deferred ACs in `cross-unit-traceability.md` stay open and are carried to the stages that can verify them: deployment-execution (manual real-space checks, live 403, release, marketplace), performance-validation (real-space latency) and observability-setup (Kandev plugin log collection).

## Loop-Back Log

### Loop-back 1 — 2026-10-07T00:00:00Z

- **Diagnosis:** T-RATE-01 Not Met (Connect retries short 429 waits up to 3 times; NFR2.1 requires an immediate `rate_limited`). T-SEC-02 Not Met (leak tests scan 8-character windows; NFR3.2 requires 4). T-PERF-03 Unverified (no 100-call timing test for `connection.get`/`set_enabled`). Open Major review findings U2 R-01 (OAuth state not bound to the starting browser) and U2 R-02 (project picker not reloaded after an API-key replace to another space).
- **Root-cause stage:** code-generation (units walking-skeleton and connection; leak-window helper in `internal/testutil` is shared by all units).
- **Planned fix:** (1) Connect uses a no-retry rate-limit policy and returns `rate_limited` with `retryAfterSeconds` at once; (2) `testutil.AssertNoLeak` scans 4-character windows; (3) add a virtual-time test timing 100 `connection.get` and 100 `connection.set_enabled` calls against the p95 ≤ 500 ms budget; (4) bind the OAuth flow to the starting browser with a verifier cookie set by the settings UI and checked by the callback in constant time; (5) remount/reload the project picker when the connected space or connection epoch changes.
- **Estimated impact:** effort a few hours of agent work plus re-approval of 5 plans and re-review of 5 units; financial cost none beyond model usage; risk low–medium (OAuth binding touches UI and webhook; 4-character windows may need care to avoid false positives).
- **Approved by:** human chose "Retry with fix" at the halt-and-ask.

### Loop-back 2 — 2026-10-07T00:20:00Z

- **Diagnosis:** T-COMPAT-01 Not Met. The packaged-host contract test fails 2 of 7 runs with `connection.get answered 500`: actions that arrive before the SDK injects the Kandev Host (background broker dial, `pluginsdk/serve.go`) hit `errNoHost` in `internal/plugin/runtime.go` and map to `internal`.
- **Root-cause stage:** code-generation (unit walking-skeleton; `internal/plugin/runtime.go`, `internal/plugin/host_port.go`).
- **Planned fix:** the Runtime signals Host readiness from `SetHost`; `hostStores.get` and the host port wait for the Host (bounded, honouring the action's context deadline, e.g. up to 5 s) instead of failing immediately; still fail closed with `errNoHost` when the wait expires. Add a test where an action is handled before `SetHost` and succeeds once the Host arrives, and one where the wait times out. Run the contract test 10 times.
- **Estimated impact:** effort about 1 hour of agent work plus re-approval of 5 plans; financial cost none beyond model usage; risk low.
- **Approved by:** human chose "Retry with fix" at the halt-and-ask.
