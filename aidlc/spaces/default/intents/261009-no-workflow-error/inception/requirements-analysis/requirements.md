# Requirements — 261009-no-workflow-error

Scope: bugfix (Minimal depth). Brownfield: Kandev plugin for Nulab Backlog (Go backend, TypeScript UI bundle), built against Kandev v0.96.0.

## Intent Analysis

Initial description: "Khi click tao task xuat hien loi "Kandev has no workflow for this workspace yet" du da co workflow. Retouch lai UIUX cho xuat error de no hoat dong tot tren ca mobile va pc" [desc]

The user wants two outcomes:

1. **Creating a task works whenever the workspace has a workflow.** Today the "+ Task" menu on issue and PR rows reports "Kandev has no workflow for this workspace yet." even though a workflow exists.
2. **Errors on the plugin's Backlog page read well on both phone and desktop.** Today errors are small inline text that squeezes or overflows the row layout on narrow screens.

Root cause (from the code knowledge base, `architecture.md` § Task Creation Context): `ui/src/page/start-task.tsx:54-55` requires `host.context.getTaskCreationContext(workspaceId)`. Kandev v0.96.0 derives that context from its web store and returns `null` unless the workflow's steps are already loaded there, which only the Kanban board and Kandev's own integration pages do. When `/backlog` is opened directly (reload, bookmark, phone), the context is `null` and the plugin raises the error although the workspace has a workflow. A secondary factor: when the active workflow is not found, Kandev falls back to the workspace's first workflow, which may be hidden and never loaded. Kandev's `TaskCreateDialog` accepts `workflowId: null` and then resolves the workflow and loads its steps itself.

## Functional Requirements

### FR1 — "+ Task" opens Kandev's task dialog whenever a workflow exists

- **FR1.1** When `getTaskCreationContext(workspaceId)` returns a context, "+ Task" opens `TaskCreateDialog` with that context's `workflowId`, `defaultStepId` and `steps` (current behaviour, unchanged). [Q1]
- **FR1.2** When the context is `null`, "+ Task" asks the plugin backend whether the workspace has at least one workflow (FR3). If it has, "+ Task" opens `TaskCreateDialog` with `workflowId: null` and no steps, so Kandev selects the workflow and loads its steps; the title and description are prefilled exactly as in FR1.1. [Q1, Q5, Q7]
- **FR1.3** When the backend reports that the workspace has no workflow, the dialog is not opened and an error toast says that Kandev has no workflow for this workspace yet. [Q5, Q3]
- **FR1.4** When the backend check itself fails (network, permission, host unavailable), "+ Task" still opens the dialog with `workflowId: null` rather than blocking the user, so a transient check failure never reproduces the original bug. [assumption]
- **FR1.5** After Kandev creates the task, linking it to the Backlog issue or PR keeps working as today (`issues.link`, `git.prs.link`, `scm.prs.link`) for both dialog modes. [desc]
- **FR1.6** Applies to every "+ Task" entry point built from `createStartTask`: Backlog issue rows and PR rows of every source-control provider. [Q2]

Acceptance criteria:

```
Given a workspace with a workflow and /backlog opened directly (Kandev context is null)
When the user picks a "+ Task" quick action on an issue row
Then Kandev's task dialog opens prefilled with the action's title and description
And no "no workflow" error is shown

Given a workspace with a workflow whose steps Kandev has already loaded
When the user picks a "+ Task" quick action
Then the dialog opens with that workflow and its default step, as before

Given a workspace with no workflow
When the user picks a "+ Task" quick action
Then the dialog does not open
And an error toast says Kandev has no workflow for this workspace yet

Given the task was created but linking it to the issue fails
When Kandev reports success for the task
Then the task is kept and an error toast says the task was not linked
```

### FR2 — Error display retouch on the Backlog page

- **FR2.1** Errors caused by a user action (a click, a submit outside a dialog) are shown as a Kandev toast via `host.toast.error`, with no inline error text inside list rows. This includes both "+ Task" errors (no workflow, task not linked). [Q3, Q6]
- **FR2.2** Errors that describe the page state (a list failed to load, the integration is not connected, Git access missing) are shown as a full-width inline alert at the top of the affected section. The alert wraps its text, never forces horizontal scrolling, and keeps any retry or reconnect control reachable at 320 px viewport width. [Q6]
- **FR2.3** Errors inside dialogs (link dialogs, save-query dialog, watch forms) are shown as a full-width inline alert inside the dialog body, above the dialog actions, wrapping on narrow screens. [Q6]
- **FR2.4** All inline alerts share one presentation (one reusable component or one shared class set), use `role="alert"`, and use Kandev's destructive colour tokens so they match the host theme in light and dark modes. [Q4, Q6]
- **FR2.5** The retouch covers every error display rendered by the plugin's Backlog page: issue list and issue panel, PR lists (Backlog Git and SCM providers), link-task dialog, save-query dialog, watch forms, Git access and review-provider prompts, and the integration switch. The Settings page is out of scope. [Q4]
- **FR2.6** Error message texts keep coming from `ui/src/messages/en.ts`; no raw backend or Backlog response content and no secret is shown (team rule). [team.md]

