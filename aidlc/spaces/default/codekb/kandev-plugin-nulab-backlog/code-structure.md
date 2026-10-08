# Code Structure — kandev-plugin-nulab-backlog

## Layout

| Path | Classification | Notes |
|---|---|---|
| `server/main.go` | entrypoint | `pluginsdk.Serve(plugin.NewRuntime())` only |
| `internal/plugin/` | adapter | action routing (`runtime.go` `HandleAction`, `guarded()`), host adapters (`host_port.go`: `hostStores`, `hostPort`, `issueHost`; `scm_actions.go`: `scmHost`), Backlog Git credential (`credential.go`), webhooks, events; `*_actions.go` per area |
| `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/` | domain services | see [component-inventory.md](component-inventory.md) |
| `internal/backlog/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` | HTTP clients | stdlib `net/http`, `testdata/` fixtures |
| `internal/redact/`, `internal/testutil/` | utility / test helper | |
| `internal/pkgverify/`, `cmd/verifypkg/`, `internal/ci/`, `cmd/ci/` | build tooling | package verifier; CI checks |
| `ui/src/` | UI bundle source | `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, `layout`, `testing/`; registrations in `index.ts` |
| `manifest.yaml` | plugin contract | see [api-documentation.md](api-documentation.md) |
| `Makefile`, `.golangci.yml`, `.github/workflows/` | build / CI | every CI step is a Make target |
| `aidlc/`, `.claude/`, `docs/` | records / framework / docs | non-app |

## SCM Package (`internal/scm/`) — intent area

| File | Holds |
|---|---|
| `service.go` | `Service`, `MethodToken`/`MethodCLI`, error codes, `ProviderView`/`view`, `Providers` (always three, `:165-175`), `SetToken`, `UseCLI`, `credential`, `Test`, `RemoveToken` (keeps data disabled, `:392-406`), `SearchRepos`, `SetMapping` |
| `store.go` | `Store` over host state, schema version 1 (`load` refuses others, `:229-233`); `Settings` per provider (`:38-50`), `Mapping`, `Link`, `Query`, `Watch`; keys `scm.settings`, `scm.links`, `scm.dismissed`, `scm.queries`, `scm.watches`, `scm.ledger`, `scm.index` |
| `types.go` | `Provider` (`github`, `gitlab`, `bitbucket`), `ParseProvider`, action inputs |
| `client.go` | `Credential` (token hidden in `fmt`), `Client` interface (5 methods, credential per call), `User{ID, Name}` |
| `prs.go`, `links.go`, `queries.go`, `watcher.go` | PR lists ("me" filter on `AccountID`), links with mapping guard, saved queries, watcher (`RunWatch`, `RefreshLinks` looping over `Providers`, `runDue`) |
| `cli_token.go`, `httpx.go`, `errors.go` | CLI runner and cache keyed by `cliKey{provider, login}`, shared HTTP helper, sentinel errors (skimmed) |

`internal/plugin/scm_actions.go`: 24 `scm.*` action key constants, `scmClients`, `wireSCM`, `scmHandlers`, `classifySCM`, `scmHost.CreateTask`.

## Issues Package (`internal/issues/`)

`types.go` (records incl. `Link`, `ParseIssueKey`), `service.go` (issue list/detail, task creation via `HostPort.CreateTask`, `SearchTasks`, `Link`/`Unlink`, `Links`), plus sync, watches, quick actions, queries (not re-read in this run).

## UI Source (`ui/src/`) — intent area

| Path | Role |
|---|---|
| `settings/SettingsScreen.tsx` | Settings screen; builds the Backlog Git access form (`:424-438`) and passes it into the Source Control section |
| `settings/source-control-section.tsx` (488 lines) | `ProjectRepos` (repo search, manual input, mapping lines per Backlog project), `ProviderCard` (status, token form, CLI login + gh account picker, Test, Remove, mappings), `createSourceControlSection` (layout `:460-485`) |
| `settings/section-parts.tsx`, `settings/state.ts` | shared section pieces (`ListError`, ...) and settings state helpers |
| `settings/source-control-section.test.tsx` | 25 tests; fixture always returns three providers |
| `git/git-state.ts` | `ProviderView` TS mirror, `providerName`, `scmNotice`, `loadProviders`, `usableProviders`, `scmRepoOptions` (skimmed) |
| `git/pr-list.tsx`, `git/watch-form.tsx`, `git/scm-*.tsx`, `issues/issue-prs.tsx` | consumers of the provider list (skimmed) |
| `page/start-task.tsx` | issue "Start task" menu (Kandev `TaskCreateDialog`, then `issues.link`) |
| `messages/en.ts` | message catalogue, the only source of UI text (`scm*` keys `:365-398`, `sourceControlDescription`) |
| `testing/harness.ts` | fake host and assertions for Vitest |

## Build and Packaging

`Makefile` targets: `check-format`, `vet`, `lint`, `test` (`-race`), `coverage` (80% floor over `./internal/... ./server/...`), `check-secrets`, `build` (4 platforms), `ui-build`, `package`, `verify-package`, `contract-test`, `release-preflight`. Every Go target first checks `../kandev` HEAD equals `.kandev-sdk-ref`. The platform set lives in `manifest.yaml`, `Makefile` `PLATFORMS`, `internal/pkgverify`, `internal/plugin/manifest_test.go`.

## Code Patterns

- Ports and adapters; domain errors mapped to SDK codes only in `internal/plugin`.
- Errors wrapped with `%w`; `context.Context` first on I/O; bounded reads.
- A secret is registered with `redact.WithSecrets(ctx, token)` as soon as it is read.
- Injected seams for tests: `Service.Now` (clock), `Service.CLI` (command runner), fake `Client`/secrets/state via `newHarness`.
- UI factories `createXxx(host, ...)`, host UI kit via `hostUi(host)`, layout tokens `BUTTON`/`FIELD`/`ROW`/`STACK`, test ids prefixed `backlog-` (`backlog-scm-<provider>-*`), catalogue-only text.
- Doc comments cite FR/NFR/AC ids of the intent that introduced them.
