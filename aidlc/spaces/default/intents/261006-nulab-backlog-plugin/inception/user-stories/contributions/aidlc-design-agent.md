**Collaborator:** aidlc-design-agent

## Contribution

Viewpoint: user experience, persona fidelity, accessibility (WCAG 2.1 AA, NFR9) and phone use (FR5.4). Checked against `wireframes.md` (W1–W10) and the open feedback in `reviews/review-01.md` (R-01 to R-07). The new AC IDs below continue each story's numbering so the lead can merge them straight into `stories.md`.

### 1. Persona fidelity

- **Keep one persona, Minh**, matching Q1 = C and Q7 = B. Propose adding three rows to the persona table, because they directly affect the ACs:
  - *Context — shared connection*: when Minh connects, changes the space or disconnects, **everyone in the workspace** is affected (FR1.1). This is why the confirmation dialogs in US1.5 and US1.6 must state the consequences clearly.
  - *Context — devices*: Minh works mainly on a computer, but checks the task board on a phone when away from the desk (FR5.4, W7). The persona does not mention phones yet, so US5.4 has nothing to rest on.
  - *Context — data*: Backlog is a Japanese product; issue titles may be in Japanese, Vietnamese or very long. The UI must handle CJK text and long titles (see section 4).
- **AC7.1.2** says "admin installs the package" — this departs from the single persona. Proposal: "When I (with Kandev admin rights) install the package through Settings > Plugins".
- **US7.3, US7.4**: per Q7 = B, the "so that …" part should state the value Minh gets when *using* the plugin, not the value for a maintainer. Proposals:
  - US7.3: "…, so that the plugin build I install carries no known bugs and does not leak the API key."
  - US7.4: "…, so that I can install the plugin on the team's Kandev without upgrading Kandev."

### 2. Missing UI states (per W1–W10)

| Story | Proposed AC (Given/When/Then) | Reason |
|---|---|---|
| US1.1 | **AC1.1.4** Given I clicked Connect, When the plugin is waiting for Backlog, Then the button changes to "Connecting..." and is locked, and clicking many times does not create two connections. | The "Connecting" state of W1; prevents duplicate submits. |
| US1.1 | **AC1.1.5** Given the space address has the right format but cannot be reached (network error, timeout), When I click Connect, Then the message says clearly "could not connect to the space" — different from "invalid API key" — and nothing is stored. | The user must know which field to fix. |
| US1.4 | Add to **AC1.4.2**: the "Sign in again" state has a sign-in-again button right on the settings page, and other Backlog pages show a link to the settings page (not just an error). | R-05; a state must have a way out. |
| US1.5 | **AC1.5.3** Given it is connected, When I click Disconnect, Then the confirmation dialog states that the credentials will be deleted and **everyone in the workspace** will lose the Backlog connection; the default focus is on the Cancel button. | W1 "Disconnect"; shared connection. |
| US1.5 | Add to **AC1.5.1**: during the check, the Test connection button is locked and reads "Testing..."; the result is read out to the screen reader (`aria-live="polite"` region). | The running state is missing. |
| US1.6 | Add to **AC1.6.2**: the space switch confirmation dialog states **exact numbers** "N issue/PR links and M PR watches will be turned off" before the user confirms. | Preventing errors beats reporting them; the consequence cannot be undone. |
| US2.1 | **AC2.1.4** Given the list is loading, When waiting for Backlog, Then placeholder rows show, the filters stay visible, and the screen reader is told "loading". | The "Loading" state of W2. |
| US2.1 | **AC2.1.5** Given the workspace is not connected, When I open the Issues page, Then "Backlog is not connected" shows with a link to the settings page, and no error shows. | The "Not connected" state of W2. |
| US2.1 | **AC2.1.6** Given I am on page 1, When I click Next, Then focus moves to the top of the list and the line "Showing 21-40 of 57" is announced to the screen reader. | Focus management when paging. |
| US3.1 | **AC3.1.4** Given I click "Create task", When the task is being created, Then the button is locked ("Creating..."); when done, the message "Created T-18" shows with a link to the task, and the issue row updates with the task key right away. | Prevents duplicates from double clicks; clear success feedback. |
| US3.1 | Add to **AC3.1.2**: the warning dialog has three choices "Open T-17", "Create a new task anyway", "Cancel"; the default focus is not on "Create a new task anyway". | The safe choice is the default. |
| US3.3 | **AC3.3.4** Given the "Link to task" dialog is open, When the search box matches no task, Then "No tasks found" shows and the Link button is locked. | W3 lacks an empty state. |
| US3.3 | **AC3.3.5** Given the "Link to task" dialog is open, When I press Esc or Cancel, Then the dialog closes, nothing changes, and focus returns to the "Link to task" button of that same row. | The W3 accessibility note is not an AC yet. |
| US3.4 | **AC3.4.3** Given the suggestion list is open, When I press Esc, Then the list closes and the typed text is kept; the number of suggestions is announced to the screen reader. | Standard combobox pattern. |
| US4.1 | **AC4.1.4** Given an issue/PR label on the task card, When I move keyboard focus onto it or tap the label, Then I see the "updated at …" time (not only on mouse hover). | FR4.4 currently only covers the list page (AC4.2.3), not the labels. |
| US4.1 | **AC4.1.5** Given the most recent cycles all failed, When I view the task card, Then the label still shows the last known status with the old "updated at …", and a text sign that the data may be stale. | Users cannot tell stale data from fresh data (R-05). |
| US4.2 | Add to **AC4.2.3**: while refreshing, the button is locked and has status text; if Backlog is limiting API calls, "Backlog is limiting requests" shows instead of staying silent. | R-05. |
| US5.1 | **AC5.1.4** Given repositories cannot be loaded from Backlog, When I open the Repository field while creating a task, Then other repository sources can still be chosen and the Backlog entry shows an error message with Retry. | W8 lacks an error state; a plugin error must not block Kandev's task creation flow. |
| US5.3 | Add to **AC5.3.1**: the Title field is required, with an error right below the field if left empty; while creating, the "Creating..." button is locked; when done, there is a link to open the PR on Backlog. | W10 lacks validation and submitting states. |
| US6.1 | **AC6.1.4** Given there is no watch yet, When I open the PR watches page, Then "No PR watches yet" shows with a New watch button. | The empty state of W4. |
| US6.1 | **AC6.1.5** Given the New watch form, When I click Save with no name or no repository, Then an error shows right below the missing field and no watch is created. | R-03: there is no New/Edit watch form yet. |
| US6.2 | **AC6.2.4** Given a watch is creating tasks gradually under the 10-tasks-per-cycle limit, When I view the PR watches page, Then the watch row shows progress in text, for example "Created 10/25 tasks, the rest in later cycles", with the last run time. | Without this, users will think the watch is broken when they only see 10 tasks (Q6 = B). |
| US6.3 | **AC6.3.3** Given Backlog errors or is limiting API calls, When I choose a query, Then an error message shows with Retry, like the Issues page. | R-03, R-05: W5 lacks loading and error states. |
| US7.2 | Add to **AC7.2.2**: the message has a "Retrying in N seconds" countdown (W2), read to the screen reader once, not re-read every second. | Avoids disturbing screen reader users. |

