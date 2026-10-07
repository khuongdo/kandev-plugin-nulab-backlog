# Code Structure — kandev-plugin-nulab-backlog

## Repository Layout

| Path | Kind | Notes |
|---|---|---|
| `server/main.go` | entrypoint | Only `pluginsdk.Serve(plugin.NewRuntime())` |
| `internal/plugin/` | Go adapter | `runtime.go` (Runtime, `handlers` map, `guarded`, combined `gateway` port), `git_actions.go`, `issue_actions.go`, `credential.go` (`ResolveGitCredential`, `GetGitCredentialBinding`), `host_port.go` (`hostPort`, `issueHost`, Repository), `config.go`, `webhook.go`, `events.go`, `references.go` |
| `internal/backlog/` | Go gateway | Backlog API v2 client pinned to `https://<SpaceHost>` (`client.go:260-265`), Git smart-HTTP probe `CheckGitAccess`, per-group rate limiter, OAuth; fixtures in `testdata/` |
| `internal/connection/` | Go domain | `address.go`, `apikey.go`, `oauth.go`, `lifecycle.go` (Snapshot/Current/Test/Disconnect), `change.go` (space change, `ConnectionChanged`), `projects.go`, `git_credential.go` (`GitCredential`, `gitSecret`, `gitCheck`), `store.go` (state/secret keys, switch, Git secret), `service.go` (Gateway, ConfigReader) |
| `internal/issues/` | Go domain (U3) | Issue list, tasks, links, sync, watches, quick actions, saved issue queries |
| `internal/git/` | Go domain (U4) | `types.go` (`ProviderID`, `RepoRef`, `Reference`, `LinkKey`, `stateIDs`, query validation), `service.go` (repository provider, `DefaultBranch`/`Branches` heuristics, link/create/status, `ResolveCredential`), `prs.go`, `host.go` (`nulab_backlog_pr` metadata), `events.go` (`ConnectionChanged` handling), `store.go`, `watcher.go` |
| `internal/redact/` | Go utility | Masks secrets and Backlog URL query strings |
| `internal/ci/`, `cmd/ci/`, `internal/pkgverify/`, `cmd/verifypkg/`, `internal/testutil/` | tooling / test helper | Secret scan, release preflight, marketplace entry, packaged-host contract driver, package verification, fake keys |
| `ui/` | TypeScript UI | See module map |
| `manifest.yaml` | plugin manifest | 55 actions, webhook, capabilities, one `repository_providers` entry, `config_schema` |
| `docs/`, `aidlc/` | docs / AI-DLC records | Not product code |

## UI Module Map (`ui/src/`)

| Directory | Main files | Role |
|---|---|---|
| (root) | `index.ts` (registrations, lines 49-75 for Git/issue providers), `layout.ts`, `host-ui.ts`, `icons.tsx` | Registrations and shared kit |
| `brand/`, `page/`, `issues/`, `switch/`, `messages/`, `testing/` | see previous store; not re-scanned | Icon, `/backlog` page, issue list/panel, enable switch (`PLUGIN_ID` in `enabled-events.ts`), catalogue, fake host |
| `git/` | `repository-provider.ts` (`BACKLOG_GIT_URL` matcher), `review-provider.tsx`, `git-state.ts` (`loadRepoOptions` from Backlog selected projects), `git-access.tsx` (Git username/password form), `pr-link.ts`, `create-pr.ts`, `pr-list.tsx`, `pr-toolbar.tsx`, `save-query-dialog.tsx`, `watch-form.tsx` | Kandev repository and review providers, PR link/create, PR list, Git access |
| `settings/` | `SettingsScreen.tsx` (stacked `SettingsSection`s; Git access section at 398-409), `connected-panel.tsx`, `project-picker.tsx`, `pr-watches-section.tsx`, `issue-watches-section.tsx`, `saved-queries-section.tsx`, `quick-actions-section.tsx`, `section-parts.tsx`, `use-list.ts`, `state.ts`, `oauth.ts`, `confirm-dialog.tsx` | Integration settings card content |

## Code Patterns

- **Go**: tests beside code, table-driven with `t.Run`, Backlog faked with `httptest`, injected clock and wait, errors wrapped with `%w`, `doc.go` per package, exported identifiers cite requirement IDs. Secret-bearing types implement `String`/`GoString`/`Format` to hide secrets (`connection.GitCredential`, `backlog.Credentials`). Domain state is a workspace-scoped `{schemaVersion, items}` document per key; consumer-side ports (`git.Gateway`, `git.Connection`) with one production implementation each.
- **UI**: `createXxx(host, messages, ...)` factories; JSX through `h = host.jsx`; host components through `hostUi(host)`; each module has a sibling test.
- **Naming**: Go `snake_case.go`; TS kebab-case with two historical PascalCase files (`SettingsScreen.tsx`, `BacklogPage.tsx`).
