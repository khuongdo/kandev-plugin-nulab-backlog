# Code Generation Plan — issues (U3)

Inputs:

- `unit-of-work`: the U3 boundary. It covers the whole IssueIntegration component: screens M2, M2m, M3 and M8, the issue badge on the card (M6), the `#` source (M10), the issue status sync cycle, list performance, and accessibility and language support for the plugin screens. It also uses `unit-of-work-story-map` (13 stories, in order) and `unit-of-work-dependency` (issues depends on connection only; issue links belong to `issues`, ADR-004).
- `stories`: US2.1, US2.2, US3.1, US3.3, US2.3, US3.2, US3.4, US4.1, US4.2, US8.1, US8.2, US8.5, US3.5. It also covers the issue-link parts of these ACs, which U2 deferred: AC1.4.3 (the settings link from other Backlog pages), AC1.5.4, AC1.7.1, AC1.7.2, AC1.8.1, AC1.8.2, AC1.8.3, AC1.9.1, AC1.9.2, AC8.3.2 (the issue cycle log) and AC8.4.4 (on the Issues page).
- `requirements`: FR1.7, FR2.1–FR2.3, FR3.1–FR3.5, FR4.1–FR4.4, NFR1, NFR2, NFR3, NFR5, NFR8, NFR9, NFR10, NFR11; assumption A4.
- `components` (IssueIntegration, Connection, BacklogGateway, KandevAdapter, PluginUI) and `decisions` (ADR-002, ADR-003, ADR-004, ADR-005, ADR-006, ADR-007).
- `contract-summary`: C1 (U3 calls), C2 (`HostPort`), C3 (`ConnectionReader`, `ConnectionChanged`), C5 (`issues.*`, `IssueQuery`, `IssuePage`), C7 (Backlog issue endpoints), C8 (`events: [task.deleted]`, `reference_sources`, `task-sidebar`, `task-card-tags`). Also the contract-review findings R-03, R-04 and R-05, which the bolt plan gives to U3, and the open questions on labels and on the `issues.createTask` time budget.
- `mockups` (M1 interval field, M2, M2m, M3, M6, M8, M10, M12), `interaction-spec` (BacklogListState, ConfirmDialog, IssueRow/IssueCard, BacklogSidebarPanel), `design-system-mapping`, `accessibility-checklist`.
- `bolt-plan` (B4 Definition of Done) and `external-dependency-map` (X7).
- U2 and U4 `code-generation-plan.md` and `code-summary.md`, and the U1/U2/U4/U5 code on `main` at `c3ca400`.
- `team.md`, `project.md`, `phases/construction.md`.

U3 has no functional, NFR or infrastructure design; the user chose to go straight to code generation. Every design choice in this plan is listed under Assumptions & Open Questions.

Stories in U3, in build order: US2.1, US2.2, US3.1, US3.3, US2.3, US3.2, US3.4, US4.1, US4.2, US8.1, US8.2, US8.5, US3.5. All are Must except US8.5 and US3.5 (Should). US3.5 stays Should per `project.md`.

## Kandev SDK Facts Used by This Plan

Read from `../kandev` at v0.96.0: `apps/backend/pkg/pluginsdk/{plugin.go,host.go,types.go,data_types.go}`, `apps/packages/plugin-sdk/src/index.ts`, `docs/public/plugins-manifest.md`, `docs/public/plugins-authoring.md`, `apps/backend/internal/plugins/{mentions.go,host_data.go,host_write.go}`, `apps/backend/internal/plugins/manifest/{manifest.go,validate.go}`, `apps/backend/internal/mentions/{types.go,service.go}`, `apps/backend/internal/events/types.go`.

- **Task actions and task menu.** `registry.registerTaskMenuAction({id, label, icon?, group: "edit" | "primary", visible?(PluginTaskMenuContext), items?, run(PluginTaskMenuContext)})`. `PluginTaskMenuContext` = `{workspaceId, taskId, taskTitle, workflowStepId, presentation}`. `visible` and `items` must be synchronous and cheap; a rejected `run` is caught and logged. `registerTaskAction({placement: "link", ...})` adds an entry to Kandev's **Link** submenu, which U4 already uses for PRs.
- **Task-to-issue link dialog.** `host.openTaskLinkDialog({title, description, inputLabel, placeholder?, emptyError, failureMessage, successMessage, inputTestId?, errorTestId?, submitTestId?, onSubmit(reference, signal)})`. The dialog is opened from a *task* and takes one text reference. Kandev has no task-picker dialog; M3 (choosing a task from an issue row) is a plugin dialog built from the `host.ui` kit, `host.openModal` or the U2 `confirm-dialog.tsx` pattern.
- **Issue picker / `#` suggestions.** The manifest declares `reference_sources: [{source, provider, kind, display_name, kind_label, order?}]`. `source`, `provider` and `kind` must match `^[a-z0-9][a-z0-9._:-]{0,127}$`. Labels are non-empty, trimmed and at most 100 characters. Each source and each provider/kind pair is unique. The Go plugin implements `pluginsdk.EntityReferenceHandler`:
  - `SearchEntityReferences(ctx, *SearchEntityReferencesRequest{Source, WorkspaceID, Query, Limit})` → `{Candidates: [{ProviderLocalID, Title, URL, Attributes}]}`. Kandev uses `ProviderLocalID` as both id and key.
  - `AuthorizeEntityReference(ctx, *AuthorizeEntityReferenceRequest{Source, WorkspaceID, Purpose, Reference{version, ref, provider, kind, id, key, title, url, scope}})` → `{Allowed, Reason}`.
  - Kandev gives each search and each authorization **1.5 s** (`defaultProviderTimeout`, `referenceAuthorizationTimeout`). `Limit` defaults to 5, max 10 (`mentions.DefaultLimit`/`MaxLimit`). Search results are display-only. At submission Kandev calls Authorize again; an error, a timeout or a denial fails closed. Kandev renders the source menu, the chips, and the arrow/Enter selection.
- **Task creation from an issue (Go).** `Host.Tasks().Create(ctx, CreateTaskInput{WorkspaceID, WorkflowID, WorkflowStepID*, Title, Description, Priority ("critical"|"high"|"medium"|"low", empty = medium), Repositories, Metadata, ...})` returns `*Task{ID, Identifier, ...}` and needs `api_write: [tasks]`. Metadata is nested under `"plugin:nulab-backlog"`; U4's `hostPort` already handles this. `Tasks().List(ctx, TaskFilter{WorkspaceIDs, IncludeArchived}, Page)` and `Tasks().Get` need `api_read: [tasks]`. `Get` on a missing task returns gRPC `NotFound`. `Tasks().Update` only changes title, description, state and priority. **`Task.Labels` is deprecated:** "provider-specific labels belong in plugin-owned task state and UI slot data". So there is no `SetTaskLabels`; badges are drawn from the plugin's own link data (ADR-004). The browser side has `host.context.getTaskCreationContext(workspaceId)` → `{workflowId, defaultStepId, steps, repositories}`.
- **Task metadata and comments.** The host has no API to read or write Kandev task comments, and no API to write task metadata after creation. U3 needs neither: issue comments are read live from Backlog (Q9).
- **Pages, routes and slots.** `registerRoute(path, Component, {topbar})` and `registerNavItem({..., section: "integrations"})` (U1/U4 use them at `/backlog`, `/backlog/watches`, `/backlog/dashboard`). `registerComponent(slot, Component)` has these relevant slots:
  - `task-card-tags`: its own row on the Kanban card; `slotProps` = `{taskId, workspaceId, workflowStepId}`.
  - `task-row-metadata`: `{taskId, workspaceId, workflowStepId, surface}`.
  - `task-sidebar`: bottom of the task-detail sidebar; **no `slotProps`**, so no task id.

  `registerTaskPanel({id, title, titleKey?, icon?, Component, mobileEnabled?, visible?})` gives the component `{panelId, taskId, sessionId, sessionKind, presentation, conversation}`.
