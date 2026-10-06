# Code Generation Plan — git-pr (U4)

Inputs:

- `unit-of-work`: the U4 boundary — the whole GitIntegration component, Git credential storage (US5.5, stored by Connection but built with this unit), the repository source, linking and creating pull requests, the PR badge, PR watches and saved queries, screens M4, M5, M7, M9 and M11, and U4's own timer (ADR-003). U4 receives `ConnectionChanged` from U2. Also `unit-of-work-story-map` and `unit-of-work-dependency` (git-pr depends on connection only).
- `stories`: US5.5, US5.1, US5.6, US5.2, US5.4, US5.3, US6.1, US6.2, US6.3; the Git, PR-link and PR-watch parts of AC1.5.4, AC1.8.1, AC1.8.2, AC1.9.1 and AC1.9.2 that U2 deferred; AC1.7.1; AC7.3.2 (Git error flow).
- `requirements`: FR1.6, FR1.7, FR5.1–FR5.4, FR6.1–FR6.3, NFR2, NFR3, NFR4, NFR5, NFR8, NFR9, NFR10, NFR11; assumption A1 (pull requests exist on the plan).
- `components` (GitIntegration, Connection, BacklogGateway, KandevAdapter, PluginUI) and `decisions` (ADR-002, ADR-003, ADR-004, ADR-005, ADR-007).
- `contract-summary`: C1, C2 (`HostPort`, `ResolveGitCredential`), C3 (`ConnectionReader`, `GitCredential`, `ConnectionChanged`), C5 (`connection.setGitCredential`, `git.*`), C7 (Backlog git endpoints), C8 (`repository_providers`, review provider); contract-review findings R-04, R-05, R-06, R-07, R-09.
- `mockups` (M1 Git block, M4, M5, M6, M7, M9, M11, M12), `interaction-spec`, `accessibility-checklist`.
- `bolt-plan` (B5 Definition of Done) and `external-dependency-map` (X6).
- U2 `code-generation-plan.md` and `code-summary.md`; U1/U2/U5 code in the working tree on top of `main` at `99760f8`.
- `team.md`, `project.md`, `phases/construction.md`.

U4 has no functional, NFR or infrastructure design; the user chose to go straight to code generation. Every design choice in this plan is listed under Assumptions & Open Questions.

Stories in U4, in build order: US5.5, US5.1, US5.6, US5.2, US5.4, US5.3, US6.1, US6.2, US6.3.

## Kandev SDK Facts Used by This Plan

Read from `../kandev` at v0.96.0 (`apps/backend/pkg/pluginsdk`, `apps/packages/plugin-sdk/src/index.ts`, `docs/public/plugins-manifest.md`, `docs/public/plugins-authoring.md`, `apps/backend/internal/plugins/repository_provider.go`, `apps/backend/internal/backendapp/git_credentials.go`).

- **Repository provider ownership.** The manifest field `repository_providers: [<id>]` lists provider ids this plugin owns (canonical lowercase; an id owned by another plugin is rejected). The bundle registers `registry.registerRepositoryProvider({id, label, icon, listRepositories, listBranches, inspectURL, matchesURL?, supportsDraft?, createChangeRequest?})`. `listRepositories` receives `{workspaceId, query?, cursor?, limit?, signal}` and returns `RepositoryInspection[]` or `{repositories, nextCursor?}`; a `RepositoryInspection` has `providerId`, `providerHost`, `providerScope?`, `ownerOrProject`, `repositoryId`, `repositoryName`, `cloneUrl`, `defaultBranch?`; Kandev overwrites `providerId`.
- **Backend actions with fixed names.** `repositories.inspect` (workspace scope; body `{url}`; reply `{repository: {provider_id, provider_host, provider_scope, provider_repository_id, owner_or_project, name, clone_url, default_branch}}` or `{"matched": false}`; clone URL HTTPS, no credentials, same origin as `provider_host`). `repositories.branches` (workspace scope; `body.repository` holds the stored snake_case descriptor; reply `{"branches": [{"name", "commit"?, "is_default"?}]}`, ≤ 10,000 entries, ≤ 1 MiB).
- **Git credentials (Go).** Optional interfaces `pluginsdk.GitCredentialResolver` and `pluginsdk.GitCredentialBinder` (`GitCredentialHandler`): `ResolveGitCredential(ctx, *ResolveGitCredentialRequest{ProviderID, WorkspaceID, TaskID, SessionID, RepositoryID, Host, Path}) (*ResolveGitCredentialResponse{Username, Secret, ExpiresAt}, error)` and `GetGitCredentialBinding(ctx, *GitCredentialBindingRequest{…}) (*GitCredentialBindingResponse{Binding}, error)`. Empty username/secret or empty binding = revoked; `ExpiresAt` RFC 3339 or empty, past values rejected; leases revoked when the binding changes. Docs: "plugins should reject missing task, session, or repository identity". The host derives `Host`/`Path` from the clone URL and checks origin against `provider_host`.
- **Linking a PR from the task menu.** `registry.registerTaskAction({id, label, icon?, placement: "link", visible?, singleTaskOnly?, run(TaskContext)})` and `host.openTaskLinkDialog({title, description, inputLabel, placeholder?, emptyError, failureMessage, successMessage, inputTestId?, errorTestId?, submitTestId?, onSubmit(reference, signal)})`. Kandev owns the Link submenu, dialog, validation display, submit state and toast.
- **PR badge, status and unlink.** `registry.registerReviewProvider({id, label, changeRequestNoun, order, getSnapshot(taskId), subscribe, refresh(taskId, signal), getAssociationSnapshot?(workspaceId), subscribeAssociations?, refreshAssociations?, unlink?(ctx), ReviewPanel, ...})`. `ReviewSummary` = `providerId, reviewKey, title, url, connectionScope, repositoryId, changeRequestNumber, state, statusBadge?{label, tone?}, taskStatus?{number, state: open|merged|closed|draft, pipelineState, checks[], review?, error?, updatedAt?}`. `ReviewTaskAssociation` = `providerId, taskId, reviewKey, connectionScope, repositoryId, changeRequestNumber`. Kandev renders the card/sidebar/topbar indicators and refreshes active status every 90 s; plugins must not run their own PR status poller.
- **Creating a PR.** `createChangeRequest({workspaceId, taskId, sessionId, repositoryId, repository, title, body, baseBranch?, draft, signal})` → `{url, provider?, output?, linked?, associationError?}`. Kandev keeps the native Create PR dialog and push-before-create flow, invoking the provider only after a push. The provider forwards `sessionId` as a task-action selector; the backend uses `VerifiedActionContext.HeadBranch`, never a browser branch. `supportsDraft: false`.
- **Verified action context.** `VerifiedActionContext{ActorID, WorkspaceID, TaskID, RepositoryID, SessionID, HeadBranch}`. A task-scoped action may name a repository attached to the task and a session of that task; with both, Kandev derives `HeadBranch`. 15 s per action; reply ≤ 1 MiB; allowed headers `Content-Type`, `Cache-Control`, `ETag`, `Retry-After`; status 200–599.
- **Host data API.** `Host.Tasks().Create(ctx, CreateTaskInput{WorkspaceID, WorkflowID, WorkflowStepID*, Title, Description, Priority, Repositories, Metadata, ...})` needs `api_write: [tasks]`; `Host.Tasks().List(ctx, TaskFilter{WorkspaceIDs, IncludeArchived, ...}, Page)` needs `api_read: [tasks]` and returns `Task.Metadata`; `Host.Repositories().List(ctx, workspaceID, Page)` returns `Repository{ID, ProviderID, ProviderRepositoryID, ProviderHost, ProviderScope, OwnerOrProject, ProviderName, RemoteURL, DefaultBranch*}` and needs `api_read: [repositories]`. Browser: `host.context.getTaskCreationContext(workspaceId)` → `{workflowId, defaultStepId, steps, repositories}`.
- **Timers.** No host timer service; a plugin runs its own goroutines (C8, ADR-003). State has no transactions; event delivery is sequential per plugin.
- **UI kit and slots.** `host.ui` includes Dialog*, Select*, Checkbox, Input, Label, Button, Table*, Badge, Spinner, Alert*, plus `ChangeRequestDetail`, `IntegrationSaveQueryDialog`, `IntegrationRepositoryFilter`, `IntegrationCursorPagination`; `host.openModal({title?, description?, content, size?, dismissible?})`; `registerRoute` and `registerNavItem(section: "integrations")` (used by U1).
- **Capabilities.** `state: true` and `secrets: true` already declared. U4 adds `api_read: ["tasks", "repositories"]` and `api_write: ["tasks"]`. `min_kandev_version` is already 0.96.0.
- **Dependencies.** Standard library only (`net/http`, `encoding/base64`, `sync`, `time`, `testing/synctest`). No Go module or npm package added.

