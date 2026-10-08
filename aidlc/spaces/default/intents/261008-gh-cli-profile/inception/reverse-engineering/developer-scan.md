# Developer Code Scan: kandev-plugin-nulab-backlog (intent 261008-gh-cli-profile)

Scan type: full rescan of `./` at depth Minimal, with extra depth on the intent area (gh CLI credential, per-workspace source control settings, task creation from Backlog issues). Base commit: `ca8146c` (v0.5.2 records) on branch `feature/gh-cli-profile-scope-q1o`. Scan date: 2026-10-08.

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - internal/scm/cli_token.go
  - internal/scm/service.go
  - internal/scm/types.go
  - internal/scm/store.go
  - internal/scm/client.go
  - internal/github/client.go
  - internal/plugin/scm_actions.go
  - internal/plugin/runtime.go
  - internal/plugin/host_port.go
  - internal/plugin/credential.go
  - internal/issues/service.go
  - internal/issues/types.go
  - ui/src/settings/source-control-section.tsx
  - ui/src/page/start-task.tsx
  - manifest.yaml
  - go.mod
  - .kandev-sdk-ref
- **Skimmed only**:
  - internal/backlog/, internal/bitbucket/, internal/gitlab/ (provider HTTP clients, same `scm.Client` shape)
  - internal/ci/, cmd/ci/, cmd/verifypkg/, internal/pkgverify/ (CI and package-verification tooling)
  - internal/connection/ (Backlog connection, API key / OAuth, Backlog Git credential)
  - internal/git/ (Backlog Git: PRs, links, watches, `ResolveCredential` for provider `nulab-backlog` only)
  - internal/issues/ other files (sync, watches, quick actions, links)
  - internal/scm/ other files (prs.go, queries.go, links.go, watcher.go, httpx.go, errors.go)
  - internal/plugin/ other files (git_actions.go, issue_actions.go, webhook.go, references.go, events.go, config.go)
  - internal/redact/, internal/testutil/, server/
  - ui/src/ other directories (brand, git, issues, page, switch, messages, testing)
  - .github/workflows/, Makefile, .golangci.yml
- **External evidence (not coverage)**: Kandev checkout at `~/repo/kandev` (tag v0.96.0, commit f099a46dc, matches `.kandev-sdk-ref`): `apps/backend/internal/github/gh_accounts.go`, `apps/backend/internal/github/auth_resolver.go`, `apps/backend/internal/orchestrator/executor/executor_credentials.go`, `apps/backend/pkg/pluginsdk/{host.go,data_types.go,plugin.go,types.go}`.

### Packages Found
- `server` — executable entry — Go — `main.go` calls `pluginsdk.Serve(plugin.NewRuntime())`
- `internal/plugin` — adapter — Go — only package importing `pluginsdk`; action routing (`HandleAction`), Host adapters (`hostStores`, `hostPort`, `issueHost`, `scmHost`), Git credential handler, webhooks
- `internal/connection` — service — Go — Backlog space connection (API key, OAuth), on/off switch, project selection, Backlog Git credential
- `internal/backlog` — client — Go — Backlog REST client (stdlib HTTP, rate limiter, OAuth)
- `internal/issues` — service — Go — issue list/detail, task creation from issue, links, sync, issue watches, quick actions
- `internal/git` — service — Go — Backlog Git PRs, links, saved queries, PR watches, Git credential leases for `nulab-backlog` repositories
- `internal/scm` — service — Go — GitHub/GitLab/Bitbucket source control: token or CLI credential, repo mappings, PR list/link, queries, watches
- `internal/github`, `internal/gitlab`, `internal/bitbucket` — clients — Go — provider REST clients implementing `scm.Client`
- `internal/redact` — utility — Go — secret redaction for logs/errors (`WithSecrets`, redacting slog handler)
- `internal/pkgverify`, `internal/ci`, `cmd/verifypkg`, `cmd/ci` — tooling — Go — package verification and CI checks (workflows, manifest, release, marketplace, secrets)
- `internal/testutil` — test helper — Go
- `ui` — frontend bundle — TypeScript/TSX — Kandev plugin UI (Settings screen, Backlog page, issue panel, start-task menu)

