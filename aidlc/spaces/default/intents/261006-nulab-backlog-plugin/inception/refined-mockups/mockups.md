# Refined Mockups — Kandev Plugin for Nulab Backlog

Inputs:

- `wireframes` (W1–W10) and `user-flow` (F1–F7) from the rough-mockups step.
- `stories` (US1.1–US8.5).
- `requirements` (FR, NFR).
- `team-practices`.
- Answers Q1–Q6 in `refined-mockups-questions.md`.

## General Principles

- **Style**: use Kandev's components and style, with no separate identity. The component mapping is in `design-system-mapping.md`.
- **Display text**: the mockups write text in English for illustration. The real text follows the language in use in Kandev (US8.5).
- **Phones**: checked at 320px width. Multi-column lists switch to vertically stacked cards, and filters collapse into a "Filters" button (Q3).
- **Five states**: every screen has all of empty, loading, with data, error, and partial or edge. The state table for each screen is in `interaction-spec.md`.
- **Accessibility**: each screen has one note line about heading level, content region (landmark) and keyboard entry point. The full checklist is in `accessibility-checklist.md`.

## Screen Table

| ID | Screen | Story | Rough-mockup source |
|----|----------|-------|-----------------|
| M1 | Connect space and Git info | US1.1–US1.9, US5.5 | W1 |
| M2 | Issue list (desktop) | US2.1–US2.3, US3.1 | W2 |
| M2m | Issue list (phone) | US2.1, US2.2 | New (Q3) |
| M3 | Link task dialog | US3.3 | W3 |
| M4 | PR watches and watch form | US6.1, US6.2 | W4 |
| M5 | Dashboard and save query form | US6.3 | W5 |
| M6 | Task card, labels and task menu | US2.3, US4.1, US5.2–US5.4 | W6, W7 |
| M7 | Link pull request dialog | US5.2 | New (Q1) |
| M8 | Backlog block in the task detail sidebar | US3.2, US3.5 | New (Q2) |
| M9 | Choose repository when creating a task | US5.1 | W8 |
| M10 | `#` reference suggestions | US3.4 | W9 |
| M11 | Create pull request form | US5.3 | W10 |
| M12 | "No longer connected" state and restore | US1.5, US1.8, US1.9 | New (Q4) |

The mockups from the rough-mockups step (W1–W10 in `wireframes.md`) keep their layout. The sections below only record what **changes or is added**, with mockups for the new screens.

## M1. Connect Space (from W1)

Additions:

- The **Git username** and **Git password** fields sit in a separate, collapsible block named "Git access (optional)". This block is only needed when the source code is on Backlog (US5.5).
- While checking, the button changes to "Connecting..." or "Testing..." and is disabled. The result is announced to screen readers.
- The **"Sign in again"** state has a sign-in-again button right on the page (US1.4).
- **Project selection** is a checkbox list with a search field. Deselecting a project in use shows a confirmation dialog stating the number of links and watches that will be disabled (US1.9).
- **Update interval**: a minutes field, default 5, minimum 1 (US4.2).
- The **Disconnect** and **change space** confirmation dialogs state the number of links and watches that will be disabled, and say clearly that this affects **everyone in the workspace**. Default focus is on Cancel.

Accessibility note: h1 "Backlog"; main region; Tab goes to the space address field first.

## M2. Issue List (from W2)

Additions:

- Linked rows still have "Create task" and "Link to task", grouped into the row's "…" menu to reduce clutter. The task key shows as a link (US2.3).
- At the top of the page there is an "updated at …" line and a refresh button (US4.2).
- Long or Japanese titles are truncated to 2 lines with a "…". The full title shows when the row has focus or is tapped.

Accessibility note: h1 "Backlog issues"; main region; Tab goes to the Project filter.

## M2m. Issue List on Phones (new)

