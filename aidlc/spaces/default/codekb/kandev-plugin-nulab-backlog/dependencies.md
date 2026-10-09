# Dependencies — kandev-plugin-nulab-backlog

## External (Go, `go.mod`)

- `github.com/kandev/kandev` — `replace` to `../kandev/apps/backend`; Make targets refuse to build unless `../kandev` HEAD equals `.kandev-sdk-ref`. `../kandev` is absent in fresh worktrees (this run linked it to `~/repo/kandev` at v0.96.0, outside git).
- `github.com/stretchr/testify` v1.12.1 — tests.
- `gopkg.in/yaml.v3` v3.0.1 — manifest parsing in tooling and tests.

## External (UI dev, `ui/package.json`)

esbuild, typescript, vitest, jsdom, eslint, typescript-eslint, prettier, react types, axe-core. No runtime npm dependency: React, the UI kit and the registry come from the host. Versions: [technology-stack.md](technology-stack.md).

## External Services and Host Tools

- Kandev host: plugin RPC, state, secrets, tasks, repositories, UI registry and UI kit ([api-documentation.md](api-documentation.md)).
- Backlog REST v2; GitHub, GitLab, Bitbucket REST.
- `gh` / `glab` on the Kandev server — optional, only for the CLI login method.
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

## Internal (UI, intent area: start task and notices)

```
page/BacklogPage -> issues/issues-page -> page/start-task (row action, with RowMenu)
git/pr-list, git/scm-pr-list -> page/start-task
page/start-task -> host.context.getTaskCreationContext, host IntegrationStartTaskMenu, host TaskCreateDialog,
                   issues.link | git.prs.link | scm.prs.link, messages/en (errorWorkflow, taskNotLinked)
settings/issue-watch-dialog, git/watch-form, git/scm-watch-form -> host.context.getTaskCreationContext, messages/en (errorWorkflow)
issues/task-menu -> host.toast
```

Host-side (Kandev v0.96.0, read-only): `getTaskCreationContext` depends on the web store's `workflows` (loaded by the layout) and `kanban` / `kanbanMulti` steps (loaded only by the board, task page and first-party integration pages). See [architecture.md](architecture.md#task-creation-context-host-derived-kandev-v0960).

## Internal (UI, settings area, pre-0.6.0)

```
settings/SettingsScreen -> settings/source-control-section (passes the Backlog Git form)
settings/source-control-section -> git/git-state (ProviderView, providerName, scmNotice, loadProviders),
                                   messages/en, settings/section-parts, settings/state, host-ui, layout
git/pr-list, git/watch-form -> git/git-state (usableProviders)  # depend on "many providers"
page/start-task -> host TaskCreateDialog, issues.link action
```