### Build System
- **Type**: Go modules (Go 1.26) + Makefile; npm for `ui/` (Vitest, tsc, ESLint, Prettier)
- **Config Files**: `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`, `.kandev-sdk-ref`, `manifest.yaml`, `ui/package.json`, `ui/tsconfig.json`, `ui/vitest.config.ts`, `ui/eslint.config.js`, `ui/.prettierrc`, `.nvmrc`
- **Build Dependencies**: `server` → `internal/plugin` → {`connection`, `backlog`, `git`, `issues`, `scm`, `github`, `gitlab`, `bitbucket`, `redact`}; `scm` → {`connection`, `redact`}; `github`/`gitlab`/`bitbucket` → `scm`; `git`, `issues` → {`backlog`, `connection`}. `go.mod` replaces `github.com/kandev/kandev => ../kandev/apps/backend`; `../kandev` is absent in this worktree.

### APIs Discovered
- Plugin browser actions — `manifest.yaml` `actions:` + `internal/plugin/*_actions.go` — about 60 keys; 23 `scm.*` keys including `scm.providers.list|set_token|test|remove|use_cli`
- Git credential gRPC — `internal/plugin/credential.go` — `ResolveGitCredential`, `GetGitCredentialBinding` (provider `nulab-backlog` only, `manifest.yaml` `repository_providers: ["nulab-backlog"]`)
- Webhooks — `internal/plugin/webhook.go`, `manifest.yaml` `webhooks:`
- Outbound: Backlog REST, GitHub REST (`/user`, `/user/repos`, repos, pulls), GitLab REST, Bitbucket REST
- Kandev Host data API used — `Tasks().Create/List/Get`, `Repositories().List`, secrets, state, config (`api_read: [tasks, repositories]`, `api_write: [tasks]`)

### Frameworks & Libraries
- `github.com/kandev/kandev/pkg/pluginsdk` — v0.96.0 (commit f099a46dc via replace) — plugin runtime, Host API, hashicorp go-plugin gRPC
- `github.com/stretchr/testify` — v1.12.1 — test assertions
- `gopkg.in/yaml.v3` — v3.0.1 — manifest parsing in tooling/tests
- `@kandev/plugin-sdk` (UI) — Host UI components (`TaskCreateDialog`, `IntegrationStartTaskMenu`), `host.api.invokeAction`
- `vitest` ^5.0.3 — UI tests

### Test Coverage
- **Test Directories**: `*_test.go` beside sources in every `internal/*` package; fixtures in `internal/*/testdata/`; UI tests `ui/src/**/*.test.ts(x)`
- **Test Frameworks**: Go `testing` + testify (`-race` in CI), `httptest` fake servers; Vitest for UI
- **Coverage Config**: present — `make coverage` with 80% floor over `./internal/...` and `./server/...`
- **Go baseline (this scan)**: `go test -count=1 ./...` passed, 13 packages with tests, 1383 passing test results (tests plus subtests), 0 failures, 0 skips. Run with Go 1.26.8 and a scratch `-modfile` whose `replace` pointed at `~/repo/kandev/apps/backend` (v0.96.0), because `../kandev` is missing; no repo file was changed. Not run with `-race`. UI Vitest suite not run.

### Code Quality Indicators
- **Linting**: golangci-lint with gosec (`.golangci.yml`), `gofmt`, `go vet`; ESLint + Prettier + `tsc --noEmit` strict in `ui/`
- **CI/CD**: `.github/workflows/ci.yml`, `release.yml`, `secrets.yml`
- **Documentation**: README present; package `doc.go` in each package; doc comments reference FR/NFR ids from earlier intents

