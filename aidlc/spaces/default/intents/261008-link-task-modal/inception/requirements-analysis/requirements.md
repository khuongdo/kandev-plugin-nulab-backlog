# Requirements — Link Task modal, GitHub-style

Intent: `261008-link-task-modal` (scope: bugfix, depth: Minimal).

## Intent Analysis

Initial description: "Sưa modal của Link Task bắt chước giống cách mà github integration làm" [desc]. In English: fix the Link Task modal so it works like Kandev's GitHub integration.

Today the plugin links a Kandev task to a Backlog issue from one side only, the issue side. On the `/backlog` page, the issue row "..." menu opens "Link to task". That is a custom dialog with a search box, a clickable task list and a plain error paragraph (`ui/src/issues/link-task-dialog.tsx`) [business-overview, architecture]. Kandev's GitHub integration links from the task side instead. The task's **Link** menu has "GitHub Issue", which opens a host dialog with one "issue URL or number" field, a description line, an inline red error, Cancel + Save/Saving..., Enter to submit, and a success toast. Plugins get the same dialog through `host.openTaskLinkDialog`, and the plugin already uses it for "Link Backlog pull request" (`ui/src/git/pr-link.ts`) [architecture].

The user's goal is for linking a Backlog issue to look and behave like GitHub's. That means two things: the same task-side entry point, and the same dialog look and feedback on the existing issue-side flow [Q1].

## Functional Requirements

### FR1 — Task-side "Link Backlog issue" action [Q1]

- **FR1.1** The plugin shall register a task action with `placement: "link"` labelled "Link Backlog issue". It appears in the task's Link menu next to GitHub Issue and "Link Backlog pull request", wherever Kandev renders plugin link actions (Kanban card menu, task switcher / sidebar).
- **FR1.2** Choosing it shall open the Kandev host link dialog through `host.openTaskLinkDialog`, with plugin copy: a title, a description line, the input label "Issue", a placeholder showing both accepted forms, an empty-input error, a failure message and a success message. The test ids follow the PR action pattern: `backlog-link-issue-input`, `backlog-link-issue-error`, `backlog-link-issue-submit`.
- **FR1.3** Submitting shall call the existing `issues.link` action for that task with the resolved issue key, passing the dialog's abort `signal`.
- **FR1.4** On success, the host shows the success toast and closes the dialog. The plugin shall then refresh the shared links store for the workspace, so the issue badge, the `/backlog` page and the task panel show the new link without waiting for the 60 s timer or a window focus.
- **FR1.5** On failure, the dialog shall show a specific inline message: issue not found, issue key not in a selected project, task already linked to another issue (conflict), not connected / reconnect required, rate-limited. Messages contain no Backlog response body and no secrets.

Acceptance (Given/When/Then):
- Given an unlinked task in a workspace where Backlog is connected, when the user opens the task's Link menu, then "Link Backlog issue" is listed.
- Given the dialog is open, when the user enters `PROJ-123` (an existing issue in a selected project) and presses Enter or Save, then the task is linked, a success toast appears, the dialog closes, and the task's issue badge shows `PROJ-123` after the links-store refresh.
- Given the dialog is open, when the user submits an empty field, then the empty-input error shows inline and no action is called.
- Given `issues.link` returns not found, when the user submits `PROJ-999`, then an inline "not found" message naming `PROJ-999` shows and the dialog stays open.

### FR2 — Accepted input: issue key or full issue URL [Q2]

- **FR2.1** The dialog shall accept an issue key (`PROJ-123`, case as typed, trimmed).
- **FR2.2** The dialog shall accept a full Backlog issue URL of the form `https://<space-host>/view/<ISSUE-KEY>`, with optional trailing `#...` / `?...`. The UI extracts `<ISSUE-KEY>` and sends only the key to `issues.link`. The backend contract does not change.
- **FR2.3** The URL host shall be an `https` host under `backlog.com`, `backlog.jp` or `backlogtool.com` (project Mandated rule). Any other host or path, or text that is neither a key nor a valid URL, shall show an inline "not a Backlog issue key or link" message without calling the backend.

Acceptance:
- Given the dialog, when the user pastes `https://acme.backlog.com/view/PROJ-123`, then `issues.link` is called with `issueKey: "PROJ-123"`.
- Given the dialog, when the user pastes `http://acme.backlog.com/view/PROJ-123` or `https://example.com/view/PROJ-123`, then an inline validation message shows and `issues.link` is not called.

### FR3 — Hide the action while the task is linked [Q3]

