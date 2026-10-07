# Code Generation Plan — github-parity-actions

Zero-Unit express run: one implementation iteration. Scope comes from `inception/requirements-analysis/requirements.md` (FR1-FR5, NFR1-NFR5) and the code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`. Brownfield: files are changed in place, and no duplicates are created.

## Design Decisions

- **Quick actions live in `internal/issues`**. They go in a new `quick_actions.go` stored under the workspace state key `issues.quick_actions` as `{schemaVersion, issue: [...], pr: [...]}`. The Go side holds the defaults, so the default text exists in one place only. When a stored list is empty, that kind falls back to the defaults. "Reset" saves an empty list for that kind, so there is no separate reset action. We add no new Go package, because the existing issues store helpers cover this.
- **New actions** (all `authenticated`, added to `manifest.yaml` and `internal/plugin/manifest_test.go`):
  - `issues.quick_actions.get` returns `{issue, pr}`, already resolved, so defaults are applied.
  - `issues.quick_actions.save` takes body `{kind: "issue"|"pr", actions: [...]}`.
  - `issues.queries.list`, `issues.queries.save`, `issues.queries.delete`, `issues.queries.set_default` handle saved issue queries.
  - `git.queries.set_default` takes body `{id, isDefault}`.
- **Saved issue queries**: new store key `issues.queries`, at most 50, with name rules matching `git.QueryInput`. Fields: `id, name, projectKey?, statusIds[], assignee ("" | "me" | numeric id string), keyword, isDefault`.
- **Default flag on saved queries**: an optional `isDefault` is added to both `git.Query` and the issue query, read as false when missing, so no migration is needed. `set_default` clears every other default in the same read-modify-write, which gives at most one default per kind. Save keeps the existing `isDefault` of the edited row.
- **`issues.list` assignee `me`**: `issues.Query` gains `assignee string` (`""` or `"me"`). The server resolves `me` with one `gateway.Myself` call into `AssigneeIDs`, the same way `SaveWatch` does. The browser never handles the numeric id.
- **Built-in presets** (not stored, UI only):
  - Issues: "Assigned to me, open" = `assignee: "me"` plus every status id from `issues.filters` except Backlog's Closed (id 4). This settles the requirements' open question; the default is status id 4.
  - PRs: "Open, assigned to me" = `statuses: ["open"]`, `assignee: "me"`, on the first repository returned by `git.repositories.list`.
- **Quick action launch** (UI): `host.ui.IntegrationStartTaskMenu` sits in the `ChangeRequestRow` `action` slot. Picking an action opens `host.ui.TaskCreateDialog`, prefilled with title `"<label>: <title>"` (truncated to 100 chars) and the interpolated prompt (`{{url}}`, `{{title}}`; unknown placeholders are kept). The workflow comes from `host.context.getTaskCreationContext`. On `onSuccess(task)` the plugin calls `issues.link` `{taskId, body:{issueKey}}` or `git.prs.link` `{taskId, body:{reference: <PR URL>}}`. If the link call fails, a notice says "task created but not linked". The old row item "Create task" is removed, and "Link to task" stays in the row "..." menu.
- **Layout**:
  - `IntegrationScopeBar` replaces the `Tabs`. It holds the kind switch, the built-in preset pill, and the Saved menu with a star toggle, with classes `px-4 py-2 sm:px-6`.
  - The toolbar row is laid out like `IntegrationListToolbar` (`border-b px-4 py-2.5 sm:px-6`): title and count on the left, filters, then last-updated and a ghost refresh button on the right.
  - Lists use `px-3 py-4 md:px-6`, and issue rows use `ChangeRequestRow`.
  - The settings card, switch, nav entry and `settingsHref()` stay unchanged (FR5.4).

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "express",
  "test_strategy": "minimal",
  "project_type": "brownfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    },
    {
      "layer": "project",
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07)"
    }
  ],
  "obligations": {
    "strategy": "minimal",
    "strategy_volume": [
      "One verifiable test per requirement at the narrowest effective level.",
      "At least one happy-path unit test per component.",
      "Unit tests are the default; a bugfix/security scope floor may require an integration or E2E regression when that is the narrowest level that reproduces the defect."
    ],
    "scope_floor": [
      "Keep the existing test suite green.",
      "This scope adds no extra new-test floor beyond the selected test strategy."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Verify the existing test runner/configuration and record the exact unit-scoped command.",
    "runner_ready_before_first_test": true,
    "testable_layers": [
      "Data model / database behavior",
      "Repository / data access",
      "Business logic",
      "API / endpoint",
      "Frontend behavior"
    ],
    "steps": [
      "Project structure and production configuration skeleton.",
      "Verify the existing test runner/configuration and record the exact unit-scoped command.",
      "Data model / database behavior - Red: write the failing tests and record the failing command output.",
      "Data model / database behavior - Green: implement only enough behavior to pass.",
      "Data model / database behavior - Refactor: improve the implementation while tests stay green.",
      "Repository / data access - Red: write the failing tests and record the failing command output.",
      "Repository / data access - Green: implement only enough behavior to pass.",
      "Repository / data access - Refactor: improve the implementation while tests stay green.",
      "Business logic - Red: write the failing tests and record the failing command output.",
      "Business logic - Green: implement only enough behavior to pass.",
      "Business logic - Refactor: improve the implementation while tests stay green.",
      "API / endpoint - Red: write the failing tests and record the failing command output.",
      "API / endpoint - Green: implement only enough behavior to pass.",
      "API / endpoint - Refactor: improve the implementation while tests stay green.",
      "Frontend behavior - Red: write the failing tests and record the failing command output.",
      "Frontend behavior - Green: implement only enough behavior to pass.",
      "Frontend behavior - Refactor: improve the implementation while tests stay green.",
      "Environment/build configuration.",
      "Documentation and traceability."
    ]
  },
  "input_sha256": "sha256:4442c5fad82de74b31fc5e573b07e7444027e7437773b5002fb422bdfb2749f7",
  "contract_sha256": "sha256:6db87aafc6be218a4f8f762cfd33d3fec2e691661f16cdc25e6dcc4236038b2d"
}
```