### 3. Contradictions and unclear points for the lead to decide

- **AC2.3.2 vs AC3.1.2/FR3.4**: AC2.3.2 implies only *unlinked* issues have the "Create task" and "Link to task" buttons (W2 too: a linked row only has Open). But FR3.4 and AC3.1.2 allow an issue to have many tasks. Proposal: change AC2.3.1 so a linked row shows the task keys (as links to the tasks) **and still has** "Create task" and "Link to task"; note for Refined Mockups to put these two buttons in the row's "…" menu to reduce clutter.
- **AC3.3.2 — blocking after selection**: propose preventing the error right in the dialog: a task already linked to another issue shows with the text "Linked to PROJ-118" and cannot be selected. The plugin still rejects on the server side as in the current AC (as a second line of defence).
- **US3.1 "one click"**: R-06 asked whether a repository must be chosen before creating a task — not yet answered. The story says "one click", so propose stating it as an assumption: the task is created without an intermediate dialog; the repository is chosen later in the task. If not, AC3.1.1 must change.
- **US5.2 — the PR linking dialog has no screen (R-02, Must)**: AC5.2.1 says "choose PR #42" but no W covers this. A dependency on Refined Mockups must be recorded, and the lead must decide: how many PRs can a task be linked to? Can one PR be linked to many tasks? FR5.2 does not say.
- **US5.3 — "Closes PROJ-120"** (R-07): the prefilled description must not contain a keyword that could make Backlog close the issue, because sync is one-way. Propose AC5.3.1 state "the description has a reference to PROJ-120" rather than leaving it open.
- **US6.1 lacks "run" and "edit"**: FR6.1 lists create, edit, delete, **run**, pause, resume; the story has no AC for Run and Edit. Proposals:
  - **AC6.1.6** Given any watch, When I click Run, Then the watch runs one cycle immediately and the last run time changes accordingly.
  - **AC6.1.7** Given a watch, When I edit the filter and save, Then the next cycle uses the new filter and tasks already created stay as they are.
  - Add to **AC6.1.3**: the delete confirmation dialog states clearly that tasks created by the watch are **not** deleted.
- **US1.7 — unselecting a project that has links**: no AC yet. Propose adding to Open Questions: are the links and watches of an unselected project turned off as in AC1.6.2, or only hidden from the lists?

### 4. Accessibility and phones