- **Events.** `capabilities.events: ["task.deleted"]` delivers `Plugin.OnEvent(ctx, *Event{EventID, EventType, OccurredAt, WorkspaceID, Payload})`. A returned error is retried after 5, 15 and 45 s. Delivery is best effort, so the plugin also reconciles from `Tasks().List` (C8, docs §5).
- **Language.** `registry.registerTranslations({en: {...}, <locale>: {...}})` takes flat catalogs: at most 1,000 messages per locale, values up to 4,096 characters, `{{name}}` interpolation. `host.i18n.locale`, `host.i18n.t(key, {defaultValue})` and `host.i18n.useTranslation()` give locale-aware lookup. A missing message falls back to the plugin's English catalog.
- **UI kit.** `host.ui` includes Table*, Badge, Button, Input, Label, Select*, Checkbox, Skeleton, Spinner, Alert*, Empty*, Pagination*, Drawer*, Sheet*, Collapsible*, Tooltip*, Dialog*, `IntegrationCursorPagination`, `IntegrationListToolbar`. Also `host.useResponsiveBreakpoint()` → `{isMobile}`, `host.toast`, `host.utils.formatRelativeTime`, `host.navigate`, `host.openModal`.
- **Action rules (unchanged).** Keys match `^[a-z0-9][a-z0-9._-]*$`. Each action gets 15 s. A task-scoped action receives a verified `TaskID`. Only `Content-Type`, `Cache-Control`, `ETag` and `Retry-After` headers reach the browser.
- **Capabilities.** Already declared: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`. U3 adds only `events: ["task.deleted"]` and `reference_sources`. `min_kandev_version` stays `0.96.0`.
- **Dependencies.** Standard library only (`net/http`, `net/url`, `sync`, `time`, `testing/synctest`). No Go module or npm package is added. `google.golang.org/grpc/status` is not imported (U4 deviation 2), so task existence is checked with `Tasks().List`, not by reading the `NotFound` code of `Get`.
- **Bitbucket plugin.** <https://github.com/kdlbs/kandev-plugin-bitbucket> was not re-read for this plan. It implements the same `reference_sources` and Link surfaces for PRs (the authoring guide's screenshots), and U4 already followed it.

### Backlog API v2 facts (developer.nulab.com, read 2026-10-06)

- **Issue list.** `GET /api/v2/issues` with `projectId[]`, `statusId[]`, `assigneeId[]`, `priorityId[]`, `id[]`, `keyword`, `sort` (incl. `updated`), `order` (`asc`|`desc`, default `desc`), `offset`, `count` (1–100, default 20). Each issue has `id, projectId, issueKey, keyId, issueType, summary, description, priority{id,name}, status{id,name}, assignee, startDate, dueDate, createdUser, created, updatedUser, updated, attachments, ...`. It filters by **project id**, not key.
- **Issue count.** `GET /api/v2/issues/count` takes the same filters → `{"count": 43}`.
- **Issue get.** `GET /api/v2/issues/:issueIdOrKey` (U4 already calls it).
- **Comments.** `GET /api/v2/issues/:issueIdOrKey/comments` with `minId`, `maxId`, `count` (1–100, default 20), `order` (default `desc`) → `[{id, content, changeLog, createdUser{name,...}, created, updated, ...}]`.
- **Attachments.** `GET /api/v2/issues/:issueIdOrKey/attachments` → `[{id, name, size (bytes), createdUser, created}]`. The file itself is `GET /api/v2/issues/:issueIdOrKey/attachments/:attachmentId` (binary, `application/octet-stream`).
- **Project statuses.** `GET /api/v2/projects/:projectIdOrKey/statuses` → `[{id, projectId, name, color, displayOrder}]`.
- **Project users.** `GET /api/v2/projects/:projectIdOrKey/users` → `[{id, userId, name, ...}]`.
- **Priorities.** `GET /api/v2/priorities` example: `2 High, 3 Normal, 4 Low`.
- **Rate-limit groups (U2 `group.go`).** The issue list and count are **Search** (queued per host, 1 s apart). Every other GET is **Read** (not queued, still retries on 429). U3 makes **no** Update call: no POST, PATCH or DELETE to Backlog, AC3.2.3 and FR4.3.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "feature",
  "test_strategy": "standard",
  "project_type": "greenfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    }
  ],
  "obligations": {
    "strategy": "standard",
    "strategy_volume": [
      "Five to eight tests per component.",
      "Unit tests plus integration tests for key boundaries.",
      "Add E2E, performance, or security tests when requirements demand them."
    ],
    "scope_floor": [
      "Meet an 80% line-coverage floor.",
      "Run the selected tests in CI before merge."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
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
      "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
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
  "input_sha256": "sha256:c08268ec87b4a805bbc352db69f8ea8730f7297e5e63201ea588835bf2850378",
  "contract_sha256": "sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc"
}
```

## Plan Steps

Each TDD layer goes Red → Green + Refactor. Each Red step says "Run the tests and record the failing output", and that output goes into `construction/issues/code-generation/code-summary.md`.

**Test naming rule.** Every new Go test function in U3 starts with `TestU3_`, so the unit filter is `-run '^TestU3_'`. No existing U1, U2, U4 or U5 test starts with `TestU3_`. U5's `check-secrets` treats a run of 32 or more mixed-case letters and digits as a credential, so every underscore-separated part of a test name stays under 32 characters, for example `TestU3_Sync_KeepsLinkOnNotFound`. Vitest files for U3 live in `ui/src/issues/` and are listed by exact path in `unit-test-instructions.md`.

**Guards for every Backlog call (applies to every step).**

- Each action goes through the existing plugin guard (`RequireEnabled`, 409 `integration_disabled`). The two Kandev RPCs (`SearchEntityReferences`, `AuthorizeEntityReference`) and the sync worker call `RequireEnabled` themselves and fail closed.
- Each call reads `Current(ctx, ws)` and uses only `SelectedProjects`.
- Calls are Interactive (3 s rate-limit budget) except the sync cycle, which is Background.
- Results are dropped when the epoch moved before the write (AC1.8.3).

### Step 1 — Runner readiness and structure (existing runner)

