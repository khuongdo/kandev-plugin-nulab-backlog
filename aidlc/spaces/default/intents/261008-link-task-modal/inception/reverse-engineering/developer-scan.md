# Developer Code Scan: Link Task modal (intent 261008-link-task-modal)

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - ui/src/issues/link-task-dialog.tsx
  - ui/src/issues/link-task-dialog.test.tsx
  - ui/src/issues/issues-page.tsx (dialog wiring: lines 27, 118, 145, 395-418, 615-622)
  - ui/src/git/pr-link.ts
  - ui/src/git/pr-link.test.ts
  - ui/src/issues/task-menu.ts
  - ui/src/index.ts
  - ui/src/host-ui.ts
  - ui/src/layout.ts
  - ui/src/messages/en.ts (link-related keys)
  - ui/src/issues/issues-state.ts (issueNotice)
  - internal/plugin/issue_actions.go
  - internal/issues/service.go (SearchTasks, Link, Unlink: lines 495-625)
  - internal/issues/types.go (ParseIssueKey, issueKeyPattern: lines 56-66)
- **Skimmed only**:
  - ui/src/page/ (start-task.tsx link call path)
  - ui/src/issues/ (issue-badge, issue-panel, links-store)
  - ui/src/settings/, ui/src/switch/, ui/src/testing/ (harness: fakeHost stubs openTaskLinkDialog at line 883)
  - internal/plugin/ (runtime.go error mapping at lines 344-348, manifest action entries)
  - internal/issues/ (remaining service, store, watcher files)

### Packages Found
- `ui/src/issues` — UI module — TypeScript/React (host React) — Backlog issues list, Link to task dialog, issue badge, issue panel, Unlink task-menu action
- `ui/src/git` — UI module — TypeScript — Backlog Git/PR features; `pr-link.ts` is the task-side "Link Backlog pull request" action built on the host's `openTaskLinkDialog`
- `ui/src/page` — UI module — TypeScript — `/backlog` page, Start task menu (create task then `issues.link`)
- `internal/plugin` — Go package — action routing (`issueHandlers`), error mapping to SDK codes; only package importing `pluginsdk`
- `internal/issues` — Go package — issue service: `SearchTasks`, `Link`, `Unlink`, `Links`, link store

### Build System
- **Type**: Go modules (backend) + npm/esbuild (UI bundle), driven by `Makefile`
- **Config Files**: go.mod, Makefile, manifest.yaml, ui/package.json, ui/tsconfig.json, ui/vitest.config.ts, ui/eslint.config.js
- **Build Dependencies**: `ui` resolves `@kandev/plugin-sdk` through a tsconfig path to `../../kandev/apps/packages/plugin-sdk/src/index.ts`; go.mod `replace`s `github.com/kandev/kandev` with `../kandev`. Neither `../kandev` nor `ui/node_modules` exists in this worktree, and `go` is not on PATH, so no Go or Vitest baseline was run in this scan.

### APIs Discovered
- Plugin actions (manifest.yaml + internal/plugin/issue_actions.go) used by the Link Task flow:
  - `issues.tasks.search` — workspace scope — body `{query}` → `{tasks: [{taskId, taskKey?, title, linkedIssueKey?}]}`; at most 20 rows, case-insensitive match on title or key (service.go:503-528)
  - `issues.link` — task scope (task id from verified context) — body `{issueKey}` → Link; key must match `^([A-Z][A-Z0-9_]*)-([1-9][0-9]{0,8})$` (types.go:56), project must be selected, task already linked to another issue → `ErrConflict` (service.go:542-583). Accepts an issue key only, not an issue URL.
  - `issues.unlink` — task scope — no body; `ErrNotLinked` when no link (service.go:600-609)
- Host API (SDK `apps/packages/plugin-sdk/src/index.ts`):
  - `openTaskLinkDialog({title, description, inputLabel, placeholder?, emptyError, failureMessage, successMessage, inputTestId?, errorTestId?, submitTestId?, onSubmit(reference, signal)})` (line 755)
  - `registerTaskAction({id, label, icon?, placement: "link", group?, visible?, singleTaskOnly?, run})` (line 842)

### Frameworks & Libraries
- Go — 1.26.0 (go.mod) — backend
- github.com/kandev/kandev (pluginsdk) — replaced to ../kandev, pinned v0.96.0 — plugin runtime
- github.com/stretchr/testify — v1.12.1 — Go tests
- React (host-provided at runtime via `host.React`/`host.jsx`; dev dep ^19.3.0) — UI
- TypeScript — ~6.0.3; esbuild ^0.28.2; Vitest ^5.0.3; jsdom ^30.1.2; axe-core ^4.14.0; ESLint ^10.12.0; Prettier ^3.9.9
- Host UI kit (`host.ui`: Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, Input, Label, Button) — the plugin ships no CSS and uses host utility classes (layout.ts)

