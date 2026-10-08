# Dependencies — kandev-plugin-nulab-backlog

## External (Go, `go.mod`)

- `github.com/kandev/kandev` — `replace` to `../kandev/apps/backend`; Make targets refuse to build unless `../kandev` HEAD equals `.kandev-sdk-ref`. `../kandev` is absent in fresh worktrees.
- `github.com/stretchr/testify` v1.12.1 — tests.
- `gopkg.in/yaml.v3` v3.0.1 — manifest parsing in tooling and tests.

## External (UI dev, `ui/package.json`)

esbuild, typescript, vitest, jsdom, eslint, prettier, react types, axe-core. No runtime npm dependency: React, the UI kit and the registry come from the host.

## External Services and Host Tools

- Kandev host: plugin RPC, state, secrets, tasks, repositories, UI registry ([api-documentation.md](api-documentation.md)).
- Backlog REST v2; GitHub, GitLab, Bitbucket REST.
- `gh` / `glab` on the Kandev server — optional runtime dependency, only for the CLI login method.
- CI: GitHub Actions (SHA-pinned), Kandev checkouts at `.kandev-sdk-ref` and `v<min_kandev_version>`, `gh` in release jobs.

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

Acyclic; only `plugin` and `server` import `pluginsdk`. Provider clients import `scm` (they implement `scm.Client`); `plugin` builds them and passes them to `scm.NewService`.

## Internal (UI, intent area)

```
settings/source-control-section -> git/git-state (ProviderView), messages/en
page/start-task -> host TaskCreateDialog, issues.link action
```