### Backlog API v2 facts (developer.nulab.com, read 2026-10-06)

- **Repository list.** `GET /api/v2/projects/:projectIdOrKey/git/repositories` → `id, projectId, name, description, hookUrl, httpUrl, sshUrl, displayOrder, pushedAt, createdUser, created, updatedUser, updated`. There is no default-branch field.
- **Pull request list.** `GET /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName/pullRequests` with `statusId[]`, `assigneeId[]`, `issueId[]`, `createdUserId[]`, `offset`, `count` (1–100, default 20). A PR has `id, projectId, repositoryId, number, summary, description, base, branch, status{id,name}, assignee, issue{id}, createdUser, created, updated, mergeAt, closeAt`. The docs example shows only status 1 = "Open".
- **Add pull request.** `POST` to the same path, `application/x-www-form-urlencoded`. Required `summary`, `description`, `base`, `branch`; optional `issueId`, `assigneeId`, `notifiedUserId[]`, `attachmentId[]`. Reply is the created PR.
- **Related endpoints.** Get Git Repository, Get Number of Pull Requests, Get Pull Request (`…/pullRequests/:number`, per C7).
- **Rate-limit groups** (U2): PR and repository GETs are Read; Add Pull Request is Update.

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

Each TDD layer goes Red → Green → Refactor. Each Red step records its failing command output in `construction/git-pr/code-generation/code-summary.md`.

**Test naming rule.** Every new Go test function in U4 starts with `TestU4_`, so the unit filter is `-run '^TestU4_'`. No existing U1, U2 or U5 test starts with `TestU4_`. U5's `check-secrets` treats a run of 32 or more mixed-case letters and digits as a credential, so every underscore-separated part of a test name stays under 32 characters (for example `TestU4_Watch_CreatesAtMostTenPerCycle`). Vitest files for U4 live in `ui/src/git/` and are listed by exact path.

### Step 1 — Project structure and configuration skeleton

- [x] Add the GitIntegration package `internal/git` with only `doc.go`. It does not import `pluginsdk`; only `internal/plugin` imports `pluginsdk`.
- [x] Add fake-only fixtures in `internal/backlog/testdata/`: `repositories_ok.json` (`web-app`, `api`, `httpUrl` on `example-space.backlog.com`), `repositories_empty.json`, `pullrequests_ok.json` (3 PRs in statuses 1, 2, 3), `pullrequest_ok.json`, `pullrequest_created.json`, `issue_ok.json` (`PROJ-120` with a numeric id). Lists of 25 PRs are built inside tests. Fixtures hold no secret and pass `make check-secrets`.
- [x] Every Git password in tests comes from `testutil.Token(t)`; `testutil` needs no change.
- [x] Do not touch `Makefile`, `.github/workflows/*`, `internal/ci`, `cmd/`, `go.mod`, `go.sum` or `ui/package.json`.
- Stories: all U4 stories (fixtures). NFRs: NFR4.

### Step 2 — Test runner readiness (already in place from U1)

- [x] Confirm before the first Red step: from the repository root, `go test -race ./internal/backlog/... ./internal/connection/... ./internal/git/... ./internal/plugin/... -run '^TestU4_'` (each package `ok … [no tests to run]`, `internal/git` `[no test files]`); from `ui/`, `npx vitest run src/git/git-state.test.ts src/git/git-access.test.tsx src/git/repository-provider.test.ts src/git/pr-link.test.ts src/git/review-provider.test.tsx src/git/create-pr.test.ts src/git/watches-page.test.tsx src/git/watch-form.test.tsx src/git/dashboard-page.test.tsx --passWithNoTests`.
- [x] Confirm the U2 regression commands from U2 `unit-test-instructions.md` are green.
- Stories: all.

