# Contract Summary — Kandev Plugin for Nulab Backlog

Inputs:

- `unit-of-work` (U1–U5) and `unit-of-work-dependency` (dependency graph, integration points), from the Units Generation step.
- `components` (6 components, entities, ADR-001 to ADR-007), from the Domain Design step.
- `requirements` (FR1–FR7, NFR1–NFR11).
- The answers to Q1–Q6 in `contract-design-questions.md`.
- Kandev plugin documentation and source code (`docs/public/plugins-manifest.md`, `docs/public/plugins-authoring.md`, `apps/backend/proto/kandev/plugin/v1/plugin.proto`, `pkg/pluginsdk`) in the `../kandev` checkout.

A contract is an agreement at a boundary: which data crosses it, in what shape, through which mechanism, and what happens when there is an error.

There are three kinds of boundary:

- **Between units.** All five units sit in one package and one Go process (`unit-of-work`). So they connect through Go interfaces and in-process events, not over the network. Only `ci-release` connects to `walking-skeleton` through `Makefile` targets.
- **Between the UI and the backend.** The UI calls plugin actions through the Kandev server.
- **With the outside.** This covers Backlog API v2 (the plugin is the caller), the Kandev server (events, host calls, state store, secret store), and the browser returning to the OAuth callback address.

## Contract Table

| # | Provider Unit | Consumer | Mechanism | Owner |
|---|---------------|----------|-----------|-------|
| C1 | walking-skeleton | connection, issues, git-pr | Go interface `backlog.Client` and error type `backlog.Error` (in-process) | walking-skeleton (U2 adds more calls) |
| C2 | walking-skeleton | issues, git-pr | Go host port interface `HostPort` (ADR-007) | walking-skeleton |
| C3 | connection | issues, git-pr | Go interface `ConnectionReader` and in-process event `ConnectionChanged` (ADR-005) | connection |
| C4 | walking-skeleton | ci-release | `Makefile` targets and package path (shared schema) | walking-skeleton |
| C5 | walking-skeleton, connection, issues, git-pr | PluginUI (in the same package) | Plugin HTTP actions through Kandev: `POST /api/plugins/{id}/actions/{key}` | Each unit owns its own actions |
| C6 | connection | External: browser (OAuth redirect from Backlog) | Public plugin webhook: `GET /api/plugins/{id}/webhooks/oauth-callback` | connection |
| C7 | External: Backlog API v2 | walking-skeleton, connection, issues, git-pr (only through BacklogGateway) | REST/HTTPS | Owned by Nulab; the plugin follows v2 |
| C8 | External: Kandev server | walking-skeleton (KandevAdapter) | gRPC over go-plugin, manifest, bus events | Owned by Kandev; the plugin declares it in `manifest.yaml` |

`ci-release` has no runtime contract. It only uses C4 and checks C8 with a contract test on the minimum Kandev version.

## General Rules for All Contracts

### Versioning [Q2]