### Technical Debt Signals
- `internal/scm/cli_token.go:28-36` — `cliCommand` is fixed to `gh auth token --hostname github.com` (and `glab config get token --host gitlab.com`) with no `--user`, so it always returns the gh active account's token.
- `internal/scm/cli_token.go:59-63` — `cliCache` is keyed by provider only and documented as "not per workspace"; one lock covers every CLI run (`ponytail:` comment).
- `internal/scm/service.go:250-271` (`UseCLI`) and `:276-298` (`credential`) — the CLI path reads whoever is active at call time; the stored `Account`/`AccountID` is not used to pick the account.
- `internal/scm/service.go:310-335` (`Test`) — in CLI mode, a test after the user runs `gh auth switch` silently rewrites `Account`/`AccountID` to the new active login, so the workspace's identity drifts and the PR "me" filter (`prs.go:61`, `watcher.go:58`) follows it.
- `internal/plugin/host_port.go:120-130` (`issueHost.CreateTask`) and `scm_actions.go:268-275` (`scmHost.CreateTask`) — plugin-created tasks pass no `Repositories` or `Launch`; worktree and agent environment are fully decided by Kandev.
- `internal/plugin/host_port.go:139` — `workflowRefused` matches gRPC text (`ponytail:` comment).

## Handoff Summary
- **Intent-relevant finding**: The plugin's gh CLI credential is the gh *active account*, server-wide. `internal/scm/cli_token.go:31` runs `gh auth token --hostname github.com` with no `--user`, cached per provider for 5 minutes (`cliTTL`, `cliCache` at `:59-63`). Per-workspace state already exists: `scm.Settings` per workspace (`internal/scm/store.go:42-50`, state key `scm.settings`) stores `Source: "cli"`, `Account` (display name) and `AccountID` (GitHub login, `internal/github/client.go:54-64`). So a per-workspace profile needs: (1) list accounts (`gh auth status --json hosts`), (2) a chosen login per workspace in `Settings`, (3) `gh auth token --hostname github.com --user <login>` in `cliCommand` (gh 2.97.0 on this host supports `--user`), (4) a cache key of provider+login, (5) a new action key (e.g. `scm.providers.cli_accounts`) in `manifest.yaml` and `scm_actions.go`, and (6) an account picker beside "Use gh CLI login" in `ui/src/settings/source-control-section.tsx:233,302`. Kandev v0.96.0 does exactly this for its own GitHub integration (`~/repo/kandev/apps/backend/internal/github/gh_accounts.go`: `ListGHAccounts` parses `gh auth status --json hosts`; `ResolveGHAccountToken(host, login)` uses `--user` when `gh auth token --help` lists it, otherwise requires the login to be active; it strips `GH_TOKEN`/`GITHUB_TOKEN` from the child env and never switches the active account).
- **Risks / follow-up**:
  - **Worktree gh CLI is outside the plugin's reach.** A task from a Backlog issue is created either by Kandev's own `TaskCreateDialog` (`ui/src/page/start-task.tsx:100-115`, plugin only links afterwards) or by `Tasks().Create` with no repositories/launch (`host_port.go:120-130`). The agent's `GH_TOKEN`/`GITHUB_TOKEN` in the worktree is set by Kandev's executor from Kandev's own workspace GitHub connection or the executor profile env (`~/repo/kandev/apps/backend/internal/orchestrator/executor/executor_credentials.go`, `executor_host_gh_bridge.go`). pluginsdk v0.96.0 offers no env injection: `CreateTaskInput` has only `Repositories` and `Launch{AgentProfileID, ExecutorProfileID, Prompt, PlanMode}`; `ExecutorProfiles()` is read-only; the plugin Git credential handler only serves `nulab-backlog` repositories. The requirement "gh CLI in the worktree uses the chosen profile" therefore cannot be met by plugin code alone for GitHub repositories. Options to put to the user: rely on Kandev's GitHub integration set to the same gh account; let the plugin pick an executor profile (`Launch.ExecutorProfileID`) whose env carries the right token (only for tasks the plugin creates itself, not `TaskCreateDialog`); or document it as a limitation.
  - Existing CLI-mode records have no chosen login; the migration default must be explicit (for example keep `AccountID` as the chosen login, which preserves today's account).
  - gh older than the `--user` flag: follow Kandev's fallback (require the login to be active, else `cli_unavailable`).
  - Never echo gh stdout/stderr (`gh auth status` can print token locations); keep NFR1 redaction and the no-shell, 10 s, 4 KiB limits in `runCLI`.
  - GitLab (`glab`) has the same single-account behaviour; the intent names only gh, so scope GitLab out explicitly or ask.
  - Test baseline needs `../kandev` → v0.96.0; this worktree lacks it (project rule), so later stages should link it or use the same scratch `-modfile` approach.
