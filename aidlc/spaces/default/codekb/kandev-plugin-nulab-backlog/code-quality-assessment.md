# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- Go: ~108 test files, about 704 `func Test` (count from the original full scan, v0.4.2), co-located in every `internal/*` package (`httptest` fakes, `testdata/` JSON), run with `-race`.
- UI: 33 Vitest files under `ui/src/` (about 283 `it` cases, original full scan), co-located, jsdom + axe-core, shared fake host `ui/src/testing/harness.ts` (stubs `openTaskLinkDialog`).
- `make coverage`: 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile at `build/coverage.out`.
- SCM tests: fake `Client`, `fakeSecrets`, `fakeState` and `newHarness` (`internal/scm/harness_test.go:232-250`); GitHub client tests use an `httptest` fake GitHub with user/repo/pull fixtures and 401, 403 rate limit, 404, 429 cases; plugin-level SCM tests in `internal/plugin/actions_scm_test.go`; secret-leak tests in `internal/scm/httpx_test.go` and `internal/plugin/actions_test.go`; UI tests in `ui/src/settings/source-control-section.test.tsx`.
- Link flow tests:
  - `ui/src/issues/link-task-dialog.test.tsx` (6 cases): labelled dialog with focus in search (axe clean); search call shape and "Linked to" disabled option (AC3.3.3); link call and `onLinked`/`onClose` (AC3.3.1-2); empty state with Link disabled; Esc/Cancel return focus to the opener (AC3.3.4, AC8.2.3); catalogue-only text.
  - `ui/src/git/pr-link.test.ts`: the pattern for an `openTaskLinkDialog` action — assert the options object, then call `options.onSubmit` directly.
  - `ui/src/issues/task-menu.test.ts`: Unlink visibility and refresh.
- Packaging: `pkgverify_test.go`, `contract_test.go`; the real install path runs only in `make contract-test`.
- **Baseline (run 3, 2026-10-08, commit `d3d17e5`, intent `261008-gh-cli-auth`)**: `go test -race ./internal/scm/... ./internal/github/... ./internal/connection/...` -> **407 passed in 3 packages**; `go test -race ./internal/plugin/... ./internal/git/...` -> **302 passed in 2 packages** (709 total, 0 failures). No `-coverprofile`; UI (Vitest) baseline not recorded. Environment: Go 1.26.8 at `~/.local/go/bin` (not on `PATH`); `../kandev` is missing in this worktree, so the run used a temporary `go.work` in the session scratchpad that points the SDK at `~/repo/kandev/apps/backend` (v0.96.0, matches `.kandev-sdk-ref`).
- The `261008-link-task-modal` scan recorded no baseline: `go` was not on `PATH`, and neither `../kandev` nor `ui/node_modules` existed in its worktree. The project rule asks for Go 1.26.x and `../kandev` at `v0.96.0` (e.g. `~/repo/kandev`, at `f099a46`) before Reverse Engineering; record the Go and Vitest baseline before Code Generation.
- Earlier runs (261007, 261008-ci-path-filter, 261008-fix-uiux-backlog) recorded no baseline.

## Linting

gofmt, go vet, golangci-lint (`.golangci.yml`, +gosec), `tsc --noEmit` strict, ESLint, Prettier, actionlint, `cmd/ci workflows` policy (SHA-pinned actions, top-level `permissions` exactly `contents: read`, no `pull_request_target`, write permission only in `release.yml` job `publish`). Existing gosec conventions: `//nolint:gosec // G204` on `exec.CommandContext` with fixed arguments (`internal/ci/changes.go:41`); `//nolint:gosec // G101` on the `scm.providers.set_token` action key; `//nolint:gosec // G117` on `Credential.Token`.

## CI/CD

