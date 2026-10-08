# Component Inventory — kandev-plugin-nulab-backlog

Health: healthy / at-risk. "At-risk" means touched by the active intent `261008-link-task-modal` (UI Bundle, Issues, KandevAdapter); notes on SCM credential work come from intent `261008-gh-cli-auth` (implemented in v0.5.1, PR #19). Depth of reading per component: [reverse-engineering-timestamp.md](reverse-engineering-timestamp.md#scope-of-analysis). Components not re-read since an earlier run keep that run's description.

## Runtime Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())`. Health: healthy.

### KandevAdapter
- `internal/plugin/`. Action/webhook/event handlers, host port (secrets, state), references, Backlog Git credential (`credential.go`, `ResolveGitCredential`); only package (with `server`) importing `pluginsdk`. `issue_actions.go` routes `issues.tasks.search`, `issues.link`, `issues.unlink`, `issues.links.list`; `runtime.go` maps domain errors to SDK codes. Builds the SCM clients (`scmClients`) and wires SCM and its watcher (`wireSCM`); maps SCM errors in `classifySCM`. Depends on: Connection, Issues, Git, SCM, GitHub Client, GitLab and Bitbucket Clients, BacklogGateway, Redact. Health: at-risk only if `issues.link` must accept an issue URL (see [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-link-task-modal)). The `gh-cli-auth` scan needed a new `scm.*` admin action, its manifest entry and an error mapping for an unavailable CLI; v0.5.1 added `scm.providers.use_cli`.

### BacklogGateway
- `internal/backlog/`. Backlog REST v2 client and OAuth token. Depends on: Redact. Health: healthy.

### Connection
- `internal/connection/`. Address validation, API key and OAuth connect, per-workspace enable switch, lifecycle, project selection, Backlog Git credential, persisted state. Also supplies `FieldError`, `ErrStore` and the `Code*` outcome codes that SCM and KandevAdapter reuse. Depends on: BacklogGateway, Redact. Health: healthy (Backlog-only; not the GitHub link).

### Issues
- `internal/issues/`. Issue list/filters, tasks from issues, task search for linking, links (`Link` with `Summary`, `LinkView`, `Link`/`Unlink`, `ParseIssueKey`), status sync, watches, quick actions, saved queries. Depends on: BacklogGateway, Connection, Redact. Health: at-risk (`issues.link` takes an issue key only).

### Git
- `internal/git/`. Backlog Git repository provider, PR links/create/status, watches, queries. Depends on: BacklogGateway, Connection, Redact. Health: healthy (unaffected by the current intent).

### SCM
- `internal/scm/`. Provider-neutral source control for GitHub, GitLab and Bitbucket: provider settings (`Settings`), token storage in host secrets, the single credential read path (`credential()`), repo search, project-to-repo mappings, PR lists, links, queries, watches and the background `Watcher`. Depends on: Connection (error/outcome types), Redact; reaches the provider clients only through the `scm.Client` interface and host secrets/state through ports. Health: at-risk (the credential source, `Settings`, `ProviderView`, `RemoveToken` and the `ErrNoToken` wording change for GitHub CLI login). Details: [architecture.md](architecture.md#scm-provider-connection-and-credentials).

### GitHub Client
- `internal/github/`. Read-only GitHub REST client (`scm.Client` implementation) for `https://api.github.com` with `Authorization: Bearer <token>`; `httptest` fixtures in `testdata/`. Depends on: SCM (`scm.API`, `scm.Credential`). Health: healthy (credential-agnostic; no change needed for a token from `gh`).

### GitLab and Bitbucket Clients
- `internal/gitlab/`, `internal/bitbucket/`. Read-only REST clients implementing `scm.Client`, same pattern as the GitHub Client. Depends on: SCM. Health: healthy (token-only; out of scope unless requirements say otherwise).

### Redact
- `internal/redact/`. Context-scoped secret masking (`WithSecrets`, `Logger`) and Backlog URL query-string masking. Health: healthy.

## Build-Time Components

### CI Workflows
- `.github/workflows/ci.yml` (jobs `checks`, `packaged-host-contract`), `.github/workflows/release.yml` (jobs `verify`, `contract`, `publish`) and `secrets.yml` (credential scan). GitHub Actions, every action pinned to a full SHA. Depends on: Build Makefile. Health: healthy. Details (as of run 1): [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).

### Build Makefile
- `Makefile`. Standard targets that CI and release call (`check-format`, `vet`, `lint`, `test`, `coverage`, `check-secrets`, `build`, `package`, `verify-package`, `contract-test`, `release-preflight`). Depends on: CI Tooling, PackageVerify, the sibling `../kandev` checkout. Health: healthy. Targets: [code-structure.md](code-structure.md#build-and-packaging).

### PackageVerify
- `internal/pkgverify/`, `cmd/verifypkg/`. Offline package verifier; hard-codes the executable set; no size check. Health: healthy.

### CI Tooling
- `internal/ci/`, `cmd/ci/`. Contract test driver (loopback install, 30 s client timeout), manifest, marketplace, release preflight, secret scan and workflow policy; `changes.go` (CI change detection) held the repo's only `os/exec` call before v0.5.1 added `internal/scm/cli_token.go`. Health: healthy.

### TestUtil
- `internal/testutil/`. Test helpers. Health: healthy.

## UI Components (TypeScript)

### UI Bundle
- `ui/src/` -> `ui/bundle.js`. Registration in `index.ts`; areas `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, test harness `testing/`. React and the UI kit from the host; no plugin CSS. Link-related parts: `issues/link-task-dialog.tsx` (issue-side task picker, the modal this intent changes), `issues/issues-page.tsx` (opens it), `issues/task-menu.ts` (Unlink), `issues/links-store.ts`, `git/pr-link.ts` (task-side host-dialog pattern). Health: at-risk — the Link Task modal differs from the GitHub link dialog in direction, input, copy, error styling and success feedback. Details: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-link-task-modal). The Source control section (`settings/source-control-section.tsx`, `ProviderCard`) offered only a token form at the `gh-cli-auth` scan; v0.5.1 added the "Use gh / glab CLI login" control with matching `scm*` messages and the `ProviderView` TS type. File map: [code-structure.md](code-structure.md#ui-source-uisrc).
