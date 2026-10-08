# API Documentation — kandev-plugin-nulab-backlog

## Kandev Host Contract (`manifest.yaml`)

- `version: "0.5.0"`, `api_version: 2`, `runtime.type: binary`, 4 executables since v0.4.2 (`linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`), `min_kandev_version: "0.96.0"`.
- Capabilities: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`, `events: [task.deleted]`. There is no capability or permission for starting processes; a native plugin binary can exec without declaring anything.
- `repository_providers: [nulab-backlog]`, 1 `reference_sources` entry (`nulab-backlog-issues`), `config_schema` (OAuth client id/secret, public base URL), `ui.bundle: /ui/bundle.js`.

## Plugin Actions

Counts from the v0.4.2 scan (not recounted for v0.5.0):

| Prefix | Count | Notes |
|---|---|---|
| `connection.*` | 9 | connect API key / OAuth, get, set_enabled, disconnect, projects, Git credential; admin-only where they change state |
| `repositories.*` | 2 | repository provider |
| `git.*` | 19 | Backlog PRs, watches, queries (incl. `git.prs.link`, body `{reference}`) |
| `issues.*` | 25 | list, tasks, links, sync, watches, quick actions, queries |
| `scm.*` | 22 | GitHub/GitLab/Bitbucket context; see [below](#scm-actions) |

Each action declares `scope` (workspace/task), `access` (authenticated/admin), `max_body_bytes` (8-256 KiB). Every key in `manifest.yaml` must have a runtime handler and vice versa (`TestU2_ManifestActionKeysMatchTheRuntime`, `internal/plugin/manifest_test.go:136-142`); a new action needs both.

### SCM Actions

`manifest.yaml` lines 270-360, handlers in `internal/plugin/scm_actions.go`, all `scope: workspace`, `max_body_bytes: 16384`.

| Key | Access | Body | Service call | Returns |
|---|---|---|---|---|
| `scm.providers.list` | authenticated | — | `Providers` | `{ providers: ProviderView[] }` (all three) |
| `scm.providers.set_token` | admin | `TokenInput{provider, token, username?}` | `SetToken` (validates with `CurrentUser`, stores secret) | `ProviderView` |
| `scm.providers.test` | admin | `{provider}` | `Test` (calls `CurrentUser`; refusal is recorded as `lastError`, not an error) | `ProviderView` |
| `scm.providers.remove` | admin | `{provider}` | `RemoveToken` (deletes secret; mappings and watches stay) | `ProviderView` |
| `scm.repos.search` | admin | `{provider, query}` | `SearchRepos` | repos |
| `scm.mappings.set` | admin | `MappingInput` | `SetMapping` (checks repos with the token) | view |
| `scm.prs.list`, `scm.prs.link`, `scm.prs.unlink`, `scm.links.list`, `scm.task_prs.list` | authenticated | per action | PR lists and links (`task_prs` reads Kandev's own PR data, no plugin token) | per action |
| `scm.queries.*` (5), `scm.watches.*` (6) | authenticated | per action | saved queries and watches | per action |

`ProviderView` (Go `internal/scm/service.go:101`, TS mirror `ui/src/git/git-state.ts`): `provider`, `state` (`not_configured` / `connected` / `error`), `account?`, `lastError?` (`invalid_token`, `missing_scope`, `rate_limited`, `unreachable`), `mappings`. It never contains the token and is readable by every member, so any field added to it must stay non-secret (NFR1).

SCM error mapping (`classifySCM`): HTTP 401/403 -> `reconnect_required`; 404 or `ErrNotFound` -> `not_found`; 429 -> `rate_limited` with `retryAfterSeconds`; other HTTP failures and `ErrHostRefused` -> `unreachable`; `ErrConflict` -> `conflict`; `ErrNoToken` -> `validation` with field `token`.

### Link actions used by the Link Task flow (`internal/plugin/issue_actions.go`, service `internal/issues/service.go`)

| Action | Scope | Request body | Response / errors |
|---|---|---|---|
| `issues.tasks.search` | workspace | `{query}` | `{tasks: [{taskId, taskKey?, title, linkedIssueKey?}]}`, max 20 rows, case-insensitive match on title or key |
| `issues.link` | task (task id from verified context) | `{issueKey}` | `Link`. Key must match `^([A-Z][A-Z0-9_]*)-([1-9][0-9]{0,8})$` (`ParseIssueKey`, `types.go:56`) -> else validation; a project must be selected; task linked to another issue -> `ErrConflict`. **Accepts an issue key only, not an issue URL.** |
| `issues.unlink` | task | none | `ErrNotLinked` when the task has no link |

Domain errors are mapped to SDK codes only in `internal/plugin/runtime.go` (around lines 344-348). The UI turns failures into notices with `issueNotice` (`ui/src/issues/issues-state.ts`).

### `issues.links.list` (workspace, authenticated)

Handler in `internal/plugin/issue_actions.go`, service `Issues.Links`. Response `{ links: LinkView[] }`:

| Field | Type | Notes |
|---|---|---|
| `taskId`, `taskKey?` | string | Kandev task |
| `issueKey`, `spaceHost` | string | Backlog issue |
| `summary?` | string | stored on `Link` since v0.5.0 (hover text) |
| `state` | string | link state |
| `status?`, `statusUpdatedAt?` | string | last synced status |
| `stale`, `unavailable` | bool | sync health |
| `url` | string | `https://<space>/view/<KEY>` |

The TS mirror is `LinkView` in `ui/src/issues/issues-state.ts`.

## Webhooks

- `oauth-callback`: GET, public, 1 KiB body limit.

