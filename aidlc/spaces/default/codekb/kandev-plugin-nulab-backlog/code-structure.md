# Code Structure — kandev-plugin-nulab-backlog

## Layout

| Path | Classification | Notes |
|---|---|---|
| `server/main.go` | entrypoint | `pluginsdk.Serve(plugin.NewRuntime())` only |
| `internal/plugin/` | adapter | action routing (`runtime.go` `HandleAction`), host adapters (`host_port.go`: `hostStores`, `hostPort`, `issueHost`; `scm_actions.go`: `scmHost`), Backlog Git credential (`credential.go`), webhooks, events; `*_actions.go` per area |
| `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/` | domain services | see [component-inventory.md](component-inventory.md) |
| `internal/backlog/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` | HTTP clients | stdlib `net/http`, `testdata/` fixtures |
| `internal/redact/`, `internal/testutil/` | utility / test helper | |
| `internal/pkgverify/`, `cmd/verifypkg/`, `internal/ci/`, `cmd/ci/` | build tooling | package verifier; CI checks (workflows, secrets, contract, release preflight, marketplace) |
| `ui/src/` | UI bundle source | `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, `testing/`; registrations in `index.ts` |
| `manifest.yaml` | plugin contract | see [api-documentation.md](api-documentation.md) |
| `Makefile`, `.golangci.yml`, `.github/workflows/` | build / CI | every CI step is a Make target |
| `aidlc/`, `.claude/`, `docs/` | records / framework / docs | non-app |

## SCM Package (`internal/scm/`) — intent area

| File | Holds |
|---|---|
| `cli_token.go` | `CLIRunner` type, `cliCommand` (fixed `gh` / `glab` args, no `--user`), `runCLI` (no shell, 10 s, 4 KiB, stderr dropped), `cliCache` (per provider, 5 min TTL, one mutex), `cliToken`, `forgetCLI` |
| `service.go` | `Service` (`clients`, `conn`, `secrets`, `tasks`, `store`, `Now`, `CLI`, `cli`), `MethodToken`/`MethodCLI`, error codes (`cli_unavailable`, ...), `ProviderView`/`view`, `Providers`, `SetToken`, `UseCLI`, `credential`, `Test`, `RemoveToken`, `SearchRepos`, `SetMapping` |
| `store.go` | `Store` over host state (schema version 1), `Settings` (`Source`, `HasToken`, `Account`, `AccountID`, `LastError`, `Mappings`), `Link`, `Query`, `Watch` |
| `types.go` | `Provider`, `ParseProvider`, inputs (`TokenInput`, ...) |
| `client.go` | `Credential` (token hidden in `fmt`), `Client` interface (5 methods, credential per call), `User{ID, Name}` |
| `errors.go` | `ErrNoToken`, `ErrCLIUnavailable`, `ErrNotFound`, `ErrConflict`, `HTTPError` |
| `prs.go`, `links.go`, `queries.go`, `watcher.go`, `httpx.go` | PR lists ("me" filter on `AccountID`), links, saved queries, 1-minute watcher, shared HTTP helper (skimmed) |

`internal/github/client.go`: Bearer auth; `CurrentUser` maps `login` -> `User.ID`, `name` (or login) -> `User.Name`.

`internal/plugin/scm_actions.go`: action key constants (incl. `scm.providers.use_cli`), `scmClients`, `wireSCM`, handlers, `classifySCM`, `scmHost.CreateTask`.

## Issues Package (`internal/issues/`)

`types.go` (records incl. `Link`, `ParseIssueKey`), `service.go` (issue list/detail, task creation via `HostPort.CreateTask`, `SearchTasks`, `Link`/`Unlink`, `Links`), plus sync, watches, quick actions, queries (skimmed).

## UI Source (`ui/src/`) — intent area

| Path | Role |
|---|---|
| `settings/source-control-section.tsx` | `ProviderCard` per provider: token form, Test, Remove, mappings; "Use gh / glab CLI login" button (`useCli`, line 233; rendered line 302) calling `scm.providers.use_cli` |
| `page/start-task.tsx` | issue "Start task" menu: renders Kandev `TaskCreateDialog`, then calls `issues.link` (lines 100-115) |
| `git/git-state.ts` | `ProviderView` TS mirror, `usableProviders` |
| `messages/en.ts` | message catalogue (`scm*` strings) |
| `testing/harness.ts` | fake host for Vitest |

## Build and Packaging

`Makefile` targets: `check-format`, `vet`, `lint`, `test` (`-race`), `coverage` (80% floor over `./internal/... ./server/...`), `check-secrets`, `build` (4 platforms), `ui-build`, `package`, `verify-package`, `contract-test`, `release-preflight`. Every Go target first checks `../kandev` HEAD equals `.kandev-sdk-ref`. The platform set lives in `manifest.yaml`, `Makefile` `PLATFORMS`, `internal/pkgverify`, `internal/plugin/manifest_test.go`.

## Code Patterns

- Ports and adapters; domain errors mapped to SDK codes only in `internal/plugin`.
- Errors wrapped with `%w`; `context.Context` first on I/O; bounded reads.
- A secret is registered with `redact.WithSecrets(ctx, token)` as soon as it is read.
- Injected seams for tests: `Service.Now` (clock), `Service.CLI` (command runner), fake `Client`/secrets/state via `newHarness`.
- `exec.CommandContext` with fixed arguments carries `//nolint:gosec // G204`.
- UI factories `createXxx(host, ...)`, test ids prefixed `backlog-`, catalogue-only text.
- Doc comments cite FR/NFR/AC ids of the intent that introduced them.
