# Code Summary — git-pr (U4)

All 18 plan steps were carried out in order under the TDD Testing Contract (`contract_sha256 sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`). Each layer went Red, then Green, then Refactor; the Red output of every layer is below. From `make clean`, `make check-format vet lint test coverage check-secrets build package verify-package` passes. Go line coverage is **92.9%** (floor 80%, only `server/main.go` excluded); `internal/git` is **90.7%** (target 85%). `go mod tidy` leaves no diff. `make contract-test KANDEV_MIN_DIR=../kandev-min` passes on Kandev v0.96.0.

U4 adds one Go package (`internal/git`) and no Go module or npm dependency. Only `internal/plugin` imports `pluginsdk`. `Makefile`, `.github/workflows/*`, `.golangci.yml`, `internal/ci`, `cmd/`, `go.mod`, `go.sum` and `ui/package.json` were not touched. `README.md` only got a section appended at the end. The uncommitted U2 and U5 work is intact and all U1/U2/U5 tests stay green.

## Files

`source-manifest.json` lists every path U4 created or changed (81 paths).

| Area | Files |
|------|-------|
| BacklogGateway | `internal/backlog/repositories.go` (`Repository`, `parseRepositories`: https clone URL on the space host only), `pullrequests.go` (`PullRequest` with the description hidden from every format verb, `PullRequestState`, `PullRequestQuery.Values`, `NewPullRequest.Form`, `Issue`), `client.go` (`Repositories`, `PullRequests`, `PullRequest`, `CreatePullRequest`, `Issue`, `CheckGitAccess` on the one `send` path; `request` gains `query`, Basic auth and `statusOnly`); tests `repositories_types_test.go`, `pullrequests_types_test.go`, `git_client_test.go`; fixtures `testdata/repositories_ok.json`, `repositories_empty.json`, `pullrequests_ok.json`, `pullrequest_ok.json`, `pullrequest_created.json`, `issue_ok.json` |
| Connection (built with U4, R-09) | `internal/connection/git_credential.go` (`GitCredentialInput`, `ValidateGitCredential`, `GitCredential`, `gitSecret`, `SetGitCredential`, `GitCredential`, the `gitCheck` probe), `store.go` (`SaveGit`, `readGit`, `deleteGit`; deleted in `Disconnect` and on a host change; `Load` sets `hasGitCredential`; `View.GitCheck`), `service.go` (wider `Gateway`, `ErrNoGitCredential` → `validation`/`gitCredential`), `lifecycle.go` (`Test` adds `gitCheck`); tests `git_credential_types_test.go`, `git_credential_test.go`, `store_test.go` (U4 cases); `fakes_test.go` records write order and can fail one key's delete |
| GitIntegration | `internal/git/doc.go`, `types.go` (`ParseReference`, `ParseClonePath`, `LinkKey`, `WatchInput`/`QueryInput` validation, `RelatedIssueKey`, `PRDescription`, `BranchCandidates`, `DefaultBranch`), `store.go` (workspace documents `git.links`, `git.watches`, `git.ledger`, `git.queries` and the instance index `git.watch_index`, `schemaVersion: 1`, caps, per-workspace mutex, 1 s calls), `host.go` (`HostPort`, `NewTask`, `KandevRepository`), `service.go` (repository source, link/unlink/associations, status, create, credential scope and binding, impact, watches, queries), `watcher.go` (one worker goroutine, ledger → create → task id → link, `watch_cycle` log), `events.go` (`Listen`, `OnConnectionChanged`, `ReconcileAll`); tests `types_test.go`, `store_test.go`, `repos_test.go`, `links_test.go`, `status_test.go`, `create_test.go`, `resolver_test.go`, `watches_test.go`, `watcher_test.go`, `events_test.go`, `impact_test.go`, `queries_test.go`, `leak_test.go`, helpers `fakes_test.go`, `harness_test.go` |
| KandevAdapter | `internal/plugin/git_actions.go` (20 handlers merged into `handlers`; the verified action context travels on the context), `credential.go` (`ResolveGitCredential`, `GetGitCredentialBinding`), `host_port.go` (`Tasks().Create`/`List`, `Repositories().List`, paged), `runtime.go` (Git service and watcher wiring, `Start`/`Close`, `classify` with `not_found` and `conflict` + `pullRequestNumber`); tests `actions_u4_test.go`, `credential_test.go`, `runtime_u4_test.go`, `manifest_test.go` (U4 cases) |
| Manifest | `manifest.yaml`: 20 actions (`max_body_bytes: 16384`), `repository_providers: ["nulab-backlog"]`, `api_read: [tasks, repositories]`, `api_write: [tasks]` |
| UI | `ui/src/git/git-state.ts`, `git-access.tsx`, `repository-provider.ts`, `pr-link.ts`, `review-provider.tsx`, `create-pr.ts`, `watches-page.tsx`, `watch-form.tsx`, `dashboard-page.tsx` and their nine test files; `ui/src/index.ts` (provider, task action, review provider, `/backlog/watches`, `/backlog/dashboard`), `messages/en.ts`, `settings/connected-panel.tsx`, `project-picker.tsx`, `SettingsScreen.tsx`, `state.ts`; tests `index.test.ts`, `settings/connected-panel.test.tsx`, `project-picker.test.tsx`; `testing/harness.ts` extended |
| Docs | `README.md` (appended "Backlog Git and pull requests"), `docs/manual-checks/TEMPLATE.md` (B5 demo steps 15–20) |