## Kandev Host API Used by the Plugin (pluginsdk v0.96.0)

`Host` methods cover state, config, secrets (`GetSecret`, `SetSecret`, `DeleteSecret`, `RevealSecret`), events, tasks, sessions, workspaces, workflows, repositories, messages and a utility agent. **No method exposes Kandev's own GitHub credential or runs a process** (external reference `~/repo/kandev/apps/backend/pkg/pluginsdk/host.go`, shallow).

## UI Registration API (Kandev plugin registry, consumed by `ui/src/index.ts`)

Used: `registerIntegrationSettings`, `registerNavItem`, `registerRoute`, `registerComponent(slot, ...)`, `registerTaskAction`, `registerTaskMenuAction`, `registerTaskPanel`, `registerRepositoryProvider`, `registerReviewProvider`, messages. Surface mapping and gating: [architecture.md](architecture.md#ui-surfaces-host-slots).

Relevant Kandev v0.96.0 SDK surface (`apps/packages/plugin-sdk/src/index.ts`, external):

- `registerTaskAction({id, label, icon?, placement: "link", group?, visible?, singleTaskOnly?, run})` (line 842) — adds an item to the task's Link submenu.
- `host.openTaskLinkDialog({title, description, inputLabel, placeholder?, emptyError, failureMessage, successMessage, inputTestId?, errorTestId?, submitTestId?, onSubmit(reference, signal)})` (line 755) — host-rendered single-field link form (Enter submits, Save/Saving..., inline destructive error from the thrown `Error` message, success toast). No Unlink button, no prefill.
- `NavItem = { id, label, path, icon, section }` — no visibility/`requires` field.
- `TaskRowMetadataSlotProps = { taskId, workspaceId, workflowStepId, surface: "sidebar" | "task-list" }` for slot `task-row-metadata`.

## Kandev Install API (consumed by operators and by the contract test)

- `POST /api/plugins/install`, either multipart field `package` (web UI upload, `internal/ci/contract.go`, expects 201) or JSON `{"url": "<package url>"}` (backend downloads, 100 MiB cap).
- `GET /ready`, `GET /api/plugins/<id>`.
- Error mapping in the install handler: 409 / 400 / 500 only; it never returns 502.

### Known issue: upload install over a slow link

The multipart body is read fully before install, bounded by the server `ReadTimeout` (default 30 s, `server.readTimeout` / `KANDEV_SERVER_READTIMEOUT`). An upload slower than the timeout is cut: backend returns 400 `missing multipart field "package"`, the `tailscale serve` proxy returns 502. Details: [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502).

## GitHub Actions Triggers and Required Checks

Recorded by run 1 of `261008-ci-path-filter`, before its change (PR #14 later added app-path gating; not re-verified since):

| Workflow | Trigger | Concurrency | Jobs |
|---|---|---|---|
| `ci.yml` | `pull_request: branches: [main]`, `push: branches: [main]` | none | `checks`; `packaged-host-contract` (`needs: checks`, 30 min) |
| `release.yml` | `push: tags: ['v*']` | `group: release`, `cancel-in-progress: false` | `verify`; `contract` (`needs: verify`); `publish` (`needs: [verify, contract]`) |

- Both workflows: top-level `permissions: contents: read`. Only `release` / `publish` has `contents: write`, `id-token: write`, `attestations: write`.
- `checks`: checkout plugin + Kandev at SDK ref, setup-go, setup-node, `npm ci`, `go mod tidy` diff check, `make check-format vet lint test coverage check-secrets build package verify-package`, upload `plugin-package`.
- `packaged-host-contract`: checks out Kandev at SDK ref and at `v<min_kandev_version>`, downloads the package, `sha256sum -c`, `make verify-package contract-test KANDEV_MIN_DIR=../kandev-min`.
- `verify`: as `checks` with `fetch-depth: 0`, plus `make release-preflight TAG=$TAG` (`GH_TOKEN`), uploads `release-package`. `contract`: as `packaged-host-contract` on `release-package`. `publish`: attests `dist/nulab-backlog-*.tar.gz`, `gh release create --verify-tag --generate-notes` (prerelease when the tag has `-`).
- GitHub applies `paths` / `paths-ignore` filters only to branch pushes and pull requests, not to tag pushes.

### Required status checks (`main` ruleset)

Repository ruleset `24580280` (read via `gh api`, 2026-10-08):

- `required_status_checks`: contexts **`checks`** and **`packaged-host-contract`** (integration 15368 = GitHub Actions), `strict_required_status_checks_policy: false`.
- Other rules: `deletion`, `non_fast_forward`, `pull_request` with `allowed_merge_methods: [squash]` and 0 required approvals.
- Classic branch protection is not used (`branches/main/protection` returns 404).
- A required check matches by job name. A job skipped by `if:` reports as success; a workflow that never starts reports nothing, leaving the check "Expected - waiting".

## Outbound APIs

- Backlog REST v2 (`internal/backlog`): ~17 paths — `users/myself`, `projects`, statuses, project users, `issues`, `issues/count`, comments, attachments, git repositories, pull requests and count, `oauth2/token`.
- GitHub REST (`internal/github`, `https://api.github.com`, `Authorization: Bearer <token>`, read-only): `GET /user`, `/user/repos`, `/repos/{o}/{r}`, `/repos/{o}/{r}/pulls`, `/repos/{o}/{r}/pulls/{n}`. Any bearer token GitHub accepts works, including one printed by `gh auth token`.
- GitLab / Bitbucket REST (read-only): repos, PRs/MRs, user (not re-read in run 3).