- **FR3.1** The action's `visible` callback shall return `false` when the shared links store already holds a link for the task. If the store has not loaded the workspace yet, it starts the load and returns `false`, the same way as the existing "Unlink Backlog issue" item.
- **FR3.2** Changing the linked issue stays a two-step flow: "Unlink Backlog issue" (existing task-menu item, unchanged), then "Link Backlog issue".

Acceptance:
- Given a task linked to `PROJ-1`, when the user opens the task's Link menu, then "Link Backlog issue" is not listed and the task menu still shows "Unlink Backlog issue".
- Given the user unlinked that task, when the user opens the Link menu after the store refresh, then "Link Backlog issue" is listed again.

### FR4 — Restyle the issue-side "Link to task" dialog to the GitHub dialog shell [Q1]

The issue-side entry point ("..." → "Link to task" on `/backlog` issue rows) and its task search and list stay. Only the dialog shell and feedback change, to match Kandev's GitHub link dialog (`task-github-issue-dialog.tsx` / `task-change-request-link-form.tsx`):

- **FR4.1** `DialogContent` width `w-[calc(100vw-2rem)] sm:max-w-lg`.
- **FR4.2** The header has the title "Link {key} to a task" (unchanged) plus a `DialogDescription` line explaining what to do (search a task and choose it).
- **FR4.3** The body is a `Label` + search `Input`, focused on open, then the task list. Errors show inline as `text-xs text-destructive` with `role="alert"`, directly under the field / list, in place of the unstyled paragraph.
- **FR4.4** The footer has Cancel (outline) and a primary submit labelled "Save" / "Saving...", matching GitHub, in place of "Link" / "Linking...". The content is a `<form>` so Enter submits when a task is chosen. With no task chosen, the submit stays disabled and Enter does nothing.
- **FR4.5** On success, show a success toast ("Backlog issue linked." or equivalent), refresh the shared links store for the workspace, update the page row (existing `onLinked`), and close.
- **FR4.6** Keep existing behaviour: a task linked to another issue cannot be chosen (AC3.3.3), Esc closes, focus returns to the opener, stale search replies are dropped, and the dialog cannot be closed while saving. Existing `backlog-link-task-*` test ids stay.

Acceptance:
- Given the issue-row "Link to task" dialog with a task chosen, when the user presses Enter in the search field, then the link is saved, a success toast appears, the dialog closes, and the badge for that task shows the issue key without a manual refresh.
- Given `issues.link` fails with a conflict, when the user saves, then a red inline error shows under the list and the dialog stays open with Save enabled again.

## Non-Functional Requirements

- **NFR1 (Security)** Error text shown in either dialog never contains API keys, tokens or raw Backlog response content (team Code Style / project Mandated redaction rule). URL parsing accepts only `https` Backlog hosts (FR2.3).
- **NFR2 (Accessibility)** Both dialogs have a labelled input, a dialog title and description, and an inline error with `role="alert"`. The existing axe check for the issue-side dialog still passes.
- **NFR3 (Consistency)** All new user-visible text lives in `ui/src/messages/en.ts`. Existing catalogue-only text rules hold.
- **NFR4 (Performance)** No new polling. The task-side `visible` check reads the existing shared links store synchronously (one `issues.links.list` call per workspace, unchanged).

## Constraints

- Kandev host API is pinned to v0.96.0. `openTaskLinkDialog` offers no Unlink button and no prefill, which is why FR3 hides the action instead of offering "Change".
- Only `internal/plugin` may import `pluginsdk`. This change is expected to touch only the TypeScript UI (`ui/src/issues/`, `ui/src/index.ts`, `ui/src/messages/en.ts`). The backend `issues.link` contract (key only) stays [Q2].
- Testing posture: TDD with Vitest for the UI. The bugfix scope requires targeted regression tests and the existing suite to stay green.

## Assumptions

- `TaskContext` passed to a link action's `run` and `visible` carries `workspaceId` and `taskId`, as the PR action and Unlink action already use [assumption].
- A Backlog issue URL is `https://<space-host>/view/<ISSUE-KEY>` [assumption, matches the plugin's existing issue links].
- The success toast for the issue-side dialog uses `host.toast.success`, already used by the Unlink action [assumption].

## Out of Scope

- Changing the backend `issues.link` input (URL or number parsing on the server).
- A bare-number input (`123`) [Q2 chose key or URL only].
- Replacing an existing link in one step, or an Unlink button inside either dialog [Q3].
- Removing the issue-side "Link to task" entry [Q1 chose both].

## Open Questions

None.
