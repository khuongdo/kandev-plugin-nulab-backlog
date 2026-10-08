# Requirements Analysis Questions — Link Task modal (GitHub-style)

Initial request: "Sưa modal của Link Task bắt chước giống cách mà github integration làm".

Context: today the plugin links from the issue side: on the `/backlog` page, the issue row "..." menu opens "Link to task", a custom dialog with a search box and a clickable task list (`ui/src/issues/link-task-dialog.tsx`). Kandev's GitHub integration links the other way. You open the task's **Link** menu, pick "GitHub Issue", and a host dialog asks for one field, "issue URL or number", with Cancel / Save and a success toast. Plugins get that same host dialog through `openTaskLinkDialog`. The plugin already uses it for "Link Backlog pull request" (`ui/src/git/pr-link.ts`), but not for Backlog issues.

## Question 1
What should "mimic the GitHub integration" mean for the Link Task modal?

A. Add a "Link Backlog issue" item to the task's Link menu that opens the host dialog (exactly like GitHub Issue / Backlog pull request), and keep the existing issue-side "Link to task" dialog unchanged
B. Add that task-side "Link Backlog issue" item, and remove the issue-side "Link to task" entry and dialog
C. Keep only the issue-side "Link to task" flow, but restyle its dialog to the GitHub dialog shell (description line, wider `sm:max-w-lg`, red inline error under the field, Cancel + Save/Saving..., Enter submits, success toast)
D. Both: add the task-side "Link Backlog issue" item AND restyle the issue-side dialog as in C
X. Other (please specify)

[Answer]: D

## Question 2
In the task-side dialog, what can the user type to identify the Backlog issue? The backend currently accepts only an issue key such as `PROJ-123`. (Ignore this question if you chose C in Question 1.)

A. Issue key (`PROJ-123`) or the full Backlog issue URL (`https://<space>.backlog.com/view/PROJ-123`); the URL is parsed to the key
B. Issue key only
C. Issue key, full URL, or a bare number (`123`) when only one Backlog project is selected for the workspace
X. Other (please specify)

[Answer]: A

## Question 3
When the task is already linked to a Backlog issue, how should the task-side "Link Backlog issue" item behave? The host dialog cannot show an Unlink button or prefill the current issue. Today the backend refuses to link a task to a second issue (conflict), and unlinking is the separate "Unlink Backlog issue" task-menu item.

A. Hide "Link Backlog issue" while the task is linked; the user unlinks first with the existing "Unlink Backlog issue" item
B. Always show it; linking to a different issue shows an inline error "This task is already linked to PROJ-1. Unlink it first."
C. Always show it; linking to a different issue replaces the existing link (backend change)
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Q1 (D): do both. Add a "Link Backlog issue" item to the task's Link menu that opens the host dialog through `openTaskLinkDialog`, like GitHub Issue and "Link Backlog pull request". Also keep the issue-side "Link to task" flow, with its search and task list, but restyle its dialog to the GitHub dialog shell: description line, `sm:max-w-lg` width, red inline error under the field, Cancel + Save/Saving..., Enter submits, success toast.
- Q2 (A): the task-side dialog accepts an issue key (`PROJ-123`) or a full Backlog issue URL. The UI extracts the key from the URL before calling `issues.link`.
- Q3 (A): hide "Link Backlog issue" while the task is linked. To change the issue, the user first uses the existing "Unlink Backlog issue" task-menu item.

Does this all look correct before I generate the requirements artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