- [x] Add the IssueIntegration package `internal/issues` with only `doc.go`. It does not import `pluginsdk` and does not import `internal/git` (U3 must not depend on U4).
- [x] Add fake-only fixtures shaped exactly like Backlog payloads in `internal/backlog/testdata/`:
  - `issues_ok.json`: `PROJ-118`, `PROJ-120` and `PROJ-123`, with `status`, `priority`, `assignee`, `dueDate`, `updated`, and one 200-character Japanese summary.
  - `issues_count_ok.json` (`{"count": 57}`), `issue_detail_ok.json`, `comments_ok.json`, `attachments_ok.json` (`spec.pdf` 1.2 MB, `dump.zip` 48 MB), `statuses_ok.json`, `project_users_ok.json` (`Test User`, `Lan`).

  Lists of 57 issues and 150 comments are generated inside tests. The U4 `issue_ok.json` is unchanged.
- [x] Do not touch `Makefile`, `.github/workflows/*`, `.golangci.yml`, `internal/ci`, `cmd/`, `go.mod`, `go.sum` or `ui/package.json`.
- [x] Confirm the runner before the first Red step, from the repository root: `go test -race ./internal/backlog/... ./internal/plugin/... -run '^TestU3_'`. Each package should report `ok … [no tests to run]`. From Step 1 on, add `./internal/issues/...` (`[no test files]`). From `ui/`, run the 8-file Vitest command in `unit-test-instructions.md` with `--passWithNoTests`.
- [x] Confirm that the U2 and U4 regression commands in their `unit-test-instructions.md` are green.
- Stories: all U3 (fixtures). NFRs: NFR4, NFR8.

### Step 2 — Data model, Red: Backlog issue types and IssueIntegration entities

- [x] `internal/backlog/issues_types_test.go`:
  - Decoding `Issue{ID, ProjectID, IssueKey, Summary, Description, StatusID, StatusName, PriorityID, AssigneeID, AssigneeName, DueDate, Updated}`, `Comment{ID, Content, AuthorName, Created}`, `Attachment{ID, Name, Size}`, `Status{ID, Name}`, `ProjectUser{ID, Name}`, and the `{count}` body.
  - A non-numeric id or an empty key is rejected as `Unreachable`/`body`.
  - `IssueQuery.Values()` encodes `projectId[]`, `statusId[]`, `assigneeId[]`, `keyword`, `sort=updated`, `order=desc`, `offset`, and `count` clamped to 1–100.
  - AC2.2.4 table: the keyword encodes Japanese, Vietnamese, `&`, `%` and spaces correctly.
  - `ValidIssueRef` accepts `^[A-Z][A-Z0-9_]*-[1-9][0-9]{0,8}$` or a numeric id. Anything else is refused before a path is built (path injection).
  - `%v` of an `Issue` and of a `Comment` shows no description or content body.
- [x] `internal/issues/types_test.go` (table-driven):
  - `ParseIssueKey` (project part plus number).
  - `ValidateQuery(raw, selected)`: `page ≥ 1`; `pageSize` 1–100, default 20; `projectKeys` must be a subset of the selected projects, else field `projectKeys`; at most 50 status and 50 assignee ids; keyword ≤ 100 runes, trimmed.
  - `ShowingRange(page, size, total)` → `1-20 of 57`, `41-57 of 57`, `hasNext`.
  - `PriorityFor(id)`: 2 → `high`, 3 → `medium`, 4 → `low`, other → `medium` (AC3.1.1).
  - `NewTaskFor(issue, host)`: title = summary; description = issue description plus `Backlog: https://<host>/view/<KEY>`; priority mapped.
  - `ValidatePollMinutes`: AC4.2.2 table of 30 s (`0.5`), `0`, negative, letters and empty, each refused with field `minutes`; `1`–`1440` accepted; default 5.
  - `Link{IssueKey, IssueID, ProjectKey, SpaceHost, TaskID, TaskKey, State: active|not_connected, Unavailable, LastKnownStatus, StatusUpdatedAt, FailCount, ConnectionEpoch, CreatedAt}` with `Stale()` = `FailCount ≥ 3` (AC4.1.4).
  - `Matches(issue, query)`: for the exact-key row (AC2.2.1, AC2.2.2).
  - `SuggestMatch(issue, q)`: the key starts with q, or the title contains q, case-insensitive (AC3.4.1).
- [x] Run the tests and record the failing output.
- Stories: US2.1, US2.2, US3.1, US3.4, US4.1, US4.2. Contracts: C1, C5 (`IssueQuery`, R-04). ACs: AC2.1.1, AC2.1.2, AC2.2.1, AC2.2.2, AC2.2.4, AC3.1.1, AC3.4.1, AC4.1.4, AC4.2.1, AC4.2.2.

### Step 3 — Data model, Green and Refactor

- [x] Add `internal/backlog/issues.go` with the types and parsers. Move U4's `Issue`/`parseIssue` from `pullrequests.go` into it and widen the struct; this change only adds fields, and U4's tests stay green.
- [x] Add `internal/issues/types.go`.
- [x] Refactor while green.
- Contracts: C1, C5. NFRs: NFR3.

### Step 4 — Gateway and data access, Red: Backlog issue calls and the IssueIntegration store

Tests use the `httptest` fake Backlog, the in-memory `RoundTripper` from `limiter_test.go` for timing, and injected `Now`/`Wait` under `testing/synctest`.

- [x] `internal/backlog/issues_client_test.go`, for each of these calls:
  - `Issues(ctx, creds, class, IssueQuery)` and `IssueCount(...)`: exact query string; **Search** group, queued and 1 s apart from a concurrent count (`-race`).
  - `Issue(ctx, creds, class, ref)`: an invalid ref makes 0 requests.
  - `IssueComments(ctx, creds, class, ref, CommentQuery{MaxID, Count})`: `count` 1–100, `order=desc`.
  - `IssueAttachments(ctx, creds, class, ref)`, `ProjectStatuses(ctx, creds, projectKey)`, `ProjectUsers(ctx, creds, projectKey)`: **Read** group, not queued.

  For each call, check 401, 403, 404, 429 (`Background` waits the full time; `Interactive` keeps the 3 s budget) and 500 with a bait body that never appears in errors or logs (AC7.3.2). Check that a fake that never answers returns `Unreachable`/timeout within the 10 s limit while another call is still served (AC8.1.3). Check that a counter over every test sees **0** non-GET requests (AC3.2.3).
- [x] `internal/issues/store_test.go`:
  - Workspace state documents `issues.links` and `issues.settings` (`{pollMinutes, lastCycleAt}`) round-trip with `schemaVersion: 1`.
  - An unknown `schemaVersion` is never overwritten (`ErrStore`).
  - The instance key `issues.index` lists the workspaces that have links.
  - Caps: 1,000 links (`validation` field `limit`).
  - Store calls keep the 1 s limit; workspaces are independent; parallel link writes in one workspace are race-free (`-race`).
- [x] Run the tests and record the failing output.
- Stories: US2.1, US2.2, US3.2, US3.4, US3.5, US4.1, US4.2, US8.1. Contracts: C1, C7. ADR-002, ADR-003, ADR-004. NFRs: NFR2, NFR3, NFR5. ACs: AC2.2.4, AC3.2.3, AC7.3.2 (list path), AC8.1.3.

### Step 5 — Gateway and data access, Green and Refactor

- [x] Add the calls to `internal/backlog/issues.go` through the existing `Client.send` path. `group.go` is unchanged: the issue list and count are already Search.
- [x] Add `internal/issues/store.go`. It follows the U4 `internal/git/store.go` pattern (one get/put per document, a per-workspace mutex, `errUnchanged`). `ponytail:` the generic document helper is copied, not shared; extract it when a third package needs it.
- [x] Refactor while green.
- Contracts: C1, C7. NFRs: NFR2, NFR3, NFR5.

