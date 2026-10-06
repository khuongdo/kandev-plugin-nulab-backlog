# Interaction Spec — Kandev Plugin for Nulab Backlog

Inputs: `mockups.md` (M1–M12), `stories` (US…), `requirements` (FR, NFR), `wireframes` and `user-flow` from the rough-mockups step, `team-practices`.

Format follows `component-spec-template.md`. Display text is written in English for illustration; the real display text follows Kandev's language.

## Shared Components

### BacklogListState

| Field | Value |
|---|---|
| Component | BacklogListState |
| Description | Shared state frame for every list: issues, watches, dashboard, repositories |
| Category | feedback |

#### States

| State | Description | Trigger |
|---|---|---|
| loading | Pulsing placeholder rows; filters stay visible | When loading or refreshing starts |
| success | List has data, with "updated at …" | Load complete |
| empty | "No … match these filters" and a Reset filters button | Empty result |
| error | Error of one of three kinds: reconnect needed, rate limited, unreachable. Has a Retry button | Backlog returns an error or times out |
| rate-limited | "Backlog is limiting requests. Retrying in N seconds", retries automatically | Receives a 429 code (US8.4) |
| not-connected | "Backlog is not connected" and a link to M1 | Not connected yet or disconnected |
| no-project | "No project selected" and a link to M1 | No project selected yet |

#### Responsive Behaviour

| Breakpoint | Behaviour |
|---|---|
| mobile (<768px) | Table switches to vertically stacked cards; filters collapse into a "Filters (n)" button that opens a drawer |
| tablet (768–1024px) | Condensed table: hides the "Updated" column |
| desktop (>1024px) | Full table |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | Lists use `list`/`listitem`; messages use `status` |
| Keyboard interaction | Tab through filters → search field → each row → pagination |
| Label / aria-label | Every filter has a visible label |
| Contrast ratio | WCAG AA (4.5:1 for text, 3:1 for components) |
| Screen reader | Reads "loading" once; reads results and errors via `aria-live="polite"`; the rate-limit countdown is read only once |
| Focus management | After changing page, focus goes to the top of the list |

### ConfirmDialog

| Field | Value |
|---|---|
| Component | ConfirmDialog |
| Description | Confirmation dialog for actions with consequences: Disconnect, changing space, deselecting a project, deleting a watch, creating a second task |
| Category | feedback |

#### States

| State | Description | Trigger |
|---|---|---|
| default | States the specific consequence, for example "5 links and 2 PR watches will be disabled for everyone in this workspace" | Dialog opens |
| loading | The confirm button changes to "Working..." and is disabled | After clicking confirm |
| error | Error shown inside the dialog; the old connection is kept | Action fails |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | `alertdialog` |
| Keyboard interaction | Default focus on Cancel; Esc to close |
| Focus management | Focus stays inside the dialog; on close it returns to the button that opened it |

## Per-Screen Components

### ConnectionForm (M1)

| Field | Value |
|---|---|
| Component | ConnectionForm |
| Description | Connect a space with an API key or OAuth; enter Git info (optional); select projects; set the update interval |
| Category | input |

#### States

| State | Description | Trigger |
|---|---|---|
| default | Not connected; has a short explanation line | Page opened for the first time |
| connecting | "Connecting..." button disabled | Click Connect |
| connected | "Connected as … @ …", with Test connection and Disconnect | Connection succeeds |
| error-key | Error under the API key field: "Invalid key" | Backlog returns 401 |
| error-address | Error under the address field, stating the correct format | Wrong address (US1.2) |
| error-network | "Couldn't reach Backlog", with Retry | 5xx or timeout |
| sign-in-again | Sign-in-again button right on the page | Refresh token no longer valid |
| restored | Message that links and watches were restored (M12) | Reconnected to the same old space |

#### Props / Inputs

| Prop | Type | Required | Default | Description |
|---|---|---|---|---|
| spaceHost | string | yes | — | Host normalised to lowercase |
| authMethod | `apiKey` \| `oauth` | yes | `apiKey` | Sign-in method |
| hasApiKey | boolean | no | false | Only a flag; the key is never returned |
| hasGitCredential | boolean | no | false | Only a flag |
| projects | string[] | no | [] | Keys of the selected projects |
| pollMinutes | number | no | 5 | Update interval, minimum 1 |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | `form`; the "Sign-in method" group uses `radiogroup` |
| Keyboard interaction | Tab in reading order; Enter to submit |
| Label / aria-label | Every input has a visible label; errors are linked to the field with `aria-describedby` |
| Screen reader | Connection or test results are read via `aria-live="polite"` |
| Focus management | On error, focus moves to the first field with an error |

### IssueRow / IssueCard (M2, M2m)

| Field | Value |
|---|---|
| Component | IssueRow (desktop) / IssueCard (mobile) |
| Description | One issue with status, assignee, linked tasks, and actions |
| Category | display |

#### States