### Test Coverage
- **Test Directories**: co-located `*_test.go` in internal/issues and internal/plugin (testdata under each); co-located `*.test.ts(x)` in ui/src
- **Test Frameworks**: Go `testing` + testify `require`; Vitest + jsdom + axe-core with the shared harness ui/src/testing/harness.ts (`fakeHost`, `mount`, `expectOnlyCatalogueText`, `axeViolations`)
- **Coverage Config**: present (`make coverage`, 80% Go line floor); not run in this scan

Link Task dialog tests (ui/src/issues/link-task-dialog.test.tsx, 6 cases): labelled dialog with focus in search (axe clean); search call shape and "Linked to" disabled option (AC3.3.3); link call and `onLinked`/`onClose` (AC3.3.1-2); empty state with Link disabled; Esc/Cancel return focus to opener (AC3.3.4, AC8.2.3); catalogue-only text. PR link tests (ui/src/git/pr-link.test.ts) show the pattern for testing a `openTaskLinkDialog` action: assert the options object, then call `options.onSubmit` directly.

### Code Quality Indicators
- **Linting**: golangci-lint + gosec, go vet, gofmt (Makefile); ESLint (ui/eslint.config.js), Prettier, `tsc --noEmit` strict
- **CI/CD**: .github/workflows/ci.yml, release.yml, secrets.yml
- **Documentation**: README present; components carry doc comments citing story/AC ids

### Technical Debt Signals
- ui/src/issues/link-task-dialog.tsx:25-29 doc comment says "Kandev has no task picker, so the plugin lists the tasks". True for a task picker, but Kandev 0.96.0 does offer `openTaskLinkDialog` for task-side linking, which the plugin already uses for PRs (ui/src/git/pr-link.ts:24) and not for issues.
- Focus handling is hand-written: `document.querySelector('[data-testid="backlog-link-task-search"]')?.focus()` and focus restore in an effect cleanup (link-task-dialog.tsx:48-52), where the host form uses `autoFocus` and Radix `onCloseAutoFocus`.
- The search effect fires on every keystroke with no debounce (link-task-dialog.tsx:54-60); a sequence counter discards stale replies, but there is no loading state, so the list area is blank until the first reply.
- Error text is a bare `<p role="alert">` with no destructive styling (link-task-dialog.tsx:127-131); the host form uses `text-xs text-destructive`.
- No success toast after linking (the host form toasts `successMessage`).
- The submit button label is "Link"/"Linking..." while the GitHub dialogs use "Save"/"Saving...".
- Options are host `Button`s inside `<ul>` with `aria-pressed`, without visual separation, key styling, or truncation for long titles (link-task-dialog.tsx:101-125).
- internal/issues/service.go:586 carries a `ponytail:` note: `taskKey` lists all tasks per link.

## Reference: Kandev GitHub integration link UX (read-only, /home/k_do_webfrontier/repo/kandev at v0.96.0)

Key fact: the GitHub integration has no "link this issue to an existing task" picker on its issue list. GitHub issue rows (apps/web/components/github/my-github/issue-list.tsx:79-90) show only `TaskRowIndicator` and `IntegrationStartTaskMenu` (create a task, which then auto-links via `linkTaskIssue`, quick-task-launcher.tsx:231). Linking an existing task goes the other way: from the task, through the task's Link submenu, by typing an issue reference.

1. **Entry point**: the task's Link submenu (apps/web/components/kanban-card-link-submenu.tsx:47-134): label `Link` with `IconLink`; built-in items "GitHub Pull Request" (`IconGitPullRequest`) and "GitHub Issue" (`IconCircleDot`); then plugin `placement: "link"` actions with the plugin icon (lines 101-116, `task-context-link-plugin-<id>`). Also reachable from the task switcher (task-switcher-link-menu.tsx:47-68). Dialogs are mounted in task-actions-menu-dialogs.tsx:201-215.
2. **GitHub issue dialog** (apps/web/components/task/task-github-issue-dialog.tsx):
   - `DialogContent className="w-[calc(100vw-2rem)] sm:max-w-lg"` with `onCloseAutoFocus={createFocusReturnHandler(focusReturnRef)}` (lines 227-230)
   - Header: title "Link GitHub issue", or "Change GitHub issue" when already linked (line 233); `DialogDescription` "Use a full issue URL or number for {owner}/{repo}." when the task has exactly one GitHub repo, else "Use a full GitHub issue URL for this task." (lines 235-239)
   - Body (`space-y-2`): `Label` "Issue", one `Input` (prefilled with the current issue URL on open, lines 186-191), placeholder "#1470 or github.com/owner/repo/issues/1470", disabled while submitting; inline error `text-xs text-destructive` (lines 99-114)
   - Footer `gap-2 sm:justify-between`: left "Unlink" (outline) only when linked; right "Cancel" (outline) and "Save"/"Saving..." (primary); all buttons `cursor-pointer` (lines 127-161)
   - Validation: empty input → "Enter a GitHub issue URL or number." (line 195); a bare number is resolved against the inferred repo (lines 74-85)
   - Success: toast "GitHub issue linked" (variant success) and close; failure: inline error from the server message or "Failed to link GitHub issue." (lines 200-206); Unlink toasts "GitHub issue unlinked" (line 216)