### Step 6 — Business logic, Red: IssueIntegration service, sync cycle and events

Tests use a fake gateway with per-path and per-group counters, a fake `Connection` (`Current`, `Credentials`, `Subscribe`, `RequireEnabled`), a fake `HostPort`, and `synctest` for timers.

- [x] List (`list_test.go`):
  - `List(ws, query)` reads `Current` and maps the selected project keys to ids with one `Projects` call, cached per (workspace, epoch).
  - It then calls `Issues` and `IssueCount` (Interactive). The `IssuePage` has `total`, `refreshedAt`, `connectionEpoch`, and items `{issueKey, summary, status, assignee, updatedAt, url, linkedTasks: [{taskId, taskKey}]}` (AC2.1.1, AC2.3.1).
  - Filter table (AC2.2.1): every row satisfies all active filters.
  - On page 1, an exact key typed as the keyword gets one `Issue` call and goes on top when it passes the filters and the search did not return it (AC2.2.2). An unknown key gives an empty page, not an error (AC2.2.3).
  - No selected project → `ErrNoProject` → `validation` field `projectKeys` (AC1.7.2). Only selected projects are queried (AC1.7.1). Not connected → `reconnect_required`.
  - With a fixed 300 ms fake delay per request, the reply is ready within **2.5 virtual seconds** (AC8.1.1).
  - Errors keep their codes, with no Backlog body (AC2.1.4).
- [x] Filters (`filters_test.go`): `Filters(ws)` returns each selected project with `ProjectStatuses` (deduplicated by id) and the union of `ProjectUsers`, using one call per project.
- [x] Create task (`create_test.go`):
  - `CreateTask(ws, issueKey, force)` takes a per-(workspace, key) lock: a second concurrent call → `ErrConflict` with 0 extra creates (AC3.1.2, `-race`).
  - If the issue has active links and `force` is false → `ErrConflict` (AC3.1.3). With `force`, a second task is created.
  - It reads the issue with one `Issue` call, checks the epoch, and creates the task through `HostPort.CreateTask` with the workflow from the request. The title, description and priority follow Step 2. The link stores `taskKey` = `Task.Identifier`, and the reply is `{taskId, taskKey, issueKey}` (AC3.1.1).
  - A Backlog error or a Kandev refusal gives no task and no link (AC3.1.4).
  - A link write failure after the create returns `ErrStore` with the created task id in the log only. `ponytail`: no automatic relink.
- [x] Link and unlink (`links_test.go`):
  - `SearchTasks(ws, query, issueKey)` returns at most 20 workspace tasks `{taskId, taskKey, title, linkedIssueKey}` from `HostPort.ListTasks`.
  - `Link(ws, taskId, issueKey)` checks the issue exists in a selected project (one `Issue` call). It allows a second task for the same issue (AC3.3.2), refuses a task linked to another issue with `ErrConflict` (AC3.3.3, FR3.4), and is idempotent for the same pair.
  - `Unlink(ws, taskId)` removes only that link, with no Backlog write (AC3.3.5).
  - `Links(ws)` returns every link for the badges, with state, status, `statusUpdatedAt` and `stale`.
- [x] Detail (`detail_test.go`):
  - `Detail(ws, taskId)` makes one `Issue` and one `IssueAttachments` call (Interactive) and returns `{issue: {key, summary, status, assignee, priority, dueDate, url}, linkState, statusUpdatedAt, attachments: [{name, size, tooLarge}]}` (AC3.2.1). `tooLarge` = size > 10 MiB (AC3.5.2).
  - A due-date change on the fake shows on the next call, and the fake host sees 0 `Update` calls (AC3.2.2).
  - 404/403 → `ErrIssueUnavailable` → `not_found`. A not-connected link → `linkState: not_connected` with 0 Backlog calls.
  - An attachments failure still returns the issue, with `attachmentsError` set.
  - `Comments(ws, taskId, maxId)`: newest first, 20 per page, `nextMaxId` while more remain; 150 comments → 8 pages (AC3.5.1). A failure keeps its code (AC3.5.3).
- [x] Suggestions (`suggest_test.go`):
  - `Suggest(ws, query, limit)`: limit 1–10, default 5. An exact key gets one `Issue` call plus one keyword `Issues` call (Search, count = limit). Results follow `SuggestMatch`, the exact match first, then deduplicated.
  - Candidates are `{ProviderLocalID: key, Title: summary, URL}` (AC3.4.1).
  - Not connected, switch off, no match or an unknown key give `[]` and no error (AC3.4.3).
  - Five queries within 250 virtual ms for one workspace make exactly **1** Backlog search, for the last query; the earlier calls return `[]` (AC3.4.2).
  - The whole call stays under a 1.2 s deadline.
  - `Authorize(ws, reference)` allows only a key in a selected project of the connected space, with a live `Issue` 200 within 1.2 s. Every other case fails closed.
- [x] Sync (`sync_test.go`):
  - One `Syncer{Tick: 1 min, Now}` worker serves both the ticks and `Refresh` requests.
  - A workspace runs when `now ≥ lastCycleAt + pollMinutes`: by default 5, and after setting 2 the cycles are 2 virtual minutes apart (AC4.2.1).
  - A status change on the fake shows on the link after exactly one cycle (AC4.1.1).
  - 404/403 table → `Unavailable` and the link is kept; other links in the same cycle still update (AC4.1.2).
  - After 3 failed cycles the link keeps its last status with `stale` (AC4.1.4).
  - A cycle makes 0 non-GET requests (AC4.1.3).
  - A new `Syncer` over the same store continues from `lastCycleAt` (AC4.1.5).
  - `Refresh` while a cycle runs never makes the cycles overlap: the fake's in-flight counter stays ≤ 1 (AC4.2.4, `-race`). `Refresh` returns `{updatedCount, refreshedAt}` (AC4.2.3).
  - All calls are Background, and a 429 waits without blocking `issues.list`.
  - With the switch off, or when not connected, a cycle makes 0 requests (AC1.5.4).
  - An epoch change during the cycle drops its results (AC1.8.3).
  - Each cycle lists the workspace's tasks once through `HostPort.ListTasks` (archived tasks included), removes links whose task is gone (AC2.3.2), and refreshes `taskKey`.
  - Exactly one `issue_sync_cycle` log line per cycle, with `workspaceCount`, `updated`, `errors`, `durationMs` and `rateLimitWaits`, and no secret (AC8.3.2).
- [x] Events (`events_test.go`):
  - `disconnected` → all links `not_connected`, and the next cycle makes 0 requests (AC1.5.4).
  - `space_changed` → links of the old host become `not_connected` (AC1.8.2).
  - `projects_changed` → only links of deselected projects change; the other projects are not affected (AC1.9.2).
  - `Restore` with a matching host and project → links become active again (M12).
  - A lower epoch is ignored. Startup and every cycle reconcile with `Current()`.
  - `OnTaskDeleted(ws, taskId)` removes that task's links and is idempotent (R-05).
