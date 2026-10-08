# Developer Code Scan - 261007-source-control-agnostic

Focused scan for the intent "link the Backlog integration with multiple source-control services (GitHub, Backlog Git, Bitbucket, ...) and configure per-service linking conditions (auth, repo, space) in Settings". Scanned at commit `5bf88b9` (v0.3.0) on branch `feature/source-control-agnos-2jr`. The existing CodeKB is STALE; this scan records only what was read in this run.

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - internal/git/ (`types.go`, `service.go`, `prs.go`, `host.go`, `events.go`, `store.go` head, `watcher.go` head, `doc.go`)
  - internal/connection/ (`git_credential.go`, `store.go` keys/records/Git secret functions, `change.go`, `service.go` Gateway/ConfigReader, `lifecycle.go` Snapshot/Current/Test/Disconnect, `doc.go`)
  - internal/plugin/ (`runtime.go`, `credential.go`, `git_actions.go`, `host_port.go` Repository)
  - ui/src/git/ (`repository-provider.ts`, `review-provider.tsx`, `git-state.ts`, `git-access.tsx` head; action call sites in `pr-link.ts`, `create-pr.ts`, `pr-list.tsx`, `watch-form.tsx`)
  - ui/src/settings/ (`SettingsScreen.tsx` section layout; action call sites in `connected-panel.tsx`, `pr-watches-section.tsx`, `saved-queries-section.tsx`)
  - manifest.yaml
- **Skimmed only**:
  - internal/backlog/ (exported surface of `pullrequests.go`, `repositories.go`, `types.go`; `CheckGitAccess` and the host guard in `client.go`)
  - internal/issues/ (import graph only; confirmed it does not import `internal/git`)
  - ui/src/index.ts (registration calls only)
  - Kandev SDK at `/home/k_do_webfrontier/repo/kandev` tag v0.96.0 (`apps/backend/pkg/pluginsdk/plugin.go`, `types.go`; `internal/plugins/manifest/validate.go`, `internal/plugins/registry.go`, `internal/plugins/service_lifecycle.go`, `internal/backendapp/git_credentials.go`; `apps/web/lib/plugins/registry.ts`, `registry-provider-ownership.ts`) - outside the repo, read only to identify extension points
  - Not read: internal/ci, internal/pkgverify, internal/redact, internal/testutil, ui/src/issues, ui/src/page, ui/src/switch, .github/workflows (names only)

### Packages Found
- `internal/git` - library (component "GitIntegration", U4) - Go - Backlog Git repository provider, PR link/unlink/create/status, PR list, PR watches + watcher goroutine, saved PR queries, Git credential lease resolution. Hard-bound to Backlog (see Technical Debt Signals).
- `internal/connection` - library - Go - the single Backlog connection per workspace (space host, API key or OAuth, selected projects, epoch), the integration switch, and the Git credential secret `backlog.git.<workspaceId>`; publishes in-process `ConnectionChanged` events.
- `internal/plugin` - adapter - Go - the only `pluginsdk` importer; action routing table, error-code mapping, host-store adapters, `ResolveGitCredential` / `GetGitCredentialBinding` RPCs, `HostPort` to Kandev repositories/tasks.
- `internal/backlog` - HTTP client - Go - Backlog REST v2 + Git smart-HTTP probe; requests are pinned to `https://<creds.SpaceHost>` (`client.go:260-265`).
- `internal/issues` - library - Go - Backlog issues (U3); depends on `backlog`, `connection`, `redact`; independent of `internal/git`.
- `ui/src/git` - UI module - TypeScript/JSX - Kandev repository provider and review provider registrations, PR link task action, create-PR flow, PR list, watch form, Git access form.
- `ui/src/settings` - UI module - TypeScript/JSX - the Integrations settings screen with framed sections (connection, PR watches, issue watches, saved queries, quick actions, issue sync, Git access, projects).

Internal import graph (non-test): `plugin -> {connection, git, issues, backlog, redact, pluginsdk}`; `git -> {connection, backlog, redact}`; `issues -> {connection, backlog, redact}`; `connection -> {backlog, redact}`; `backlog -> {redact}`.

