## Developer Code Scan Results

Intent: `261007-opt-in-default` (bugfix, Minimal, brownfield). Commit scanned: `2b4325f` (v0.1.1).
Bug report (verbatim): "hiện tại đang opt-out by default, chuyển sang opt-in sau khi cài đặt plugin".
This is a FOCUSED scan of everything that decides whether the plugin is active after installation.

### Scan Coverage
- **Analyzed deeply**:
  - manifest.yaml
  - server/main.go
  - internal/plugin/runtime.go (Start/Close, `guarded`, the single `RequireEnabled` guard in `HandleAction`)
  - internal/plugin/credential.go, internal/plugin/events.go, internal/plugin/references.go
  - internal/connection/store.go (`switchRecord`, `LoadSwitch`, `SaveSwitch`, `Load`)
  - internal/connection/service.go (`SetEnabled`, `RequireEnabled`, `ErrIntegrationDisabled`, outcome mapping)
  - internal/connection/lifecycle.go (`view.Enabled = true` after guarded actions), internal/connection/oauth.go (switch checks in start/callback)
  - internal/issues/sync.go, internal/issues/watcher.go, internal/issues/service.go (`Suggest`, `Authorize` switch checks)
  - internal/git/watcher.go (`cycleWatch`, `createOne` switch checks)
  - ui/src/index.ts (registrations, `publishSwitches`)
  - ui/src/switch/enabled-events.ts, ui/src/switch/integration-switch.tsx, ui/src/switch/switch.test.tsx
  - ui/src/settings/SettingsScreen.tsx (off rendering, `subscribeEnabled`), ui/src/settings/state.ts (`isOff`, `enabledChanged`, `integration_disabled`)
  - ui/src/page/BacklogPage.tsx (`pageState` off state), ui/src/messages/en.ts (switch/off messages)
  - internal/connection/store_test.go (switch tests)