- [x] Impact (`impact_test.go`): `Impact(ws, projectKeys)` → `{issueLinks}` for all active links or only the given projects (AC1.8.1, AC1.9.1 issue part).
- [x] Settings (`settings_test.go`): `SetPollInterval(ws, raw)` validates and stores; `Settings(ws)` returns `{pollMinutes, lastCycleAt}`.
- [x] Leak test (`leak_test.go`): the API key, the token and bait bodies never appear in logs, errors, replies, task descriptions or candidates.
- [x] Run the tests and record the failing output.
- Stories: all U3 backend stories, plus the issue parts of US1.5, US1.7, US1.8 and US1.9. Contracts: C2, C3, C5. ADR-003, ADR-004, ADR-005. Findings: R-03, R-05. ACs: AC1.5.4, AC1.7.1, AC1.7.2, AC1.8.1–3, AC1.9.1–2 (issue parts), AC2.1.1, AC2.1.4, AC2.2.1–3, AC2.3.1–2, AC3.1.1–4, AC3.2.1–3, AC3.3.2, AC3.3.3, AC3.3.5, AC3.4.1–3, AC3.5.1–3, AC4.1.1–5, AC4.2.1, AC4.2.3, AC4.2.4, AC8.1.1, AC8.3.2. NFRs: NFR1, NFR2, NFR3, NFR5, NFR11.

### Step 7 — Business logic, Green and Refactor

- [x] `internal/issues/service.go` holds:
  - `Gateway`: a consumer-side interface over `Projects`, `Issues`, `IssueCount`, `Issue`, `IssueComments`, `IssueAttachments`, `ProjectStatuses` and `ProjectUsers`.
  - `Connection`: the same five methods U4 uses minus `GitCredential`.
  - `HostPort`: `CreateTask(ctx, NewTask) (TaskRef{ID, Key}, error)` and `ListTasks(ctx, ws) ([]TaskInfo{ID, Key, Title}, error)`.
  - The store.

  It provides list, filters, create, link, detail, comments, suggest/authorize, impact and settings.
- [x] `internal/issues/sync.go`: `Syncer` with `Start`/`Stop`/`Refresh` and one goroutine. `ponytail:` one `Issue` GET per link per cycle; switch to `id[]` batches when workspaces exceed about 200 links.
- [x] `internal/issues/events.go`: `Listen`, `OnConnectionChanged`, `OnTaskDeleted`, `ReconcileAll`.
- [x] Refactor while green.
- Contracts: C2, C3. NFRs: NFR2, NFR3, NFR5, NFR11.

### Step 8 — API / endpoint, Red: actions, reference RPCs, events, manifest, wiring

- [x] `internal/plugin/actions_u3_test.go`, through `HandleAction` with the fake Host, for these actions:
  - `issues.list`, `issues.filters`, `issues.create_task`, `issues.tasks.search`, `issues.links.list`, `issues.refresh`, `issues.impact`, `issues.settings.get` (workspace scope);
  - `issues.link`, `issues.unlink`, `issues.get`, `issues.comments` (task scope; the task id comes from the verified context, never the body);
  - `issues.set_poll_interval` (admin).

  Every action is refused with 409 `integration_disabled` while Backlog is off. Bad bodies → 400 `validation` with `field`. Unavailable or unlinked → 404 `not_found`. Duplicate or other-issue link → 409 `conflict`. 429 → `Retry-After` and `retryAfterSeconds`. A 500 writes exactly one `action_failed` line with no secret.
- [x] `internal/plugin/references_test.go`:
  - `Runtime` satisfies `pluginsdk.EntityReferenceHandler`.
  - `SearchEntityReferences` returns candidates only for the source `nulab-backlog-issues`; for an unknown source, the switch off or not connected it returns none and no error.
  - `AuthorizeEntityReference` allows only a valid live issue and fails closed on any error, with a fixed `Reason` and no secret.
- [x] `internal/plugin/events_test.go`: `OnEvent` with `task.deleted` (payload `task_id`, workspace from `Event.WorkspaceID`) removes the links. Other event types are accepted and ignored. A store failure returns an error, so Kandev retries.
- [x] `internal/plugin/host_port_u3_test.go`: the issue adapter maps `Tasks().Create` (returning `ID` and `Identifier`) and a paged `Tasks().List` with `IncludeArchived`. It reuses U4's paging and nesting helpers, so U4's `TestU4_HostPort_*` tests stay green. A missing Host gives `errNoHost`.
- [x] `internal/plugin/manifest_test.go`, U3 cases:
  - The 13 `issues.*` keys with the scope and access from Step 9 and `max_body_bytes: 16384`.
  - `capabilities.events: ["task.deleted"]`.
  - `reference_sources: [{source: nulab-backlog-issues, provider: nulab-backlog, kind: issue, display_name: "Backlog issues", kind_label: "Issue"}]`, matching Kandev's identity pattern.
- [x] `internal/plugin/runtime_u3_test.go`: `Runtime.Start` subscribes the issues service and starts the `Syncer`; `Close` stops it; `newRuntime` starts nothing. A real `connection.Service` switch turned off stops the cycle.
- [x] Run the tests and record the failing output.
- Stories: all U3. Contracts: C2, C5, C8. Findings: R-04, R-05. ACs: AC3.2.3, AC3.4.1–3, AC2.3.2, AC8.3.1 (U3 actions).

### Step 9 — API / endpoint, Green and Refactor

- [x] Add these files:
  - `internal/plugin/issue_actions.go`: handlers merged into `handlers` in `init()`, like `git_actions.go`, reading the verified task with `verified(ctx)`.
  - `internal/plugin/references.go`: the two RPCs.
  - `internal/plugin/events.go`: `OnEvent`.
  - Additions to `host_port.go`: a shared `create`/`eachTask` helper used by both the U4 and U3 adapters, with no second paging loop or metadata namespace.
- [x] Extend `runtime.go`:
  - Wire `issues.Service` and `Syncer` in `newRuntime`, and start and stop them in `Start`/`Close`.
  - Add three `classify` cases: `issues.ErrNotFound` → `not_found`; `issues.ErrConflict` and `issues.ErrStale` → `conflict`.
  - Add the `gateway` interface embedding `issues.Gateway`.
- [x] `manifest.yaml`:
  - Every `issues.*` action is workspace/authenticated except `issues.link`, `issues.unlink`, `issues.get` and `issues.comments` (task/authenticated) and `issues.set_poll_interval` (workspace/admin).
  - Every action has `max_body_bytes: 16384`.
  - Add `capabilities.events` and `reference_sources`.
- [x] Refactor while green.
- Contracts: C5, C8. NFRs: NFR3, NFR11.

### Step 10 — Frontend behaviour, Red: M2, M2m, M3, M6, M8, M1 interval, M12/settings counts, language

Every test uses the shared fake `host` (`ui/src/testing/harness.ts`). The harness is extended with `registerTaskPanel`, `registerTaskMenuAction`, `registerComponent`, `registerTranslations`, `i18n {locale, t}`, `useResponsiveBreakpoint`, `toast` and `ui.Skeleton`. Every test also runs `axe-core` with no violations, checks a `data-testid` on every interactive element, and checks that no literal text appears outside the catalogue (AC8.2.2).