### Build System
- **Type**: Go modules (Go 1.26.0) + npm/esbuild for the UI bundle; Makefile drives both.
- **Config Files**: `go.mod` (`replace github.com/kandev/kandev => ../kandev/apps/backend`), `.kandev-sdk-ref` (`f099a46d...` = Kandev v0.96.0), `Makefile` (targets `check-sdk`, `check-format`, `vet`, `lint`, `test`, `coverage`, `check-secrets`, `ui-build`, `build`, `package`, `verify-package`, `contract-test`, `release-preflight`, `marketplace-entry`), `ui/package.json`, `ui/tsconfig.json` (path alias `@kandev/plugin-sdk -> ../../kandev/apps/packages/plugin-sdk/src/index.ts`), `manifest.yaml`.
- **Build Dependencies**: `server/main.go -> internal/plugin -> internal/{git,issues,connection,backlog,redact}`; UI bundle -> `@kandev/plugin-sdk` source via path alias.

### APIs Discovered
- Plugin actions (manifest.yaml) - 55 action keys total. Source-control relevant (22): `connection.set_git_credential`, `repositories.inspect`, `repositories.branches` (Kandev-called names for a repository provider), `git.repositories.list`, `git.prs.{link,unlink,create,status,list}`, `git.links.list`, `git.impact`, `git.watches.{list,save,delete,run,pause,resume}`, `git.queries.{list,save,delete,run,set_default}`. All are workspace or task scoped; only `connection.set_git_credential` is `admin`.
- Plugin gRPC extension (Go) - `internal/plugin/credential.go` - 2 methods: `ResolveGitCredential`, `GetGitCredentialBinding` (implements `pluginsdk.GitCredentialHandler`).
- Kandev UI registrations - `ui/src/index.ts:49-75` - `registerIntegrationSettings`, `registerRepositoryProvider`, `registerReviewProvider`, `registerTaskAction` (PR link), plus issue-side registrations.
- Manifest declarations - `repository_providers: ["nulab-backlog"]` (`manifest.yaml:31`), `capabilities.secrets: true`, `api_read: [tasks, repositories]`, `config_schema` with operator-level `oauth_client_id`, `oauth_client_secret`, `public_base_url` (`manifest.yaml:282`), webhook `oauth-callback`.
- Outbound: Backlog REST `/api/v2/projects/{key}/git/repositories[/{repo}/pullRequests[/count|/{n}]]`, `/api/v2/users/myself`, `/api/v2/issues/{key}`, and Git smart-HTTP `/git/{PROJ}/{repo}.git/info/refs?service=git-upload-pack` (Basic auth) for the Git check.

### Frameworks & Libraries
- Kandev plugin SDK (Go `pkg/pluginsdk`) - v0.96.0 (commit f099a46d) - plugin runtime, Host (state, secrets, config, tasks, repositories), Git credential extension.
- `@kandev/plugin-sdk` (TS, source alias) - v0.96.0 - UI registry, host UI components (`ChangeRequestDetail`, `IntegrationRepositoryFilter`, `SettingsSection`, ...).
- `github.com/stretchr/testify` - v1.12.1 - test assertions.
- `gopkg.in/yaml.v3` - v3.0.1 - manifest parsing in tests/CI tooling.
- UI dev: TypeScript ~6.0.3, esbuild ^0.28.2, Vitest ^5.0.3, ESLint ^10.12.0, Prettier ^3.9.9, React 19 types, jsdom, axe-core.
- No HTTP client library and no vendor SDK (stdlib `net/http`), consistent with the team rule.

### Test Coverage
- **Test Directories**: co-located `*_test.go` in every `internal/*` package (95 Go test files; `internal/git` has 17 incl. `harness_test.go`, `fakes_test.go`, `leak_test.go`; `internal/connection` has 20 incl. `git_credential_test.go`); `internal/backlog/testdata/` JSON fixtures; co-located `*.test.ts(x)` in `ui/src` (31 files; `ui/src/git` 7, `ui/src/settings` 7).
- **Test Frameworks**: Go `testing` + testify `require`, `httptest` fake Backlog; Vitest + jsdom (+ axe-core) for the UI.
- **Coverage Config**: present - `make coverage` (80% floor over `./internal/...` and `./server/...`).
- **Baseline**: NOT recorded in this run. The Go toolchain is not installed in this environment (`go: command not found`), `ui/node_modules` is absent, and this worktree's parent has no `../kandev` checkout, so the `replace` in `go.mod` and the UI `tsconfig` path alias do not resolve here (a v0.96.0 checkout exists at `/home/k_do_webfrontier/repo/kandev`).

### Code Quality Indicators
- **Linting**: `golangci-lint` + `gosec` (`.golangci.yml`), `go vet`, `gofmt`; UI `tsc --noEmit` strict, ESLint (`ui/eslint.config.js`), Prettier (`ui/.prettierrc`).
- **CI/CD**: `.github/workflows/ci.yml`, `.github/workflows/release.yml`; packaged-host contract test via `make contract-test`.
- **Documentation**: every Go package has a `doc.go`; exported identifiers carry doc comments with requirement IDs (ACx.y, BRx, FRx); secret-bearing types implement `String`/`GoString`/`Format` to hide secrets (`connection.GitCredential`, `backlog.Credentials`, `backlog.PullRequest`).

