# Code Summary — Opt-in by default after installation

Intent `261007-opt-in-default` (bugfix, Minimal test strategy, brownfield, zero-Unit). Methodology: TDD per the Testing Contract (`sha256:85ae4f86…`).

## Files Modified

Production code:
- `internal/connection/store.go`: `LoadSwitch` returns `false, nil` when there is no switch record (root cause D1). The doc comment says opt-in and that it supersedes BR7.1 of intent 261006. A record that cannot be decoded is still an error (NFR3.9).
- `internal/ci/contract.go`: the packaged-host contract driver now expects a fresh install to report `enabled: false` and to refuse `connection.connect_api_key` with `409 integration_disabled`. It then turns the switch on (`connection.set_enabled`) and still checks the real connect path (`400 validation spaceUrl`). `connectValidation` became `connectBait(wantStatus, wantCode, wantField)`, and a new `enable` helper was added. The bait key is still redacted from every message (D5).
- `ui/src/index.ts`, `ui/src/switch/integration-switch.tsx`: the fallbacks are now `view.enabled === true` (D2, D3).
- `ui/src/page/BacklogPage.tsx` (`pageState`) and `ui/src/settings/state.ts` (`isOff`): a view without `enabled` is now treated as off, so the page and Settings match the switch (FR3 acceptance; see Deviations). The `ConnectionView.enabled` doc comment was updated.
- `README.md`: added the "Upgrading from v0.1.x: Backlog is now off by default" note. "Turn Backlog on or off" now says off by default, admin turns it on, then connects. The install steps now include turning the switch on (D4).

Tests:
- `internal/connection/fakes_test.go`: added the `switchOn(t, state, ws...)` helper, which turns workspaces on through `SaveSwitch` and resets the write counter. Added `fakeState.nonSwitch()`.
- `internal/connection/service_test.go`: `newHarness` turns `ws`, `ws-a` and `ws-b` on. `TestSecondConnectInTheSameWorkspaceIsAConflict` now fails fast (a `select` on `done`) instead of hanging. `TestConnectCancelledWritesNothing…` uses `nonSwitch()`.
- `internal/connection/oauth_flow_test.go`: the four "nothing stored" checks use `nonSwitch()`. They are just as strict, because only the harness's own switch records are excluded.
- `internal/connection/store_test.go`: `TestSwitchWithNoRecordIsEnabled` is inverted to `TestSwitchWithNoRecordIsOff`. Added `TestSwitchWithNoRecordViewReportsOff`.
- `internal/plugin/actions_test.go`: `newRig` is now `newFreshRig()` plus `switchOn(t, host, "ws-1")`, a plugin-side helper that goes through `connection.Store.SaveSwitch`.
- `internal/plugin/actions_u3_test.go`, `internal/plugin/actions_u4_test.go`: the U3 and U4 rigs turn `ws-1` on.
- `internal/plugin/opt_in_test.go` (new): six regressions. Fresh install is off and refuses connect with no Backlog call. Admin turns it on, then connect works, and an explicit on survives a restart. An upgraded workspace keeps its connection but is off and resumes without reconnecting. The Git credential RPC is refused. Issue sync is idle. The PR watcher is idle.
- `internal/ci/contract_test.go`: the fake Kandev now models the switch (off after install, `set_enabled`, 409 while off). Happy-path order: get, connect (409), set_enabled, connect (400). New failure cases: enabled on a fresh install, connect not refused while off, set_enabled refused, and a key-leak case while off.
- `ui/src/switch/switch.test.tsx`, `ui/src/index.test.ts`, `ui/src/page/backlog-page.test.tsx`, `ui/src/settings/settings.test.tsx`: tests for a view without `enabled` (switch off and published off, page Off alert, Settings Off copy). The local settings fixtures now say `enabled: true` explicitly.

## Key Decisions

