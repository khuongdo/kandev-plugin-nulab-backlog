# Wireframes (sketches) — Kandev Plugin for Nulab Backlog

Inputs: `intent-statement` (Kandev users who use Backlog), `scope-document` (in-scope and out-of-scope capabilities), `intent-backlog` (items IB-1 to IB-14).

## General principles

- Use Kandev's UI components and style, no separate identity. [Q4]
- Two entry points, like the Bitbucket plugin: the integration settings page and the task menu. [Q1]
- Plus a **Backlog** entry under "Integrations" on the sidebar, with three sub-pages: Issues, PR watches, Dashboard. [Q8]
- Runs on desktop and phone. [Q5]
- Display text follows the language Kandev is using. The text in the sketches below is in English only for ease of illustration. [Q6]
- Basic WCAG 2.1 AA accessibility:
  - fully operable by keyboard;
  - every input has a label;
  - status is never shown by colour alone, always with text. [Q7]
- These are low-fidelity sketches, not the final design. The real layout will follow the UI frame Kandev provides to plugins. [assumption]

## Information architecture

```
Kandev
  +-- Settings > Integrations > Backlog       (W1: connect space)
  +-- Sidebar > Integrations > Backlog
  |     +-- Issues                             (W2, W3)
  |     +-- PR watches                         (W4)
  |     +-- Dashboard                          (W5)
  +-- Task board / task menu                   (W6, W7, W10)
  +-- New task > Repository                    (W8)
  +-- Task composer: # reference               (W9)
```

## W1. Connect a Backlog space (IB-2, IB-3)

A single form. [Q3] The user enters the space address and chooses the sign-in method:

- **API key**: shows the API key field.
- **OAuth**: replaces the field with a "Sign in with Nulab" button, which opens the Backlog sign-in page and then returns.

After saving, the API key is never shown again.

```
+--------------------------------------------------------------------+
| Settings > Integrations > Backlog                                  |
|--------------------------------------------------------------------|
| Backlog space address                                              |
| [ myteam.backlog.com                                ]              |
| Example: myteam.backlog.com, myteam.backlog.jp, x.backlogtool.com  |
|                                                                    |
| Sign-in method                                                     |
| ( ) API key        ( ) OAuth (sign in with Nulab)                  |
|                                                                    |
| API key   [ ****************************            ]              |
| How to create an API key in Backlog  (help link)                   |
|                                                                    |
|                                               [ Connect ]          |
|--------------------------------------------------------------------|
| When connected: Connected as Nguyen Van A @ myteam.backlog.com     |
|                               [ Test connection ] [ Disconnect ]   |
+--------------------------------------------------------------------+
```

| State | Display |
|-----------|----------|
| Not connected (empty) | The form above, with one line explaining what the plugin does after connecting |
| Connecting | The Connect button changes to "Connecting..." and is disabled |
| Connected | User name, space address, Test connection and Disconnect buttons |
| Error | Message right under the related field, stating the reason (wrong space address, wrong API key, OAuth cancelled) and how to fix it |
| Disconnect | Confirmation dialog, stating clearly that the credentials will be deleted |

Accessibility: h1 heading "Backlog"; main region; Tab goes to the space address field first.

## W2. Issue list (IB-9, IB-10)

A list page with filters by project, status, assignee and keyword (issue keys such as `PROJ-123` also work). [Q2] Each row has "Create task" and "Link to task" buttons. A linked issue shows the task key.

```
+--------------------------------------------------------------------+
| Integrations > Backlog > Issues                                    |
|--------------------------------------------------------------------|
| Project [ All     v]  Status [ Open  v]  Assignee [ Me    v]       |
| Search  [ keyword or PROJ-123                    ] [ Search ]      |
|--------------------------------------------------------------------|
| PROJ-123  Fix login timeout          In Progress      Linked: T-12 |
|           Assignee: A   Updated 2h ago                  [ Open ]   |
| PROJ-120  Add CSV export             Open                          |
|           Assignee: B         [ Create task ] [ Link to task ]     |
| PROJ-118  Update docs                Open                          |
|           Assignee: -         [ Create task ] [ Link to task ]     |
|--------------------------------------------------------------------|
| Showing 1-20 of 57                          [ < Prev ] [ Next > ]  |
+--------------------------------------------------------------------+
```

| State | Display |
|-----------|----------|
| Empty | "No issues match these filters", with a Reset filters button |
| Loading | Blinking placeholder rows, filters keep their place |
| Error | "Couldn't load issues from Backlog", with the reason and a Retry button |
| API rate limit hit | "Backlog is limiting requests. Retrying in N seconds", retries automatically |
| Not connected | Link to W1 |

Accessibility: h1 heading "Backlog issues"; each row is a list item; Tab moves through the filters, the search box, then each row in turn.

## W3. Link an issue to a task (IB-10)

Clicking "Link to task" in W2 opens a task search dialog.