### Technical Debt Signals
These are the places where the code assumes the only source-control service is Backlog Git. Each one must change, or be wrapped, to support more services:
- **One provider id everywhere**: `git.ProviderID = "nulab-backlog"` (`internal/git/types.go:19`) is used as the repository provider, review provider and credential provider id. `PLUGIN_ID = "nulab-backlog"` (`ui/src/switch/enabled-events.ts:3`) is also the id of the UI repository/review provider registrations. The manifest declares exactly one `repository_providers` entry (`manifest.yaml:31`).
- **The Git domain model uses Backlog identity**: `Link`, `Watch`, `Query`, `RepoRef`, `Reference` are keyed by `SpaceHost` + Backlog `ProjectKey` + repo name + Backlog numeric `RepositoryID` (`internal/git/store.go` Link/Watch; `types.go` RepoRef/Reference). `LinkKey = spaceHost|repositoryId|number` (`types.go` `LinkKey`). The validation regexes are Backlog-specific: `projectKeyRe ^[A-Z][A-Z0-9_]{0,24}$`, the clone path `/git/<PROJ>/<repo>`, and the PR URL `/git/<PROJ>/<repo>/pullRequests/<n>`.
- **Backlog-specific behaviour hard-wired in `git.Service`**: the PR state map `open=1, closed=2, merged=3` (`types.go` `stateIDs`); `DefaultBranch` guesses the default branch from recent PR bases, falling back to `"master"`, and `Branches` lists branches from the newest 100 PRs because Backlog has no branch or default-branch API (`service.go:222-290`); `RelatedIssueKey` and the `PRDescription` "Related: KEY" rule; PR create attaches a Backlog `IssueID`.
- **One gateway interface, one implementation**: `git.Gateway` (`service.go:59-67`) takes `backlog.Credentials` and returns `backlog.*` types. `internal/plugin/runtime.go` builds one `backlog.Client` that satisfies the `connection`, `git` and `issues` gateways together (`runtime.go` `type gateway interface`).
- **Git scope comes from the issue tracker connection**: every Git operation reads `conn.Current()` (Backlog `SpaceHost` + `SelectedProjects`) and `conn.Credentials()` (Backlog API key/OAuth token). Repositories, PR links, watches and credential leases are only allowed for the connected Backlog space and its selected issue projects (`service.go` `ListRepositories`, `Inspect`, `Branches`, `Link`, `CreatePR`, `ResolveCredential`; `events.go` `apply` disables items whose host/project are no longer covered).
- **One Git credential per workspace, tied to the Backlog space**: secret key `backlog.git.<workspaceId>` holds `{username, password, spaceHost, revision}` (`connection/store.go:43`, `git_credential.go` `gitSecret`). It counts as valid only when `spaceHost` equals the connected Backlog host (`store.go:683-685`). It is deleted on disconnect and on a space change (`store.go:273, 383`). The binding is `<connectionEpoch>.<revision>` (`git_credential.go` `GitCredential`). `connection.test` probes Git access only against Backlog smart-HTTP (`git_credential.go` `gitCheck`; `backlog/client.go:174`).
- **Settings has a single Git section**: the "Git access" section (`ui/src/settings/SettingsScreen.tsx:398-409`) is a username/password form for Backlog Git only. It appears only while Backlog is connected and only for admins. No setting chooses a provider, a repository, or an owner/workspace for another service. The repository picker comes from Backlog's selected projects (`git-state.ts` `loadRepoOptions`; `pr-list.tsx` `repo: "PROJ/repo"`).
- **UI URL matcher is Backlog-only**: `BACKLOG_GIT_URL` (`ui/src/git/repository-provider.ts`) matches only `*.backlog.com|backlog.jp|backlogtool.com/git/`.
- **Action names are not namespaced by provider**: `git.*` and `repositories.*` actions take no provider argument. `repositories.inspect/branches` are the fixed names Kandev calls for this plugin's provider, so a second provider in the same plugin needs payload-level dispatch on `provider_id` (the descriptor already carries `provider_id`).
- **Error mapping is Backlog-typed**: `git.errorCode`/`isKind` and `connection.Classify` read `*backlog.Error` kinds. A second vendor client needs its own error type mapped into the same codes, or a shared error type.
- Store documents are single capped JSON lists per workspace (`git.links` max 500, `git.watches` 50, `git.queries` 50, `schemaVersion: 1`) and have no provider field. Adding a provider dimension needs a schema bump or a field that defaults to Backlog when absent, so stored v0.3.0 data still reads correctly.