### Step 3 — Data model, Red: Backlog git types, GitIntegration entities, Git secret

- [x] `internal/backlog/repositories_types_test.go`: decoding `Repository{ID, ProjectID, Name, HTTPURL}`; a non-numeric id, an empty name, or an `httpUrl` that is not https on the requested host is rejected as `Unreachable`/`body`.
- [x] `internal/backlog/pullrequests_types_test.go`: decoding `PullRequest{Number, Summary, Base, Branch, StatusID, AssigneeName, IssueID, Created}`; `PullRequestState` 1 → open, 2 → closed, 3 → merged, other → unknown; `PullRequestQuery.Values()` encodes `statusId[]`, `assigneeId[]`, `issueId[]`, `createdUserId[]`, `offset`, `count` clamped to 1–100; `NewPullRequest.Form()` always sends `summary`, `description`, `base`, `branch`, and `issueId` only when set; `%v` of a `PullRequest` shows no description body.
- [x] `internal/git/types_test.go` (table-driven): `ParseReference` accepts `https://<space>/git/<PROJ>/<repo>/pullRequests/<n>`, `<repo>#<n>`, and `<n>` when the task has exactly one Backlog repository; rejects another host, a non-PR URL, `0`, negatives, letters, over 9 digits (field `reference`). `ParseClonePath` accepts `/git/<PROJ>/<repo>.git` and `/git/<PROJ>/<repo>`. `LinkKey`/`LedgerKey` = spaceHost|repositoryId|number. `Watch.Validate` (name 1–100 after trim; repository in a selected project, AC6.1.2; statuses open/closed/merged; assignee and creator `anyone`/`me`; optional linked issue key). `Query.Validate`. `RelatedIssueKey(title, body, selected)` = first `[A-Z][A-Z0-9_]*-\d+` whose project is selected. `PRDescription(body, key)`: non-empty body kept; empty body → `Related: <KEY>` or the title; never contains `close`/`fix`/`resolve` followed by the key (AC5.3.1). `BranchCandidates(prs)`: de-duplicated `base`/`branch` names, most used first.
- [x] `internal/connection/git_credential_types_test.go`: `ValidateGitCredential` (username 1–100, no control characters; password 1–256; fields `gitUsername`, `gitPassword`); `gitSecret{Username, Password, SpaceHost, Revision}` JSON round-trip; `GitCredential` hides the password under `%v` and `%#v`.
- [x] Run the tests and record the failing output.
- Stories: US5.5, US5.1, US5.2, US5.3, US5.4, US6.1, US6.3. Contracts: C1, C3, C5 (R-04). ACs: AC5.2.2, AC5.3.1, AC5.4.1, AC6.1.1, AC6.1.2.

### Step 4 — Data model, Green and Refactor

- [x] Implement `internal/backlog/repositories.go` and `pullrequests.go`, `internal/git/types.go`, and `internal/connection/git_credential.go` (types and validation only).
- [x] Refactor while green.
- Contracts: C1, C3, C5. NFRs: NFR3.

### Step 5 — Repository / data access, Red: Git secret storage and GitIntegration state

- [x] `internal/connection/store_test.go`, U4 cases: `SaveGit` writes secret key `backlog.git.<ws>` bound to the connected `spaceHost` with `Revision` = previous + 1; `Load` sets `hasGitCredential: true` only when the Git secret's host equals the connected host; the password never appears in the view; `Disconnect` deletes the Git secret right after the connection secret and before the disconnect record (AC1.5.4), and a failed delete writes no record and returns `ErrStore`; connecting to a different host deletes the Git secret before the new connection is written (AC1.8.2); same-host replacement, project saves and token refresh keep it.
- [x] `internal/git/store_test.go`: workspace state docs `git.links`, `git.watches`, `git.ledger`, `git.queries` round-trip with `schemaVersion: 1`; an unknown `schemaVersion` is never overwritten (`ErrStore`); instance key `git.watch_index` lists workspaces that have watches; caps 500 links, 50 watches, 50 queries (`validation`); 1-second store limit; workspaces independent.
- [x] Run the tests and record the failing output.
- Stories: US5.5, US5.2, US6.1, US6.2, US6.3, US1.5/US1.8 (Git parts). ADR-004, ADR-005. ACs: AC5.5.1, AC1.5.4 (Git), AC1.8.2 (Git).

### Step 6 — Repository / data access, Green and Refactor

- [x] `internal/connection/store.go`: `SaveGit`, `readGit`, `deleteGit`; call `deleteGit` in `Disconnect` and in `saveConnection` on a host change; `Load` reads the Git secret only to set `HasGitCredential`.
- [x] `internal/git/store.go`: one get/put per doc using the U2 store-call pattern; one per-workspace `sync.Mutex` for read-modify-write.
- [x] Refactor while green.
- Contracts: C3. NFRs: NFR3, NFR5.

### Step 7 — Business logic, Red: BacklogGateway git calls

Tests use the `httptest` fake Backlog and injected `Now`/`Wait` under `testing/synctest`.

- [x] `internal/backlog/git_client_test.go`: `Repositories(ctx, creds, projectKey)` (path, credential headers, Read group not queued, 401/403/404/500 with a bait body that never appears in errors or logs); `PullRequests(ctx, creds, class, projectKey, repo, query)` (exact query string; `Background` waits the full 429 time; `Interactive` keeps the 3 s budget); `PullRequest(ctx, creds, class, projectKey, repo, number)` (404 → `NotFound`); `CreatePullRequest(ctx, creds, projectKey, repo, NewPullRequest)` (exact form body; Update group queued and spaced 1 s from a concurrent token refresh, `-race`; 400 → `Invalid`); `Issue(ctx, creds, class, issueKey)` (C1 signature; 404 → `NotFound`); `CheckGitAccess(ctx, spaceHost, username, password, projectKey, repo)` (`GET /git/<PROJ>/<repo>.git/info/refs?service=git-upload-pack` with `Authorization: Basic …` over https; 200 → nil; 401/403 → `Unauthorized`; password and Basic header value redacted from logs and errors, AC5.5.2, AC7.3.2).
- [x] Run the tests and record the failing output.
- Stories: US5.1, US5.2, US5.3, US5.4, US5.5, US6.2, US6.3. Contracts: C1, C7. NFRs: NFR2, NFR3, NFR5, NFR11. ACs: AC5.5.2, AC7.3.2 (Git), AC7.3.3.