Acceptance criteria:

```
Given a 360 px wide viewport and an issue row
When a "+ Task" error occurs
Then a toast appears and the row's title column keeps its width (no overflow)

Given the issue list fails to load
When the Backlog page renders
Then a full-width alert with the message is shown above the list on both 360 px and 1280 px widths
And the page has no horizontal scroll

Given a dialog submit fails
When the error is shown
Then it appears as a wrapping alert inside the dialog above its buttons
```

### FR3 — Backend check for workspace workflows

- **FR3.1** The plugin manifest adds the capability `api_read: ["workflows"]` next to the existing `tasks` and `repositories`. [Q7]
- **FR3.2** A new backend action (working name `workflows.status`) takes `workspaceId` and returns whether the workspace has at least one workflow, using `Host.Workflows().List` with a page size of 1. It returns only a boolean, no workflow data. [Q7]
- **FR3.3** The action follows existing conventions: waits (bounded) for the asynchronously injected Host, takes `context.Context`, maps host errors to `pluginsdk` error codes in `internal/plugin` only, and never leaks host error text. [project.md, team.md]
- **FR3.4** The UI calls the action only when the Kandev context is `null` (FR1.2), never on page load. [Q7]

## Non-Functional Requirements

- **NFR1 — Responsiveness:** every error display in FR2 renders without horizontal page scroll and without truncating its message at viewport widths of 320 px, 360 px, 768 px and 1280 px.
- **NFR2 — Latency:** when the context is `null`, the dialog opens (or the no-workflow toast appears) within 1 s of the click on a normal connection; the backend check makes at most one host call.
- **NFR3 — Accessibility:** inline alerts use `role="alert"`; toasts use Kandev's toast (already announced); colour is not the only signal (text is always present).
- **NFR4 — Security:** no API key, token, backend error text or Backlog response body appears in any error message (team rule; the existing redaction test must stay green).
- **NFR5 — Quality:** TDD per team Testing Posture; a targeted regression test reproduces the original bug (context `null` + workspace with a workflow → dialog opens, no error) and fails before the fix; Go tests run with `-race`; Go coverage floor 80% stays met; the existing Go and Vitest suites stay green.
- **NFR6 — Compatibility:** works on the pinned minimum Kandev version (v0.96.0); the packaged-host contract test passes with the new capability.

## Constraints

- Kandev's browser API (`PluginContextApi`) exposes no workflow list; the workflow list is reachable only through the backend Host API (`WorkflowReader.List`). [Q7]
- Kandev owns `TaskCreateDialog` and toasts; the plugin cannot change their rendering. [refined-mockups correction in project.md]
- Adding a manifest capability can make the admin re-approve the plugin's permissions on upgrade; this is accepted (Q7=A) and must be stated in the release's upgrade notes. [Q7]
- The unit tests use a test double for `TaskCreateDialog`; the `workflowId: null` behaviour must also be checked against real Kandev v0.96.0 (contract test or manual check). [codekb]

## Assumptions

- Kandev v0.96.0's `TaskCreateDialog` with `workflowId: null` selects a workflow of the given workspace and loads its steps (read from Kandev source `components/task-create-dialog-submit.tsx`; to verify in Build and Test).
- FR1.4 (open the dialog when the check fails) is the safer default because the bug being fixed is a false "no workflow" error.
- "Backlog page" in Q4 means everything rendered under the plugin's `/backlog` route, including its dialogs.

## Out of Scope

- The three watch dialogs' workflow lookup (`settings/issue-watch-dialog.tsx`, `git/watch-form.tsx`, `git/scm-watch-form.tsx`) — follow-up intent. Their error display, where rendered on the Backlog page (`git/watch-form.tsx`, `git/scm-watch-form.tsx`), is in FR2 scope. [Q2, Q4]
- Error displays on the Settings page. [Q4]
- Changing Kandev itself.

## Open Questions

- Exact action name and response shape for FR3.2 — settled in Code Generation.
- Whether the release is a patch or minor version — settled at Deployment (a new capability may argue for minor).
