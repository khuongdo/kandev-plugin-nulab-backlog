# Developer Code Scan: Source Control settings (intent 261008-source-control-settings)

Focused scan for the refactor of the Source Control settings page: clearer separation between services, a repo-scope label that names the service, and at most one source control service active at a time. The CodeKB store at `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` is STALE and was not edited.

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - `ui/src/settings/source-control-section.tsx`
  - `ui/src/settings/source-control-section.test.tsx` (structure, fixtures, every test name)
  - `ui/src/settings/SettingsScreen.tsx`
  - `ui/src/settings/section-parts.tsx`
  - `ui/src/settings/state.ts`
  - `internal/scm/service.go`
  - `internal/scm/store.go`
  - `internal/scm/types.go`
  - `internal/scm/client.go`
  - `internal/scm/prs.go`
  - `internal/scm/queries.go` (mapping guards, list/save)
  - `internal/scm/watcher.go` (RunWatch, RefreshLinks, runDue)
  - `internal/scm/links.go` (Link, mapping guard)
  - `internal/plugin/scm_actions.go`
- **Skimmed only**:
  - `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` (function index, base URLs, testdata list; each implements `scm.Client`)
  - `internal/scm/cli_token.go`, `internal/scm/httpx.go`, `internal/scm/errors.go` and the remaining `internal/scm/*_test.go` (test names only)
  - `internal/plugin/runtime.go` (only `guarded()` and SCM wiring), `internal/plugin/credential.go` (head only), the other `internal/plugin/*_test.go` (test names only)
  - `ui/src/settings/sections.test.tsx`, `ui/src/settings/settings.test.tsx` (grep for source-control test ids)
  - Outside the snapshot, shallow only: `ui/src/git/git-state.ts` (ProviderView, providerName, scmNotice, loadProviders, usableProviders, scmRepoOptions), `ui/src/git/pr-list.tsx` and `ui/src/git/watch-form.tsx` (provider selectors), `ui/src/git/scm-pr-list.tsx`, `ui/src/git/scm-watch-form.tsx`, `ui/src/issues/issue-prs.tsx`, `ui/src/page/start-task.tsx` (grep hits only), `ui/src/messages/en.ts` (the `scm*` keys and `sourceControlDescription`), `manifest.yaml` (scm.* action block), `Makefile` (targets), `ui/package.json`, `.github/workflows/` (file names)

### Packages Found
- `internal/scm` — library/service — Go — provider-agnostic source control component: per-workspace provider settings, token/CLI credentials, repo mappings, PR list, links, saved queries, watches, refresh watcher
- `internal/github`, `internal/gitlab`, `internal/bitbucket` — API clients — Go — read-only `scm.Client` implementations for api.github.com, gitlab.com/api/v4, api.bitbucket.org/2.0
- `internal/plugin` — adapter — Go — Kandev SDK runtime; `scm_actions.go` maps `scm.*` action keys to `scm.Service` and classifies errors
- `ui/src/settings` — UI — TypeScript (host React via `host.jsx`, no bundled React) — Backlog settings screen; `source-control-section.tsx` is the page under refactor
- `ui/src/git` (shallow) — UI — TypeScript — shared SCM view types/helpers and the PR list / watch form consumers of the provider list

### Build System
- **Type**: Go modules + Make; npm + esbuild for the UI bundle
- **Config Files**: `go.mod` (`replace github.com/kandev/kandev => ../kandev/apps/backend`), `.kandev-sdk-ref` (f099a46d = v0.96.0), `Makefile` (check-format, vet, lint, test, coverage, build, package, verify-package, contract-test, ui-build), `ui/package.json`, `ui/tsconfig.json`, `ui/vitest.config.ts`, `ui/eslint.config.js`, `ui/.prettierrc`, `manifest.yaml`
- **Build Dependencies**: `internal/plugin → internal/scm → internal/connection, internal/redact`; `internal/plugin → internal/github|gitlab|bitbucket → internal/scm`; UI `settings/source-control-section.tsx → git/git-state.ts, messages/en.ts, settings/section-parts.tsx, settings/state.ts, host-ui, layout`
- **Environment note**: `../kandev` was absent; linked to `~/repo/kandev` (HEAD f099a46d, tag v0.96.0, matches `.kandev-sdk-ref`) per the project Testing Posture learning. `ui/node_modules` was installed with `npm ci` for the baseline.