- **One root-cause fix.** All guarded actions, RPCs, the OAuth webhook and the workers already go through `RequireEnabled` → `LoadSwitch`, so flipping the default is the whole backend change. Step 7 needed no other production change.
- **Test harnesses switch on explicitly, through `SaveSwitch`.** We did not hide this in the fakes. A fresh rig (`newFreshRig`, or a plain `newFakeState`) behaves like a real fresh install. No assertion was removed or loosened.
- **An upgrade from v0.1.x is simulated by deleting the switch record** (`forgetSwitch`) after connecting. This is the FR5 scenario.
- **The contract test keeps the real connect path.** It proves the 409 refusal first, then turns the switch on and checks the existing `400 validation spaceUrl` path.
- **Admin-only `set_enabled` (FR2 (d)) is enforced by the Kandev host** through `access: admin` in `manifest.yaml`. The existing `internal/plugin/manifest_test.go` asserts that. The plugin cannot test the host's enforcement in a unit test, so no new test was added for it.

## Test Coverage Summary

- Go `go test -race ./...`: **1169 passed, 0 failed** (baseline 1159). The net +10 are the new regressions and contract cases.
- Vitest: **291/291 passed** (baseline 286), 29 files.
- `make coverage`: **92.8%** (floor 80%, excluded only `server/main.go`). The profile is under `build/`; there is no `coverage.out` at the root.
- `make check-format`, `make vet`, `make lint` (golangci-lint v2.14.0 + gosec, `tsc --noEmit`, ESLint, actionlint, workflow policy), `make check-secrets`: all pass.
- `make contract-test KANDEV_MIN_DIR=../kandev` (real Kandev v0.96.0, packaged plugin 0.1.1): **10/10 passed**.

## Deviations from Plan

- **Step 10 touched two more UI files** (`ui/src/page/BacklogPage.tsx`, `ui/src/settings/state.ts`). The FR3 acceptance ("the switch and the page show the Off state" for a view without `enabled`) and Step 9's Red test ("page shows the existing Off copy") cannot pass with the two listed fallbacks alone, because `pageState` and `isOff` were strict `=== false`. Both now treat anything but an explicit `true` as off. The Off wording is unchanged (FR4). `isOff(undefined)` stays false, so nothing changes before the view loads.
- **Step 2 touched `internal/connection/oauth_flow_test.go`** (the four "nothing stored" checks now use `nonSwitch()`), and also the settings test fixtures in Step 10. The `internal/issues` and `internal/git` rigs did not need changes: they use their own consumer-side fakes and do not read the real store.
- **Step 6, Red after Green.** Step 4 had already landed, so the Red evidence for Step 6 was recorded by temporarily restoring the old default (as the plan allows). The file was restored right after.
- **Observed once, not caused by this change:** two or three intermittent failures in untouched axe accessibility tests (`issues-page.test.tsx`, `backlog-lists.test.tsx`) in 2 of about 12 full Vitest runs. Every later run, including 6 in a row, was green. These look like an existing timing flake under load.

## Red/Green Log

| Step | Command | Red result (before) | Green result (after) |
|---|---|---|---|
| 2 (helper, old default) | `go test -race -timeout 120s ./internal/connection/ ./internal/plugin/` with the default temporarily flipped | 45 connection failures plus a hang in `TestSecondConnectInTheSameWorkspaceIsAConflict` (2 min timeout), and 67 plugin failures | Old default: both packages ok. Flipped default: only `TestSwitchWithNoRecordIsEnabled` fails (the Step 3 target) |
| 3 | `go test -race ./internal/connection/ -run 'TestSwitchWithNoRecord'` | `TestSwitchWithNoRecordIsOff` FAIL "Should be false"; `TestSwitchWithNoRecordViewReportsOff` FAIL "Should be false" | ok after Step 4 |
| 6 | `go test -race ./internal/plugin/ -run TestOptIn` with the old default temporarily restored | 5/6 FAIL: FreshInstall (enabled true), UpgradedWorkspace (enabled true), GitCredential ("An error is expected but got nil"), IssueSync (issue reads), PRWatcher ("[PROJ/web-app] should have 0 item(s)"). AdminTurnsOn passes by design (it is the after-opt-in path) | 6/6 PASS with the new default |
| 9 | `npx vitest run src/switch/switch.test.tsx src/index.test.ts src/page/backlog-page.test.tsx src/settings/settings.test.tsx` | 5 failed / 63 passed (publishSwitches, switch card, pageState, page alert, Settings Off) | 291/291 after Step 10 and the fixture update |
| 12 | `go test -race ./internal/ci/` | HappyPath, 3 new failure cases, 4 leak cases and `TestRunWorkflowsAndContract` FAIL (driver still wanted `enabled true`) | ok; then `make contract-test` 10/10 |