### Step 8 — Business logic, Green and Refactor: BacklogGateway

- [x] Add the six calls to `internal/backlog/client.go` through the existing `send` path, with `Background` where the call class is given. `CheckGitAccess` uses a `request` variant carrying a Basic header and its `secrets`; no second HTTP path.
- [x] Refactor while green. `group.go` unchanged.
- Contracts: C1, C7. NFRs: NFR2, NFR3.

### Step 9 — Business logic, Red: Connection Git credentials (built with U4)

- [x] `internal/connection/git_credential_test.go`: `SetGitCredential(ctx, ws, in)` validates, requires a connection (`ErrNotConnected` → `reconnect_required`), stores under the workspace write lock, returns the view with `hasGitCredential: true` and no password (AC5.5.1). `GitCredential(ctx, ws)` returns `(GitCredential, binding, error)`: `ErrNoGitCredential` when none is stored or the host differs; `binding` = `"<connectionEpoch>.<revision>"`, changing on save, replacement, space change and disconnect. `Test(ctx, ws)` with a stored Git credential calls `Repositories` for the first selected project then `CheckGitAccess` on its first repository; the view gets `gitCheck: ok | invalid | untested`; a Git 401 does not fail the connection test (AC5.5.2). After `Disconnect` there is no API key, token or Git password and exactly one `disconnected` event (AC1.5.4); after a space change the old Git password is gone (AC1.8.2). Leak test: no 8-character window of the Git password in logs, errors, views or events.
- [x] Run the tests and record the failing output.
- Stories: US5.5, US1.5, US1.8. Contracts: C3, C5. Finding: R-09. ACs: AC5.5.1, AC5.5.2, AC1.5.4 (Git), AC1.8.2 (Git).

### Step 10 — Business logic, Green and Refactor: Connection

- [x] Implement `Service.SetGitCredential`, `Service.GitCredential` and the `gitCheck` part of `Test`. Widen `Gateway` with `Repositories` and `CheckGitAccess`. Add `ErrNoGitCredential` and fields `gitUsername`/`gitPassword` to `Classify`; missing Git access → `validation` field `gitCredential`.
- [x] Refactor while green.
- Contracts: C3, C5. NFRs: NFR3.

### Step 11 — Business logic, Red: GitIntegration service, watcher and events

Tests use a fake `HostPort`, a fake gateway with request counters, a fake `ConnectionReader` (`Current`, `Credentials`, `GitCredential`, `Subscribe`), and `synctest` for timers.

- [x] Repository source (`repos_test.go`, AC5.1.1, AC5.1.2, AC1.7.1): `ListRepositories(ws, query, cursor)` lists only selected projects' repositories as `RepositoryInspection` fields (`providerHost` `https://<space>`, `providerScope` = spaceHost, `repositoryId` = Backlog id, `ownerOrProject` = project key, `cloneUrl` = `httpUrl`); the query filters by name; the cursor is opaque and bound to query and scope, a mismatched cursor is rejected before any request; gateway errors pass through. `Inspect(ws, url)` returns the snake_case descriptor or `matched:false` for another host, an unselected project or a non-Backlog URL. `Branches(ws, repository)` returns `BranchCandidates` from the newest 100 PRs, or an empty list.
- [x] Linking (`links_test.go`, AC5.2.1–AC5.2.3): `Link(ws, taskId, reference, taskRepos)` resolves the PR with one `PullRequest` call and stores the link; a missing PR → `ErrPRNotFound` → 404 `not_found`; linking twice is idempotent; `Unlink` removes only that association with no Backlog write; `Associations(ws)` lists only active links.
- [x] Status (`status_test.go`, AC5.4.1, AC5.4.3): `Status(ws, taskId)` returns one `ReviewSummary` per active link (`taskStatus.state` open/closed/merged, table-driven; `statusBadge.label` "Open – Lan"; `pipelineState: neutral`; `checks: []`); a failed fetch → label "Status unknown" and `taskStatus.error`, other links still returned; not-connected links → label "Not connected".
- [x] Create (`create_test.go`, AC5.3.2, AC5.3.4): `CreatePR(ws, taskId, kandevRepoID, headBranch, title, body, baseBranch)` resolves the Backlog repository from `HostPort.Repository` (provider `nulab-backlog` only); empty title → `validation` field `title`; empty `headBranch` → `validation` field `branch` (AC5.3.3); lists open PRs and an open PR with the same `branch` → `ErrOpenPRExists{Number}` with 0 create requests; otherwise resolves `RelatedIssueKey` with one `Issue` call, then `CreatePullRequest` with base (or the Kandev default branch), branch and `issueId`; links the task and returns the PR URL; if the link store fails after the create, returns `linked:false` and never creates again.
- [x] Credential scope (`resolver_test.go`, AC5.6.1, AC5.6.3): `ResolveCredential(scope)` requires provider `nulab-backlog`, non-empty task, session and repository, a host equal to the connected host, and a path whose project is selected; returns `{Username, Secret, ExpiresAt: now+15m}`; with no Git credential → `ErrNoGitCredential` whose text says to update Git access in Settings and holds no password; every refusal holds no secret. `Binding(scope)` returns the connection binding, or empty when not connected.
- [x] Watches (`watches_test.go`, AC6.1.1–AC6.1.5): Save creates an Active watch with numeric `assigneeId`/`createdUserId` for `me` (one `Myself` call) and `issueId` for a linked issue key (one `Issue` call), storing `workflowId`/`workflowStepId` from the request; Edit keeps counts and ledger; Pause → no request over 5 virtual cycles, Resume → requests again; Delete removes the watch and keeps ledger and tasks; a not-connected watch cannot be resumed (`conflict`).
- [x] Watcher (`watcher_test.go`, AC6.2.1–AC6.2.5, AC6.1.4, AC1.8.3): one worker goroutine handles ticks (every 5 virtual minutes) and `Run` requests; 3 matching open PRs → exactly 3 linked tasks; 25 PRs → 10, 10, 5 across cycles, oldest number first, with `createdCount`/`pendingCount`; a ledgered PR whose task was deleted is never re-created; a failure at the 5th of 10 continues from the 5th next cycle; a reserved ledger entry without a task id is resolved next cycle via task metadata `nulab_backlog_pr`, created again only when none is found; a restart (new watcher over the same store) creates no duplicates; `Run` racing a tick (`-race`) → one task per PR; an epoch change during a cycle drops its results; calls are `Background` and a 429 waits without blocking actions; one `watch_cycle` log line per cycle with `watchCount`, `created`, `errors`, `durationMs`, `rateLimitWaits`, no secret.
- [x] ConnectionChanged (`events_test.go`, AC1.5.4, AC1.8.2, AC1.9.2 Git parts, M12): `disconnected` → all links and watches `not_connected`, next cycle 0 requests; `space_changed` → old-host items `not_connected`; `projects_changed` → only deselected projects' items change; `Restore` with matching host/project → links active, watches **Paused**; a lower epoch than the last handled is ignored; startup reconciliation with `Current()` applies the same rules.
- [x] Impact (`impact_test.go`, AC1.8.1, AC1.9.1 PR part): `Impact(ws, projectKeys)` → `{prLinks, prWatches}` for all active items or only the given projects.
- [x] Saved queries (`queries_test.go`, AC6.3.1, AC6.3.2): save, list, delete; `RunQuery` returns at most 20 PR rows with state and linked task ids; empty → `[]`; errors keep their codes (`rate_limited` with `retryAfterSeconds`, `unreachable`).
- [x] Leak test (`leak_test.go`): API key, token, Git password and bait bodies never appear in logs, errors, replies or task descriptions.
- [x] Run the tests and record the failing output.
- Stories: all U4 plus Git parts of US1.5, US1.8, US1.9. Contracts: C2, C3, C5. ADR-003, ADR-004, ADR-005. ACs: AC5.1.1–2, AC5.2.1–3, AC5.3.2–4, AC5.4.1, AC5.4.3, AC5.6.1, AC5.6.3, AC6.1.1–5, AC6.2.1–5, AC6.3.1–2, AC1.5.4, AC1.7.1, AC1.8.1–3, AC1.9.1–2 (PR/watch parts), AC7.3.2. NFRs: NFR2, NFR3, NFR5, NFR11.