US8.2 currently packs all of NFR9 into one AC (AC8.2.1), which is hard to test and does not cover the automated scan that NFR9 lists in its pass criterion. Proposal to split:

- Change the scope of **AC8.2.1**: "every Backlog screen and dialog (W1–W10, plus the confirmation dialogs, the PR linking dialog and the watch form added in Refined Mockups)".
- **AC8.2.3** Given every Backlog screen, When an automated scanner runs (for example axe), Then there are no level A/AA issues. (Traces to the NFR9 pass criterion.)
- **AC8.2.4** Given any element receives focus, When using the Tab key, Then the focus ring is clearly visible, and Tab order follows reading order.
- **AC8.2.5** Given any dialog (W3, PR linking, W10, the confirmation dialogs), When the dialog opens, Then focus stays inside the dialog, Esc closes the dialog, and focus returns to the button that opened it.
- **AC8.2.6** Given a background action finishes (loading done, error, task created, being API limited), When the result appears, Then the screen reader is told through an `aria-live="polite"` region.
- **AC8.2.7** Given a colored issue/PR status label, When contrast is measured, Then the text meets at least 4.5:1 and the label border 3:1.
- **AC8.2.8** Given Kandev is set to a language the plugin has no translation for yet, When I open a Backlog page, Then the text shows in English, without raw translation keys (for example `backlog.issues.empty`). (The missing error/edge case of AC8.2.2.)

For US5.4 and phones:

- **AC5.4.2**: change 375px to **320px** (the WCAG 1.4.10 reflow level), and add: buttons and menu items on the phone have a touch target of at least 44x44px.
- **AC5.4.4** Given W7 shortens the PR label on the phone (only "PR #42 Open"), When using a screen reader or opening the task detail, Then the full information "Pull request #42, Open, 1 of 2 people approved" is still available.
- **AC5.3.2**: a locked menu item must still receive keyboard focus (`aria-disabled`, not `disabled`) so keyboard and screen reader users can read the reason "Branch not pushed yet".
- Note for Refined Mockups (R-04): on the phone, W2, W4, W5 turn multi-column tables into vertically stacked cards, and the filters collapse into a "Filters" button. No extra story needed; it should be attached to the scope of AC8.2.1/AC5.4.2.
- Long and CJK text: add an edge case to **AC2.1.1** — Given an issue title of 200 characters or in Japanese, When I view the list, Then the title wraps or is shortened with "…" without breaking the layout, and the full title is still readable.

### 5. Writing

- **The language of UI strings in the ACs is mixed**: AC1.1.2, AC1.7.2, AC3.1.2 quote Vietnamese strings; AC2.1.2, AC2.1.3, AC1.3.2 quote English strings. Propose adding to the "Conventions" section: "UI strings in quotes are illustrative; the real text follows the Kandev language (NFR10). Tests compare by meaning or by translation key, not word for word." If one standard language is needed for the ACs, use English like the wireframes.
- **Terms** (R-07): "PR watch", "OAuth", "API key", "Integrations" have no explanation. Propose adding a short glossary table at the top of `stories.md`, or pointing to the glossary that will exist in Refined Mockups.
- **US8.1** "opens fast, so that I do not have to wait": acceptable because AC8.1.1 already has a measurable threshold; better rewritten as "shows within 3 seconds for 95% of opens" to match NFR1.
- **AC1.6.2** "shows the label 'no longer connected'": should state where the label shows (task card, watch row) and what action comes with it (unlink or leave as is).

## Positions

- AGREE: One persona, Minh, for every story, including release — matches Q1 = C and Q7 = B as confirmed by the user.
- AGREE: Group stories by FR1–FR7 and mark [Skeleton] on US1.1, US1.2, US7.1, US7.2 — matches Q2 and the walking skeleton in `team.md`.
- AGREE: AC5.4.3 requires status to always have text, not only color — matches NFR9 and the wireframes' principles.
- OBJECT: AC2.3.2 implies a linked issue loses the "Create task" button — contradicts AC3.1.2 and FR3.4 (one issue, many tasks).
- OBJECT: AC3.3.2 only rejects after selection — it should be prevented right in the W3 dialog (error prevention), keeping server-side rejection as a second layer.
- OBJECT: US6.1 lacks ACs for "run" and "edit" — FR6.1 lists both, so the story does not fully trace the requirement.
- OBJECT: AC5.4.2 uses 375px — the WCAG 2.1 AA reflow check level (1.4.10) is 320px.
- OBJECT: AC7.1.2 uses the "admin" role — departs from the single persona settled in Q1/Q7.
- OBJECT: US8.2 packs all of NFR9 into one AC and does not cover the automated scan that NFR9 lists in its pass criterion — it needs splitting as in section 4.
- OBJECT: US5.2 (Must) is written as if the PR linking screen already exists — R-02 is still open; a dependency on Refined Mockups must be recorded and the number of PRs per task settled.