## Plan Steps

The "Data model / database" layer has no database here. Its role is played by the plugin state documents, which belong to the Repository layer, so that layer is merged into the Repository steps.

### Structure and runner
- [x] **Step 1** — Structure check (no new package). Confirm at Kandev `v0.96.0` (`../kandev/apps/web/lib/plugins/host-api.ts`) that the plugin kit exposes `IntegrationStartTaskMenu`, `TaskCreateDialog`, `IntegrationScopeBar` and `ChangeRequestRow` (`action` prop) with the props used below. Record the citations in `code-summary.md`. If a component is missing, stop and report; do not build a replacement. (FR1.1, FR5.1, NFR2)
- [x] **Step 2** — Runner readiness: run the exact scoped commands in `unit-test-instructions.md` once on the unchanged tree (Go with `PATH=$HOME/.local/go/bin:$PATH`, Vitest in `ui/`). All must pass. (NFR3)

### Repository / data access (plugin state documents)
- [x] **Step 3** — Red: tests in `internal/issues/quick_actions_test.go` (store round-trip; empty list falls back to defaults; the 6 default ids and labels) and `internal/issues/queries_test.go` (save/list/delete; limit 50; `set_default` leaves exactly one default; missing `isDefault` reads false). Add to `internal/git/queries_test.go`: `set_default` on PR queries moves the star; save keeps `isDefault`. Record the failing output. (FR2.1, FR2.2, FR3.3, FR4.3, FR4.4, NFR2)
- [x] **Step 4** — Green:
  - `internal/issues/quick_actions.go`: types `QuickAction{ID, Label, Hint, Icon, PromptTemplate}`, `QuickActions{Issue, PR}`, defaults, store load/update on key `issues.quick_actions`.
  - `internal/issues/queries.go`: type `IssueQuery`, store key `issues.queries`, max 50.
  - `internal/git/types.go`: `QueryInput.IsDefault bool \`json:"isDefault,omitempty"\``.
  - `internal/git/service.go`: `SetQueryDefault`.