- [x] `ui/src/issues/issues-state.test.ts`:
  - The list state machine: loading, success, empty, error with one of the three reasons, rate-limited with a countdown announced once (AC8.4.4), not-connected, no-project, and sign-in-again with a settings link (AC1.4.3).
  - The "Showing x-y of z" text; the badge label `PROJ-120 · Resolved`; the "not connected" and "may be out of date" texts.
  - Screen-reader text "PROJ-123, Fix login timeout, In Progress, assignee A, linked to T-12 and T-15".
  - An error code → notice mapping that never shows a server body.
- [x] `ui/src/issues/issues-page.test.tsx` (M2, M2m):
  - Loading shows `Skeleton` rows with the filters still visible and "loading" announced once (AC2.1.5).
  - Twenty rows with key, title, status, assignee and updated time; "Showing 1-20 of 57"; Next goes to page 3 with "41-57 of 57", Next disabled, focus at the top of the list and the line announced (AC2.1.1, AC2.1.2).
  - The project, status and assignee filters send `issues.list` bodies. The search waits 400 ms (`vi.useFakeTimers`) and sends one request.
  - Empty shows "No issues match these filters" with Reset filters (AC2.1.3, AC2.2.3). An error shows Retry (AC2.1.4). Not connected and no project show settings links (AC1.7.2).
  - A long Japanese title is clamped to 2 lines and the full text is in `title` and readable on focus (AC2.1.6).
  - Linked rows show task-key links plus the "…" menu with Create task and Link to task (AC2.3.1).
  - Create task shows "Creating…" and is disabled; double clicks make one call; it then shows "Created T-18" with a link and updates the row (AC3.1.2).
  - A linked row opens the three-choice dialog: Open T-17, Create another task, Cancel. Focus starts on Cancel, and Create another sends `force: true` (AC3.1.3). An error shows and no row changes (AC3.1.4).
  - Refresh calls `issues.refresh` then reloads, is disabled while running, and updates the "updated at" line (AC4.2.3).
  - With `isMobile`, the page shows cards with a "Filters (n)" drawer button (M2m).
- [x] `ui/src/issues/link-task-dialog.test.tsx` (M3):
  - The dialog has `role="dialog"` with an h2, and focus starts on the search field.
  - It searches with `issues.tasks.search`. A task linked to another issue shows "Linked to PROJ-118" and is `aria-disabled` (AC3.3.3).
  - Link calls `issues.link` with `taskId` and the row then shows T-17 (AC3.3.1, AC3.3.2).
  - No match shows "No tasks found" with Link disabled. Esc and Cancel change nothing and give focus back to that row's Link to task button (AC3.3.4, AC8.2.3).
- [x] `ui/src/issues/issue-badge.test.tsx` (M6, `task-card-tags`):
  - One `issues.links.list` call per workspace is shared by every card.
  - The badge shows text and is not shown by colour alone. Focus or tap shows "updated at …" (AC4.1.4). "may be out of date" appears when stale.
  - A not-connected link shows "PROJ-120 – not connected" and "Reconnect <host> to restore this link" (AC1.8.2, M12).
  - An unavailable link shows "Issue unavailable" (AC4.1.2). After unlink the badge disappears (AC3.3.5).
- [x] `ui/src/issues/task-menu.test.ts`: "Unlink Backlog issue" (`registerTaskMenuAction`, group `primary`) is visible only for a linked task, synchronously from the shared links store. It calls `issues.unlink` with `taskId` and refreshes the store (AC3.3.5).
- [x] `ui/src/issues/issue-panel.test.tsx` (M8, `registerTaskPanel`):
  - The panel shows key, title, status, assignee, priority, due date read in full, "updated at …" and Open in Backlog, all read-only (AC3.2.1).
  - Reopening refetches (AC3.2.2). An unavailable issue and a not-connected link have their own states.
  - Comments (150) show newest first with Load more (`issues.comments` with `maxId`) (AC3.5.1).
  - Attachments show name and size; one over the limit shows "too large to preview" with a link to Backlog (AC3.5.2).
  - A comments error shows Retry while the other sections still show (AC3.5.3).
  - Subsections are buttons with `aria-expanded`, and the expanded state persists per user in `localStorage` keyed by plugin.
- [x] `ui/src/issues/poll-interval.test.tsx` (M1): a labelled minutes field with default 5. Invalid values show "The minimum is 1 minute" under the field with `aria-describedby` (AC4.2.2). Save calls `issues.set_poll_interval` and shows "Saving…".
- [x] `ui/src/issues/i18n.test.ts` (US8.5):
  - `registerMessages(registry, host)` registers the `en` catalogue (converted to the `{{name}}` form only where needed).
  - `messagesFor(host)` returns `host.i18n.t(key, {defaultValue: en[key]})` for every key.
  - A pseudo-locale catalogue shows its text (AC8.5.1). An unknown locale shows English, never a raw key (AC8.5.2).
- [x] Update `ui/src/settings/connected-panel.test.tsx` and `project-picker.test.tsx` only for these changes:
  - the poll-interval block is mounted;
  - the disconnect and change-space dialogs read "N issue links, M PR links and K PR watches will be turned off", summing `issues.impact` and `git.impact` (AC1.8.1);
  - the deselect confirmation shows the issue-link count for the unselected projects (AC1.9.1).

  Update `ui/src/page/backlog-page.test.tsx` because `/backlog` renders the issues page once connected. Update `ui/src/index.test.ts` for the new registrations.
- [x] Run the tests and record the failing output.
- Stories: US2.1, US2.2, US2.3, US3.1, US3.2, US3.3, US3.5, US4.1, US4.2, US8.2, US8.5, plus the issue parts of US1.4, US1.7, US1.8 and US1.9. ACs: AC1.4.3, AC1.7.2, AC1.8.1, AC1.8.2, AC1.9.1, AC2.1.1–6, AC2.2.3, AC2.3.1, AC3.1.2–4, AC3.2.1–2, AC3.3.1–5, AC3.5.1–3, AC4.1.2, AC4.1.4, AC4.2.2–3, AC8.2.2–3, AC8.4.4, AC8.5.1–2. NFRs: NFR1, NFR9, NFR10.

### Step 11 — Frontend behaviour, Green and Refactor

- [x] Add kebab-case files in `ui/src/issues/`: `issues-state.ts`, `issues-page.tsx`, `link-task-dialog.tsx`, `links-store.ts`, `issue-badge.tsx`, `task-menu.ts`, `issue-panel.tsx`, `poll-interval.tsx`, `i18n.ts`.
- [x] `ui/src/index.ts`:
  - call `registerMessages` first;
  - `/backlog` renders `createIssuesPage(host)`, which reuses U1's `pageState` and `settingsHref` for the not-connected states;
  - `registerComponent("task-card-tags", IssueBadge)`, `registerTaskMenuAction(unlink)` and `registerTaskPanel(issuePanel)`;
  - nav labels come through `host.i18n.t`.
- [x] Mount `poll-interval.tsx` in `connected-panel.tsx`. Add the issue counts to the existing confirm dialogs. Add the message keys to `messages/en.ts`. Reuse U2's `confirm-dialog.tsx` for the three-choice dialog: "Open T-17" is a link in its body, confirm = "Create another task", and focus starts on Cancel.
- [x] Use only host UI kit components, the U1 spacing classes and native inputs, as U2 and U4 did.
- [x] Refactor while green.
- Rules: U1 BR6.1–BR6.5. NFRs: NFR9, NFR10.