- **Skimmed only**:
  - internal/ci/ (contract.go `actions` step and contract_test.go fixtures only)
  - internal/plugin/*_test.go, internal/connection/*_test.go, internal/issues/*_test.go, internal/git/*_test.go (test names and switch-related tests only)
  - ui/src/index.test.ts, ui/src/testing/harness.ts (fixture `enabled` values), ui/src/issues/, ui/src/git/ (only `integration_disabled` handling)
  - README.md ("Turn Backlog on or off" section)
  - Kandev v0.96.0 host web registry (../kandev/apps/web/lib/plugins/registry.ts, app-sidebar integration badge) and backend manifest schema
  - internal/backlog/, internal/redact/, internal/pkgverify/, cmd/, .github/ (not relevant to the default)
  - Prior functional design of the switch: aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/functional-design/ (BR7.1 origin)

### What is currently opt-out by default

The plugin has exactly ONE on/off control: the per-workspace **IntegrationSwitch** (state key `integration`, scope `workspace`, record `{schemaVersion, enabled, changedAt}`). Everything else is already gated by it. The opt-out behaviour comes from a single backend default plus two UI fallbacks:

| # | Location | Current behaviour | Why it is opt-out |
|---|----------|-------------------|-------------------|
| D1 | internal/connection/store.go:212-220 `LoadSwitch` | `if !found { return true, nil }` — comment "No record means on (BR7.1)" | A fresh install has no switch record, so every workspace is ON. This is the root cause. |
| D2 | ui/src/index.ts:32 `publishSwitches` | `publishEnabled(host, ws, view.enabled !== false)` | A view without `enabled` is published to Kandev as on. Only a fallback: the backend always sends a boolean. |
| D3 | ui/src/switch/integration-switch.tsx:45 | `const value = view.enabled !== false` | Same fallback for the card switch's initial value. |
| D4 | README.md:38 | "Backlog is on by default in every workspace. A Kandev admin can turn it off ..." | User-facing documentation of the opt-out default. |
| D5 | internal/ci/contract.go:205-206 (packaged-host contract test) | Requires `connection.get` on a fresh host to answer `not_connected` **and `enabled: true`**, then calls `connection.connect_api_key` expecting `400 validation spaceUrl` | Encodes the default in CI. With opt-in, `connect_api_key` is a guarded action and would answer `409 integration_disabled`, so both assertions break. |
| D6 | Design record BR7.1 (261006 functional-design entities.md:63 `default: true`, functional-spec.md:60) | "an absent record means enabled" | Origin of D1; superseded by this intent. |

Not opt-out (already safe, verified):
- `ui/src/settings/state.ts:162 isOff` and `ui/src/page/BacklogPage.tsx:27` use `view.enabled === false` (strict); they follow whatever the backend returns.
- Kandev host (v0.96.0) `registry.isIntegrationEnabled` is `=== true`; an unpublished plugin is treated as off (no "enabled" badge). No manifest field exists to request a default-on/default-off state; installing activates the plugin process, which the plugin cannot change.
- Background workers: `NewRuntime` always calls `Start()` (PR watcher, issue syncer, issue watcher). They only iterate workspaces present in the plugin's own indexes (`store.Index`, `WatchIndex`) — i.e. workspaces that already connected or saved watches — and every per-workspace path calls `RequireEnabled` first and skips on `ErrIntegrationDisabled` (issues/sync.go:235, issues/watcher.go:169/224, git/watcher.go:203/303). With the switch off they make no Backlog call. No watch, saved query or link is created automatically; issue poll interval default (5 min, issues/types.go:42) only applies to linked issues.
- Every action except `connection.get` and `connection.set_enabled` is guarded (runtime.go:47-50, 241-246); Git credential RPCs (credential.go:27, 42), `#` suggestions/authorization (issues/service.go:776, 839) and the OAuth start/callback (connection/oauth.go:169, 250, 267) check the switch. `OnEvent(task.deleted)` is intentionally unguarded (cleanup only).
- UI registrations (Integrations card + switch, nav item, `/backlog` route, repository/review providers, task action, badge, task menu, panel) are unconditional by design (BR5.4/BR7.6/BR7.8) so the switch is always reachable. When off, Settings shows `backlog-off` ("... Turn it on with the switch above to connect."), the page shows `pageOff`, and other surfaces map `409 integration_disabled` (settings/state.ts:120/213, issues/issues-state.ts:59/88, git/git-state.ts:65). The off copy already reads correctly for an opt-in flow.

Conclusion: flipping D1 (absent record → off) makes the backend opt-in everywhere at once; D2/D3 should be aligned to `=== true` for consistency; D4/D5 must be updated.

### Packages Found
- internal/connection — library — Go — connection record/secret, IntegrationSwitch store and `RequireEnabled` guard
- internal/plugin — service adapter — Go — the only `pluginsdk` importer; action routing, guard, worker lifecycle
- internal/issues — library — Go — issue list/links, sync worker, issue watches
- internal/git — library — Go — repository provider, PR links, PR watcher
- server — binary entrypoint — Go — `pluginsdk.Serve(plugin.NewRuntime())`
- ui — browser bundle — TypeScript/React (host-provided React) — settings card, switch, `/backlog` page, task surfaces

### Build System
- **Type**: Go modules + Make; npm (ui/)
- **Config Files**: go.mod (replace `github.com/kandev/kandev => ../kandev/apps/backend`), .kandev-sdk-ref (`f099a46...` = v0.96.0), Makefile, ui/package.json, ui/tsconfig.json, ui/vitest.config.ts
- **Build Dependencies**: plugin → connection, issues, git, backlog, redact; issues/git → connection (switch + snapshot via consumer-side interfaces)

### APIs Discovered
- Plugin actions — manifest.yaml / internal/plugin/runtime.go handlers + git_actions.go + issue_actions.go — 48 actions; 46 guarded by the switch, 2 unguarded (`connection.get`, `connection.set_enabled` [admin])
- gRPC plugin hooks — internal/plugin — `ResolveGitCredential`, `GetGitCredentialBinding`, `SearchEntityReferences`, `AuthorizeEntityReference`, `OnEvent`
- Webhook — `oauth-callback` (public GET), checks the switch

### Frameworks & Libraries
- Kandev plugin SDK — v0.96.0 (pinned checkout) — host API, Serve, UI host API (`setIntegrationEnabled`, `ui.IntegrationEnabledControl`)
- testify/require — tests
- Vitest + jsdom + axe — UI tests

### Test Coverage
- **Test Directories**: internal/*/ (`*_test.go`, testdata/), ui/src/**/*.test.ts(x)
- **Test Frameworks**: Go `testing` + testify; Vitest
- **Coverage Config**: present (Makefile `coverage`, 80% floor over ./internal/... and ./server/...)
- **Baseline (2026-10-07, commit 2b4325f)**:
  - `go test -race -count=1 ./...` with Go 1.26.8 and ../kandev → v0.96.0: all 9 packages with tests **ok** (1159 passing tests incl. subtests). No coverage.out written.
  - `npx vitest run` in ui/: **286 passed (286)**, 29 files.
