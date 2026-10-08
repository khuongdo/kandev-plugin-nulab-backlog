# Requirements — Opt-in by default after installation

## Intent Analysis

Initial description (verbatim) [desc]: "hiện tại đang opt-out by default, chuyển sang opt-in sau khi cài đặt plugin"

- **Type**: bug fix (behaviour change of one default). **Scope**: single rule with a cross-cutting effect. **Complexity**: simple. **Depth**: Minimal.
- **Goal**: installing the plugin must not activate Backlog in any workspace. Backlog becomes active in a workspace only after an admin explicitly turns the integration switch on.
- **Root cause** (code KB, `business-overview.md` § IntegrationSwitch; `code-quality-assessment.md` D1): `Store.LoadSwitch` (`internal/connection/store.go:212-220`) returns `true` when a workspace has no switch record (design rule BR7.1 of intent 261006). Every guarded action, RPC, webhook and background worker already goes through `RequireEnabled`, so this one default decides opt-out vs opt-in.

## Functional Requirements

- **FR1 — Off by default.** A workspace that has no switch record is treated as OFF. [desc]
  - **FR1.1** `LoadSwitch` returns `false` (not an error) when no switch record exists; an undecodable record stays an error (NFR3.9 of intent 261006 unchanged).
  - **FR1.2** While a workspace has no record, every action and RPC guarded by `RequireEnabled` refuses with `integration_disabled`, and the background workers (issue sync/watch, Git watcher) do no work for it — exactly the existing OFF behaviour.
  - Acceptance:
    - Given a freshly installed plugin and a workspace with no switch record, When any guarded action (e.g. `connection.connect_api_key`) is called, Then it is refused with `409 integration_disabled` and no Backlog request is made.
    - Given a workspace with no switch record, When the settings view is loaded (`connection.get`), Then it reports `enabled: false`.
- **FR2 — Opt in by admin.** Turning Backlog on stays the existing admin-only `connection.set_enabled` action, which remains unguarded and writes only the switch record (BR7.4 unchanged). [desc]
  - Acceptance:
    - Given a workspace with no switch record, When an admin turns the switch on and then connects with an API key, Then the connection succeeds and guarded features work.
    - Given a non-admin member, When they try to turn the switch on, Then the request is refused as today.
    - Given a workspace whose admin turned the switch on, When the plugin restarts or is upgraded, Then the workspace stays ON (explicit records are never overwritten).
- **FR3 — UI fallbacks match.** Where the UI falls back when `enabled` is missing from the view, it treats the workspace as OFF (`ui/src/index.ts:32`, `ui/src/switch/integration-switch.tsx:45`: `view.enabled !== false` → `view.enabled === true`). [desc]
  - Acceptance: Given a settings view without an `enabled` field, When the switch and the page render, Then they show the Off state.
- **FR4 — Existing Off state unchanged.** The existing Off state, its wording ("Turn it on with the switch above to connect") and the members' Off view are kept as-is; no new prompt is added. [Q2]
  - Acceptance: Given an off workspace, When an admin opens Settings, Then the existing Off copy is shown and no new banner or prompt appears.
- **FR5 — Upgrade from v0.1.0 / v0.1.1 is a documented behaviour change.** Workspaces that connected on v0.1.x but never touched the switch (no record) turn OFF after the upgrade; their connection is kept, and sync, watches and Git credentials pause until an admin turns the switch on. No grandfathering or migration is added. [Q1]
  - **FR5.1** The README states that Backlog is off by default after installation and that an admin turns it on, then connects (replaces "on by default" at `README.md:38`).
  - **FR5.2** The release/upgrade notes for the release containing this fix state the upgrade effect above and the one-step remedy (admin turns the switch on).
  - Acceptance: Given a v0.1.x workspace with a saved connection and no switch record, When the plugin is upgraded, Then the workspace shows Off, the saved connection is still present, and turning the switch on restores sync without reconnecting.
- **FR6 — Automated checks follow the new default.** The packaged-host contract test (`internal/ci/contract.go:195-207`) expects the new default (`enabled: false`, and `connect_api_key` refused with `409 integration_disabled` before opt-in, or the test turns the switch on first), and the test that asserted "no record means enabled" (`TestSwitchWithNoRecordIsEnabled`) asserts the opposite. [desc]

## Non-Functional Requirements

- **NFR1 — Regression test.** A targeted regression test proves a workspace with no switch record is OFF and that guarded calls are refused (bugfix testing floor, TDD per team Testing Posture: write the failing test first).
- **NFR2 — Existing suite stays green.** `go test -race ./...` and Vitest pass after the change (baseline: Go 1159 tests, Vitest 286/286); the 80% Go coverage floor holds; existing tests that implicitly relied on the ON default are updated to turn the switch on explicitly (one shared test helper), without weakening any assertion.
- **NFR3 — Security unchanged.** No secret appears in logs, errors or UI responses (existing redaction rules); `set_enabled` stays admin-only.

## Constraints

- Kandev v0.96.0 has no manifest field for a default-off install; installing always starts the plugin process, so opt-in is implemented only through the per-workspace switch (code KB risk 4).
- Registrations (settings card, switch, nav entry, `/backlog` route) stay independent of the switch (BR5.4/BR7.6/BR7.8 unchanged).
- Team practices apply: TDD, `-race`, 80% coverage floor, contract test on the minimum Kandev version, no real credentials in tests.

## Assumptions

- [assumption] Background workers already skip OFF workspaces through `RequireEnabled`, so no worker code changes are needed (developer scan, D1). To be confirmed by the regression and existing worker off-state tests.

## Out of Scope

- Grandfathering or migrating v0.1.x workspaces (Q1 = A).
- New prompts, banners or wording for the Off state (Q2 = A).
- Any change to the switch's admin rules, the connection flow, or the registrations.

## Open Questions

None.

## Sources

- [desc] Initial description: project-description.json.
- [Q1], [Q2] `requirements-analysis-questions.md`.
- Code KB: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/business-overview.md`, `architecture.md`, `code-structure.md`, `code-quality-assessment.md` (D1–D6, risks 1–6).