### Step 12 — Environment and build configuration

- [x] No `Makefile` or CI change: `GO_PKGS` already covers `./internal/...`, and Vitest runs every file.
- [x] From `make clean`, run `make check-format vet lint test coverage check-secrets build package verify-package`. Go coverage must be ≥ 80% with only `server/main.go` excluded, and `internal/issues` ≥ 85%. `go mod tidy` must leave no diff.
- [x] Confirm that the packaged `manifest.yaml` carries the `issues.*` actions, `events` and `reference_sources`, and that `verifypkg` reports OK.
- [x] Run `make contract-test KANDEV_MIN_DIR=../kandev-min` and confirm that Kandev v0.96.0 accepts the manifest (reference source, `task.deleted` subscription) and starts the plugin.
- Stories: US7.3 (the gate applies), US8.1. NFRs: NFR6, NFR8.

### Step 13 — Documentation and traceability

- [x] `README.md`: append a "Backlog issues" section covering the Issues page, Create task, Link to task, `#` references, the Backlog panel in the task (opened from the "+" panel menu), the badge, the sync interval (default 5 minutes, minimum 1) and that the plugin never writes to Backlog. Only append.
- [x] `docs/manual-checks/TEMPLATE.md`: add the B4 demo steps:
  1. Search `PROJ-12` and create a task.
  2. See the badge.
  3. Change the status on Backlog and see the badge update within one cycle.
  4. Delete the task and see the link disappear.
  5. Measure the 20-row list 20 times (AC8.1.2 `[manual]`).
  6. Do the keyboard and screen-reader pass and the 320 px check (AC8.2.1, AC8.2.4 `[manual]`).
- [x] Write `construction/issues/code-generation/code-summary.md`, `source-manifest.json` and `traceability.json`, including the upstream amendments listed under Assumptions.
- Stories: all U3.

## Story-to-Step Map

| Story / AC part | Steps |
|-----------------|-------|
| US2.1 View the issue list | 1–7, 8–11 |
| US2.2 Filter and search issues | 2–7, 8–11 |
| US3.1 Create a task from an issue | 2, 3, 6–11 |
| US3.3 Link and unlink an issue | 4–11 |
| US2.3 Linked tasks in the list | 4–7, 10, 11 |
| US3.2 Issue information in the task | 2–11 |
| US3.4 `#` references | 2–9 |
| US4.1 Status sync cycle | 2–11 |
| US4.2 Interval and manual refresh | 2–11 |
| US8.1 Lists within 3 seconds | 4–7, 12, 13 |
| US8.2 Keyboard and screen reader | 10, 11, 13 |
| US8.5 Text follows the Kandev language | 10, 11 |
| US3.5 Comments and attachments | 2–11 |
| US1.5 / US1.8 / US1.9 (issue-link parts: AC1.5.4, AC1.8.1–3, AC1.9.1–2) | 6, 7, 10, 11 |
| US1.7 AC1.7.1, AC1.7.2 | 6, 7, 10, 11 |
| US1.4 AC1.4.3 (settings link from the Issues page) | 10, 11 |
| US8.3 AC8.3.2 (issue cycle log) | 6, 7 |
| US8.4 AC8.4.4 (Issues page) | 10, 11 |
| US7.3 AC7.3.2 (list error flow) | 4, 5, 12 |

**Shared files touched:**

- Go:
  - `manifest.yaml`: the `issues.*` actions, `capabilities.events`, `reference_sources`.
  - `internal/plugin/runtime.go`: wiring, `Start`/`Close`, three `classify` cases, the `gateway` interface.
  - `internal/plugin/host_port.go`: a shared helper extracted; U4 behaviour unchanged.
  - `internal/plugin/manifest_test.go`: U3 cases.
  - `internal/backlog/pullrequests.go` and `client.go`: `Issue` moved to `issues.go`, signature unchanged.
- UI:
  - `ui/src/index.ts`, `ui/src/messages/en.ts`, `ui/src/testing/harness.ts`;
  - `ui/src/page/BacklogPage.tsx` and its test (the `/backlog` body);
  - `ui/src/settings/{connected-panel.tsx,project-picker.tsx,SettingsScreen.tsx}` and their tests (interval mount and issue counts);
  - `ui/src/index.test.ts`.
- Docs: `README.md` (one appended section), `docs/manual-checks/TEMPLATE.md`.
- Not touched: `Makefile`, `.github/workflows/*`, `.golangci.yml`, `go.mod`, `go.sum`, `ui/package.json`, `internal/ci`, `cmd/`, `internal/connection/*`, `internal/git/*`.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `unit-of-work.md`, `unit-of-work-story-map.md`, `unit-of-work-dependency.md`; `stories.md`; `requirements.md`; `components.md`, `decisions.md`; `contract-summary.md` and `contract-design/reviews/review-01.md` (R-03, R-04, R-05); `mockups.md`, `interaction-spec.md`, `design-system-mapping.md`, `accessibility-checklist.md`; `bolt-plan.md` (B4), `external-dependency-map.md` (X7).
- U2 and U4: `code-generation-plan.md`, `code-summary.md`, and the code on `main` at `c3ca400`.
- `team.md`, `project.md`, `phases/construction.md`; `.claude/aidlc-common/stages/construction/code-generation.md` Step 2.
- Kandev v0.96.0: `apps/backend/pkg/pluginsdk/{plugin.go,host.go,types.go,data_types.go}`; `apps/backend/internal/plugins/{mentions.go,host_data.go,host_write.go}`; `apps/backend/internal/plugins/manifest/{manifest.go,validate.go}`; `apps/backend/internal/mentions/{types.go,service.go}`; `apps/backend/internal/events/types.go`; `apps/packages/plugin-sdk/src/index.ts`; `docs/public/plugins-manifest.md`; `docs/public/plugins-authoring.md`.
- Backlog (read 2026-10-06): <https://developer.nulab.com/docs/backlog/api/2/get-issue-list/>, <https://developer.nulab.com/docs/backlog/api/2/count-issue/>, <https://developer.nulab.com/docs/backlog/api/2/get-comment-list/>, <https://developer.nulab.com/docs/backlog/api/2/get-list-of-issue-attachments/>, <https://developer.nulab.com/docs/backlog/api/2/get-issue-attachment/>, <https://developer.nulab.com/docs/backlog/api/2/get-status-list-of-project/>, <https://developer.nulab.com/docs/backlog/api/2/get-project-user-list/>, <https://developer.nulab.com/docs/backlog/api/2/get-priority-list/>.

## Assumptions & Open Questions