```
+--------------------------------------------------------------------+
| Integrations > Backlog > Issues (320px)                            |
|--------------------------------------------------------------------|
| [ Filters (2) ]   [ Search            ]                            |
| +--------------------------------------------------------------+   |
| | PROJ-123  In Progress                                        |   |
| | Fix login timeout when the session                           |   |
| | expires during upload                                        |   |
| | Assignee: A    Linked: T-12, T-15                            |   |
| | [ Create task ]  [ Link to task ]                            |   |
| +--------------------------------------------------------------+   |
| Showing 1-20 of 57                        [ < ]  [ > ]             |
+--------------------------------------------------------------------+
```

- The "Filters (2)" button opens a drawer with the filters. The number in brackets is the number of active filters.
- Each issue is a card. Buttons have a touch target of at least 44x44px.
- PR watches (M4) and Dashboard (M5) use the same card pattern on phones.

Accessibility note: h1 "Backlog issues"; each card is one item in the list; Tab goes to the Filters button.

## M3. Link Task Dialog (from W3)

Additions:

- A task already linked to another issue shows the text "Linked to PROJ-118" and cannot be selected (US3.3).
- If no task matches, show "No tasks found" and disable the Link button.

Accessibility note: dialog (role dialog) with an h2 heading; focus goes to the task search field; Esc to close, focus returns to the button that opened it.

## M4. PR Watches and Watch Form (from W4)

The "New watch" and "Edit watch" forms open as dialogs (Q6). The filters only have the fields Backlog supports (US6.1).

```
+--------------------------------------------------------------------+
| New PR watch                                                       |
|--------------------------------------------------------------------|
| Name *         [ My open PRs                         ]             |
| Repository *   [ web-app                           v]              |
| Status         [x] Open  [ ] Closed  [ ] Merged                    |
| Assignee       ( ) Anyone  (*) Me  ( ) Choose user...              |
| Linked issue   [ optional: PROJ-120                ]               |
| Creator        [ optional                          v]              |
|                                                                    |
| Creates a Kandev task for each new matching PR,                    |
| at most 10 tasks per cycle.                                        |
|                                          [ Cancel ] [ Save ]       |
+--------------------------------------------------------------------+
```

On the watch list:

- Each row has Run, Pause or Resume, Edit, Delete buttons. Status is written as text: Active, Paused or "Not connected".
- Each row has progress, for example "Created 10/25 tasks, the rest in later cycles", and the last run time (US6.2).
- The Delete confirmation dialog states clearly that tasks already created are **not** deleted.

Accessibility note: h1 "PR watches"; main region; Tab goes to the "New watch" button.

## M5. Dashboard and Save Query Form (from W5)

```
+--------------------------------------------------------------------+
| Save PR query                                                      |
|--------------------------------------------------------------------|
| Name *         [ Open PRs in web-app                 ]             |
| Repository *   [ web-app                           v]              |
| Status         [x] Open  [ ] Closed  [ ] Merged                    |
| Assignee       (*) Anyone  ( ) Me  ( ) Choose user...              |
|                                                                    |
|                                          [ Cancel ] [ Save ]       |
+--------------------------------------------------------------------+
```

- The saved query list has Edit and Delete.
- It has empty, loading and error states, the same as M2.

Accessibility note: h1 "Backlog dashboard"; main region; Tab goes to the query select field.

## M6. Task Card and Task Menu (from W6, W7)

Changes:

- The PR label drops the "1/2 approved" part because Backlog has no reviewers. The label only shows status and assignee, for example "PR #42 Open – Lan" (US5.4).
- The issue label shows "updated at …" when the label has focus or is tapped. If the data is stale it shows the text "may be out of date" (US4.1).
- On phones the label is shortened to "PR #42 Open", but screen readers still read it in full.
- Disabled menu items can still receive focus, so the reason can be read.

Accessibility note: labels are text, not shown by color alone; the menu follows the menu button pattern.

## M7. Link Pull Request Dialog (new)

The user enters a PR number or pastes a PR link (Q1). The plugin checks the PR and shows the result before Link is clicked.

```
+--------------------------------------------------------------------+
| Link Backlog pull request to T-17                                  |
|--------------------------------------------------------------------|
| Pull request number or URL                                         |
| [ 42 or https://myteam.backlog.com/.../pullRequests/42         ]   |
| Repository  [ web-app             v]   (needed for a number only)  |
|                                                                    |
| Found: #42 Add CSV export   Open   task/export-csv -> main         |
|                                                                    |
|                                            [ Cancel ] [ Link ]     |
|--------------------------------------------------------------------|
| Error: Pull request #999 not found in web-app                      |
+--------------------------------------------------------------------+
```

