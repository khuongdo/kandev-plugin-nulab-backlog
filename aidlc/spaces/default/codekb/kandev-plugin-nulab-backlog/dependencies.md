# Dependencies — kandev-plugin-nulab-backlog

## External (Go, `go.mod`)

- `github.com/kandev/kandev` — `replace` to `../kandev/apps/backend`; build fails unless `../kandev` HEAD equals `.kandev-sdk-ref`.
- `github.com/stretchr/testify` v1.12.1 — tests.
- `gopkg.in/yaml.v3` v3.0.1 — manifest parsing in `pkgverify` and `ci`.
- Indirect via SDK: `hashicorp/go-plugin` v1.8.0, `google.golang.org/grpc` v1.83.1.

## External (UI dev, `ui/package.json`)

esbuild, typescript, vitest, jsdom, eslint + typescript-eslint, prettier, react/react-dom (types and tests only, never bundled), axe-core. No runtime npm dependency: React, the UI kit and the registry come from the host at runtime. Versions: [technology-stack.md](technology-stack.md).

## External (CI)

- GitHub Actions listed with SHAs in [technology-stack.md](technology-stack.md#github-actions-pinned-by-full-sha). No third-party (non-`actions/`) action is used today; any new one must be SHA-pinned or `make lint` fails.
- Tools fetched by `go run` in `Makefile`: golangci-lint v2.14.0, actionlint v1.7.12.
- CI checks out `kandev/kandev` twice: at `.kandev-sdk-ref` (build) and at `v<min_kandev_version>` (contract test).
- `gh` CLI with `GH_TOKEN` in `release-preflight` and `publish`.

## External Services

- Kandev host (install API, plugin RPC, state, secrets, UI registry and slots): [api-documentation.md](api-documentation.md).
- Backlog REST v2, GitHub, GitLab, Bitbucket REST.
- Packaging tool: Kandev `cmd/plugin-pack` from the sibling checkout.
- GitHub repository ruleset `24580280` (required checks): [api-documentation.md](api-documentation.md#required-status-checks-main-ruleset).

## Internal (cross-package)

```
server -> plugin
plugin -> connection, issues, git, scm, backlog, redact, pluginsdk
connection -> backlog, redact
issues -> backlog, connection, redact
git -> backlog, connection, redact
scm -> github, gitlab, bitbucket
backlog -> redact
cmd/verifypkg -> pkgverify
cmd/ci -> ci
```

The graph is acyclic; only `plugin` and `server` import `pluginsdk`.

Build-time chain: `ci.yml` / `release.yml` -> `Makefile` targets -> `cmd/ci`, `cmd/verifypkg`, Kandev `plugin-pack`.

## Internal (UI, `ui/src/`)

```
index.ts -> issues/{issue-badge, links-store, ...}, settings/SettingsScreen, switch/*, page/*, git/*, messages
issues/issue-badge -> issues/issues-state, issues/links-store (type), host-ui, layout, messages/en
settings/SettingsScreen -> settings/{issue-watches-section, pr-watches-section, project-picker, ...}, switch/enabled-events
settings/issue-watches-section -> settings/{section-parts, issue-watch-dialog, confirm-dialog, use-list}, git/git-state, messages/en
switch/enabled-events <- index.ts, switch/integration-switch, settings/SettingsScreen, page/BacklogPage, git/*
```

`LinksStore` is created once in `index.ts` and shared by the badge, the `/backlog` page and the task panel.