- [x] **Step 5** — Refactor while green: reuse the existing `update`/`load` helpers and add no new helper copies (the store-helper `ponytail` notes stay true).

### Business logic
- [x] **Step 6** — Red:
  - Validation tests: label required and at most 100 runes; prompt at most 4000; at most 20 per kind; icon in the allowed set; kind `issue|pr`; issue query name required and at most 100 runes, assignee `""|me|<id>`, status ids at most 50, keyword at most 100.
  - `internal/issues/list_test.go`: `assignee: "me"` makes one `Myself` call and sends `assigneeId[]` = the connected user to the fake Backlog server; an unknown assignee value gives `validation`.
  - Record the failing output. (FR2.4, FR4.2, NFR1)
- [x] **Step 7** — Green: `Validate` methods; `issues.Query.Assignee` with resolution in `Service.List`; `Service.QuickActions` / `SaveQuickActions` / `ListQueries` / `SaveQuery` / `DeleteQuery` / `SetQueryDefault` in `internal/issues`.
- [x] **Step 8** — Refactor while green.

### API / endpoint
- [x] **Step 9** — Red:
  - `internal/plugin/actions_u3_test.go` / `actions_u4_test.go`: the 7 new actions decode bodies and return the expected shapes; a bad body gives `validation`.
  - `internal/plugin/manifest_test.go`: the new keys exist with `authenticated` access and match `^[a-z0-9][a-z0-9._-]*$`.
  - Extend the existing leak/redaction test so the new actions' responses and errors contain no API key or token.
  - Record the failing output. (FR2.3, FR3.3, FR4.3, FR4.4, NFR1, NFR2)
- [x] **Step 10** — Green: handlers in `internal/plugin/issue_actions.go` and `git_actions.go`; entries in `manifest.yaml`.
- [x] **Step 11** — Refactor while green; run `go test -race ./internal/...` in full.

### Frontend behavior
- [x] **Step 12** — Harness: in `ui/src/testing/harness.ts`, add fakes for `IntegrationStartTaskMenu` (renders one button per preset with `itemTestId`), `TaskCreateDialog` (renders when `open`, exposes the initial title/description, and a confirm that calls `onSuccess({id})`) and `IntegrationScopeBar` (kind buttons, preset pills, saved items with a star toggle). Make the `ChangeRequestRow` fake render its `action` prop. (test infrastructure)
- [x] **Step 13** — Red, quick actions: new `ui/src/page/quick-actions.test.ts` covering `interpolate` (`{{url}}`/`{{title}}`, unknown placeholder kept) and `taskTitle` (truncation). New `ui/src/page/start-task.test.tsx`: picking "Investigate" on an issue row opens the dialog with the prefilled title and prompt; on success `issues.link` is called with `taskId`; a link failure shows the "created but not linked" notice; the PR row calls `git.prs.link` with the PR URL. Record the failing output. (FR1.1-FR1.4)
- [x] **Step 14** — Green: `ui/src/page/quick-actions.ts` (types, `interpolate`, `taskTitle`, load via `issues.quick_actions.get`) and `ui/src/page/start-task.tsx` (menu, dialog and link). Wire it into the issue rows (`ui/src/issues/issues-page.tsx`, rows switched to `ChangeRequestRow`; the "Create task" menu item and its now-unused code removed, "Link to task" kept) and the PR rows (`ui/src/git/pr-list.tsx`). Add messages to `ui/src/messages/en.ts` / `ui/src/issues/i18n.ts`.
- [x] **Step 15** — Red, quick-actions settings: in `ui/src/settings/sections.test.tsx`, the "Quick actions" section shows the Issues tab with the defaults; edit, add, delete and Save call `issues.quick_actions.save` with the right body; Reset saves an empty list and shows the defaults again; an over-long label shows the validation message. (FR2.3, FR2.4)
- [x] **Step 16** — Green: new `ui/src/settings/quick-actions-section.tsx`, mounted in `SettingsScreen.tsx`.
- [x] **Step 17** — Red, default queries and scope bar: `ui/src/page/backlog-lists.test.tsx` / `backlog-page.test.tsx`:
  - Opening `/backlog` with no starred issue query sends `issues.list` with `assignee: "me"` and the non-closed status ids.
  - With a starred issue query, its filters are sent.
  - Switching to Pull requests with no starred PR query sends `git.prs.list` for the first repository with `statuses:["open"]`, `assignee:"me"`; with a starred PR query, that query's filters are sent.
  - Starring a saved query calls `*.queries.set_default`.
  - Saving issue filters calls `issues.queries.save`.
  - The scope bar replaces the tabs.

  Record the failing output. (FR3.1-FR3.4, FR4.1, FR4.3-FR4.5, FR5.1)
