# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

Measured 2026-10-07 at commit `2b4325f`, no tracked source file changed.

| Area | Suite | Baseline |
|---|---|---|
| Go | `_test.go` files beside the code; fake fixtures in `internal/backlog/testdata/`; `make coverage` floor 80% over `./internal/... ./server/...` (writes `build/coverage.out`), only exclusion `server/main.go` | `go test -race -cover ./internal/... ./server/...` with go1.26.8 and `../kandev` at `v0.96.0`: all 9 test packages pass. backlog 96.1%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.7%, redact 97.4%, testutil 88.0%. Total 92.7%. Profile kept outside the repo |
| UI | 29 `*.test.{ts,tsx}` files, `ui/src/testing/harness.ts`, axe-core a11y checks | `vitest run` 286/286 pass; `tsc --noEmit`, `eslint .`, `prettier --check .` clean |

No UI coverage floor exists.

## Linting, CI/CD, Documentation

- Lint: Go `gofmt`, `go vet`, golangci-lint (default + `gosec`, exceptions as `//nolint:<rule> // reason`), `go mod tidy` check; UI ESLint, Prettier, `tsc --strict`; workflows actionlint + `go run ./cmd/ci workflows`. No `eslint-disable` and no `TODO/FIXME/HACK` in `ui/src`.
- CI: `ci.yml` `checks` job runs the Makefile targets; `packaged-host-contract` job on the minimum Kandev version. `release.yml`: `verify`, `contract`, `publish` with provenance.
- Docs: `README.md`, `docs/brand/backlog-logo.md`, `docs/manual-checks/`. Go package doc comments complete; TS doc comments cite requirement IDs; comments explain host quirks (metadata namespace, async host injection).

## Technical Debt

- **Hand-made host equivalents**: `ui/src/git/pr-toolbar.tsx` re-implements `IntegrationListToolbar` (no border, no `px-4 sm:px-6`, no search because Backlog's PR API has none; the host component requires `customQuery` props); `ui/src/git/save-query-dialog.tsx` parallels `IntegrationSaveQueryDialog`; `BacklogPage.tsx` uses plain `Tabs` instead of `IntegrationScopeBar`.
- **Inconsistent lists**: issues use `Table` + own toolbar `ROW` + `<h2>`; PRs use `ChangeRequestRow` + `PrToolbar`. GitHub uses `ChangeRequestRow` + `IntegrationListToolbar` for both.
- **No page padding**: `/backlog` content is edge to edge.
- **Large files**: `ui/src/issues/issues-page.tsx` (625), `ui/src/settings/SettingsScreen.tsx` (441), `ui/src/git/pr-list.tsx` (386), `internal/git/service.go` (925), `internal/issues/service.go` (893).
- **Go ceilings already marked `ponytail:`**: process-wide rate-limit queue `internal/backlog/client.go:346`; store helpers copied from `internal/git/store.go` into `internal/issues/store.go` (extract when a third package needs them); unpruned ledger `internal/git/store.go:153`; one GET per linked issue `internal/issues/sync.go:26`; gRPC code read by text `internal/plugin/host_port.go:119-125`; no relink after a failed link write `internal/issues/service.go:381`. UI: first 50 repositories only `ui/src/git/git-state.ts:89`.

## Intent 261007-github-parity-actions Risks

Item numbers follow the intent ([business-overview.md](business-overview.md#current-intent-261007-github-parity-actions-express-minimal)); reference behaviour in [architecture.md](architecture.md#external-reference-kandev-github-integration-v0960).

1. **Quick actions (item 1)**
   - No prompt templates anywhere. Two launch options: (a) host `TaskCreateDialog` prefilled with `"<label>: <title>"` and the interpolated prompt (user can edit; Kandev creates the task) followed by `issues.link` with the created task id — loses the atomic link-after-create and the duplicate "already linked" check; (b) keep `issues.create_task` and add a preset/prompt field with Go-side `{{url}}`/`{{title}}` interpolation, optionally `StartAgent`/`Launch.Prompt`. Needs a requirements decision.
   - PR rows have no task path today; a PR quick action needs either `TaskCreateDialog` + `git.prs.link` or a new server action.
   - Presets are user data in plugin state: a new store key and actions (manifest + `manifest_test.go`), name/prompt length validation, no secrets in prompts or logs. Default texts stay untranslated (persisted seed), as in Kandev. A new store is the third user of the copied store helpers.
   - GitHub edits presets with a page-level Save/Discard (`useSettingsSaveContributor`); plugin settings sections save per dialog. Decide which model to follow.
2. **Default queries (item 2)**
   - A preinstalled PR query cannot be a stored `git.queries` row as-is: `QueryInput.Validate` requires a repository of a selected project, unknown at install time, and `git.prs.list` takes one repository per call. Options: a built-in unstored default (e.g. open, assigned to me, over a chosen/first repository), or relaxing the repository rule.
   - Issues have no saved queries: new store and actions (or a shared key) are needed. A GitHub-like "Assigned to me" default needs a "me" assignee for `issues.list` (numeric ids only today; `Connection` stores `ConnectedUserID`).
   - Default selection on open: today nothing is applied (`START` in `pr-list.tsx`); GitHub opens the default saved query, else the first built-in preset.
3. **Layout (item 3)**: add `px-4 sm:px-6` (bars) and `px-3 md:px-6` (results) in the plugin, since `PageShell` adds none; place refresh/last-updated right in the toolbar and the "+ Task" menu at the row end (`ChangeRequestRow` `action`). Adopting `IntegrationListToolbar` needs a query input that Backlog PRs cannot honour.
4. **Test impact**: `ui/src/page/backlog-lists.test.tsx`, `ui/src/issues/issues-page.test.tsx`, `ui/src/settings/sections.test.tsx`, `ui/src/index.test.ts`, `internal/git/queries_test.go`, `internal/issues/create_test.go`, `internal/plugin/actions_u3_test.go` / `actions_u4_test.go`, `internal/plugin/manifest_test.go`. `ui/src/testing/harness.ts` must fake every newly used host component (`IntegrationStartTaskMenu`, `IntegrationScopeBar`, `IntegrationListToolbar`, `TaskCreateDialog`, `Textarea`) and honour `ChangeRequestRow` `action`.
5. **Locked behaviour to keep**: settings card, switch and nav entry independent of the enabled switch (BR5.4/BR7.6/BR7.8); `settingsHref()` format; existing action keys unchanged (`^[a-z0-9][a-z0-9._-]*$`); secret redaction.
6. **Environment**: Go 1.26.x and `../kandev` at `f099a46` (absent by default in a fresh worktree); Node 22 for parity with CI; keep `coverage.out` out of the repo root.