### Step 12 — Business logic, Green and Refactor: GitIntegration

- [x] `internal/git/service.go`: `Service` holding the `Gateway` (consumer-side interface over the six client calls plus `Myself`), the `Connection` (consumer-side interface over `Current`, `Credentials`, `GitCredential`, `Subscribe`), the `HostPort` and the store; repository, link, status, create, resolver, impact and query methods.
- [x] `internal/git/host.go`: `HostPort` with `CreateTask(ctx, NewTask) (TaskRef, error)`, `FindTaskByMetadata(ctx, ws, key, value) (string, bool, error)`, `Repository(ctx, ws, repoID) (KandevRepository, error)`.
- [x] `internal/git/watcher.go`: `Watcher{Every, Now}` with one goroutine and a `Run` channel; per watch: reserve ledger → `CreateTask` → write task id → add link; `Start`/`Stop`.
- [x] `internal/git/events.go`: `OnConnectionChanged` and `Reconcile`.
- [x] Refactor while green.
- Contracts: C2, C3. NFRs: NFR2, NFR3, NFR5, NFR11.

### Step 13 — API / endpoint, Red: actions, credential RPCs, manifest, wiring

- [x] `internal/plugin/actions_u4_test.go`, through `HandleAction` with the fake Host: `connection.set_git_credential` (admin); `repositories.inspect`, `repositories.branches`, `git.repositories.list`; `git.prs.link`, `git.prs.unlink`, `git.prs.create` (task scope; `HeadBranch` and `RepositoryID` from the verified context, never the body); `git.prs.status`, `git.links.list`, `git.impact`; `git.watches.list|save|delete|run|pause|resume`; `git.queries.list|save|delete|run`. Every action refused with 409 `integration_disabled` while Backlog is off; not found → 404 `not_found`; open PR → 409 `conflict` with `pullRequestNumber`; bad bodies → 400 `validation` with field; 429 → `Retry-After`; one `action_failed` line on a 500 with no secret.
- [x] `internal/plugin/credential_test.go`: `Runtime` satisfies `pluginsdk.GitCredentialHandler`; `ResolveGitCredential` returns username and secret only for a valid scope; each refusal returns a gRPC error with no secret and logs `git_credential_refused` with a reason; `GetGitCredentialBinding` returns the binding, or empty after a disconnect.
- [x] `internal/plugin/manifest_test.go`, U4 cases: new actions with scope/access/`max_body_bytes` from Step 14, keys matching `^[a-z0-9][a-z0-9._-]*$`; `repository_providers: ["nulab-backlog"]`; `capabilities.api_read: [tasks, repositories]`, `api_write: [tasks]`.
- [x] `internal/plugin/runtime_u4_test.go`: `newRuntime` subscribes the Git service to `ConnectionChanged`, starts the watcher with an injected interval, `Close` stops it; the `HostPort` adapter maps `Tasks().Create`/`List` and `Repositories().List`; a missing Host gives `ErrHostUnavailable`.
- [x] Run the tests and record the failing output.
- Stories: all U4. Contracts: C2, C5, C8. Findings: R-04, R-06, R-09. ACs: AC5.3.3, AC5.6.1, AC5.6.3, AC8.3.1 (U4 actions).

### Step 14 — API / endpoint, Green and Refactor