- There is no separate version number for actions or interfaces. The UI and the backend always sit in the same package, so they always match.
- A breaking change (renaming an action, removing a field, changing a field's meaning) bumps the package's major version per semver.
- An additive change (adding an action, adding an optional field) is not breaking. Receivers ignore unknown fields.
- Plugin data stored in Kandev's state store has a `schemaVersion` field. On startup, the plugin migrates old data to the new version automatically. If migration fails, the plugin goes into an error state and does not overwrite the data.

### Shared Error Codes [Q3]

Every action returns errors in one single shape. The UI picks the message to show by `code` and by the current language (NFR10). Errors never contain Backlog response content or secrets (NFR3).

| `code` | HTTP status | When | Extra fields |
|--------|-------------|---------|-------------|
| `reconnect_required` | 401 | Backlog returns 401, the refresh token is no longer valid, or not connected yet | — |
| `rate_limited` | 429 | Rate limited by Backlog beyond the allowed time (see below) | `retryAfterSeconds`; with a `Retry-After` header |
| `unreachable` | 503 | Network error, over 10 seconds, or Backlog returns 5xx | — |
| `not_found` | 404 | Issue, PR, repository or task does not exist | — |
| `validation` | 400 | Invalid input (for example a malformed space address) | `field` |
| `conflict` | 409 | Rule violation (for example the task already has another issue, an open PR already exists for the branch) | — |
| `internal` | 500 | Any other error | — |

Error body:

```yaml
# shared-schema: ActionError (JSON)
ActionError:
  type: object
  required: [error]
  properties:
    error:
      type: object
      required: [code]
      properties:
        code: { type: string, enum: [reconnect_required, rate_limited, unreachable, not_found, validation, conflict, internal] }
        retryAfterSeconds: { type: integer, minimum: 1 }   # only with rate_limited
        field: { type: string }                             # only with validation
        requestId: { type: string }                         # for matching logs, contains no secrets
```

Error mapping:

- Only `internal/plugin` (KandevAdapter) maps errors to HTTP status (`components`, ADR-007).
- `HandleAction` always returns a `PluginActionResponse` with a `status`. It never returns a Go error, because Kandev turns every Go error into a generic 503 and the UI would lose the error code.
- The Kandev toolkit has a built-in `upstream` → 502 code. The plugin does not use it; Backlog-side errors become `unreachable` or `reconnect_required`.

### Timeouts and Retries [Q4, Q5]

- Every HTTP call to Backlog has a 10-second limit (AC8.1.3).
- On a 429, the plugin waits per `X-RateLimit-Reset`. If that header is missing it uses `Retry-After`, and if that is also missing it waits 60 seconds. The plugin retries at most 3 times, then returns `rate_limited` (AC8.4.2, NFR2).
- **User-clicked commands** (every action in C5) do not wait more than 3 seconds because of rate limiting. If a longer wait is needed, the action returns `rate_limited` right away with `retryAfterSeconds`, and the UI retries automatically when the time is up (NFR1).
- **Background tasks** (the issue sync cycle, PR watch) wait the full time, within the 3-retry limit above.
- Kandev gives each action at most 15 seconds. A user command spends at most 3 seconds waiting on the rate limit plus 10 seconds calling Backlog, so 13 seconds. That always stays under Kandev's 15-second limit.
- An action may only chain Backlog calls if the total time stays under 15 seconds. Actions that need several calls (for example creating a task with copies of comments and files, FR3.2) must have their own time budget. This budget is settled at the functional design step.
- Every function takes a `context.Context`, and returns `context.Canceled` or `DeadlineExceeded` unchanged.

## C1 — `backlog.Client` (walking-skeleton → connection, issues, git-pr)

BacklogGateway is the only anti-corruption layer (ADR-002). It holds no credentials; the caller passes them in on every call. U1 builds the client with one call (`Myself`). U2, U3 and U4 add calls, but keep the signatures and the error type unchanged.

```yaml
# shared-schema: Go package internal/backlog
package: internal/backlog
types:
  Credentials:
    SpaceHost: string        # validated: https, *.backlog.com | *.backlog.jp | *.backlogtool.com, lowercase
    APIKey: string           # empty when using OAuth
    AccessToken: string      # empty when using an API key
  Error:                     # the only error type Client returns (errors.As)
    Kind: enum [Unauthorized, RateLimited, Unreachable, NotFound, Invalid, Conflict]
    Status: int              # HTTP status from Backlog, 0 on a network error
    RetryAfter: duration     # only with RateLimited
    # No body field, full URL or secret. Error() is already redacted (internal/redact).
  CallClass: enum [Interactive, Background]   # Interactive: waits at most 3 seconds on the rate limit [Q5]
interface Client:
  # U1
  Myself(ctx, Credentials) -> (User, error)
  # U2
  ExchangeOAuthCode(ctx, spaceHost, code, codeVerifier, redirectURI) -> (TokenSet, error)
  RefreshToken(ctx, spaceHost, refreshToken) -> (TokenSet, error)
  Projects(ctx, Credentials) -> ([]Project, error)
  # U3
  Issues(ctx, Credentials, CallClass, IssueQuery) -> (IssuePage, error)
  IssueCount(ctx, Credentials, CallClass, IssueQuery) -> (int, error)
  Issue(ctx, Credentials, CallClass, issueIDOrKey) -> (Issue, error)
  IssueComments(ctx, Credentials, CallClass, issueIDOrKey, limit) -> ([]Comment, error)
  IssueAttachments(ctx, Credentials, CallClass, issueIDOrKey) -> ([]Attachment, error)
  ProjectStatuses(ctx, Credentials, projectIDOrKey) -> ([]Status, error)
  # U4
  Repositories(ctx, Credentials, projectIDOrKey) -> ([]Repository, error)
  PullRequests(ctx, Credentials, CallClass, projectIDOrKey, repoIDOrName, PullRequestQuery) -> ([]PullRequest, error)
  PullRequest(ctx, Credentials, CallClass, projectIDOrKey, repoIDOrName, number) -> (PullRequest, error)
  CreatePullRequest(ctx, Credentials, projectIDOrKey, repoIDOrName, NewPullRequest) -> (PullRequest, error)
guarantees:
  - Search-group and update-group calls run sequentially in a shared queue, at least 1 second apart, no matter which component the call comes from (ADR-003).
  - Does not follow redirects to another host.
  - Reads responses through io.LimitReader.
  - Writes structured logs for errors and waits, without secrets (NFR11).
```

## C2 — `HostPort` (walking-skeleton → issues, git-pr)

KandevAdapter implements this port. The business components only see the interface and do not import `pluginsdk` (ADR-007). `issues` and `git-pr` use this port through an indirect dependency: `issues` → `connection` → `walking-skeleton`.

```yaml
# shared-schema: interface declared on the consumer side (internal/issues, internal/git)
interface HostPort:
  CreateTask(ctx, NewTask) -> (TaskRef, error)       # NewTask: workspaceId, title, description, sourceURL, metadata
  TaskExists(ctx, taskId) -> (bool, error)           # GetTask; codes.NotFound -> (false, nil)
  SetTaskLabels(ctx, taskId, labels []string) -> error
  # U4
  ResolveGitCredential(ctx, repositoryId) -> (GitCredential, error)
errors:
  - ErrHostUnavailable: Kandev does not respond; background tasks skip that item and retry in the next cycle.
  - ErrHostDenied: missing permission (capability or workspace grant not yet approved); reported to the user through the internal code with a log entry.
rules:
  - If task creation fails, no link is created (no orphan links, `components` IssueIntegration).
  - Labels on the card are only a display copy, rebuilt from IssueLink/PullRequestLink (ADR-004).
```

## C3 — `ConnectionReader` and `ConnectionChanged` (connection → issues, git-pr)

```yaml
# shared-schema: Go package internal/connection
interface ConnectionReader:
  Current(ctx) -> (Snapshot, error)                  # ErrNotConnected when not connected
  Credentials(ctx) -> (backlog.Credentials, error)   # refreshes the OAuth token automatically if it is about to expire; concurrent calls refresh only once
  GitCredential(ctx) -> (GitCredential, error)       # U4; ErrNoGitCredential if not stored
Snapshot:
  SpaceHost: string
  AuthMethod: enum [api_key, oauth]
  ConnectionEpoch: int64           # increases on every space change, disconnect or project change
  SelectedProjects: [{ProjectKey, ProjectID, ProjectName}]
  PollMinutes: int                 # default 5, minimum 1
event ConnectionChanged:           # emitted in-process, sent to every subscriber
  ConnectionEpoch: int64
  Reason: enum [connected, disconnected, space_changed, projects_changed, credentials_replaced]
  SpaceHost: string                # empty when disconnected
  SelectedProjectKeys: [string]
  Restore: bool                    # true when reconnecting to the same old space or reselecting the same old projects
delivery:
  - Connection writes the new epoch to the state store first, then emits the event.
  - Secrets are deleted in Connection before the disconnected event is emitted (ADR-005).
  - Receivers process events in epoch order. A receiver ignores events and Backlog call results with an epoch lower than the one it holds (AC1.8.3).
  - On startup, each receiver reads Current() and compares it with its stored epoch, so it does not depend on whether the event arrived.
  - Receivers must not block: heavy work moves to the receiver's own goroutine.
```

## C4 — `Makefile` Targets (walking-skeleton → ci-release)

```yaml
# shared-schema: Makefile targets (exit code 0 = pass)
targets:
  check-format:   gofmt (and Prettier for the UI), does not modify files
  vet:            go vet ./...
  lint:           golangci-lint (default + gosec), tsc --noEmit, ESLint
  test:           go test -race ./internal/... ./server/... ; Vitest
  coverage:       go test -race -coverprofile; fails if < 80% (exclusions listed in the Makefile)
  build:          binaries for linux, darwin, windows on amd64/arm64 (5 binaries, NFR7)
  package:        builds the plugin package in dist/
  verify-package: verifies the package in dist/
outputs:
  dist/<plugin-id>-<version>.<ext>: plugin package
  dist/checksums.txt
rules:
  - CI calls exactly these targets, and does not repeat its own commands.
  - Renaming or removing a target is a breaking change for ci-release.
```

## C5 — Plugin Actions (UI → backend) [Q1]

Each operation is its own action with a clear name, declared in the manifest's `actions[]`. The UI calls it through `host.api.invokeAction(key, input)`. Kandev wraps the call as `{workspaceId, taskId, repositoryId, body}`, checks the login session, then passes it to the plugin's `HandleAction`. Request and response are JSON with fixed types. Kandev does not check types, so the plugin checks them itself and returns `validation` when they are wrong. Every error is returned as `ActionError`.

No action returns a secret. The UI only receives the flags `hasApiKey`, `hasOAuthToken` and `hasGitCredential`.

```yaml
openapi: 3.0.3
info: { title: Backlog plugin actions, version: "plugin package (semver)" }
paths:
  /api/plugins/{id}/actions/connection.get:            # U1; scope workspace
    post: { responses: { "200": { $ref: "#/components/schemas/ConnectionView" } } }
  /api/plugins/{id}/actions/connection.connectApiKey:  # U1; scope workspace, access admin
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [spaceUrl, apiKey], properties: { spaceUrl: {type: string}, apiKey: {type: string, writeOnly: true} } } } } }
      responses: { "200": { $ref: "#/components/schemas/ConnectionView" }, "400": { description: "validation (field: spaceUrl | apiKey)" } }
  /api/plugins/{id}/actions/connection.startOAuth:     # U2; returns the Backlog sign-in URL
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [spaceUrl], properties: { spaceUrl: {type: string} } } } } }
      responses: { "200": { content: { application/json: { schema: { type: object, properties: { authorizeUrl: {type: string} } } } } } }
  /api/plugins/{id}/actions/connection.test:           # U2
    post: { responses: { "200": { $ref: "#/components/schemas/ConnectionView" } } }
  /api/plugins/{id}/actions/connection.disconnect:     # U2; access admin
    post: { responses: { "200": { $ref: "#/components/schemas/ConnectionView" } } }
  /api/plugins/{id}/actions/connection.setProjects:    # U2
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [projectKeys], properties: { projectKeys: {type: array, items: {type: string}} } } } } }
      responses: { "200": { $ref: "#/components/schemas/ConnectionView" } }
  /api/plugins/{id}/actions/connection.listProjects:   # U2; projects available on the space
    post: { responses: { "200": { description: "[{projectKey, projectName, selected}]" } } }
  /api/plugins/{id}/actions/connection.setPollInterval: # U2
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [minutes], properties: { minutes: {type: integer, minimum: 1} } } } } }
  /api/plugins/{id}/actions/connection.setGitCredential: # U4
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [username, password], properties: { username: {type: string}, password: {type: string, writeOnly: true} } } } } }
  /api/plugins/{id}/actions/issues.list:               # U3; scope workspace
    post:
      requestBody: { content: { application/json: { schema: { $ref: "#/components/schemas/IssueQuery" } } } }
      responses: { "200": { $ref: "#/components/schemas/IssuePage" } }
  /api/plugins/{id}/actions/issues.get:                # U3; scope task (M8 sidebar)
    post: { responses: { "200": { description: "IssueDetail: issue, linkState, statusUpdatedAt, comments[], attachments[]" } } }
  /api/plugins/{id}/actions/issues.createTask:         # U3
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [issueKey], properties: { issueKey: {type: string} } } } } }
      responses: { "200": { description: "{taskId, issueKey}" }, "409": { description: "conflict" } }
  /api/plugins/{id}/actions/issues.link:               # U3; scope task
    post:
      requestBody: { content: { application/json: { schema: { type: object, required: [issueKey], properties: { issueKey: {type: string} } } } } }
      responses: { "409": { description: "conflict: the task already has another issue (FR3.4)" } }
  /api/plugins/{id}/actions/issues.unlink:             # U3; scope task
    post: {}
  /api/plugins/{id}/actions/issues.refresh:            # U3; refreshes status right away (FR4.2)
    post: { responses: { "200": { description: "{updatedCount, refreshedAt}" } } }
  /api/plugins/{id}/actions/git.links.list:            # U4; same for git.prs.link, git.prs.unlink, git.prs.create,
    post: {}                                           # git.watches.list|save|delete|run|pause|resume, git.queries.list|save|delete|run
components:
  schemas:
    ConnectionView:
      type: object
      properties:
        connected: {type: boolean}
        state: {type: string, enum: [not_connected, connected, sign_in_again, error]}
        spaceHost: {type: string}
        authMethod: {type: string, enum: [api_key, oauth]}
        connectedUserName: {type: string}
        hasApiKey: {type: boolean}
        hasOAuthToken: {type: boolean}
        hasGitCredential: {type: boolean}
        selectedProjects: {type: array, items: {type: object, properties: {projectKey: {type: string}, projectName: {type: string}}}}
        pollMinutes: {type: integer}
        connectionEpoch: {type: integer}
    IssueQuery:
      type: object
      properties:
        projectKeys: {type: array, items: {type: string}}
        statusIds: {type: array, items: {type: integer}}
        assigneeIds: {type: array, items: {type: integer}}
        keyword: {type: string}              # an exact issue key match is put first
        page: {type: integer, minimum: 1}
        pageSize: {type: integer, default: 20, maximum: 100}
    IssuePage:
      type: object
      properties:
        total: {type: integer}
        refreshedAt: {type: string, format: date-time}   # FR4.4
        connectionEpoch: {type: integer}                 # the UI drops pages with an old epoch
        items:
          type: array
          items:
            type: object
            properties:
              issueKey: {type: string}
              summary: {type: string}
              status: {type: string}
              assignee: {type: string}
              updatedAt: {type: string, format: date-time}
              url: {type: string}
              linkedTaskIds: {type: array, items: {type: string}}   # FR2.3
```

Action owners by prefix:

- `connection.*` belongs to `walking-skeleton` (U1: `get`, `connectApiKey`) and `connection` (U2: the rest). Only `connection.setGitCredential` is built together with U4 (`unit-of-work`).
- `issues.*` belongs to `issues` (U3).
- `git.*` belongs to `git-pr` (U4).

The actions that write connection configuration (`connectApiKey`, `disconnect`, `setGitCredential`) use `access: admin`. The other actions use `access: authenticated`. Kandev requires `min_kandev_version` `0.91.1` or later when there is an `admin` action.

The parts Kandev renders (`#` suggestions, repository source, PR status on the card) do not go through actions. They go through the mount points in C8.

## C6 — OAuth Callback (connection → browser)

Kandev has no dedicated OAuth path for plugins, so the callback is a public webhook.

```yaml
openapi: 3.0.3
info: { title: Backlog OAuth callback, version: "plugin package (semver)" }
paths:
  /api/plugins/{id}/webhooks/oauth-callback:
    get:
      description: Backlog redirects the browser here after the consent page.
      parameters:
        - { name: code,  in: query, schema: { type: string } }
        - { name: state, in: query, required: true, schema: { type: string } }
        - { name: error, in: query, schema: { type: string } }   # access_denied when the user cancels
      responses:
        "302": { description: "Redirects to the M1 settings page with ?oauth=connected | cancelled | failed. Never puts the code, token or Backlog error into the URL." }
rules:
  - state is signed, single-use, expires after 10 minutes, and is bound to the spaceHost at start (OAuthPendingState in `components`). An invalid state returns 302 with oauth=failed.
  - The code is exchanged for a token through C1 ExchangeOAuthCode; the token is stored with Kandev's SetSecret.
  - The webhook declares access: public, because the browser returns from Backlog.
```

## C7 — Backlog API v2 (Nulab → plugin)

The plugin is the caller and follows the contract Nulab publishes. Only BacklogGateway calls these endpoints. The table below is the part the plugin uses.

```yaml
openapi: 3.0.3
info: { title: Backlog API v2 (part used by the plugin), version: v2 }
servers: [{ url: "https://{space}.backlog.com" }, { url: "https://{space}.backlog.jp" }, { url: "https://{space}.backlogtool.com" }]
paths:
  /api/v2/users/myself: { get: {} }                                              # U1
  /OAuth2AccessRequest.action: { get: {} }                                       # U2, opened by the browser
  /api/v2/oauth2/token: { post: {} }                                             # U2: authorization_code, refresh_token
  /api/v2/projects: { get: {} }                                                  # U2
  /api/v2/projects/{projectIdOrKey}/statuses: { get: {} }                        # U3
  /api/v2/issues: { get: {} }                                                    # U3
  /api/v2/issues/count: { get: {} }                                              # U3
  /api/v2/issues/{issueIdOrKey}: { get: {} }                                     # U3
  /api/v2/issues/{issueIdOrKey}/comments: { get: {} }                            # U3
  /api/v2/issues/{issueIdOrKey}/attachments: { get: {} }                         # U3
  /api/v2/projects/{projectIdOrKey}/git/repositories: { get: {} }                # U4
  /api/v2/projects/{projectIdOrKey}/git/repositories/{repoIdOrName}/pullRequests: { get: {}, post: {} }   # U4
  /api/v2/projects/{projectIdOrKey}/git/repositories/{repoIdOrName}/pullRequests/{number}: { get: {} }  # U4
rateLimitHeaders: [X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset]
errorMapping:   # HTTP -> backlog.Error.Kind
  401: Unauthorized
  403: Unauthorized
  404: NotFound
  429: RateLimited
  5xx: Unreachable
  network/timeout: Unreachable
```

- The plugin never calls the API to write to an issue (FR4.3).
- The only write call is creating a pull request, which exists only in U4.

## C8 — Kandev Server (Kandev ↔ plugin)

Kandev owns the protocol (gRPC over go-plugin). The plugin declares what it uses in `manifest.yaml`. Calling an undeclared capability is rejected with `PermissionDenied`.

```yaml
# shared-schema: the part of manifest.yaml relevant to the contracts (final values settled at Code Generation)
api_version: 2
min_kandev_version: "0.91.1"        # minimum because there is an admin action; the contract test runs on exactly this version (NFR6)
capabilities:
  events: ["task.deleted"]           # [Q6]
  state: true                        # IssueLink, PullRequestLink, watch, dedup ledger, connectionEpoch, schemaVersion
  secrets: true                      # API key, OAuth token, Git password
  api_read: [tasks]
  api_write: [tasks]                 # create task, set labels
actions: [ ...C5... ]
webhooks:
  - { key: oauth-callback, access: public }          # C6
repository_providers: [ ... ]        # U4: repository source, ResolveGitCredential (FR5.1)
reference_sources: [ ... ]           # U3: `#` suggestions (FR3.5)
ui:
  - settings route and the Integrations entry (M1)
  - task-sidebar (M8), task-card-tags (issue/PR badges), review provider (PR status, Kandev polls every 90 seconds)
```

The `task.deleted` event [Q6]:

- Kandev sends the event to `OnEvent` with `event_type = task.deleted` and a payload containing `task_id`, `workspace_id`.
- Kandev delivers on a best-effort basis only, with no guarantee. It retries after 5, 15 and 45 seconds, may drop events under overload or on restart, and does not deliver while the plugin is off.
- So the plugin uses two paths:
  1. On receiving the event, remove every IssueLink and PullRequestLink of that `task_id`. The PR watch dedup ledger entry is kept, so the deleted task is not re-created (FR6.2).
  2. Every sync cycle, call `TaskExists` (C2, based on `GetTask`) for the linked tasks. A `NotFound` result is handled the same as receiving the event.
- Events are deduplicated by `task_id` plus event type, not by `event_id`. The Kandev source code notes that `event_id` is generated anew on each delivery.

Timers: Kandev has no shared timer for plugins. The issue sync cycle and PR watch run in the plugin's own goroutines, as ADR-003 assumed.

## Contract Ownership Rules

- The providing unit owns its contracts, as in the Owner column. A consuming unit does not add calls to another unit's interface on its own. If it needs a new call, the consuming unit adds that call during the providing unit's pass, or states it clearly in its own plan.
- A breaking change in C1–C5 or C8 must be called out in the pull request, and bumps the package's major version [Q2]. Since every unit lives in the same repo, the consumers are fixed in that same pull request.
- Additive changes are always safe: receivers ignore unknown fields and enum values. The UI treats an unknown `code` as `internal`.
- C7 is owned by Nulab. A change on the Backlog side only requires edits in `internal/backlog`.
- C8 is owned by Kandev. A change on the Kandev side only requires edits in `internal/plugin`, and is detected by the contract test on the minimum version (U5).

## Open Questions

| Contract | Question | Blocks |
|----------|----------|--------|
| C7 | Does Backlog accept the API key through a header? The Backlog documentation uses the `apiKey` URL parameter, while NFR3 and ADR-002 require sending it in a header. If Backlog only accepts it in the URL, the redaction layer must also redact the URL in every error and log. This is already a requirement in `components`. | U1 |
| C6, C7 | Does Backlog OAuth support PKCE? `components` requires PKCE. If it does not, the plugin relies only on the signed single-use `state` and a client secret kept on the server side. | U2 |
| C2, C8 | Setting issue labels on a task with `SetTaskLabelsExact` needs an administrator to grant `host.v2.write:tasks` for the workspace. The alternative is to only display labels through the UI's `task-card-tags` mount point, without writing real labels. Which approach? | U3 |
| C5 | The time budget for `issues.createTask` when copying comments and attachments (FR3.2) must stay under 15 seconds. This part may have to be copied after the task is created. | U3 |
| C8 | Does Backlog accept a `localhost` callback over HTTP when testing (assumption A2 in `requirements`)? | U2 |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog, modelled on the Bitbucket plugin, built on the public Backlog API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q6]: the answers in `contract-design-questions.md`.
- `unit-of-work.md`, `unit-of-work-dependency.md` (Units Generation); `components.md`, `decisions.md` (Domain Design); `requirements.md` (Requirements Analysis); `stories.md` (AC8.1.3, AC8.4.2, AC1.8.3).
- Kandev: `docs/public/plugins-manifest.md`, `docs/public/plugins-authoring.md`, `docs/decisions/0050-plugin-external-auth-capability.md`, `apps/backend/proto/kandev/plugin/v1/plugin.proto`, `pkg/pluginsdk/action_error.go`, `pkg/pluginsdk/host.go`.

## Assumptions & Open Questions

- [assumption] The list of Backlog endpoints in C7 is taken from Nulab's public documentation and has not been called yet. The first manual check with a real space in the walking skeleton confirms `users/myself`; the other endpoints are confirmed as each unit is built.
- [assumption] The action names in C5 are proposed names. The full list for `git.*` is settled at the functional design step of U4.
- [assumption] The 10-minute lifetime of the OAuth `state` is a proposed value, not taken from any answer.
- The five questions in the "Open Questions" section.