### APIs Discovered
- Kandev plugin actions — `internal/plugin/scm_actions.go`, `manifest.yaml` — 24 `scm.*` keys. The settings page calls 8 of them:
  - `scm.providers.list` (authenticated, unguarded while Backlog is off) → `{providers: ProviderView[]}`, always all three providers
  - `scm.providers.set_token` (admin) `{provider, token, username?}` → ProviderView
  - `scm.providers.test` (admin) `{provider}` → ProviderView (a refusal is a result in `lastError`)
  - `scm.providers.remove` (admin) `{provider}` → ProviderView
  - `scm.providers.use_cli` (admin) `{provider, login?}` → ProviderView (GitHub/GitLab only)
  - `scm.providers.cli_accounts` (admin) → `{accounts: CliAccount[]}`
  - `scm.repos.search` (admin) `{provider, query}` → `{repos: {fullName,url}[]}`
  - `scm.mappings.set` (admin) `{provider, projectKey, repos[]}` → ProviderView (empty list removes the mapping; max 20 repos; each repo checked with the provider)
  - Others (prs.list, prs.link/unlink, links.list, task_prs.list, queries.*, watches.*) are consumed by PR list, watch forms and task/issue panels and all take a `provider` per item.
- Outbound REST — `internal/github|gitlab|bitbucket/client.go` — 5 methods each (`CurrentUser`, `SearchRepos`, `GetRepo`, `ListPRs`, `GetPR`)
- Kandev state store — `internal/scm/store.go` — workspace keys `scm.settings`, `scm.links`, `scm.dismissed`, `scm.queries`, `scm.watches`, `scm.ledger`; instance key `scm.index`; each `{schemaVersion: 1, items}`
- Kandev secret store — `scm.SecretKey` = `backlog.scm.<provider>.<workspace>`

### Frameworks & Libraries
- Go 1.26.x (toolchain at `~/.local/go`), Kandev `pkg/pluginsdk` v0.96.0, stdlib `net/http`, testify `require`
- UI: TypeScript ~6.0.3, host React (types `@types/react` 19.3), esbuild 0.28, Vitest 5.0, jsdom 30, axe-core 4.14, ESLint 10 + typescript-eslint, Prettier 3.9
- Host UI kit components used on the page (via `hostUi(host)`): `SettingsSection`, `Badge`, `Button`, `Input`, `Label`, `Select*`, `Alert`; layout tokens `BUTTON`, `FIELD`, `ROW`, `STACK` from `ui/src/layout`

### Test Coverage
- **Test Directories**: Go tests beside sources (`internal/scm/*_test.go`, `internal/plugin/*_test.go`, `internal/<provider>/client_test.go`), JSON fixtures in `internal/*/testdata/`; UI tests beside sources (`ui/src/settings/*.test.tsx`) with the shared harness `ui/src/testing/harness` (`fakeHost`, `mount`, `byTestId`, `axeViolations`, `expectOnlyCatalogueText`, `expectTestIds`, `pseudoCatalogue`, `rawControls`)
- **Test Frameworks**: Go `testing` + testify, `-race`; Vitest + jsdom + axe-core
- **Coverage Config**: present (`make coverage`, 80% Go line floor over `./internal/...` and `./server/...`)
- **Baseline (2026-10-08, this worktree)**:
  - `go test -race -count=1 ./internal/scm/... ./internal/github/... ./internal/gitlab/... ./internal/bitbucket/... ./internal/plugin/...` → all 5 packages `ok`
  - `npx vitest run src/settings` → 8 files, 121 tests passed (one harmless jsdom canvas warning)
  - Full `make coverage` / contract test not run at this stage
- `source-control-section.test.tsx` (25 tests): the fixture `setup()` always returns all three providers (GitHub connected, GitLab/Bitbucket `not_configured`); assertions key on test ids `backlog-scm-<provider>-*`, `backlog-scm-<provider>-map-<project>`, `backlog-scm-backlog`, and on catalogue-only text plus axe. `sections.test.tsx` asserts the `backlog-section-source-control` section and the absence of `backlog-scm-github-save` for members.
- Go: `TestProviders_ListsAllThreeNotConfigured` (service_test.go:20) pins "always three providers"; `TestSCM_Manifest_Actions` (actions_scm_test.go:355) pins the manifest action list; `TestSCM_Actions_GuardRefusesAllButProvidersList` pins the guard.

### Code Quality Indicators
- **Linting**: golangci-lint + gosec (Makefile `lint`), `go vet`, gofmt; UI ESLint (`ui/eslint.config.js`), Prettier (`ui/.prettierrc`), `tsc --noEmit` strict
- **CI/CD**: `.github/workflows/ci.yml`, `release.yml`, `secrets.yml`
- **Documentation**: Go doc comments on every exported name with FR/NFR references; TS components carry JSDoc with FR/AC references; message catalogue `ui/src/messages/en.ts` is the only source of UI text (tests enforce it)

