# Code Summary — Link Task modal, GitHub-style

## Files Created / Modified

| File | Change |
|---|---|
| `ui/src/issues/issue-link.ts` (new) | `parseIssueReference` (FR2) and `createIssueLinkAction(host, store, messages)` (FR1, FR3) |
| `ui/src/issues/issue-link.test.ts` (new) | 31 tests: parser table, action shape, dialog options, submit, refresh, error mapping, visibility |
| `ui/src/issues/link-task-dialog.tsx` | FR4 restyle: `<form className="contents">`, `DialogDescription` (`aria-describedby`), `w-[calc(100vw-2rem)] sm:max-w-lg`, inline `text-xs text-destructive` alert, Save/Saving..., success toast, optional `store` refresh |
| `ui/src/issues/link-task-dialog.test.tsx` | 6 cases became 9; labels updated to Save/Saving... |
| `ui/src/issues/issues-page.tsx` | Passes the shared links store to the dialog |
| `ui/src/index.ts` / `ui/src/index.test.ts` | Registers "Link Backlog issue" next to "Link Backlog pull request"; asserts both, in order |
| `ui/src/messages/en.ts` | 13 new messages (`linkIssue*`, `issue*`, `linkTaskDescription`); unused `link` / `linking` removed |
| `README.md` | Short note on "Link Backlog issue" next to the task-menu actions |

No Go or backend change.

## Key Implementation Decisions

- **URL check:** a URL is accepted only if it uses `https:`, its host ends in `.backlog.com`, `.backlog.jp` or `.backlogtool.com`, and its path is exactly `/view/<KEY>`. The key is upper-cased and checked against `^[A-Z][A-Z0-9_]*-[1-9][0-9]*$`. A look-alike host such as `backlog.com.example.com` is rejected.
- **Task-side errors (R-03):**
  - `not_found` → "Issue {key} was not found."
  - `conflict` → "already linked, unlink first"
  - `validation` with `field=issueKey` → project not selected
  - everything else → `issueNotice`
  - A test asserts that the backend's `detail` text never reaches the user.
- **Refresh (R-04):** after a successful link, the task action calls `store.refresh(...).catch(() => undefined)`. The dialog starts the refresh without awaiting it, so it closes immediately.
- **Visibility (R-01):** `visible` uses the shared links store, the same way the Unlink item does.
- **No shared error helper with `pr-link.ts` (Step 8):** the codes and messages differ, so sharing would have needed a new abstraction.

## Test Coverage Summary

| Run | Files | Tests |
|---|---|---|
| Baseline, full UI suite | 33 | 387 passed |
| After, scoped (issue-link, dialog, task-menu, index) | 4 | 59 passed |
| After, full UI suite | 34 | 421 passed |

`tsc --noEmit`, `eslint .` and `prettier --check .` are all green. Red evidence was recorded for Steps 3, 6 and 9. Go was not touched, so its suite and coverage are unchanged and will be re-run in Build and Test.

## Deviations From the Plan

- **Conflict message in the issue-side dialog:** it keeps the existing generic `issueNotice` text ("That cannot be done in the current state."). Only the task-side action got the specific "already linked" message.
- **Enter key test:** Enter-to-submit is verified with `form.requestSubmit()`, because jsdom does not submit a form on a dispatched Enter keydown. Browser Enter handling comes from the native `<form>`.
- **Test runner setup:** `../kandev` was linked to `~/repo/kandev` (`v0.96.0`), outside the repo.
- **Node version:** `npm ci` warns that the local Node version is v25.2.1. This is not blocking.
