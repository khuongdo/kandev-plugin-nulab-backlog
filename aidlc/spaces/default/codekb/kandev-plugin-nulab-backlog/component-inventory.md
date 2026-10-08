# Component Inventory — kandev-plugin-nulab-backlog

Health: healthy / at-risk. "At-risk" means at risk for the active intent `261008-ci-path-filter`. Depth of reading per component: [reverse-engineering-timestamp.md](reverse-engineering-timestamp.md#scope-of-analysis).

## Runtime Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())`. Health: healthy.

### KandevAdapter
- `internal/plugin/`. Action/webhook/event handlers, host port, references, credentials; only package (with `server`) importing `pluginsdk`. Depends on: Connection, Issues, Git, SCM, BacklogGateway, Redact. Health: healthy.

### BacklogGateway
- `internal/backlog/`. Backlog REST v2 client and OAuth token. Depends on: Redact. Health: healthy.

### Connection
- `internal/connection/`. Address validation, API key and OAuth connect, lifecycle, project selection, Git credential, persisted state. Depends on: BacklogGateway, Redact. Health: healthy.

### Issues
- `internal/issues/`. Issue list/filters, tasks from issues, links, sync, watches, quick actions, saved queries. Depends on: BacklogGateway, Connection, Redact. Health: healthy.

### Git
- `internal/git/`. Backlog Git repository provider, PR links/create/status, watches, queries. Depends on: BacklogGateway, Connection, Redact. Health: healthy.

### SCM
- `internal/scm/`. Provider-neutral source-control context (links, queries, watches). Depends on: SCM Clients. Health: healthy.

### SCM Clients
- `internal/github/`, `internal/gitlab/`, `internal/bitbucket/`. Read-only REST clients. Health: healthy.

### Redact
- `internal/redact/`. Masks secrets and Backlog URL query strings. Health: healthy.

## Build-Time Components

### CI Workflows
- `.github/workflows/ci.yml` (jobs `checks`, `packaged-host-contract`) and `.github/workflows/release.yml` (jobs `verify`, `contract`, `publish`). GitHub Actions, every action pinned to a full SHA. Depends on: Build Makefile. Health: at-risk (the triggers change in this intent; `checks` and `packaged-host-contract` are required status checks on `main`). Details: [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).

### Build Makefile
- `Makefile`. Standard targets that CI and release call (`check-format`, `vet`, `lint`, `test`, `coverage`, `check-secrets`, `build`, `package`, `verify-package`, `contract-test`, `release-preflight`). Depends on: CI Tooling, PackageVerify, the sibling `../kandev` checkout. Health: healthy (may be touched if release filtering moves into `release-preflight`). Targets: [code-structure.md](code-structure.md#build-and-packaging).

### PackageVerify
- `internal/pkgverify/`, `cmd/verifypkg/`. Offline package verifier; hard-codes the platform executables; no size check. Health: healthy.

### CI Tooling
- `internal/ci/`, `cmd/ci/`. Contract test driver (loopback install, 30 s client timeout), manifest, marketplace, release preflight, secret scan (whole repo, `docs/manual-checks/` allow-listed) and workflow policy (SHA pins, `contents: read`, no `pull_request_target`, write only in `release` / `publish`; does not inspect `paths`). Health: healthy.

### TestUtil
- `internal/testutil/`. Test helpers. Health: healthy.

## UI Components (TypeScript)

### UI Bundle
- `ui/src/` -> `ui/bundle.js` (226 KB in the v0.4.1 package). Registration in `index.ts`; areas `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, test harness `testing/`. React from the host. Health: healthy.