## Action schemas (R-04)

Bodies are JSON objects; the task, repository, session and head branch of task-scoped actions always come from the verified action context, never from the body.

| Action | Scope / access | Body | Reply |
|--------|----------------|------|-------|
| `connection.set_git_credential` | workspace / admin | `{gitUsername, gitPassword}` | ConnectionView (`hasGitCredential`, never the password) |
| `repositories.inspect` | workspace / authenticated | `{url}` | `{repository: {provider_id, provider_host, provider_scope, provider_repository_id, owner_or_project, name, clone_url, default_branch}}` or `{matched: false}` |
| `repositories.branches` | workspace / authenticated | `{repository: <descriptor>}` | `{branches: [{name, is_default?}]}` |
| `git.repositories.list` | workspace / authenticated | `{query?, cursor?}` | `{repositories: [RepositoryInspection], nextCursor?}` (50 per page) |
| `git.prs.link` | task / authenticated | `{reference}` | the link `{taskId, spaceHost, projectKey, repoName, repositoryId, number, title, status}` |
| `git.prs.unlink` | task / authenticated | `{reviewKey}` | `{ok: true}` |
| `git.prs.create` | task / authenticated | `{title, body?, baseBranch?}` | `{url, number, linked}` |
| `git.prs.status` | task / authenticated | `{}` | `{summaries: [ReviewSummary + base, branch, assignee]}` |
| `git.links.list` | workspace / authenticated | — | `{associations: [ReviewTaskAssociation]}` |
| `git.impact` | workspace / authenticated | `{projectKeys?}` | `{prLinks, prWatches}` |
| `git.watches.list` / `save` / `delete` / `run` / `pause` / `resume` | workspace / authenticated | `{}` / WatchInput / `{id}` | `{watches}` / Watch / `{ok}` / `{queued: true}` / Watch |
| `git.queries.list` / `save` / `delete` / `run` | workspace / authenticated | `{}` / QueryInput / `{id}` | `{queries}` / Query / `{ok}` / `{rows: [≤ 20]}` |

Error codes: `validation` (with `field`), `not_found` (404), `conflict` (409, with `pullRequestNumber` for an open PR on the branch), `integration_disabled` (409, every U4 action while Backlog is off), `reconnect_required`, `rate_limited` (with `Retry-After`), `unreachable`, `internal`.

## Key Decisions

