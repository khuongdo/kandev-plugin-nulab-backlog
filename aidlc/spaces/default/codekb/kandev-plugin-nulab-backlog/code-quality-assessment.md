# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

Measured 2026-10-07 at commit `86ae473`, no tracked file changed.

| Area | Suite | Baseline |
|---|---|---|
| Go | 86 `_test.go` files beside the code; fake fixtures in `internal/backlog/testdata/`; `make coverage` floor 80% over `./internal/... ./server/...`, only exclusion `server/main.go` | `go test -race -coverprofile=build/coverage.out ./internal/... ./server/...` with go1.26.8 and `../kandev` at `v0.96.0`: all 9 test packages pass. backlog 96.0%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.6%, redact 97.4%, testutil 88.0%, server 0.0% (excluded). Total 92.9% from the raw profile (gated figure equal or higher). Profile deleted afterwards |
| UI | 28 `*.test.{ts,tsx}` files, `ui/src/testing/harness.ts`, axe-core a11y checks | `vitest run` 229/229 pass; `tsc --noEmit`, `eslint .`, `prettier --check .` clean |

No UI coverage floor exists.

## Linting, CI/CD, Documentation

- Lint: Go `gofmt`, `go vet`, golangci-lint (default + `gosec`, exceptions as `//nolint:<rule> // reason`), `go mod tidy` check; UI ESLint (recommended + typescript-eslint, `h` allowed unused), Prettier, `tsc --strict`; workflows actionlint + `go run ./cmd/ci workflows`. No `eslint-disable` or `TODO/FIXME` in non-test UI code; `as unknown as` only in tests.
- CI: `ci.yml` `checks` job runs `make check-format vet lint test coverage check-secrets build package verify-package`; `packaged-host-contract` job on the minimum Kandev version. `release.yml`: `verify`, `contract`, `publish` with provenance.
- Docs: `README.md`, `docs/brand/backlog-logo.md` (logo source, hash, Nulab terms, pre-release check), `docs/manual-checks/`. Go package doc comments complete; TS doc comments cite requirement IDs.

## Technical Debt

- **Buttons unstyled**: none of the 42 `<Button>` usages in non-test UI code pass `variant` or `size`; every one renders as the primary filled button.
- **Raw HTML controls** (no host styling): `<select>` `issues/issues-page.tsx:261`, `git/watch-form.tsx:147,175`, `git/dashboard-page.tsx:130,174`; radio `settings/SettingsScreen.tsx:220`; checkbox `settings/project-picker.tsx:176`, `git/watch-form.tsx:199`; `<table>` `issues/issues-page.tsx:448`, `git/dashboard-page.tsx:214`; `<details>` `issues/issues-page.tsx:311`, `git/git-access.tsx:111`; plain `<button>` `issues/issue-badge.tsx:47`, `issues/issue-panel.tsx:123,180`, `issues/link-task-dialog.tsx:122`.
- **Repeated layout constants**: `STACK`/`FIELD`/`ROW` in 12 files, two different `ROW` values (`flex gap-2` vs `flex flex-wrap gap-2`).
- **Inline confirm**: `settings/confirm-dialog.tsx` is an in-flow `role="alertdialog"` with its own focus trap, not `host.ui.Dialog`.
- **Settings scattered**: connection, poll interval and Git access in `SettingsScreen`; PR watches on `/backlog/watches`; saved queries on `/backlog/dashboard`; `SettingsScreen.tsx:321` hard-codes `/backlog/watches`.
- **Icon fallback comment**: `brand/backlog-logo.tsx:31-35` suggests `"plug"`, not a curated host name (would render `IconPuzzle`).
- **Build flag mismatch**: `ui/package.json` `build` passes `--jsx-fragment=Fragment`, `make ui-build` does not; any future `<>...</>` builds under one and fails under the other.
- **`coverage.out` at repo root**: `make coverage` writes it there, against the project rule; use the `build/` form.
- **Go ceilings already marked `ponytail:`**: process-wide rate-limit queue `internal/backlog/client.go:346`; code copied from `internal/git` at `internal/issues/store.go:46` and `internal/issues/sync.go:191`; unpruned ledger `internal/git/store.go:153`; one GET per linked issue `internal/issues/sync.go:26`. UI: first 50 repositories only `ui/src/git/git-state.ts:89` (also limits watch and query repository choices).