- **Tests that encode the current default**:
  - internal/connection/store_test.go:297 `TestSwitchWithNoRecordIsEnabled` — asserts absent record = on (must be inverted).
  - internal/ci/contract.go:205 (exercised by the packaged-host contract job; contract_test.go fixtures at :54 use `"enabled":true`).
  - Implicit dependence: most connection and plugin tests connect or call guarded actions without writing a switch record. A throwaway experiment (temp copy outside the repo, D1 flipped to `false`, then deleted) gave: internal/connection **45** failing tests and one hang (`TestSecondConnectInTheSameWorkspaceIsAConflict` blocks until the package timeout), internal/plugin **67** failing; internal/ci, issues, git, backlog unaffected (issues/git use their own fake `RequireEnabled`).
  - UI tests pass explicit `enabled` values (harness.ts fixtures `enabled: true`); no UI test covers a view without `enabled`.
- **Existing tests for the off state (reusable for opt-in)**: plugin/actions_test.go:443 `TestWhileOffOnlyGetAndSetEnabledRun`, :467 `TestGuardFailsClosed`, :486 `TestGuardedActionsAreEveryActionExceptGetAndSetEnabled`; actions_u2/u3/u4 `...GuardRefusesAllWhileOff`; credential_test.go:91; runtime_u3_test.go:57; runtime_u4_test.go:199; git/watcher_test.go:192/216/231; issues/sync_test.go:186; connection/service_test.go:489-584; ui/src/switch/switch.test.tsx; ui/src/page/backlog-page.test.tsx (off view).

### Code Quality Indicators
- **Linting**: golangci-lint + gosec (.golangci.yml), gofmt, go vet; ESLint + Prettier + `tsc --noEmit` strict (ui/)
- **CI/CD**: .github/workflows/ci.yml (incl. packaged-host contract job), .github/workflows/release.yml
- **Documentation**: README.md documents the switch (and the opt-out default at line 38); doc comments cite BR/NFR IDs

### Technical Debt Signals
- Switch default lives in one place (good), but tests rely on it implicitly via the "no record" path rather than an explicit test helper, so flipping it needs a shared "turn on" step in the connection and plugin test harnesses.
- `TestSecondConnectInTheSameWorkspaceIsAConflict` hangs instead of failing when the first Connect is refused early — a test-robustness gap surfaced by the experiment.
- UI uses two different readings of a missing `enabled` (`!== false` in index.ts/integration-switch.tsx vs `=== false` in state.ts/BacklogPage.tsx).

## Handoff Summary
- **Intent-relevant finding**: The plugin is opt-out solely because `Store.LoadSwitch` treats a missing IntegrationSwitch record as on (internal/connection/store.go:219-220, design rule BR7.1). Every guarded action, RPC, webhook and background worker already routes through `RequireEnabled`, so changing that one default to off makes the whole plugin opt-in after installation; an admin then turns it on with the existing card switch (`connection.set_enabled`, admin-only). UI fallbacks at ui/src/index.ts:32 and ui/src/switch/integration-switch.tsx:45 (`!== false`) should be aligned; README.md:38 and the contract check internal/ci/contract.go:205-206 (expects `enabled: true` and a 400 from `connect_api_key` on a fresh host) must change.
- **Risks / follow-up**:
  - **Upgrade migration (needs a product decision)**: v0.1.0/v0.1.1 users who connected but never touched the switch have no switch record. A plain flip turns Backlog OFF for them on upgrade (connection kept, watches/sync paused, Git credentials refused) until an admin turns it on. Options: (a) accept and document in upgrade notes; (b) treat a workspace that already has a connection record as on (grandfather); (c) write the switch record on first successful Connect is not enough by itself because Connect is guarded.
  - Connect is guarded, so in opt-in the admin must turn the switch on before connecting; non-admin members can only see the Off state. The existing off copy already says so.
  - The packaged-host contract test (CI and `make contract-test`) must be changed together with D1: either expect `enabled:false` and assert `connect_api_key` → `409 integration_disabled`, or turn the switch on via `connection.set_enabled` before the validation step.
  - About 112 Go tests (connection + plugin) depend on the implicit default; plan a harness-level "switch on" helper rather than per-test edits, and fix the hang noted above.
  - Environment notes: a symlink `../kandev` → /home/k_do_webfrontier/repo/kandev (v0.96.0, f099a46) was created next to the repo (outside it) for the Go baseline; `ui/node_modules` was installed (gitignored). No source file was modified and no coverage.out was left.
  - The shared CodeKB (built for v0.1.1) still describes the switch as default-on; update it when this fix lands.