- **One send path.** All six new Backlog calls use `Client.send`, so they get the per-(host, group) queues, the 429 retries and the Interactive 3 s budget. The watcher's calls are `Background`. `CheckGitAccess` sends HTTP Basic auth instead of the API key header; the password and the base64 header value are added to the request's redaction set, and only the status of the Git ref list is read (`statusOnly`), so a large ref list is not an error.
- **Git secret bound to the space.** `backlog.git.<ws>` holds `{username, password, spaceHost, revision}`. It counts only when its host equals the connected host. `Disconnect` deletes it right after the connection secret and before the disconnect record (a failed delete writes no record); a host change deletes it before the new connection is written. The lease binding is `"<connectionEpoch>.<revision>"`, so saving, replacing, changing projects or the space, and disconnecting all revoke Kandev's leases.
- **Exactly-once watch tasks.** Per pull request: reserve a ledger entry → `Tasks().Create` with metadata `nulab_backlog_pr: <LinkKey>` → store the task id → link. A reserved entry without a task id is resolved by searching the workspace's tasks (archived included) for the metadata and created only when none is found. A ledger entry is never removed, so a deleted task is never re-created. One worker goroutine serialises ticks and `Run` requests.
- **Late results are dropped.** Link, create and every watch task read the connection epoch first and check `Current()` again before writing; if it moved, nothing is written (AC1.8.3).
- **Items follow the connection.** `disconnected` turns every link and watch off; other reasons turn off the items the new connection does not cover (another host or a deselected project); a restore brings covered items back with links active and watches **Paused** (M12). Events older than the last handled epoch are ignored. Every watcher cycle starts with the same reconciliation against `Current()`.
- **UI.** Kandev owns the Link dialog, the PR badge, the Create PR dialog and the repository picker; the plugin supplies data through `openTaskLinkDialog`, `registerReviewProvider` (with `ChangeRequestDetail` in the panel) and `registerRepositoryProvider`. The plugin pages (watches, dashboard) and the Git block use native `details`, `select` and checkboxes plus the host's `Button`, `Input` and `Label`; every interactive element has a `data-testid`, and axe reports no violations.

## Red evidence

Each Red run happened before the production code of its layer existed. Key lines are trimmed.

**Step 2, runner readiness.** `go test -race ./internal/backlog/... ./internal/connection/... ./internal/git/... ./internal/plugin/... -run '^TestU4_'` reported `ok … [no tests to run]` for `backlog`, `connection` and `plugin` and `[no test files]` for `internal/git`. The nine-file Vitest command with `--passWithNoTests` ran clean. The U2 regression commands were green (Vitest: 5 files, 68 tests; Go: `backlog`, `connection`, `plugin` ok).

**Step 3, data model.** `go test -race ./internal/backlog/... ./internal/connection/... ./internal/git/... -run '^TestU4_'`

```
internal/git/types_test.go:24:10: undefined: RepoRef
internal/git/types_test.go:40:16: undefined: ParseReference
internal/backlog/pullrequests_types_test.go:13:14: undefined: parsePullRequests
internal/backlog/pullrequests_types_test.go:45:26: undefined: PullRequestState
internal/connection/git_credential_types_test.go:17:14: undefined: ValidateGitCredential
internal/connection/git_credential_types_test.go:25:78: undefined: FieldGitUsername
FAIL	.../internal/backlog [build failed]
FAIL	.../internal/connection [build failed]
FAIL	.../internal/git [build failed]
```

**Step 5, repository / data access.** `go test -race ./internal/connection/... ./internal/git/... -run '^TestU4_'`

```
internal/git/store_test.go:21:7: undefined: NewStore
internal/git/store_test.go:23:10: undefined: Link
internal/git/store_test.go:26:54: undefined: LedgerEntry
internal/connection/store_test.go:766:27: store.SaveGit undefined (type *Store has no field or method SaveGit)
FAIL	.../internal/connection [build failed]
FAIL	.../internal/git [build failed]
```

**Step 7, business logic: BacklogGateway.** `go test -race ./internal/backlog/... -run '^TestU4_'`