## Intent 261007 Risks

Item numbers follow the intent ([business-overview.md](business-overview.md#current-intent-261007-uiux-github-style-refactor-minimal)).

1. **Watcher settings in Settings (item 1)**
   - The host already wraps `SettingsScreen` in an unframed section with icon and switch, so framed `host.ui.SettingsSection` blocks fit (Connection, PR watches as `Card` + `Table` + `Dialog`, Issue sync, Git access, Projects).
   - Only PR watches exist in the backend (`internal/git/watcher.go`, `git.watches.*`). "Issue watcher" maps only to the issue status sync interval (`issues.set_poll_interval`, already in `connected-panel.tsx:435`) unless a new backend feature is added — needs a requirements decision.
   - Permissions: `git.watches.*` and `git.queries.*` are `authenticated`; `SettingsScreen` switches to a member view (`state.isMember`, set only after an admin action returns 403, `settings/state.ts:165-166`) that hides the connected panel. Decide who sees and edits watches in Settings; changing `access` is a manifest and contract change.
2. **One Integrations entry (item 2)**
   - `index.ts:58-64,72-83` registers 3 nav items and routes; `index.test.ts:104-105,120-122` locks them.
   - The PR side is query-driven only (max 20 PRs of one repository, no paging). A GitHub-like PR list needs saved queries used as presets, or a new `git.prs.list`-style action (manifest, adapter, service, `TestU*_Manifest*` tests).
   - Removing `/backlog/watches` and `/backlog/dashboard` breaks bookmarks; `registerRoute` without `registerNavItem` can keep them. The "Review watches" restore notice (`SettingsScreen.tsx:315-325`) must point to the new location.
   - SDK v0.96.0 cannot reproduce `PageShell`, GitHub's scope bar/toolbar, nested nav or sidebar shortcuts; compose from `host.ui` ([architecture.md](architecture.md#external-reference-kandev-github-integration-ui)).
3. **Textbox and button style (item 3)**: replace raw controls listed above with `host.ui` equivalents and add `variant`/`size` following GitHub (`sm` header actions, `outline` secondary, `ghost` + `icon` icon-only). No `AlertDialog` in the SDK: use `Dialog`. Icon-only buttons need hand-drawn inline SVGs (no `@tabler/icons-react`) or text.
4. **Outline logo (item 4)**: redrawing the official Nulab paths as an outline is a modified logo under the terms quoted in `docs/brand/backlog-logo.md`. Safer: an original stroke-only `currentColor` icon that is not the Nulab mark, or a curated name such as `ticket`, switched at `PLUGIN_ICON`; update `backlog-logo.test.tsx:55`; `pkgverify` still forbids Nulab asset URLs. Needs a user decision.
5. **Test impact**: `index.test.ts`, `connected-panel.test.tsx:148-149` (navigate to `/backlog/watches`), `backlog-logo.test.tsx`, `watches-page.test.tsx`, `dashboard-page.test.tsx`, any test that finds a raw control by tag. The harness must stub every newly used host component (`SettingsSection`, `Card*`, `Select*`, `Checkbox`, `Table*`, `Dialog*`, `Tabs*`, `IntegrationScopeBar`, `IntegrationListToolbar`, ...); `variant`/`size` pass through to plain `<button>` in tests.
6. **Locked behaviour to keep**: settings card, switch and nav entry independent of the enabled switch (BR5.4/BR7.6/BR7.8, `index.ts:48-49`); `settingsHref()` format; action keys unchanged (`^[a-z0-9][a-z0-9._-]*$`).
7. **Environment**: keep Go 1.26.x and `../kandev` at `f099a46`; use Node 22 for parity with CI; keep `coverage.out` out of the repo root.