- [assumption] There is no U3 functional, NFR or infrastructure design (user choice). The bolt plan wanted R-03, R-04 and R-05 settled in functional design; this plan settles them in the assumptions below, and they need review at plan approval.
- [assumption] The IssueIntegration package is `internal/issues`. It declares consumer-side interfaces (`Gateway`, `Connection`, `HostPort`) because tests swap in fakes. It imports neither `pluginsdk` nor `internal/git`.
- [assumption] (labels, X7) No real Kandev labels are written: `Task.Labels` is deprecated in v0.96.0. The issue badge is the `task-card-tags` slot, drawn from `issues.links` (ADR-004). C2's `SetTaskLabels` and `TaskExists` are dropped, and X7 is not needed.
- [assumption] (R-05) Deleted tasks are handled two ways: `OnEvent(task.deleted)` removes the links, and each cycle lists the workspace's tasks once (`Tasks().List`, archived included) and drops links to missing tasks. `Tasks().Get`'s gRPC `NotFound` is not used, so `go.mod` stays unchanged. `ponytail:` one paged task list per workspace per cycle.
- [assumption] The `HostPort` for U3 is a second small adapter in `internal/plugin/host_port.go` that reuses U4's paging and the `"plugin:nulab-backlog"` metadata nesting through one shared helper. U3 writes no task metadata and needs no repository names.
- [assumption] Task from issue (AC3.1.1):
  - title = issue summary;
  - description = issue description plus `Backlog: https://<host>/view/<KEY>`;
  - priority by Backlog priority id: 2 → high, 3 → medium, 4 → low, other → medium;
  - workflow and step from `getTaskCreationContext`, sent in the body;
  - no repository;
  - the "label PROJ-120" is the plugin badge.

  Comments and attachments are not copied (Q9 overrides FR3.2's copy).
- [assumption] (budget) `issues.create_task` makes one Backlog call and one Kandev call, well under 15 s. A link write that fails after the task was created returns `internal`, and the task stays unlinked so the user can link it by hand. No automatic relink.
- [assumption] AC3.1.3: the server refuses a second task with `conflict` unless the body has `force: true`, and the UI shows the three-choice dialog before asking. No new error field is added to `actionError`.
- [assumption] M3 is a plugin dialog that lists at most 20 workspace tasks through `issues.tasks.search` (a paged host scan filtered by title or task key). A link from the task side (`openTaskLinkDialog`) is not built because no story needs it.
- [assumption] (R-04) The action schemas are defined by Steps 8–9: 13 `issues.*` keys in snake_case, following U2. `issues.set_poll_interval` is admin, matching R-01, because the interval is shared by the workspace.
- [assumption] The poll interval lives in IssueIntegration's own state (`issues.settings`), not in the Connection record. A Connection write would raise the epoch and emit `ConnectionChanged`, dropping in-flight results. Contract C3's `Snapshot.PollMinutes` and C5's `connection.setPollInterval` therefore move to `issues.*`. This needs an upstream amendment.
- [assumption] The scheduler is one worker that ticks every minute (the minimum interval) and runs each workspace when `lastCycleAt + pollMinutes` has passed. `lastCycleAt` is stored, so a restart continues the schedule (AC4.1.5). `issues.refresh` runs a cycle on the same worker and waits up to 10 s, so two polls never overlap.
- [assumption] The sync cycle makes one `Issue` GET per active link (Read group, Background), so 404 and 403 are told apart per link (AC4.1.2). A link is "may be out of date" after 3 failed checks in a row.
- [assumption] (R-06 for U3) An issue 404 or 403 means "Issue unavailable": `not_found` for `issues.get`, and a link flag in the cycle. It never means `reconnect_required`. A list-level 401 or 403 still maps to `reconnect_required`, as in U2.
- [assumption] (R-03) The plugin-side list budget is 2.5 s (AC8.1.1). It is met with `Issues` and `IssueCount` (Search, 1 s apart), plus one `Issue` GET for an exact key, and project ids cached per (workspace, epoch) from `Projects`. Background sync uses only Read calls, so it never holds the Search queue that interactive calls use. No priority lane is added to the gateway.
- [assumption] The exact-key row (AC2.2.2) is added only on page 1 and only when it passes the active filters. `total` stays Backlog's count.
- [assumption] The status filter shows each selected project's statuses, deduplicated by id. The assignee filter is the union of the selected projects' users (`ProjectUsers` is a new C1 call).
- [assumption] (M10) `#` uses `reference_sources` with source `nulab-backlog-issues`, provider `nulab-backlog` and kind `issue`.
  - N = Kandev's `Limit` (default 5, at most 10).
  - Each search and authorize call has a 1.2 s budget, inside Kandev's 1.5 s.
  - AC3.4.2 is met by a per-workspace 250 ms "latest query wins" wait in the plugin, because Kandev's composer debounce is not documented.
  - Authorize allows only a live `Issue` 200 in a selected project, as the docs require. Under rate limiting this fails closed, so the reference is dropped at submission.
  - It is unverified that Backlog's `keyword` matches issue keys; the exact-key GET covers full keys.
- [assumption] (M8) The Backlog block is a task panel (`registerTaskPanel`), because the `task-sidebar` slot gets no task id. The panel opens from the task's "+" menu instead of sitting in the sidebar, as Q2 at refined mockups asked. The expanded state is kept in `localStorage`; `host.storage` would need the `user_state` capability.
- [assumption] (AC3.5.2) The preview limit N is 10 MiB. The plugin does not proxy files. Each attachment links to the Backlog issue page; files over N add the "too large to preview" reason. Comments load 20 per page (Backlog default; at most 100, AC3.5.1).
- [assumption] (US8.5) The `en` catalogue is registered with `registerTranslations`, and screens get their strings through `host.i18n.t` with the English text as fallback. The first release ships English only (NFR10 open question), so AC8.5.1 is tested with a pseudo-locale. Strings are resolved when a screen is created; a locale change applies after a reload.
- [assumption] AC1.8.1 and AC1.9.1: U3 adds `issues.impact`, and the U2 dialogs show the sum with U4's `git.impact`. The M12 restore notice gets no restored counts, as in U4.
- [assumption] `ConnectionChanged` handling copies U4's rules for issue links: disconnect, other host or deselected project → `not_connected`; a restore with a matching host and project → active; a lower epoch is ignored; startup and every cycle reconcile with `Current()`.
- [assumption] Interactive suggestion and list calls never send a Backlog body or a secret to the UI. New Go test secrets come from `testutil.APIKey`/`testutil.Token` and are generated at run time, never committed.
- [assumption] `/backlog` becomes the issue list (U1 comment: "U3 adds the issue list"); U1's not-connected, off and incomplete states are kept.
- Upstream amendments needed:
  - C1: `Issues`, `IssueCount`, `Issue` (wider), `IssueComments(…, CommentQuery)`, `IssueAttachments`, `ProjectStatuses`, `ProjectUsers`.
  - C2: `HostPort` = `CreateTask` → `{ID, Key}`, `ListTasks`; `SetTaskLabels` and `TaskExists` removed.
  - C3: `PollMinutes` leaves `Snapshot`.
  - C5: the `issues.*` table; `connection.setPollInterval` → `issues.set_poll_interval`.
  - C8: `events: [task.deleted]`, `reference_sources`, `task-card-tags`, `registerTaskPanel`, `registerTaskMenuAction`.
- Open: AC8.1.2, AC8.2.1 and AC8.2.4 are `[manual]` and are covered by the B4 demo in `TEMPLATE.md`.
- Open (user decision at plan approval; approving the plan accepts the defaults above): English only for the first release; M8 as a task panel; the 10 MiB preview limit with attachments linking to Backlog (no in-plugin preview); `connection.setPollInterval` renamed to `issues.set_poll_interval` (admin).