- [x] Add `internal/plugin/git_actions.go` (handlers merged into the `handlers` map in `init()`), `internal/plugin/credential.go` (the two RPCs) and `internal/plugin/host_port.go` (adapter on `hostStores`). Extend `runtime.go`: wiring, `pullRequestNumber` on `actionError`, Git error codes in the mapping.
- [x] `manifest.yaml`: `connection.set_git_credential` workspace/admin; `repositories.inspect`, `repositories.branches` workspace/authenticated; every `git.*` key workspace/authenticated except `git.prs.link`, `git.prs.unlink`, `git.prs.create`, `git.prs.status` which are task/authenticated; every action `max_body_bytes: 16384`; add `repository_providers` and the two capability lists.
- [x] Refactor while green.
- Contracts: C5, C8. NFRs: NFR3, NFR11.

### Step 15 — Frontend behaviour, Red: M1 Git block, M4, M5, M6, M7, M9, M11, M12 additions

All tests use the shared fake `host` (`ui/src/testing/harness.ts`, extended with `registerRepositoryProvider`, `registerReviewProvider`, `registerTaskAction`, `openTaskLinkDialog`, `openModal`, `context.getTaskCreationContext`), run `axe-core` with no violations, check `data-testid` on every interactive element, and check no literal text outside `messages/en.ts`.

- [x] `ui/src/git/git-state.test.ts`: PR badge label and screen-reader text "Pull request #42, Open, assignee Lan"; watch row status Active/Paused/Not connected and progress "Created 10/25 tasks, the rest in later cycles"; error code → notice mapping (reconnect, rate limited with countdown, unreachable).
- [x] `ui/src/git/git-access.test.tsx` (M1): collapsible "Git access (optional)" block with labelled username and password; Save calls `connection.set_git_credential`, clears the password and shows "Saved"; `hasGitCredential` shows "Stored", never the password; "Saving…" disables the button; `gitCheck: invalid` from Test connection shows the Git authentication error, announced once (AC5.5.2).
- [x] `ui/src/git/repository-provider.test.ts` (M9): provider id `nulab-backlog`; `listRepositories` forwards query, cursor and signal to `git.repositories.list`; an action error rejects so Kandev shows its error with Retry (AC5.1.2); `inspectURL` → null on `matched:false`; `listBranches` maps `repositories.branches`; an aborted signal publishes nothing.
- [x] `ui/src/git/pr-link.test.ts` (M7): task action "Link Backlog pull request" opens `openTaskLinkDialog` with localized copy; `onSubmit` calls `git.prs.link` with the task id; `not_found` rejects with "Pull request #999 not found"; `validation` field `reference` rejects with "This link is not a pull request in <space>".
- [x] `ui/src/git/review-provider.test.tsx` (M6): `getAssociationSnapshot` from `git.links.list`; `refresh` calls `git.prs.status` and publishes summaries; `unlink` calls `git.prs.unlink` (AC5.2.3); "Status unknown" on error (AC5.4.3); `ReviewPanel` renders `host.ui.ChangeRequestDetail` with title, state, branches.
- [x] `ui/src/git/create-pr.test.ts` (M11): `createChangeRequest` calls `git.prs.create` with `taskId`, `sessionId`, `repositoryId` selectors and `{title, body, baseBranch}`, never a source branch; returns `{url, linked: true}`; a `conflict` with `pullRequestNumber` opens a confirm "Pull request #N is already open for this branch. Link it?" (Confirm → `git.prs.link`, Cancel creates nothing, AC5.3.4); `validation` field `title` rejects with "Title is required"; `supportsDraft` is `false`.
- [x] `ui/src/git/watches-page.test.tsx` (M4): states loading, empty ("No PR watches yet" with New watch), error with Retry, rate-limited, not-connected; Run, Pause, Resume, Delete call their actions; Run disabled while running; the delete confirmation says created tasks are not deleted, focus on Cancel (AC6.1.5); not-connected rows have no Resume.
- [x] `ui/src/git/watch-form.test.tsx` (M4): Name*, Repository* (selected projects only), Status checkboxes in a `fieldset`, Assignee and Creator (Anyone/Me), Linked issue; empty name or server `validation` shows a field error and moves focus (AC6.1.2); Save sends `workflowId`/`workflowStepId` from `getTaskCreationContext`; "Saving…" disables the button.
- [x] `ui/src/git/dashboard-page.test.tsx` (M5): saved queries in a select; choosing one runs `git.queries.run` and lists PRs with state and linked task links (AC6.3.1); empty shows a hint; error or rate limit shows Retry (AC6.3.2); the save-query dialog validates Name and Repository; Delete asks for confirmation.
- [x] Update `ui/src/settings/connected-panel.test.tsx` and `project-picker.test.tsx` only for: the Git block mount; disconnect/change-space dialogs showing "N PR links and M PR watches will be turned off" from `git.impact` (AC1.8.1 PR part); the deselect confirmation showing the same counts for unselected projects (AC1.9.1 PR part); the M12 notice adding "PR watches are Paused." and a "Review watches" button to `/backlog/watches`.
- [x] Run the tests and record the failing output.
- Stories: all U4 plus Git parts of US1.5, US1.8, US1.9. ACs: AC5.1.2, AC5.2.1–3, AC5.3.1, AC5.3.4, AC5.4.1, AC5.4.3, AC5.5.1–2, AC6.1.1–5, AC6.2.2, AC6.3.1–2, AC1.8.1, AC1.9.1 (PR parts). NFRs: NFR9, NFR10.

### Step 16 — Frontend behaviour, Green and Refactor

- [x] Add kebab-case files in `ui/src/git/`: `git-state.ts`, `git-access.tsx`, `repository-provider.ts`, `pr-link.ts`, `review-provider.tsx`, `create-pr.ts`, `watches-page.tsx`, `watch-form.tsx`, `dashboard-page.tsx`.
- [x] Register in `ui/src/index.ts`: the provider, the task action, the review provider, routes `/backlog/watches` and `/backlog/dashboard` with nav items in section `integrations`.
- [x] Mount `git-access.tsx` in `connected-panel.tsx`; add counts to the existing confirm dialogs; add message keys to `messages/en.ts`. Use only host UI kit components, U1 spacing classes and U2 `confirm-dialog.tsx`.
- [x] Refactor while green.
- Rules: U1 BR6.1–BR6.5. NFRs: NFR9, NFR10.