```
internal/backlog/git_client_test.go:43:18: c.Repositories undefined (type *Client has no field or method Repositories)
internal/backlog/git_client_test.go:58:16: c.PullRequests undefined (type *Client has no field or method PullRequests)
internal/backlog/git_client_test.go:62:16: c.PullRequest undefined (type *Client has no field or method PullRequest)
internal/backlog/git_client_test.go:66:16: c.Issue undefined (type *Client has no field or method Issue)
FAIL	.../internal/backlog [build failed]
```

**Step 9, business logic: Connection Git credentials.** `go test -race ./internal/connection/... -run '^TestU4_'`

```
internal/connection/git_credential_test.go:67:18: u.svc.SetGitCredential undefined (type *Service has no field or method SetGitCredential)
internal/connection/git_credential_test.go:82:24: u.svc.GitCredential undefined (type *Service has no field or method GitCredential)
FAIL	.../internal/connection [build failed]
```

**Step 11, business logic: GitIntegration.** `go test -race ./internal/git/... -run '^TestU4_'`

```
internal/git/harness_test.go:243:14: undefined: NewTask
internal/git/harness_test.go:248:23: undefined: KandevRepository
internal/git/harness_test.go:317:9: undefined: Service
internal/git/harness_test.go:418:43: undefined: Watcher
internal/git/create_test.go:11:20: undefined: CreateInput
internal/git/resolver_test.go:13:14: undefined: Scope
FAIL	.../internal/git [build failed]
```

**Step 13, API / endpoint.** `go test -race ./internal/plugin/...`

```
internal/plugin/actions_u4_test.go:370:2: undefined: actionSetGitCredential
internal/plugin/actions_u4_test.go:370:26: undefined: actionReposInspect
internal/plugin/actions_u4_test.go:371:2: undefined: actionPRLink
internal/plugin/actions_u4_test.go:371:32: undefined: actionPRCreate
internal/plugin/actions_u4_test.go:371:81: undefined: actionImpact
FAIL	.../internal/plugin [build failed]
```

**Step 15, frontend behaviour.** From `ui/`: `npx vitest run src/git/git-state.test.ts … src/git/dashboard-page.test.tsx` (the nine U4 files), then the U2 files U4 extends.

```
 FAIL  src/git/create-pr.test.ts [ src/git/create-pr.test.ts ]
Error: Failed to resolve import "./create-pr" from "src/git/create-pr.test.ts". Does the file exist?
 FAIL  src/git/git-access.test.tsx [ src/git/git-access.test.tsx ]
Error: Failed to resolve import "./git-access" from "src/git/git-access.test.tsx". Does the file exist?
 …(the same for the other seven files)
 Test Files  9 failed (9)

$ npx vitest run src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx
     × shows and announces the M12 restore notice
     × mounts the Git access block (US5.5)
     × passes the Git check of Test connection to the Git block (AC5.5.2)
     × says how many PR links and watches a disconnect turns off (AC1.8.1)
     × says how many PR links and watches a space change turns off
     × says how many PR links and watches of the unselected projects are turned off (AC1.9.1)
      Tests  6 failed | 14 passed (20)
```