## Handoff Summary
- **Intent-relevant finding**: Source control is not a separate concept in the code today. "Git" means "Backlog Git of the connected Backlog space". `internal/git.Service` depends on `connection.Connection` (Backlog space + selected projects + Backlog API credentials) and on a `Gateway` typed in `backlog.*` (`internal/git/service.go:59-78`). Its identity model (`SpaceHost|RepositoryID|Number`, `ProjectKey`, `/git/PROJ/repo` paths) and one Git secret per workspace (`backlog.git.<ws>`, valid only for the connected Backlog host: `internal/connection/store.go:43,683`; `git_credential.go`) are Backlog-specific. The Kandev SDK side is already provider-neutral and supports this intent: a plugin may declare several `repository_providers` (manifest validator `internal/plugins/manifest/validate.go:664`), register one UI repository provider and review provider per declared id (`apps/web/lib/plugins/registry.ts:481,504` with `registry-provider-ownership.ts` `claim`), and receive `ResolveGitCredential` / `GetGitCredentialBinding` with `ProviderID`, `Host` and `Path` per request (`pkg/pluginsdk/plugin.go:51-71`, `types.go:403`). The plugin already rejects any `ProviderID != "nulab-backlog"` (`internal/git/service.go:612-614, 649`), so this is the point where a per-provider dispatch would go.
- **Risks / follow-up**:
  - **Host-reserved provider ids**: Kandev v0.96.0 reserves `github`, `gitlab` and `azure_devops` (`apps/web/lib/plugins/registry-provider-ownership.ts:6`). Kandev also has native GitHub/GitLab/Azure DevOps integrations, and its GitHub credential resolver runs before the plugin resolvers (`internal/backendapp/git_credentials.go:42-48`). The plugin therefore cannot register a provider called `github`. It needs plugin-scoped ids (for example `nulab-backlog-github`), or it must link to Kandev's native GitHub repositories (provider id `github`) only as PR references, without owning cloning or credentials. Requirements must decide this.
  - **Provider ownership is exclusive across active plugins**: `ensureOwnershipAvailable` (`internal/plugins/service_lifecycle.go:117-121`) refuses to activate a plugin if another active plugin declares the same provider id. A `bitbucket` id would conflict with a separate Bitbucket plugin, so plugin-prefixed ids are safer.
  - **The project rule restricts hosts**: "ALWAYS accept only `https` space addresses under `backlog.com`, `backlog.jp`, or `backlogtool.com`" (project.md Mandated). The `backlog.Client` also pins requests to the space host. Calling GitHub/Bitbucket hosts needs a new client package (not `internal/backlog`) and an explicit rule decision that this mandate covers only the Backlog space address. The team rule also asks for stdlib-only HTTP with no vendor SDK.
  - **Backward compatibility of stored data**: v0.3.0 installs hold `git.links/watches/queries` documents and the `backlog.git.<ws>` secret with no provider field. Any new schema must read these as the Backlog Git provider. The `nulab_backlog_pr` task-metadata key (`internal/git/host.go:9`) and `reviewKey` format already exist on live tasks.
  - **Scope coupling**: Git items are switched off by Backlog `ConnectionChanged` events (project deselect, space change, disconnect - `internal/git/events.go`). Requirements must decide whether a GitHub/Bitbucket link still depends on the Backlog connection and the integration switch: today every guarded action and both credential RPCs fail closed while the switch is off (`internal/plugin/runtime.go` `guarded`, `credential.go`).
  - **Issue to PR relation**: `RelatedIssueKey` and the `IssueID` attachment on PR create are Backlog-native features. For external providers, the Backlog issue key can only go in the PR text, so issue to PR traceability differs by provider.
  - **New per-provider settings and secrets**: each provider needs its own secret keys (only `capabilities.secrets: true` is needed, no manifest change for secrets). Operator-level OAuth app credentials for a new provider would go in `config_schema`. Per-workspace settings (auth, owner/workspace, repo allowlist) would follow the existing admin-gated action pattern (`connection.set_git_credential`, `access: admin`).
  - **No test baseline**: the Go toolchain, `ui/node_modules` and a `../kandev` link are missing in this worktree, so no test baseline was recorded. Per the project Testing Posture correction, install Go 1.26.x and link `../kandev` to v0.96.0 (available at `/home/k_do_webfrontier/repo/kandev`) before Construction.