- `ci.yml`: `checks` (format, vet, lint, test, coverage, check-secrets, build, package, verify-package) and `packaged-host-contract` (installs the package on Kandev built at `v0.96.0`). A CI path filter for app-only changes was added after run 1 by `261008-ci-path-filter` (PR #14) and is not re-verified here.
- `secrets.yml`: separate credential scan.
- `release.yml`: `verify` -> `contract` -> `publish` (build provenance attestation, `gh release create`) on `v*` tag push.
- Required checks `checks` and `packaged-host-contract` on `main` via ruleset `24580280`; squash-only. Details: [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).

## Intent Findings: 261008-link-task-modal

Request: "Fix the Link Task modal to mimic how the GitHub integration does it". Reference flow: [architecture.md](architecture.md#external-reference-kandev-github-integration-link-ux-v0960-read-only).

Implemented in intent `261008-link-task-modal`: the task-side "Link Backlog issue" action (`ui/src/issues/issue-link.ts`, reading (a)) and the restyled issue-side dialog (reading (b)). The findings below describe the code as scanned before that change.

### Current modal vs GitHub integration

| Aspect | GitHub integration (Kandev v0.96.0) | Backlog plugin today (`ui/src/issues/link-task-dialog.tsx`) |
|---|---|---|
| Direction | Task -> issue (task Link submenu) | Issue -> task (`/backlog` issue row menu "Link to task", `issues-page.tsx` ~405-415, 615-622) |
| Input | One text field: URL or number | Search field + clickable task list (max 20) |
| Description line | Yes (`DialogDescription`) | None |
| Submit | Save / Saving..., Enter submits | Link / Linking..., click only |
| Error | Inline `text-xs text-destructive` under the field | `<p role="alert">` below the list, unstyled (lines 127-131) |
| Success | Toast + close | Close only (`onLinked` updates the page row) |
| Unlink | In the dialog when linked | Separate task-menu action (`task-menu.ts`) |
| Width | `w-[calc(100vw-2rem)] sm:max-w-lg` | Host default |
| Focus | `autoFocus` + `onCloseAutoFocus` | Hand-written `querySelector(...).focus()` and restore in effect cleanup (lines 48-52) |
| Existing plugin precedent | — | `git/pr-link.ts` already uses `openTaskLinkDialog` (task side) for PRs |

### Open decision (for Requirements Analysis)

"Mimic GitHub" has two readings, not exclusive:

- **(a) Task-side action**: add "Link Backlog issue" (`registerTaskAction`, `placement: "link"`) that opens `host.openTaskLinkDialog`, copying `pr-link.ts`. The host gives the exact GitHub form. Consequences: no Unlink or prefill in that dialog (Unlink stays in `task-menu.ts`); `issues.link` accepts only an issue key, so a pasted Backlog issue URL needs UI-side parsing or a backend change in `internal/issues` (`ParseIssueKey`) and `internal/plugin`; `onSubmit` must map validation / conflict / not-found / no-project failures to catalogue messages (as `pr-link.ts:48-58`); the action must call `LinksStore.refresh`, or badges stay stale.
- **(b) Restyle the issue-side dialog**: keep the task picker but adopt the GitHub shell (description, `sm:max-w-lg`, destructive inline error, Save/Saving..., `<form>` so Enter submits, `autoFocus`/`onCloseAutoFocus`, success toast). Consequence: GitHub has no task picker to copy, so the list part has no reference design.

Whether the issue-row "Link to task" entry stays is part of the same choice.

### Constraints to preserve

- Existing ACs and tests: AC3.3.1-AC3.3.4, AC8.2.3, catalogue-only text, axe check, and the `backlog-link-task-*` test ids if the dialog is kept.
- `LinksStore` refresh: today the issue-side link updates only the page row (`addTask`), not `LinksStore`; any new link path should refresh it like `task-menu.ts:38`.
- Everything used must exist in Kandev `0.96.0` (`min_kandev_version`); `openTaskLinkDialog` does.
- No test baseline was run by this scan (see [Test Coverage and Baselines](#test-coverage-and-baselines)).

## Intent Findings: 261008-gh-cli-auth

Request: add GitHub CLI login as a second way to connect the GitHub source-control provider, next to the access token. Current flow: [architecture.md](architecture.md#interaction-diagrams). Implemented in v0.5.1 (PR #19, `internal/scm/cli_token.go`, admin action `scm.providers.use_cli`, GitLab via `glab` as well); the table describes the code as scanned before that change.

| # | Area | Evidence | Change shape |
|---|---|---|---|
| 1 | Credential source | One private read path `(*scm.Service).credential` (`internal/scm/service.go:221-237`) with 7 callers, including the 1-minute watcher | Choose the source inside `credential()`: for GitHub with the CLI method, run `gh auth token --hostname github.com` instead of reading the secret; callers stay unchanged |
| 2 | Stored method | `scm.Settings` (`internal/scm/store.go:41-48`) has no notion of credential source; schema version 1, `load` checks only `schemaVersion` | Add an `omitempty` field (e.g. auth method `token` / `gh_cli`); old documents stay readable |
| 3 | Connect action | Only `scm.providers.set_token` exists; manifest/handler parity test (`internal/plugin/manifest_test.go:136-142`) | New admin action (e.g. `scm.providers.use_gh_cli`) mirroring `SetToken`: resolve the token, validate with `CurrentUser`, record account and method; add the manifest entry (key must match `^[a-z0-9][a-z0-9._-]*$`) |
| 4 | View and state | `view` derives `state` only from `HasToken` + `LastError`; UI `hasToken = state !== "not_configured"`; `usableProviders` keeps only `connected` | A CLI-backed provider must still report connected (set `HasToken`-equivalent truth) and expose the method as a non-secret `ProviderView` field so the card can show it |
| 5 | Remove | `RemoveToken` deletes the secret and clears account fields | Also clear the method; deleting a missing secret must stay harmless |
| 6 | Errors | `ErrNoToken` text says "add one under Source control"; `classifySCM` maps it to `validation`/`token` | New sentinel for an unusable CLI (e.g. not installed, not logged in, timeout) with its own mapping; keep `gh` stderr out of messages |
| 7 | UI | `ProviderCard` in `ui/src/settings/source-control-section.tsx` offers only the token form | GitHub-only control to use the CLI login, plus showing which method is active; new `scm*` strings in `ui/src/messages/en.ts`; TS `ProviderView` in `ui/src/git/git-state.ts` |

Constraints and risks:

1. **Where `gh` runs.** The token is read on the Kandev server host, from the environment of the OS account that runs Kandev (see [architecture.md](architecture.md#system-overview)). In Docker or when `gh` is absent or logged out, the action must fail with a clear typed error. The host offers no API to borrow Kandev's own GitHub credential, so the plugin must exec `gh` itself (stdlib `os/exec`).
2. **Resolve per call vs import once.** Per call keeps the token current (survives `gh auth refresh`) but runs a subprocess on every `credential()` call, including the watcher each minute — needs a short timeout and possibly a small in-memory TTL cache. Import once into the secret store means no runtime exec but the copy goes stale when `gh` rotates the token. Decision belongs to Requirements/Design.
3. **Secrets.** Trim `gh` stdout, wrap it with `redact.WithSecrets`, never log or return `gh` stderr, never persist the CLI token in plugin state (a cache, if any, stays in memory). Project rules: redact tokens in logs, errors and test output; no real credentials in tests. Add a leak test like the existing ones.
4. **Testability.** Inject the command runner as a `Service` field (like `Now`) so tests never exec a real `gh`; wire the fake in `newHarness`.
5. **Lint.** `exec.CommandContext("gh", ...)` with fixed arguments needs `//nolint:gosec // G204: fixed arguments`.
6. **Scope.** Only GitHub has this path; GitLab (`glab`) and Bitbucket stay token-only unless requirements say otherwise. `scm.task_prs.list` uses Kandev's own PR data and is unaffected.
7. **Reference shape.** Kandev itself runs `gh auth token --hostname <host> [--user <login>]` and handles older `gh` without `--user` and an empty token (Kandev `internal/github/gh_accounts.go:201-215`, external).
8. **Construction environment.** Link `../kandev` to `~/repo/kandev` (v0.96.0) or use an out-of-tree `go.work`; put Go 1.26.x on `PATH`.

## CI Path Filter Constraints

History from intent `261008-ci-path-filter`, recorded by run 1 (pre-change facts; the change shipped in PR #14):

1. **Required checks must still report.** A plain `paths` filter on `ci.yml` stops the workflow from starting, leaving `checks` and `packaged-host-contract` "Expected - waiting". A job skipped by `if:` does report success.
2. **Release filtering has no trigger-level meaning.** `paths` filters do not apply to tag pushes.
3. **`docs/manual-checks/` is a release input** (`release-preflight`, `internal/ci/release.go`).
4. **Secret scan coverage.** `check-secrets` scans the whole repo, including `aidlc/` and `docs/`; skipping CI on records-only PRs drops that scan unless a light scan stays (now `secrets.yml`).
5. **App set includes CI and build files** (`.github/workflows/**`, `Makefile`). Classification: [code-structure.md](code-structure.md#top-level-path-classification).
6. **Workflow policy.** New actions must be SHA-pinned; top-level `permissions: contents: read`.
7. **Both events.** The filter applies to `pull_request` and `push` to `main`.

## Documentation

README covers build, install (upload via Settings > Plugins), connection, CI, release, marketplace. Go packages have `doc.go`; exported Go symbols and UI factories have doc comments, often citing BR/FR/NFR/AC ids. `docs/manual-checks/` holds the first-release record. Makefile targets carry traceability IDs. At run 3 the README did not describe SCM provider connection methods beyond tokens (PR #19 added a CLI login section).

## Intent Findings: 261008-fix-uiux-backlog

History; shipped in v0.5.0 (PR #17). Verified by the `261008-link-task-modal` scan: the badge is registered for `task-card-tags`, `task-row-metadata` and `chat-top-bar`; the nav entry and route are registered only when Backlog is ON at load (FR3); `Link`/`LinkView` carry `Summary`.

| # | Request | Pre-change evidence | Change shape proposed |
|---|---|---|---|
| 1 | Backlog issue on Home > Tasks rows, hover summary, click opens issue | Badge only on `task-card-tags`; no `Summary` on `Link`/`LinkView` | Register `IssueBadge` for `task-row-metadata`; add `Summary` to `Link` and `LinkView` |
| 2 | Hide Home > Integrations entry while OFF | `registerNavItem` unconditional; no `requires`, no unregister, no late registration in Kandev v0.96.0 | Needs a decision (shipped as register-at-load only, FR3) |
| 3 | Projects right below the sign-in method | `SettingsScreen.tsx` order ended with `projects` | JSX reorder |
| 4 | Issue-watch empty message | Shared `watchesEmpty` = "No PR watches yet" | New message key |
| 5 | Remove per-list Add watch button | Empty state rendered a second `add(...)` | Drop the empty-state child |

## Known Issue: Plugin install 502

Status: addressed in v0.4.2 by dropping `windows-amd64` from the package; not re-verified since. Original analysis (intent `261007-plugin-install-502`):

- Runtime: Kandev v0.97.0 (`kandev --headless`, `:38429`) behind `tailscale serve`.
- Kandev `server.readTimeout` default 30 s (`KANDEV_SERVER_READTIMEOUT`). Multipart upload install parses the whole body before `Install`; a body slower than 30 s is cut.
- Backend logs: `POST /api/plugins/install` 400, `duration_ms` ~30000, body `{"error":"missing multipart field \"package\""}`; `tailscale serve` turns it into 502.
- Reproduced: 29.5 MB v0.4.1 upload throttled to 800 KB/s -> 502 after 33.5 s.
- Remaining option: install by URL (see [architecture.md](architecture.md#improvement-opportunities)).

## Technical Debt

- Link Task dialog (`ui/src/issues/link-task-dialog.tsx`, as scanned before `261008-link-task-modal`):
  - The doc comment (lines 25-29) says "Kandev has no task picker, so the plugin lists the tasks" — true for a picker, but Kandev 0.96.0 offers `openTaskLinkDialog` for task-side linking, used for PRs and not for issues.
  - Search fires on every keystroke with no debounce (lines 54-60); a sequence counter drops stale replies, but there is no loading state, so the list is blank until the first reply.
  - Options are host `Button`s in a `<ul>` with `aria-pressed`, without visual separation, key styling or truncation of long titles (lines 101-125).
  - Hand-written focus handling, unstyled error, no success toast, "Link" instead of "Save" (see the comparison table above).
- `internal/issues/service.go:586`: `ponytail:` note — `taskKey` lists all tasks per link.
- SCM (as scanned by `261008-gh-cli-auth`, before v0.5.1): only one credential method (`set_token`); `Settings` had no credential-source field; `ErrNoToken` wording assumed a pasted token (intent findings above).
- SCM: `ponytail:` markers in `internal/scm/store.go` (unbounded dismissed list / ledger) and `internal/scm/watcher.go` (single watcher worker).
- `ci.yml` and `release.yml` duplicate setup and contract steps (run 1).
- `release-preflight` uses `gh release list --limit 1000` (marked `ponytail:` in the Makefile).
- Plugin-owned switch bus (`ui/src/switch/enabled-events.ts`) duplicates host state because Kandev v0.96.0 cannot expose the integration switch to plugins.
- Platform list declared in four places (see [code-structure.md](code-structure.md#build-and-packaging)).
- `pkgverify` duplicates Kandev `pkgtar` rules; no size limit check.
- Contract test installs over loopback (30 s client timeout), so it never sees proxy/slow-upload failures.
- Runtime drift: Kandev 0.97.0 running vs 0.96.0 pin.
- Dev environment: `../kandev` missing in fresh worktrees and Go not on `PATH` (run 3).