Each Green step then made its layer pass, and each Refactor kept it green (for example: a fresh body error per parse, one `cycleCtx` for the watcher's counting logger, `errUnchanged` so unchanged documents are not rewritten, `withID` for the id-only handlers).

## Test, coverage and contract results

- **Go:** 132 new `TestU4_` tests (backlog 20, connection 17, git 78, plugin 17), all under `-race`; `go test -race -count=10` over the four packages is stable. Timer tests (`watcher_test.go`, `watches_test.go`, `store_test.go`, the 429 and queue tests) run in `testing/synctest` with no real sleep.
- **UI:** 52 tests in the nine `ui/src/git` files plus 6 new tests in `connected-panel.test.tsx`, `project-picker.test.tsx` and `index.test.ts`; the full Vitest run is 19 files, 161 tests, all green. `tsc --noEmit`, ESLint and Prettier are clean.
- **Gate (from `make clean`):** `make check-format vet lint test coverage check-secrets build package verify-package` → exit 0. golangci-lint (with gosec): `0 issues.` `coverage: 92.9% (floor 80%, excluded: server/main.go)`; per package: backlog 95.8%, connection 94.5%, git 90.7%, plugin 93.1%. `ci secrets: OK`. `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`. The packaged `manifest.yaml` carries `repository_providers`, `api_read`, `api_write` and the 20 new actions.
- **`go mod tidy`:** no diff.
- **Contract:** `make contract-test KANDEV_MIN_DIR=../kandev-min` → `ci contract: OK nulab-backlog on Kandev v0.96.0` (ports 38529/39529; the user's Kandev on 38429/39429 was not touched).

## Review iteration 1 repairs

Each fix was written test first (the new tests failed against the old code, then passed).

- **R-01 (nested task metadata).** Kandev v0.96.0 stores plugin task metadata under `"plugin:nulab-backlog"` next to `source` (`pluginTaskMetadata`, host_write.go). `hostPort.FindTaskByMetadata` now reads that namespace only; flat keys and other plugins' namespaces never match. The fake `Tasks().Create` nests metadata the same way. Tests: `TestU4_HostPort_ReadsNestedMetadata`, `TestU4_Watcher_CrashAfterCreateNoDuplicate` (task created, ledger write fails, fresh runtime over the same host and storage, still one task).
- **R-02 (repository name).** Kandev sets `Repository.Name` to `<owner>/<providerName>`. `hostPort.Repository` now maps `ProviderName`, else the clone URL's last path segment without `.git`, never `Name`. A Backlog repository whose derived name fails the repo-name regex (new `git.ValidRepoName`) returns `ErrRepositoryNotFound` (fail closed: short refs are refused, Create PR answers invalid repository). The fake returns `Name: "PROJ/web-app"`, `ProviderName: "web-app"`, and the fake gateway records the project/repo path of every PR call. Tests: `TestU4_HostPort_RepoNameFromProvider` (table), `TestU4_HostPort_OtherProviderKeepsWorking`, `TestU4_Actions_ShortRefsUseProviderName` (`42`, `#42`, `web-app#42`), and a path assertion in `TestU4_Actions_CreateUsesTheVerifiedContext`.
- **R-03 (Backlog switch in the watcher).** `git.Connection` gains `RequireEnabled` (implemented by `connection.Service`, unchanged). `cycleWatch` checks it first: switch off skips the watch with no Backlog call and no error; a switch-store error is a cycle error with no Backlog call (fail closed). `createOne` checks it again before each task, so a switch turned off mid-cycle stops creation (dropped like a stale result). `Watcher.Run` is unchanged: the plugin guard already refuses `watches.run` while off. Tests: `TestU4_Watcher_SwitchOffSkipsWorkspace` (3 virtual cycles plus a Run: 0 gateway calls, 0 tasks; back on: resumes), `TestU4_Watcher_SwitchReadFailsClosed`, `TestU4_Watcher_SwitchOffMidCycleStops`, `TestU4_Runtime_WatcherObeysTheSwitch` (real `connection.Service` switch through the runtime).
- R-04 to R-07 are deferred to Build and Test, as agreed.
- `make check-format vet lint test coverage`: pass; total coverage 92.9% (git 90.8%, plugin 93.3%). No shared-with-other-units file was changed.

## Deviations

1. **`default_branch` is required.** Kandev v0.96.0 (`validateRepositoryProviderInspection`) rejects a `repositories.inspect` descriptor with an empty `default_branch`, and Backlog has no such field. `Inspect` uses the most used base branch of the newest 100 pull requests, else `master`, which costs one extra Read call per inspect.
2. **Credential RPC errors are plain Go errors** (gRPC delivers them as `Unknown`) with a fixed reason and no secret. Importing `google.golang.org/grpc/status` directly would turn its `// indirect` line in `go.mod` into a direct requirement, and `go.mod` is out of scope.
3. **Credential RPCs respect the switch.** While Backlog is off, `ResolveGitCredential` refuses (`integration_disabled`) and `GetGitCredentialBinding` returns an empty binding (BR7.3). This is in addition to the plan.
4. **Background start.** `NewRuntime` calls `Runtime.Start` (ConnectionChanged subscription and watcher); `newRuntime`, used by the tests, starts nothing, so the existing `synctest` test leaves no goroutine behind. `Close` stops both. The plan put the start in `newRuntime`.
5. **Startup reconciliation runs at the start of every watcher cycle** (the first one 5 minutes after start), because Kandev injects the Host only after `NewRuntime` returns.
6. **Names.** A missing Host returns the existing `errNoHost` (the plan said `ErrHostUnavailable`); `HostPort.CreateTask` returns the task id string (the plan said `TaskRef`).
7. **Short link references** use the verified `RepositoryID`; the UI sends it when the task has exactly one Backlog repository. `<repo>#<n>` takes the project of that repository.
8. **Files outside the plan's shared list:** `ui/src/settings/SettingsScreen.tsx` (pass the view to the connected panel, the PR counts in the change-space dialog, the M12 "PR watches are Paused." line with **Review watches**) and `ui/src/index.test.ts` (the registry needs the three new methods; nav items and routes go from 1 to 3). The M12 text of U2 is unchanged; the paused line is rendered under it, so U2's state tests stay as they were.
9. **Dashboard save form.** The plugin renders its own Name + Repository form (status Open, assignee Anyone) instead of `host.ui.IntegrationSaveQueryDialog`, whose props are not typed in the v0.96.0 SDK. The watch form and Git block use native `select` and `details` for the same reason.
10. **Link placeholder** is `https://<space>/git/PROJ/repo/pullRequests/42`: `verifypkg` rejects any `backlog.com` URL in the bundle as a possible asset URL.
11. **`create-pr.ts`** builds its confirm dialog with `h(...)` instead of JSX so the file keeps its planned `.ts` name.
12. **`gitCheck`** is absent (not `untested`) when no Git credential is stored, so the UI shows nothing in that case.

## Upstream amendments (for the contract and functional docs)

- **C1:** `Repositories(ctx, creds, projectKey)`, `PullRequests(ctx, creds, class, projectKey, repo, query)`, `PullRequest(ctx, creds, class, projectKey, repo, number)`, `CreatePullRequest(ctx, creds, projectKey, repo, NewPullRequest)`, `Issue(ctx, creds, class, issueKey)`, `CheckGitAccess(ctx, spaceHost, username, password, projectKey, repo)`.
- **C2:** `HostPort` = `CreateTask` (returns the task id), `FindTaskByMetadata`, `Repository`; `ResolveGitCredential` and `GetGitCredentialBinding` live in the plugin, which implements `pluginsdk.GitCredentialHandler`.
- **C3:** `GitCredential(ctx, ws)` returns `(GitCredential, binding, error)`; `ConnectionView` gains `hasGitCredential` and, on `connection.test`, `gitCheck`.
- **C5:** the action table above, including `conflict` with `pullRequestNumber` and `not_found`.
- **C8:** `repository_providers: ["nulab-backlog"]`, `api_read: [tasks, repositories]`, `api_write: [tasks]`; `repositories.inspect` must return a non-empty `default_branch`.

## Open items

- R-01 (OAuth webhook lock growth) and R-02 (`SetProjects` epoch race) from the U2 review are unchanged; they are scheduled for Build and Test.
- X6: pull requests and Git hosting must be on the test space's plan (A1) for the B5 manual demo (TEMPLATE steps 15–20), which also covers AC5.4.2 and AC5.6.2 `[manual]`.
- Unverified against a real space: whether Backlog's Git HTTPS endpoint answers `GET /git/<PROJ>/<repo>.git/info/refs?service=git-upload-pack` with 200/401 as assumed, and whether Kandev's first clone sends the task and session ids the resolver requires (it fails closed otherwise).
- AC5.1.1 is met only in part: branch choices come from recent pull requests, because Backlog has no branch-list API; the base-branch choice needs the user's decision.
- The ledger is never pruned and each list is one capped state document (`ponytail` notes in the code).