### Step 17 — Environment and build configuration

- [x] No `Makefile` or CI change; `GO_PKGS` already covers `./internal/...` and Vitest runs every file.
- [x] From `make clean`, run `make check-format vet lint test coverage check-secrets build package verify-package`; Go coverage ≥ 80% with only `server/main.go` excluded and `internal/git` ≥ 85%; `go mod tidy` leaves no diff.
- [x] Confirm the packaged `manifest.yaml` carries `repository_providers`, the capabilities and the new actions, and `verifypkg` reports OK.
- [x] Run `make contract-test KANDEV_MIN_DIR=../kandev-min` and confirm Kandev v0.96.0 accepts the manifest and starts the plugin.
- Stories: US7.3. NFRs: NFR6, NFR8.

### Step 18 — Documentation and traceability

- [x] `README.md`: append "Backlog Git and pull requests": Git user name and password (Backlog password, or the Git password with 2FA), the repository source, linking and creating PRs, PR watches (at most 10 tasks per cycle, every 5 minutes), saved queries, assumption A1. Only append.
- [x] `docs/manual-checks/TEMPLATE.md`: B5 demo steps (create a task on a Backlog repository and let the agent push; create a PR and see the badge; a watch creates a task for a new PR; disconnect turns PR links and watches not connected, reconnect restores them with watches Paused).
- [x] Write `construction/git-pr/code-generation/code-summary.md`, `source-manifest.json` and `traceability.json`, including the upstream amendments under Assumptions.
- Stories: all U4.

## Story-to-Step Map

| Story / AC part | Steps |
|-----------------|-------|
| US5.5 Store Git credentials | 3–6, 7–10, 13–16, 18 |
| US5.1 Choose a Backlog repository | 3, 4, 7, 8, 11–16 |
| US5.6 Fetch and push with Git credentials | 9–14, 18 |
| US5.2 Link and unlink a PR | 3–8, 11–16 |
| US5.4 PR status on the card | 3, 4, 7, 8, 11–16 |
| US5.3 Create a PR from a task | 3, 4, 7, 8, 11–16 |
| US6.1 Manage PR watches | 3–6, 11–16 |
| US6.2 Watch creates tasks | 5, 6, 11–14, 15, 16 |
| US6.3 Saved queries and dashboard | 3–8, 11–16 |
| US1.5 / US1.8 / US1.9 (Git password, PR link and watch parts) | 5, 6, 9–12, 15, 16 |
| US1.7 AC1.7.1 (repositories and watches from selected projects) | 11, 12 |
| US7.3 AC7.3.2 (Git error flow) | 7–12, 17 |

**Shared files touched**: `manifest.yaml`; `README.md` (one appended section); `docs/manual-checks/TEMPLATE.md`; `internal/plugin/{runtime.go,manifest_test.go}`; `internal/connection/{store.go,lifecycle.go,service.go,store_test.go,fakes_test.go}` (Connection-owned, edited under R-09); `internal/backlog/client.go`; `ui/src/index.ts`, `ui/src/messages/en.ts`, `ui/src/testing/harness.ts`; `ui/src/settings/{connected-panel.tsx,project-picker.tsx,state.ts,connected-panel.test.tsx,project-picker.test.tsx}`. Possible overlap with U3: `backlog.Client.Issue`, `runtime.go`, `index.ts`, `en.ts`. Not touched: `Makefile`, `.github/workflows/*`, `.golangci.yml`, `go.mod`, `go.sum`, `ui/package.json`, `internal/ci`, `cmd/`.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `unit-of-work.md`, `unit-of-work-story-map.md`, `unit-of-work-dependency.md`; `stories.md`; `requirements.md`; `components.md`, `decisions.md`; `contract-summary.md` and `contract-design/reviews/review-01.md`; `mockups.md`, `interaction-spec.md`, `accessibility-checklist.md`; `bolt-plan.md` (B5), `external-dependency-map.md` (X6).
- U2: `code-generation-plan.md`, `code-summary.md`, and the code in the working tree.
- `team.md`, `project.md`, `phases/construction.md`.
- Kandev v0.96.0: `apps/backend/pkg/pluginsdk/{plugin.go,host.go,types.go,data_types.go}`; `apps/backend/internal/plugins/repository_provider.go`; `apps/backend/internal/backendapp/git_credentials.go`; `apps/packages/plugin-sdk/src/index.ts`; `docs/public/plugins-manifest.md`; `docs/public/plugins-authoring.md`.
- Bitbucket plugin <https://github.com/kdlbs/kandev-plugin-bitbucket> (README level).
- Backlog: <https://developer.nulab.com/docs/backlog/api/2/get-pull-request-list/>, <https://developer.nulab.com/docs/backlog/api/2/add-pull-request/>, <https://developer.nulab.com/docs/backlog/api/2/get-list-of-git-repositories/> (read 2026-10-06).

## Assumptions & Open Questions

