# Dependencies — kandev-plugin-nulab-backlog

## External (Go, `go.mod`)

- `github.com/kandev/kandev` — `replace` to `../kandev/apps/backend`; build fails unless `../kandev` HEAD equals `.kandev-sdk-ref`.
- `github.com/stretchr/testify` v1.12.1 — tests.
- `gopkg.in/yaml.v3` v3.0.1 — manifest parsing in `pkgverify` and `ci`.
- Indirect via SDK: `hashicorp/go-plugin` v1.8.0, `google.golang.org/grpc` v1.83.1.

## External (UI dev, `ui/package.json`)

esbuild, typescript, vitest, jsdom, eslint + typescript-eslint, prettier, react/react-dom (types and tests only, never bundled), axe-core. No runtime npm dependency: React, the UI kit, the link dialog and the registry come from the host at runtime. Versions: [technology-stack.md](technology-stack.md).

## External (CI)

- GitHub Actions listed with SHAs in [technology-stack.md](technology-stack.md#github-actions-pinned-by-full-sha-recorded-by-run-1). No third-party (non-`actions/`) action; any new one must be SHA-pinned or `make lint` fails.
- Tools fetched by `go run` in `Makefile`: golangci-lint v2.14.0, actionlint v1.7.12.
- CI checks out `kandev/kandev` twice: at `.kandev-sdk-ref` (build) and at `v<min_kandev_version>` (contract test).
- `gh` CLI with `GH_TOKEN` in `release-preflight` and `publish` (CI only, unrelated to the runtime).

## External Services and Host Tools

- Kandev host (install API, plugin RPC, state, secrets, UI registry, slots, `openTaskLinkDialog`, toasts): [api-documentation.md](api-documentation.md).
- Backlog REST v2, GitHub REST (`api.github.com`), GitLab, Bitbucket REST.
- Packaging tool: Kandev `cmd/plugin-pack` from the sibling checkout.
- GitHub repository ruleset `24580280` (required checks).
- Runtime host tool: none today. Intent `261008-gh-cli-auth` would add an **optional** runtime dependency on the `gh` CLI on the Kandev server host (`gh auth token --hostname github.com`), present only when the admin chooses that method.

## Internal (cross-package)

```
server -> plugin
plugin -> connection, issues, git, scm, github, gitlab, bitbucket, backlog, redact, pluginsdk
connection -> backlog, redact
issues -> backlog, connection, redact
git -> backlog, connection, redact
scm -> connection, redact
github, gitlab, bitbucket -> scm
backlog -> redact
cmd/verifypkg -> pkgverify
cmd/ci -> ci
```

Correction in run 3: the provider clients import `scm` (they implement `scm.Client` and use `scm.API`/`scm.Credential`); `scm` does not import them. `plugin` builds the clients (`scmClients`) and passes them to `scm.NewService`. The graph is acyclic; only `plugin` and `server` import `pluginsdk`.

Build-time chain: `ci.yml` / `release.yml` -> `Makefile` targets -> `cmd/ci`, `cmd/verifypkg`, Kandev `plugin-pack`.

## Internal (UI, `ui/src/`)

```
index.ts -> issues/{issue-badge, links-store, task-menu, issue-panel, i18n}, git/pr-link, settings/SettingsScreen, switch/*, page/*, messages
issues/issues-page -> issues/{link-task-dialog, issues-state, ...}, host-ui, layout, messages/en
issues/link-task-dialog -> host-ui, layout, messages/en, issues/issues-state (issueNotice, TaskLink), git/git-state (noticeText), settings/state (Notice)
issues/task-menu -> issues/{issues-state, links-store (type)}, git/git-state, messages/en
git/pr-link -> git/git-state, settings/state (readFailure), switch/enabled-events (PLUGIN_ID), messages/en
issues/issue-badge -> issues/issues-state, issues/links-store (type), host-ui, layout, messages/en
settings/SettingsScreen -> settings/{source-control-section, issue-watches-section, pr-watches-section, project-picker, ...}, switch/enabled-events
settings/source-control-section -> git/git-state (ProviderView, scmNotice), messages/en
settings/issue-watches-section -> settings/{section-parts, issue-watch-dialog, confirm-dialog, use-list}, git/git-state, messages/en
switch/enabled-events <- index.ts, switch/integration-switch, settings/SettingsScreen, page/BacklogPage, git/*
```

`LinksStore` is created once in `index.ts` and shared by the badge, the `/backlog` page, the task panel and the Unlink menu action. `link-task-dialog` does not receive the store.