### Technical Debt Signals
- `source-control-section.tsx` (488 lines) holds three components (`ProjectRepos`, `ProviderCard`, section) with duplicated local `field`/`input` and `button` helpers (lines 138-162 vs 302-362).
- Visual separation is weak: all four blocks (Backlog Git + 3 providers) are bare `<section className={STACK}>` siblings in one `flex flex-col gap-6` inside a single `SettingsSection` (source-control-section.tsx:466-482); there is no card frame, border or divider, and the provider name is only an `h4 text-sm font-medium` (line 367), same size as the `h5` repo heading (line 424).
- Repo-scope copy is provider-neutral: `scmMappingsHeading` "Repositories of each selected Backlog project", `scmSearchLabel` "Search repositories for {project}", `scmManualLabel` "Repository name for {project}", `scmNoRepos`/`scmMappingLine` "{project}: ...", and one shared placeholder listing all three formats `scmManualPlaceholder` "owner/name, group/project or workspace/repo" (en.ts:390-398). The service is only implied by DOM nesting.
- The word "scope" is overloaded: `scmScopes*` texts are token permission scopes (en.ts:365-369), unrelated to the repo mapping.
- `scmNotice` maps `cli_unavailable` to the literal `{cli: "gh / glab"}` (git-state.ts:194) instead of the card's CLI.
- `ponytail:` notes in scope: one settings document per list (store.go:105), dismissed/ledger never pruned (store.go:165, 181), one global CLI lock (cli_token.go:109), one watcher worker for all workspaces (watcher.go:200).

## Handoff Summary
- **Intent-relevant finding**: Nothing enforces a single source control service today; several can be connected at once. `scm.Store` keeps one `Settings` entry per provider in the `scm.settings` document (store.go:38-50). `Service.Providers` always returns all three (service.go:165-175). `SetToken`, `UseCLI` and `RemoveToken` change only their own provider's entry and secret (service.go:227-406). Every downstream feature is per-provider and works with any provider that has a credential: PR list and watch form selectors offer `"backlog"` plus every connected provider (`usableProviders`, pr-list.tsx:127-132, 217; watch-form.tsx:91-115), links refresh loops over `Providers` (watcher.go:137), and queries, watches and links each carry their own `provider`. The UI renders Backlog Git plus three unframed provider blocks in one section (source-control-section.tsx:460-485), and each block's repo-mapping area uses provider-neutral labels (en.ts:390-398).
- **Risks / follow-up**:
  - The "one at a time" rule needs a product decision before design. (a) Does Backlog Git count as a service? It is always on through the Backlog connection and also holds the Git access form (SettingsScreen.tsx:424-438). (b) Enforce it only in the UI, or in the backend? A backend invariant needs a new field, such as `activeProvider` in `scm.settings`, or a rule that connecting one provider disconnects the others. (c) What happens to the mappings, queries, watches and links of providers that are not active? The precedent is `RemoveToken`, which keeps them disabled and does not delete them (FR2.2, service.go:392-406).
  - Upgrade path: existing workspaces may already have two or three providers connected. The design must pick which one stays active, or ask the admin to choose. The `scm.settings` schema is version 1, and `load` refuses any other version and never overwrites it (store.go:229-233). Prefer an additive JSON field over bumping the schema version.
  - A new action key, such as `scm.providers.set_active`, means changes to `manifest.yaml`, `scmHandlers`, `TestSCM_Manifest_Actions` and probably the `guarded()` allow-list. The frozen `internal/plugin/testdata/v030/manifest.yaml` must stay unchanged. Admin-only actions are enforced by the manifest `access: admin`.
  - Preserve these test ids and contracts, which other suites depend on: `backlog-section-source-control`, `backlog-scm-<provider>-*`, `backlog-scm-backlog`, and catalogue-only text. Tests also assert axe passes. `TestProviders_ListsAllThreeNotConfigured` assumes the list always has three entries. If the list shape changes, that test changes too.
  - All UI text must come from `ui/src/messages/en.ts` (tests enforce this). New provider-specific labels need new keys, or `{provider}` params added to the existing `scm*` keys.
  - NFR1 still applies: never show or refill tokens. `setToken("")` runs after every action (source-control-section.tsx:252).
  - Environment: `../kandev` is now a symlink to `~/repo/kandev` (v0.96.0), and `ui/node_modules` is installed. Both sit outside git tracking.