- If a link is pasted, the plugin detects the repository itself and the Repository field is disabled.
- Links from another space, or links that are not Backlog PR links, are rejected with a reason.
- The Link button is only enabled after the PR is found.

Accessibility note: dialog with an h2 heading; focus goes to the input; the "Found" result or error is announced via the status region.

## M8. Backlog Block in the Task Detail Sidebar (new)

A collapsible block in the sidebar (Q2). Data is fetched directly from Backlog and is read-only (US3.2, US3.5).

```
+--------------------------------------------------------------------+
| Task T-17 > sidebar                                                |
|--------------------------------------------------------------------|
| v Backlog                                     updated at 10:42     |
|   PROJ-120  Add CSV export                       [Open in Backlog] |
|   Status    In Progress                                            |
|   Assignee  Lan        Priority  High        Due  2026-10-20       |
|   > Comments (150)                                    [Load more]  |
|   > Attachments (2)                                                |
|       spec.pdf  1.2 MB   [Preview]                                 |
|       dump.zip  48 MB    too large to preview  [Open in Backlog]   |
+--------------------------------------------------------------------+
```

- Newest comments show first, one more page per load.
- Files over the size limit only show the name and a link to Backlog.
- An error in the comments section does not break the other sections in the block.
- The block remembers its expanded or collapsed state for each user.

Accessibility note: the block heading is an h2 inside the complementary region; each subsection is an expand button with `aria-expanded`.

## M9–M11. Choose Repository, `#` Suggestions, Create PR (from W8–W10)

Changes:

- **M9**: If Git access is not saved yet, the Backlog entry still shows, with the text "Git access not set up" and a link to M1. If the list cannot be loaded, show an error with Retry, and other repository sources can still be chosen.
- **M10**: Selection with the arrow keys and Enter is handled by Kandev. The plugin only provides search results, and waits a beat after the user stops typing before searching.
- **M11**: The prefilled description has the reference "Related: PROJ-120", and does not use keywords that could close the issue automatically. The title field is required. On create the button changes to "Creating..." and is disabled. When done it shows a link to open the PR.

## M12. "No Longer Connected" and Restore (new)

On disconnect, space change or project deselection, links and watches switch to the "no longer connected" state:

```
+--------------------------------------------------------------------+
| Task board card                                                    |
|--------------------------------------------------------------------|
| Export report as CSV                                               |
| [BL PROJ-120 - not connected]                                      |
| Hint: reconnect myteam.backlog.com to restore this link            |
+--------------------------------------------------------------------+
```

When reconnecting to the same old space or reselecting the same old project, links and watches are restored automatically, and watches return to the Paused state (Q4):

```
+--------------------------------------------------------------------+
| Settings > Integrations > Backlog                                  |
|--------------------------------------------------------------------|
| Connected as Nguyen Van A @ myteam.backlog.com                     |
|                                                                    |
| (i) Restored 5 links and 2 PR watches from your earlier            |
|     connection to this space. PR watches are Paused.               |
|                                             [ Review watches ]     |
+--------------------------------------------------------------------+
```

Accessibility note: the restore message is announced via the status region; the "Review watches" button opens M4.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q6]: answers in `refined-mockups-questions.md`.
- `wireframes`, `user-flow`: `ideation/rough-mockups/`. `stories`: `inception/user-stories/stories.md`. `requirements`: `inception/requirements-analysis/requirements.md`. `team-practices`: `inception/practices-discovery/team-practices.md`.

## Assumptions & Open Questions

- [assumption] Kandev has a place for a plugin to mount a block in the task detail sidebar (the developer saw the `task-sidebar` slot when contributing at the story step). The exact position is decided by Kandev.
- [assumption] The PR label on the card is rendered by Kandev from data the plugin provides, so the shortening on phones may differ slightly from the mockup.
- The preview file size limit (M8) will be settled at the functional design step.