- [assumption] There is no U4 functional, NFR or infrastructure design (user choice). The `git.*` schemas the bolt plan wanted in functional design (R-04) are defined by this plan (Steps 13–14) and need review at plan approval.
- [assumption] The GitIntegration package is `internal/git` (C2). It declares consumer-side interfaces (`Gateway`, `Connection`, `HostPort`) because tests swap fakes in. Only `internal/plugin` imports `pluginsdk`.
- [assumption] The Git credential is a separate secret `backlog.git.<workspaceId>` = `{username, password, spaceHost, revision}`, bound to the connected host and deleted by `Store.Disconnect` and by `saveConnection` on a host change, before the connection writes, so ADR-005 holds and the Git parts of AC1.5.4/AC1.8.2 are met inside Connection. Same-host replacement, project changes and token refresh keep it.
- [assumption] `connection.set_git_credential` is `access: admin` (C5). No delete action: Disconnect deletes it and saving replaces it.
- [assumption] AC5.5.2 "Test connection" extends `connection.test`: with a stored Git credential it probes `GET /git/<PROJ>/<repo>.git/info/refs?service=git-upload-pack` with HTTP Basic auth on the first repository of the first selected project, through BacklogGateway (ADR-002), and reports `gitCheck`. Whether Backlog's Git HTTPS endpoint answers 401/200 this way is unverified.
- [assumption] Provider id `nulab-backlog`; `providerScope`/`connectionScope` = space host; `repositoryId` = Backlog repository id; `ownerOrProject` = project key; `cloneUrl` = Backlog `httpUrl` (https on the connected host, no credentials).
- [assumption] Backlog has no branch-list API and no default-branch field, so `repositories.branches`/`listBranches` return branch names seen in the newest 100 pull requests, or an empty list. This only partly meets AC5.1.1 and needs the user's decision.
- [assumption] `ResolveGitCredential` rejects a missing task, session or repository id, as the Kandev docs advise. If Kandev's first host-side clone sends no task/session id, cloning fails closed; check this in the B5 manual demo (AC5.6.2 `[manual]`).
- [assumption] Lease `ExpiresAt` is 15 minutes; binding `"<connectionEpoch>.<revision>"`, so credential, space, project changes or disconnect revoke leases.
- [assumption] M7 uses Kandev's `host.openTaskLinkDialog` (one reference field) instead of a plugin dialog with a Repository select and preview, since Kandev owns the Link dialog. The reference accepts a PR URL, `<repo>#<n>`, or `<n>` when the task has exactly one Backlog repository.
- [assumption] The PR badge (M6) is published through `registerReviewProvider`; Kandev renders it and refreshes every 90 s, so U4 runs no PR status poller. Label "Open – Lan". Screen-reader wording and 320 px layout are host-owned (AC5.4.2 `[manual]`).
- [assumption] M11 is Kandev's native Create PR dialog with the plugin's `createChangeRequest`. Kandev builds the title and body, so "Related: PROJ-120" (AC5.3.1) applies only when the body is empty. The related issue is the first selected-project issue key in title or body, resolved with one `Issue` call; U3's task-to-issue link is not read (U4 must not depend on U3).
- [assumption] AC5.3.4 checks open PRs for the same branch first; on a match it returns 409 `conflict` with `pullRequestNumber` and the UI offers "Link it?". Nothing is created.
- [assumption] U4 adds `backlog.Client.Issue(ctx, creds, class, issueKey)` with the C1 signature for the numeric issue id. If U3 lands it first, U4 reuses it.
- [assumption] PR status ids: 1 Open, 2 Closed, 3 Merged. Only 1 is confirmed on the docs page.
- [assumption] Watch filters: status, assignee (Anyone/Me), creator (Anyone/Me), linked issue. "Choose user…" from M4 is skipped (needs a project-users API no story requires).
- [assumption] The watch cycle is fixed at 5 minutes (FR4.2 default); U3 owns the interval setting. Watch calls are `Background`.
- [assumption] One watcher goroutine for all workspaces serialises ticks and Run requests (AC6.2.4 race-free). `git.watches.run` returns at once (`queued`) and the cycle runs on the worker, because 10 task creations may exceed the 15 s action limit. Workspaces with watches are in instance state `git.watch_index`.
- [assumption] Exactly-once task creation: reserve ledger entry (spaceHost, repositoryId, number) → `Tasks().Create` with metadata `nulab_backlog_pr: <key>` → store task id. A reserved entry with no task id is resolved next cycle by searching workspace tasks (including archived) for that metadata, created only when none is found. A deleted task's ledger entry stays, so it is never re-created (AC6.2.3). Ponytail ceiling: a paged list scan per reserved entry.
- [assumption] Watch tasks use `workflowId`/`workflowStepId` captured when the watch is saved; title "Review PR #<n>: <summary>", description = PR URL; no repository attached.
- [assumption] All U4 data lives in single workspace state documents with `schemaVersion: 1`, capped at 500 links, 50 watches, 50 queries, serialised by a per-workspace in-process mutex. Ponytail: split per key if the caps are hit.
- [assumption] Late results (AC1.8.3): a cycle or action reads the connection epoch before calling Backlog and checks `Current()` again before writing; if it moved, nothing is written.
- [assumption] On `ConnectionChanged`: `disconnected` → all items not connected; `space_changed` → other-host items not connected; `projects_changed` → deselected projects' items not connected; `Restore` with matching host/project → links back, watches back as **Paused** (M12); not-connected watches cannot be resumed; startup reconciles with `Current()`.
- [assumption] AC1.8.1/AC1.9.1 counts: U4 adds `git.impact` and shows PR-link/PR-watch counts in the U2 dialogs; U3 adds issue-link counts. The M12 notice gets "PR watches are Paused." and "Review watches", without restored counts.
- [assumption] `task.deleted` (R-05) is not handled in U4: the ledger prevents re-creation and links to deleted tasks are harmless. U3 owns the `OnEvent` fan-out.
- [assumption] Every `git.*` action is `authenticated` (C5 default). The U1 integration-switch guard blocks all of them while Backlog is off.
- [assumption] New Go test secrets (Git passwords) use `testutil.Token`, generated at run time and never committed.
- Upstream amendments needed: C1 (`Issue`, `CheckGitAccess`, `CallClass` on `PullRequests`/`PullRequest`); C2 (`HostPort` gains `FindTaskByMetadata`, `Repository`; `ResolveGitCredential` moves to the plugin implementing `pluginsdk.GitCredentialHandler`); C3 (`GitCredential` returns `(GitCredential, binding, error)`); C5 (`connection.set_git_credential`, fixed `repositories.inspect`/`repositories.branches`, the `git.*` list, `conflict` field `pullRequestNumber`, `ConnectionView.gitCheck`); C8 (`repository_providers`, `api_read`, `api_write`).
- Open: X6 — pull requests and Git hosting must be on the test space's plan (A1) for the B5 manual demo.
- Open: the AC5.1.1 base-branch choice and the Git HTTPS probe path need the user's decision or a real-space check.
