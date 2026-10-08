# Code Structure — kandev-plugin-nulab-backlog

## Layout

| Path | Classification | Notes |
|---|---|---|
| `server/main.go` | entrypoint | `pluginsdk.Serve(plugin.NewRuntime())` only |
| `internal/plugin/` | adapter | handlers map, webhook, events, host port, Backlog Git credential (`credential.go`); `issue_actions.go` holds the `issues.*` handlers, `scm_actions.go` the `scm.*` handlers; `Version`/`SDKRef` set by ldflags |
| `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/` | domain services | see [component-inventory.md](component-inventory.md) |
| `internal/backlog/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` | HTTP clients | stdlib `net/http`, `testdata/` JSON fixtures |
| `internal/redact/`, `internal/testutil/` | utility / test helper | |
| `internal/pkgverify/`, `cmd/verifypkg/` | build tooling | offline package verifier |
| `internal/ci/`, `cmd/ci/` | build tooling | `secrets`, `workflows`, `contract`, `preflight`, `marketplace`; `changes.go` is the only `os/exec` use in the repo |
| `ui/src/` | UI bundle source | see [UI source](#ui-source-uisrc) |
| `manifest.yaml` | plugin contract | see [api-documentation.md](api-documentation.md) |
| `Makefile` | build | every CI step goes through it |
| `.github/workflows/` | CI/CD | `ci.yml`, `release.yml`, `secrets.yml`; see [architecture.md](architecture.md#ci-and-release-pipeline) |
| `docs/` | docs | `brand/`, `manual-checks/` |
| `aidlc/`, `.claude/` | process records / framework | AI-DLC records, memory, codekb; AI-DLC shell |

## Top-Level Path Classification

What each top-level path means for CI and release (run 1, 2026-10-08):

| Path | Class | Why |
|---|---|---|
| `server/`, `internal/` | app | Go code; `GO_PKGS := ./internal/... ./server/...` is tested, linted, covered; `./server` is built |
| `cmd/` | app (tooling) | `cmd/ci` (workflow lint, secret scan, contract test, release preflight), `cmd/verifypkg` |
| `ui/` | app | Prettier, tsc, ESLint, Vitest, esbuild bundle into the package |
| `manifest.yaml` | app | copied into the package; `VERSION` and `MIN_KANDEV_VERSION` are read from it |
| `go.mod`, `go.sum` | app | build inputs; CI checks `go mod tidy` |
| `.kandev-sdk-ref` | app | SDK pin for `check-sdk`, ldflags `SDKRef`, CI checkout |
| `.nvmrc` | app | Node version for CI |
| `.golangci.yml` | app (lint config) | changes lint results |
| `Makefile` | app (build) | every CI step |
| `.github/workflows/` | app (CI) | `make lint` runs actionlint and `cmd/ci workflows` on these files |
| `aidlc/`, `.claude/` | non-app | read by no target except the repo-wide `check-secrets` scan |
| `docs/brand/` | non-app | brand note |
| `docs/manual-checks/` | non-app for CI, **release input** | read by `release-preflight`; allow-listed in the secret scan |
| `README.md`, `LICENSE`, `.gitignore` | non-app | read by no target |
| `build/`, `dist/` | generated | local output, not inputs |

## SCM Package (`internal/scm/`)

Provider-neutral source control for GitHub, GitLab and Bitbucket (run 3, analyzed deeply). Never imports `pluginsdk` (`doc.go`).

| File | Holds |
|---|---|
| `client.go` | `Credential{Token, Username}` (token hidden in every `fmt` verb), `Client` interface (`CurrentUser`, `SearchRepos`, `GetRepo`, `ListPRs`, `GetPR`, each taking `cred Credential`) |
| `types.go` | `Provider` (`github`, `gitlab`, `bitbucket`), `ParseProvider`, `TokenInput` and its `validate`, other inputs and views |
| `errors.go` | `ErrHostRefused`, `ErrNotFound`, `ErrConflict`, `ErrNoToken`; `HTTPError{Provider, Status, RetryAfter}` |
| `service.go` | `Service` (fields `clients`, `conn`, `secrets`, `tasks`, `store`, injectable clock `Now`), `NewService`, `SecretKey`, `ProviderView`/`view`, `Providers`, `SetToken`, private `credential`, `Test`, `RemoveToken`, `SearchRepos`, `SetMapping`/`checkRepos` |
| `store.go` | `Store` over host state (schema version 1; `load` checks only `schemaVersion`), `Settings`, `Mapping`, `Link`; `ponytail:` unbounded dismissed list |
| `httpx.go` | shared `API` helper for the three clients: https host check, bounded read, `HTTPError` |
| `prs.go`, `links.go`, `queries.go` | PR lists, PR links, saved queries (credential call sites only were re-read) |
| `watcher.go` | `Watcher` (1-minute tick, single worker, `ponytail:`), `runWatch`, `refreshProvider` |
| `harness_test.go`, `fakes_test.go` | `newHarness` (lines 232-250) builds the service with fake `Client`, `fakeSecrets`, `fakeState` |

`internal/github/`: `client.go` (`New()`; Bearer auth at lines 20-23; `/user`, `/user/repos`, `/repos/{o}/{r}`, `/repos/{o}/{r}/pulls[/{n}]`), `doc.go`, `client_test.go` with `testdata/` fixtures (user, repo, pull, 401, 403 rate limit, 404, 429).

Plugin wiring for SCM (`internal/plugin/`): `scm_actions.go` (action keys, `scmClients`, `wireSCM`, handlers, `classifySCM`, `taskPRs`), `runtime.go` (handler map, `Runtime` with `scm` and `scmWatcher`), `manifest_test.go` (action/handler parity, `TestU2_ManifestActionKeysMatchTheRuntime`, lines 136-142).

## Issues Package (`internal/issues/`)

- `types.go`: persisted records, including `Link` (issue key/id, project, space host, task, state, `LastKnownStatus`, `StatusUpdatedAt`, fail count, connection epoch) and `IssueURL`.
- `service.go`: `Service` — `Link`/`newLink` (create a link), `Links` -> `[]LinkView` (UI view with status, stale, unavailable, `url`), `Detail` (live issue read).
- `sync.go`: periodic status refresh of links (updates `LastKnownStatus`).
- Recorded by run 2 (before v0.5.0); not re-read since.

## UI Source (`ui/src/`)

| Path | Role |
|---|---|
| `index.ts` | `initialize`: all host registrations (table in [architecture.md](architecture.md#ui-surfaces-host-slots)) |
| `host-ui.ts` | typed access to the host UI kit |
| `issues/issue-badge.tsx` | `createIssueBadge` — task badge (key + status chip, link to issue) |
| `issues/issues-state.ts` | `LinkView` TS type, `badgeHref` (https + Backlog host re-check), badge text helpers |
| `issues/links-store.ts` | `LinksStore` — one `issues.links.list` per workspace, shared |
| `issues/i18n.ts`, `messages/en.ts` | message catalogue (`messagesFor` falls back to English; only `en` exists); `scm*` strings for the Source control section |
| `switch/enabled-events.ts`, `switch/integration-switch.tsx` | plugin-owned bus and switch for the per-workspace Backlog ON/OFF |
| `settings/SettingsScreen.tsx` | Backlog settings page and its section order |
| `settings/source-control-section.tsx` | `createSourceControlSection`; `ProviderCard` per provider (token form: save/replace, Test, Remove; mappings). `hasToken = view.state !== "not_configured"`; calls `scm.providers.list`, `set_token`, `test`, `remove` |
| `settings/issue-watches-section.tsx`, `settings/pr-watches-section.tsx` | watch lists |
| `settings/section-parts.tsx` | `SettingsSection`, `ListEmpty` |
| `git/git-state.ts` | `ProviderView` TS mirror, `scmNotice`, `usableProviders` (keeps `state === "connected"`) |
| `settings/*`, `git/`, `page/`, `brand/` | other settings sections, Git UI, `/backlog` page, brand assets |
| `testing/harness.ts` | shared fake host for Vitest |

## Build and Packaging

`Makefile` (GNU Make, `SHELL := /bin/bash`, `-eu -o pipefail`). Every Go target first runs `check-sdk`: `../kandev` HEAD must equal `.kandev-sdk-ref`.

| Target | Does |
|---|---|
| `check-format` | `gofmt -l server internal`; `prettier --check` in `ui/` |
| `vet` | `go vet ./...` |
| `lint` | golangci-lint v2.14.0 (`./...`), `tsc --noEmit`, ESLint, actionlint v1.7.12, `go run ./cmd/ci workflows -dir .github/workflows` |
| `test` | `go test -race ./internal/... ./server/...`; `vitest run` |
| `coverage` | 80% floor (`COVERAGE_MIN := 80`), profile under `build/`, excludes only `server/main.go` |
| `check-secrets` | `go run ./cmd/ci secrets -root .` over the whole repo |
| `build` | `CGO_ENABLED=0` cross-build of `./server` for `PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64` |
| `ui-build` | esbuild `ui/src/index.ts` -> `build/ui/bundle.js`; fails if React is bundled |
| `package` | build + ui-build -> stage -> Kandev `cmd/plugin-pack` -> `dist/nulab-backlog-<version>.tar.gz` + `dist/checksums.txt` |
| `verify-package` | `go run ./cmd/verifypkg` |
| `contract-test` | builds Kandev at `v$(MIN_KANDEV_VERSION)` from `KANDEV_MIN_DIR`, installs the package, runs `cmd/ci contract` |
| `release-preflight` | needs `TAG`; tag format, tag == manifest version, tag on `origin/main`, no existing Release, first-release record in `docs/manual-checks/` |
| `marketplace-entry`, `clean`, `help` | not used by CI |

The platform set lives in four places: `manifest.yaml` `runtime.executables`, `Makefile` `PLATFORMS`, `internal/pkgverify` `executables`, `internal/plugin/manifest_test.go` (`internal/plugin/testdata/v030/manifest.yaml` is a frozen 0.3.0 snapshot).

## Code Patterns

- Ports-and-adapters: only `internal/plugin` maps domain errors to `pluginsdk` codes (`classify`, `classifySCM`).
- Errors wrapped with `%w`; `context.Context` first on I/O; responses bounded by `io.LimitReader`.
- Secrets: a token is registered with `redact.WithSecrets(ctx, token)` as soon as it is read or received, so logs and errors through that context mask it.
- Injected dependencies for tests: SCM clock `Service.Now`; fakes for `Client`, secrets and state via `newHarness`.
- `exec.CommandContext` with fixed arguments carries `//nolint:gosec // G204` (`internal/ci/changes.go:41`).
- UI: factories `createXxx(host, ...)` return host-React components via the `h` JSX factory; test ids prefixed `backlog-`.
- Tests co-located (`_test.go`, `*.test.ts[x]`), fake servers via `httptest`, fake host via `ui/src/testing/harness.ts`.
- Package docs and Makefile targets carry traceability IDs (FR/NFR/US/AC).