- [x] **Step 18** — Green: `ui/src/page/BacklogPage.tsx` (scope bar, kind state, preset/saved selection, default-on-open), `ui/src/issues/issues-page.tsx` and `ui/src/git/pr-list.tsx` (accept the selected query; save issue query through the existing `ui/src/git/save-query-dialog.tsx` pattern), `ui/src/settings/saved-queries-section.tsx` (list issue queries next to PR queries: rename, delete, star).
- [x] **Step 19** — Red then Green, layout: one test in `ui/src/page/backlog-page.test.tsx` checks the classes. The scope bar wrapper has `px-4 py-2 sm:px-6`, the toolbar has `border-b px-4 py-2.5 sm:px-6` with the refresh button as the last toolbar child, and the list container has `px-3 py-4 md:px-6`. Then update `ui/src/layout.ts` constants and `ui/src/git/pr-toolbar.tsx` / issues toolbar to that layout. (FR5.2, FR5.3)
- [x] **Step 20** — Refactor while green: delete code made dead by the switch (old tabs wiring, unused row-menu create path), and run the full Vitest suite, `tsc`, ESLint and Prettier. Settings card, switch, nav entry and `settingsHref()` tests stay green unchanged. (FR5.4)

### Environment/build, documentation, traceability
- [x] **Step 21** — Build configuration: no new dependency, and `go mod tidy` leaves `go.mod`/`go.sum` unchanged. Bump `manifest.yaml` `version` to `0.2.0` (new features). (NFR2)
- [x] **Step 22** — Full verification: `make check-format vet lint test coverage` (coverage of at least 80%; delete `coverage.out` afterwards), then in `ui/`: `npm run typecheck`, `npm run lint`, `npm run format:check`, `npm test` (or the real script names in `ui/package.json`), then `make package verify-package`. (NFR3)
- [x] **Step 23** — Documentation and traceability: update `README.md` (quick actions, default queries) if it documents features; write `code-summary.md`, `source-manifest.json` and `traceability.json` (FR/NFR → implementation/test files).

## Requirement-to-Step Traceability

| Requirement | Steps |
|---|---|
| FR1.1-FR1.4 quick action launch | 1, 12, 13, 14 |
| FR2.1-FR2.2 defaults and fallback | 3, 4 |
| FR2.3 settings editing | 9, 10, 15, 16 |
| FR2.4 validation | 6, 7, 15 |
| FR3.1-FR3.2, FR3.4 PR default query | 17, 18 |
| FR3.3 PR star | 3, 4, 9, 10, 17, 18 |
| FR4.1, FR4.5 issue default query | 17, 18 |
| FR4.2 assignee `me` | 6, 7 |
| FR4.3-FR4.4 saved issue queries and star | 3, 4, 9, 10, 17, 18 |
| FR5.1-FR5.3 layout | 12, 17, 18, 19 |
| FR5.4 locked behaviour | 20 |
| NFR1 security | 6, 9 |
| NFR2 compatibility | 1, 3, 9, 21 |
| NFR3 quality | 2, 11, 20, 22 |
| NFR4 performance | 14, 18 (quick actions loaded once per page; one list call per default query) |
| NFR5 accessibility | 13, 14 (`aria-label` on "+ Task" and star via host components) |