```
+--------------------------------------------------------------------+
| Link PROJ-120 to a task                                            |
|--------------------------------------------------------------------|
| Search tasks [ CSV                                          ]      |
| ( ) Export report as CSV          In Progress                      |
| ( ) CSV import spike              Todo                             |
|                                                                    |
|                                           [ Cancel ] [ Link ]      |
+--------------------------------------------------------------------+
```

Accessibility: the dialog keeps focus inside; Esc closes it; focus returns to the "Link to task" button after closing.

## W4. Pull request watches (IB-13)

List of watches. Each watch has a name, repository, filter and state, with run, pause, resume, edit and delete buttons. The state is written as text (Active / Paused), not by colour alone.

```
+--------------------------------------------------------------------+
| Integrations > Backlog > PR watches                                |
|--------------------------------------------------------------------|
| [ + New watch ]                                                    |
|--------------------------------------------------------------------|
| Name            Repository      Filter            State            |
| My open PRs     web-app         assignee = me     Active           |
|                          [ Run ] [ Pause ] [ Edit ] [ Delete ]     |
| Review queue    api             status = Open     Paused           |
|                         [ Run ] [ Resume ] [ Edit ] [ Delete ]     |
+--------------------------------------------------------------------+
```

Empty state: "No PR watches yet", with a New watch button. Delete needs confirmation.

## W5. Saved dashboard (IB-13)

Pick a saved query to see the list of pull requests with their status and linked tasks.

```
+--------------------------------------------------------------------+
| Integrations > Backlog > Dashboard                                 |
|--------------------------------------------------------------------|
| Saved query [ Open PRs in web-app    v]   [ + Save query ]         |
|--------------------------------------------------------------------|
| web-app #42  Add CSV export    Open    2 reviewers   Task: T-17    |
| web-app #41  Fix login         Merged                Task: T-12    |
| api #9       Rate limit retry  Open    0 reviewers   (no task)     |
+--------------------------------------------------------------------+
```

## W6. Task board and task menu (IB-6, IB-7, IB-8)

The task card shows a Backlog issue label and a pull request label (status, number of approvals). The task menu has Backlog actions. [Q1]

```
+--------------------------------------------------------------------+
| Task board (desktop)                                               |
|--------------------------------------------------------------------|
| +-------------------------------------------------------------+    |
| | Export report as CSV                                        |    |
| | [BL PROJ-120 Open]  [PR #42 Open - 1/2 approved]            |    |
| +-------------------------------------------------------------+    |
|                                                                    |
| Task menu ( ... ):                                                 |
|   Open Backlog issue                                               |
|   Link Backlog pull request                                        |
|   Create Backlog pull request   (after the branch is pushed)       |
|   Unlink Backlog issue / pull request                              |
+--------------------------------------------------------------------+
```

## W7. Task board on phone (IB-8)

On phones, labels stack and are shortened. [Q5]

```
+--------------------------------------------------------------------+
| Task board (mobile)                                                |
|--------------------------------------------------------------------|
| Export report as CSV                                               |
| PROJ-120 Open                                                      |
| PR #42 Open                                                        |
+--------------------------------------------------------------------+
```

## W8. Pick a Backlog repository when creating a task (IB-5)

```
+--------------------------------------------------------------------+
| New task > Repository                                              |
|--------------------------------------------------------------------|
| Repository [ Backlog: myteam / web-app             v]              |
|            Backlog: myteam / api                                   |
|            (other providers ...)                                   |
| Base branch [ main v]                                              |
+--------------------------------------------------------------------+
```

## W9. `#` reference to an issue (IB-11)

Typing `#` followed by a key or keyword opens a list of issue suggestions. Use the arrow keys to pick and Enter to insert.

```
+--------------------------------------------------------------------+
| Task composer                                                      |
|--------------------------------------------------------------------|
| Please fix #PROJ-12_                                               |
|    +--------------------------------------+                        |
|    | PROJ-123  Fix login timeout          |                        |
|    | PROJ-124  Login page typo            |                        |
|    +--------------------------------------+                        |
+--------------------------------------------------------------------+
```

## W10. Create a pull request (IB-7)

Opened from the task menu after Kandev has pushed the branch. Title and description are pre-filled from the task and the linked issue.

```
+--------------------------------------------------------------------+
| Create Backlog pull request                                        |
|--------------------------------------------------------------------|
| Repository    web-app                                              |
| From branch   task/export-csv    ->   Into [ main v]               |
| Title         [ Add CSV export                          ]          |
| Description   [ Closes PROJ-120 ...                     ]          |
| Related issue PROJ-120 (from the linked issue)                     |
|                                      [ Cancel ] [ Create PR ]      |
+--------------------------------------------------------------------+
```

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q8]: answers in `rough-mockups-questions.md`.
- IB-…: items in `ideation/scope-definition/intent-backlog.md`.

## Assumptions & Open Questions

- [assumption] The UI extension points (settings page, entry under Integrations, task menu, labels on the task card, `#` suggestions) all exist in Kandev's plugin frame, because the Bitbucket plugin already uses them. Exact placement will be confirmed in the Refined Mockups step.
- [assumption] Status sync only goes one way, so there is no screen for changing issue status in Kandev (per `scope-document`).