| State | Description | Trigger |
|---|---|---|
| default | Has Create task and Link to task buttons | Issue not linked |
| linked | Task key is a link; actions are grouped into the "…" menu | Issue already has a task |
| creating | "Creating..." disabled; when done shows "Created T-18" | Click Create task |
| duplicate-warning | ConfirmDialog with three choices: Open T-17, Create another task, Cancel | Issue already has a task |
| partial | Long title truncated to 2 lines; the full title shows on focus or on tap | Title longer than 2 lines |

#### Accessibility

| Requirement | Implementation |
|---|---|
| Keyboard interaction | Each row is one item; Tab into the buttons in the row |
| Screen reader | Reads "PROJ-123, Fix login timeout, In Progress, assignee A, linked to T-12 and T-15" |
| Contrast ratio | Status labels reach 4.5:1 for text |

### PullRequestLinkDialog (M7)

| Field | Value |
|---|---|
| Component | PullRequestLinkDialog |
| Description | Link a task to a PR by entering a PR number or pasting a PR link |
| Category | input |

#### States

| State | Description | Trigger |
|---|---|---|
| default | Input empty; Link button disabled | Dialog opens |
| resolving | Checking the PR | After typing or pasting, wait a beat |
| found | "Found: #42 …"; Link button enabled | PR exists |
| not-found | "Pull request #999 not found in web-app" | Backlog returns 404 |
| invalid-url | "This link is not a pull request in myteam.backlog.com" | Link belongs to another space or is malformed |
| linking | "Linking..." button disabled | Click Link |

#### Props / Inputs

| Prop | Type | Required | Default | Description |
|---|---|---|---|---|
| taskId | string | yes | — | Task to link |
| input | string | yes | — | PR number or PR link |
| repository | string | no | the task's repository | Required when entering a number; disabled when pasting a link |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | `dialog` with an h2 heading |
| Screen reader | The found or not-found result is read via the status region |
| Focus management | On open, focus goes to the input; on close it returns to the menu item that opened it |

### BacklogSidebarPanel (M8)

| Field | Value |
|---|---|
| Component | BacklogSidebarPanel |
| Description | Issue info fetched directly from Backlog, read-only, in the task detail sidebar |
| Category | display |

#### States

| State | Description | Trigger |
|---|---|---|
| collapsed / expanded | User expands or collapses; the state is remembered per user | Click the block heading |
| loading | Placeholder rows | Open task detail |
| success | Status, assignee, priority, due date, "updated at …" | Load complete |
| comments-partial | Has a "Load more" button | More than one page of comments |
| attachment-too-large | Shows only the name and a link to Backlog | File exceeds the limit |
| section-error | A failing section (for example comments) has its own Retry; other sections still show | Error loading one section |
| unavailable | "Issue unavailable" | Backlog returns 404 or 403 |
| not-connected | "not connected" label and a hint to reconnect | Link disabled |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | `complementary` with an h2 heading; subsections are expand buttons with `aria-expanded` |
| Screen reader | Due dates are read in full, not as shortened numbers |

### WatchFormDialog, SaveQueryDialog (M4, M5)

| Field | Value |
|---|---|
| Component | WatchFormDialog / SaveQueryDialog |
| Description | Dialog to create or edit a PR watch, and to save a PR query |
| Category | input |

#### States

| State | Description | Trigger |
|---|---|---|
| default | Fields: name*, repository*, status, assignee, linked issue, creator | Dialog opens |
| validation-error | Error right under the missing or wrong field | Click Save |
| saving | "Saving..." disabled | Click Save with a valid form |

#### Accessibility

| Requirement | Implementation |
|---|---|
| ARIA role | `dialog`; status checkboxes sit inside `fieldset`/`legend` |
| Focus management | On open, focus goes to the Name field; on error, focus moves to the first field with an error |

### WatchRow (M4)

| State | Description | Trigger |
|---|---|---|
| active | Text "Active", last run time, progress "Created x/y tasks" | Watch is running |
| paused | Text "Paused"; has a Resume button | Click Pause, or just restored |
| not-connected | Text "Not connected"; no Resume | Disconnect, change space, or deselect a project |
| running | Run button disabled while running | Click Run |

## Shared Interaction Rules

- **Prevent repeated clicks**: every button that sends a request is disabled until there is a result.
- **Wait a beat while typing**: search fields and the PR input only send a request after the user stops typing for a short time, to respect the API rate limit (US8.4).
- **Notifications**: success notifications hide themselves after 5 seconds and have a related link. Error notifications do not hide themselves.
- **One-way sync**: no control in the UI allows editing issue data on Backlog.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q6]: answers in `refined-mockups-questions.md`.
- `mockups.md`, `stories.md`, `requirements.md`, `wireframes.md`, `user-flow.md`, `team-practices.md`.

## Assumptions & Open Questions

- [assumption] Some components (task menu, `#` suggestions, PR label) are rendered by Kandev. For these components the plugin only provides data, so the states here are requirements on the data the plugin returns.
- The typing wait time (debounce) and the notification auto-hide time will be settled at the functional design step.