3. **GitHub PR dialog** (apps/web/components/task/task-github-pr-dialog.tsx:65-97): same shell and description pattern, body is the shared `TaskChangeRequestLinkForm`.
4. **Shared host form** (apps/web/components/integrations/task-change-request-link-form.tsx): `<form className="contents">` so Enter submits; `Label` + `Input autoFocus`; inline error `text-xs text-destructive`; footer "Cancel" (outline) + submit "Save"/"Saving..." with `data-dialog-default-action`; empty check with `emptyError`; `AbortController` per submit, aborted on cancel/unmount; success toast then close (lines 26-150).
5. **Plugin access**: `host.openTaskLinkDialog(options)` (apps/web/lib/plugins/host-api.ts:537, 561-586) renders exactly this form in a host dialog with the plugin's title/description/copy; the plugin supplies only `onSubmit(reference, signal)` and throws an `Error` whose message becomes the inline error. No Unlink button and no prefill are available through this API.
6. **i18n strings** (apps/web/src/locales/en/task.json): `linkGithubIssue` 811, `changeGithubIssue` 240, `issue` 752, `githubIssueRefPlaceholder` 690, `enterGithubIssueUrlOrNumber` 533, `githubIssueLinked` 688, `failedToLinkGithubIssue` 590, `unlink` 1952, `useAFullIssueUrlOr` 1975, `useAFullGithubIssueUrl` 1973. Link submenu labels in kanban.json: `link` 51, `githubIssue` 39.

### Comparison with the current plugin Link Task modal
| Aspect | GitHub integration | Backlog plugin today |
|---|---|---|
| Direction | Task → issue (task Link submenu) | Issue → task (issue row "..." menu → "Link to task", issues-page.tsx:405-415) |
| Input | One text field: URL or number | Search field + clickable task list |
| Description line | Yes (`DialogDescription`) | None |
| Submit | Save / Saving..., Enter submits | Link / Linking..., click only |
| Error | Inline `text-xs text-destructive` under the field | `<p role="alert">` below the list, unstyled |
| Success | Toast + close | Close only (`onLinked` updates the row) |
| Unlink | In the dialog when linked | Separate task-menu action (task-menu.ts) |
| Width | `w-[calc(100vw-2rem)] sm:max-w-lg` | Host default |
| Existing plugin precedent | — | `pr-link.ts` already uses `openTaskLinkDialog` for PRs (task side) |

## Handoff Summary
- **Intent-relevant finding**: The current modal (ui/src/issues/link-task-dialog.tsx, opened from the issue row menu in ui/src/issues/issues-page.tsx:405-415 and 615-622) is a custom issue-side task picker. The GitHub integration has no such picker; it links from the task's Link submenu with a single "issue URL or number" field in a host dialog (task-github-issue-dialog.tsx; shared form task-change-request-link-form.tsx), which plugins get through `host.openTaskLinkDialog` + `registerTaskAction({placement: "link"})`. The plugin already follows that exact pattern for PRs in ui/src/git/pr-link.ts:13-63 (registered at ui/src/index.ts:92), but not for Backlog issues.
- **Risks / follow-up**:
  - The requirements stage must settle what "mimic GitHub" means: (a) add a task-side "Link Backlog issue" link-menu action through `openTaskLinkDialog` (copying pr-link.ts), (b) restyle the existing issue-side dialog to the GitHub dialog shell (description, `sm:max-w-lg`, destructive inline error, Save/Saving..., Enter to submit, success toast), or both. Keeping or removing the issue-row "Link to task" entry is part of that choice.
  - For (a), `issues.link` accepts only an issue key (types.go:56); accepting a pasted Backlog issue URL needs either UI-side parsing or a backend change, plus a not-found/conflict/validation message mapping like pr-link.ts:48-58. `openTaskLinkDialog` offers no Unlink and no prefill, so Unlink stays in task-menu.ts.
  - After a task-side link, the shared links store must refresh (task-menu.ts:38 does `store.refresh`), otherwise the badge and issues page will not show the new link.
  - Preserve existing ACs and tests: AC3.3.1-AC3.3.4, AC8.2.3, catalogue-only text, axe check, and the `backlog-link-task-*` test ids if the dialog is kept.
  - No test baseline ran: `go` is not installed and ../kandev and ui/node_modules are absent in this worktree (project rule asks for both before Reverse Engineering).
