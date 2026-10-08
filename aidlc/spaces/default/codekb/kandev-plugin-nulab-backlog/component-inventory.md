# Component Inventory — kandev-plugin-nulab-backlog

Health: healthy / at-risk. "At-risk" means touched by intent `261008-gh-cli-profile`. Reading depth per component: [reverse-engineering-timestamp.md](reverse-engineering-timestamp.md#scope-of-analysis).

## Runtime Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())`. Health: healthy.

### KandevAdapter
- `internal/plugin/`. Action routing, Host adapters (`hostPort`, `issueHost`, `scmHost`), webhooks, events, Backlog Git credential. Only package (with `server`) importing `pluginsdk`. Builds SCM clients (`scmClients`), wires SCM (`wireSCM`), maps errors (`classify`, `classifySCM`). Depends on: Connection, Issues, Git, SCM, GitHub Client, GitLab and Bitbucket Clients, BacklogGateway, Redact. Health: at-risk (new `scm.*` action and manifest entry; task creation passes no `Repositories`/`Launch`).

### BacklogGateway
- `internal/backlog/`. Backlog REST v2 client, rate limiter, OAuth. Depends on: Redact. Health: healthy.

### Connection
- `internal/connection/`. Backlog space connection (API key, OAuth), switch, project selection, Backlog Git credential; supplies `FieldError`, `ErrStore`, outcome codes reused by SCM. Depends on: BacklogGateway, Redact. Health: healthy.

### Issues
- `internal/issues/`. Issue list/detail, task creation from issues and watches, task search, links, sync, quick actions. Depends on: BacklogGateway, Connection, Redact. Health: healthy (task creation path is relevant context, not changed by the plugin-side part of the intent).

### Git
- `internal/git/`. Backlog Git PRs, links, queries, watches, Git credential leases for `nulab-backlog`. Depends on: BacklogGateway, Connection, Redact. Health: healthy.

### SCM
- `internal/scm/`. GitHub/GitLab/Bitbucket per workspace: `Settings`, token or CLI credential (`credential`, `cliToken`, `cliCache`), repo mappings, PR lists, links, queries, watches, `Watcher`. Depends on: Connection, Redact; providers via `scm.Client`; host via ports. Health: at-risk (CLI command has no account selector, server-wide cache, identity drift in `Test`). Details: [architecture.md](architecture.md#scm-provider-connection-and-credentials).

### GitHub Client
- `internal/github/`. Read-only GitHub REST client implementing `scm.Client`; `CurrentUser` returns `login` as `ID`. Depends on: SCM. Health: healthy (credential-agnostic).

### GitLab and Bitbucket Clients
- `internal/gitlab/`, `internal/bitbucket/`. Same shape. Depends on: SCM. Health: healthy (GitLab's `glab` has the same single-account CLI behaviour; out of scope unless requirements say otherwise).

### Redact
- `internal/redact/`. `WithSecrets`, redacting slog handler. Health: healthy.

## Build-Time Components

### CI Workflows
- `.github/workflows/ci.yml`, `release.yml`, `secrets.yml`. SHA-pinned actions. Depends on: Build Makefile. Health: healthy.

### Build Makefile
- `Makefile`, `.golangci.yml`. Depends on: CI Tooling, PackageVerify, sibling `../kandev`. Health: healthy.

### PackageVerify
- `internal/pkgverify/`, `cmd/verifypkg/`. Offline package verifier. Health: healthy.

### CI Tooling
- `internal/ci/`, `cmd/ci/`. Workflow policy, secret scan, contract test, release preflight, marketplace. Health: healthy.

### TestUtil
- `internal/testutil/`. Test helpers. Health: healthy.

## UI Components (TypeScript)

### UI Bundle
- `ui/src/` -> bundle. Settings (incl. Source control `ProviderCard` with the CLI login button), Backlog page, issue panel and badge, start-task menu (`page/start-task.tsx`, Kandev `TaskCreateDialog`), link actions. Host React and UI kit; no plugin CSS. Health: at-risk (GitHub card needs an account picker). File map: [code-structure.md](code-structure.md#ui-source-uisrc--intent-area).
