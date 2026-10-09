# Requirements Analysis Questions — 261009-no-workflow-error

Context: the "+ Task" menu (`ui/src/page/start-task.tsx:54-55`) asks Kandev v0.96.0 for `getTaskCreationContext(workspaceId)`. Kandev only returns it when the workflow's steps are already in its web store, which only the Kanban board and Kandev's own integration pages load. Opened directly on the plugin's `/backlog` route (reload, mobile), the context is `null` and the plugin shows "Kandev has no workflow for this workspace yet." even though a workflow exists. The error is a small inline span inside the row's non-shrinking action cell, which squeezes the title column and overflows on narrow screens.

## Question 1
How should "+ Task" open Kandev's task dialog when the context is not loaded?

A. Open Kandev's `TaskCreateDialog` with `workflowId: null`, so the dialog picks the workspace workflow and loads its steps itself; use the context only when it is already available (recommended)
B. Always open the dialog with `workflowId: null`, never using the context
C. Keep requiring the context, but show a better error telling the user to open the Kanban board first
X. Other (please specify)

[Answer]: A

## Question 2
The three watch dialogs (`ui/src/settings/issue-watch-dialog.tsx:118`, `ui/src/git/watch-form.tsx:168`, `ui/src/git/scm-watch-form.tsx:119`) fail the same way, but they must store a real `workflowId` / `workflowStepId`, so option 1A does not apply to them. Are they in scope for this fix?

A. No — fix only "+ Task" (issue rows and PR rows) now; the watch dialogs become a follow-up (recommended for a small bugfix)
B. Yes — also fix the watch dialogs in this bugfix (needs another way to obtain the workflow steps, larger change)
C. Only improve the error message/UI in the watch dialogs, no logic change
X. Other (please specify)

[Answer]: A

## Question 3
How should the "+ Task" errors (no workflow, task created but not linked) be shown so they work on both mobile and desktop?

A. Kandev toast (`host.toast.error`), like the Link dialog already uses for success — no layout change to the row (recommended)
B. A full-width inline alert below the row, wrapping on narrow screens
C. Both: toast plus a short inline marker on the row
X. Other (please specify)

[Answer]: A

## Question 4
Which error displays should the UI/UX retouch cover?

A. Only the "+ Task" errors on issue and PR rows (recommended for this bugfix)
B. Every error display on the plugin's Backlog page (issue list, PR list, link dialogs)
C. Every error display in the plugin, including the settings page
X. Other (please specify)

[Answer]: B

## Question 5
When the workspace really has no workflow at all, what should happen?

A. Let Kandev's dialog handle it (with 1A the dialog shows its own empty/no-workflow state); the plugin shows nothing extra
B. Plugin shows a clear error toast before opening the dialog (requires the plugin to know the workflow list)
C. Not sure — decide in Code Generation after checking Kandev's dialog behaviour
X. Other (please specify)

[Answer]: B

## Follow-up Questions

## Question 6
Q4=B widens the retouch to every error display on the Backlog page (issue list, PR lists, link dialogs, watch forms, Git access, etc.). Q3=A chose a toast for "+ Task" errors. Which style applies to the other errors?

A. Toast for errors caused by a click (an action failed); a full-width inline alert that wraps on narrow screens for errors that describe the page state (a list failed to load, not connected) and inside dialogs (recommended)
B. Toast for every error on the Backlog page
C. Full-width inline alert for every error except "+ Task"
X. Other (please specify)

[Answer]: A

## Question 7
Q5=B needs the plugin to know whether the workspace has any workflow. The browser API has no workflow list; the plugin backend can read it through Kandev's Host `Workflows().List`, which needs the manifest capability `api_read: ["workflows"]` added. Adding a capability can ask the admin to re-approve the plugin's permissions on upgrade. Accept this?

A. Yes — add `api_read: workflows` and a small backend action that answers "does this workspace have a workflow"; only call it when Kandev's context is missing (recommended for Q5=B)
B. No — drop Q5=B and use Q5=A (let Kandev's dialog handle the no-workflow case)
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Q1=A: "+ Task" uses Kandev's task-creation context when it is loaded; when it is missing, it opens Kandev's `TaskCreateDialog` with `workflowId: null` so the dialog picks the workflow and loads its steps.
- Q2=A: only "+ Task" on issue rows and PR rows gets the logic fix; the three watch dialogs' workflow lookup is a follow-up.
- Q3=A: "+ Task" errors (no workflow, task created but not linked) are shown as a Kandev toast (`host.toast.error`), with no inline text in the row.
- Q4=B: the error-display retouch covers every error display on the plugin's Backlog page (issue list, PR lists, link dialogs, watch forms, Git access, etc.), not the Settings page.
- Q5=B: when the workspace really has no workflow, the plugin shows a clear error toast and does not open the dialog.
- Q6=A: toast for errors caused by a click; a full-width inline alert that wraps on narrow screens for page-state errors (list failed to load, not connected) and for errors inside dialogs.
- Q7=A: add manifest capability `api_read: workflows` and a small backend action that tells whether the workspace has a workflow; the UI calls it only when Kandev's context is missing.

Does this all look correct before I generate the requirements artifact?

Looks correct / Request changes

[Answer]: Looks correct
