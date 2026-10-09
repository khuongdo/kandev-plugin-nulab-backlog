# Component Inventory — kandev-plugin-nulab-backlog

Health: healthy / at-risk. "At-risk" means touched by intent `261009-no-workflow-error`. SCM and settings notes below predate v0.6.0 (one source control service per workspace) and were not re-read. Reading depth per component: [reverse-engineering-timestamp.md](reverse-engineering-timestamp.md#scope-of-analysis).

## Runtime Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())`. Health: healthy.

### KandevAdapter
- `internal/plugin/`. Action routing (`guarded()` lets only `scm.providers.list` through while Backlog is off), Host adapters (`hostPort`, `issueHost`, `scmHost`), webhooks, events, Backlog Git credential. Only package (with `server`) importing `pluginsdk`. Builds SCM clients (`scmClients`), wires SCM (`wireSCM`), maps errors (`classify`, `classifySCM`). Depends on: Connection, Issues, Git, SCM, GitHub Client, GitLab and Bitbucket Clients, BacklogGateway, Redact. Health: healthy (not on the "no workflow" path; `issueHost.CreateTask` in `host_port.go` is used only by issue watches with a stored workflow).

### BacklogGateway
- `internal/backlog/`. Backlog REST v2 client, rate limiter, OAuth. Depends on: Redact. Health: healthy.

### Connection
- `internal/connection/`. Backlog space connection (API key, OAuth), switch, project selection, Backlog Git credential; supplies `FieldError`, `ErrStore`, outcome codes reused by SCM. Depends on: BacklogGateway, Redact. Health: healthy.

### Issues
- `internal/issues/`. Issue list/detail, task creation from issues and watches, task search, links, sync, quick actions. `issues.create_task` and issue watches need a concrete `workflowId` from the caller; the package cannot read workflows. Depends on: BacklogGateway, Connection, Redact. Health: healthy (not involved before the "no workflow" notice, which is raised in the UI).

### Git
- `internal/git/`. Backlog Git PRs, links, queries, watches, Git credential leases for `nulab-backlog`. Depends on: BacklogGateway, Connection, Redact. Health: healthy (its status under "one service at a time" is an open question).

### SCM
- `internal/scm/`. GitHub/GitLab/Bitbucket per workspace: per-provider `Settings`, token or CLI credential, repo mappings, PR lists, links, queries, watches, `Watcher`. Multi-provider by design; no active-provider concept. Depends on: Connection, Redact; providers via `scm.Client`; host via ports. Health: healthy (v0.6.0 added the one-service rule; not re-read). Pre-0.6.0 details: [architecture.md](architecture.md#scm-provider-model-multi-provider-today).

### GitHub Client
- `internal/github/`. Read-only GitHub REST client implementing `scm.Client`. Depends on: SCM. Health: healthy (not changed by this intent).

### GitLab and Bitbucket Clients
- `internal/gitlab/`, `internal/bitbucket/`. Same shape. Depends on: SCM. Health: healthy (not changed by this intent).

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
- `ui/src/` -> bundle. Settings (incl. the Source Control section with Backlog Git block and one `ProviderCard` per provider), Backlog page, issue panel and badge, PR list and watch forms with provider selectors, start-task menu, link actions. Host React and UI kit; no plugin CSS. Health: at-risk (start-task treats a missing host task-creation context as "no workflow"; row notices widen the non-shrinking action slot; inconsistent notice styles; watch dialogs share the context read). File map: [code-structure.md](code-structure.md#ui-page-and-issues-uisrcpage-uisrcissues--intent-area).
