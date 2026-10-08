# Requirements - 261008-fix-uiux-backlog

Scope: express. Depth: Minimal. Test Strategy: Minimal. Project type: brownfield (plugin `kandev-plugin-nulab-backlog`, Kandev minimum 0.96.0).

## Intent Analysis

- **Goal**: make the Backlog plugin feel like Kandev's built-in GitHub integration in daily use, and clean up small defects in the Backlog settings page.
- **Type**: UI/UX enhancement and bug fixes, with one small backend addition (store the issue summary on links).
- **Scope**: multi-component but local: `ui/src/` (registration, badge, task top bar, settings), `internal/issues/` (link summary), `internal/git/` (PR `taskStatus`, check only), `internal/plugin/` (no new action expected).
- **Complexity**: simple to standard. One platform limit (Kandev cannot hide a plugin menu entry after load) is handled by a load-time rule ([Q1]).

## Functional Requirements

### FR1 - Backlog issue badge on task rows (like GitHub)
- **FR1.1**: The plugin shall show the linked Backlog issue badge on each task row of Home > Tasks that has an active issue link, by registering the existing badge for Kandev's `task-row-metadata` slot ([desc], [Q2]).
- **FR1.2**: The same badge shall also show on rows of the sidebar task list (`surface: "sidebar"`) ([Q2]).
- **FR1.3**: The Kanban card badge (`task-card-tags`) shall keep working as today ([Q2]).
- **FR1.4**: Hovering or focusing the badge shall show a hover card with the issue key, the issue summary and the current status ([desc], [Q3]). When the summary is not yet known (link created before this change and not yet synced), the hover card shall show the key and status without a summary.
- **FR1.5**: Clicking the badge shall open the issue (`https://<space>/view/<KEY>`) in a new browser tab, keeping the existing https host check (backlog.com, backlog.jp, backlogtool.com), and shall not open the task row itself ([desc]).
- **FR1.6**: A task with no active link shall show no badge; an unavailable or stale link keeps its current visual hint.

Acceptance criteria (Given/When/Then):
- Given a task linked to `PROJ-12` ("Fix login", status "In Progress"), when the user opens Home > Tasks, then the row shows the `PROJ-12` badge; when the user hovers it, then a card shows `PROJ-12`, `Fix login` and `In Progress`; when the user clicks it, then the issue opens in a new tab and the task does not open.
- Given the same task, when the user looks at the sidebar task list, then the row shows the same badge.
- Given a link stored before this change, when the badge is hovered before the next sync, then the card shows the key and status and no empty summary line.

### FR2 - Store the issue summary on links
- **FR2.1**: The backend shall store the Backlog issue summary on the link when the link is created ([Q3]).
- **FR2.2**: The status refresh loop shall update the stored summary whenever it refreshes the status, so older links gain a summary at the next sync and renamed issues show the new summary.
- **FR2.3**: `issues.links.list` shall return the summary in each link view (new optional field); older stored links without it stay readable.

Acceptance criteria:
- Given a new link to an issue with summary "Fix login", when `issues.links.list` is called, then the view carries `summary: "Fix login"`.
- Given a stored link without a summary, when the sync loop refreshes it, then the stored link has the current summary.

### FR3 - Hide Backlog in Home > Integrations when OFF
- **FR3.1**: At page load, the plugin shall add the Home > Integrations entry (and its `/backlog` route) only when Backlog is ON in at least one workspace ([desc], [Q1]).
- **FR3.2**: Turning Backlog on or off after the page has loaded shall take effect in the menu after a page reload; no Kandev change is made ([Q1]).
- **FR3.3**: The Settings > Integrations card with its on/off switch shall always stay registered, so Backlog can be turned back on.
- **FR3.4**: If the ON/OFF state cannot be read at load (for example the call fails or times out), the plugin shall show the entry (fail open), so a temporary error never hides a working integration.
- **FR3.5**: This requirement supersedes the earlier rules that the entry is always shown (BR5.4, BR7.6, BR7.8) and the test "registers the entry and the route even when Backlog is off everywhere".

Acceptance criteria:
- Given Backlog is OFF in every workspace, when Kandev loads, then Home > Integrations has no Backlog entry and Settings > Integrations still shows the Backlog card.
- Given Backlog is ON in one workspace, when Kandev loads, then the Backlog entry is shown.
- Given the ON/OFF state cannot be read, when Kandev loads, then the Backlog entry is shown.

### FR4 - Backlog settings page fixes
- **FR4.1**: The Projects section (project picker) shall appear right after the Connection section (which holds the sign-in method) ([desc], [Q4]).
- **FR4.2**: An empty Issue watches list shall show an issue-specific message (for example "No issue watches yet") instead of "No PR watches yet" ([desc]).
- **FR4.3**: An empty Issue watches list and an empty PR watches list shall show no "Add watch" button inside the empty state; the "Add watch" button in the section header stays ([desc], [Q5]).

Acceptance criteria:
- Given an admin connected to a space, when the settings page renders, then the section order starts Connection, Projects, then the remaining sections in their current order.
- Given no issue watches, when the Issue watches section renders, then it shows "No issue watches yet" and exactly one "Add watch" button (the header one).
- Given no PR watches, when the PR watches section renders, then it shows "No PR watches yet" and exactly one "Add watch" button.

### FR5 - Backlog issue and PRs in the task top bar (like GitHub)

Source: gate feedback on the first draft ("status bar at the top, right of the workflow steps; for GitHub this area shows the issue and the PRs, grouped into a dropdown when a task has more than one PR"). Checked against Kandev v0.96.0 `apps/web/components/task/task-top-bar.tsx`: the right side of the task top bar (after the `WorkflowStepper` in the center) has a plugin cluster that renders the `chat-top-bar` slot (`TaskTopBarPluginActions`), and a status cluster with the first-party GitHub issue button, `PRTopbarButton`, `MRTopbarButton`, `RegisteredChangeRequestStatus` (PRs of plugin review providers whose summary carries `taskStatus`; one item shows as a single button, two or more are grouped by the host's `MultiStatusMenu` dropdown) and the Jira/Linear issue buttons. The plugin registers no `chat-top-bar` component today.

- **FR5.1**: When the open task has an active Backlog issue link, the plugin shall show a Backlog issue button in the task top bar right side, by registering a component for the `chat-top-bar` slot ([feedback]).
- **FR5.2**: The button shall show the issue key; hovering or focusing it shall show the same card as FR1.4 (key, summary, status); clicking it shall open the issue in a new browser tab (same URL and host check as FR1.5).
- **FR5.3**: When the task has no active issue link, or Backlog is OFF for the workspace, or the task is archived, the slot shall render nothing.
- **FR5.4**: Backlog pull requests linked to the task shall show in the same top-bar status area through the existing review provider (`taskStatus` on each summary): one PR as a single button, two or more grouped into the host's dropdown. The plugin shall not draw its own PR dropdown; it shall make sure every linked PR summary carries `taskStatus` so the host shows it.
- **FR5.5**: On phones (`presentation: "mobile"`) the issue button shall follow the host's shared menu rules (44px touch target, 16px icon).
- **FR5.6**: The issue button shall reuse the shared per-workspace links store (NFR2); opening a task shall not add an extra `issues.links.list` call when the store is already loaded.

Acceptance criteria:
- Given an open task linked to `PROJ-12` ("Fix login", "In Progress"), when the task page shows, then the top bar right side has a `PROJ-12` button; when hovered, then the card shows `PROJ-12`, `Fix login`, `In Progress`; when clicked, then the issue opens in a new tab.
- Given a task with one linked Backlog PR, when the task page shows, then the top bar shows one PR status button; given two linked Backlog PRs, then the top bar shows one dropdown listing both.
- Given a task with no issue link, when the task page shows, then no Backlog issue button is shown.

## Non-Functional Requirements

- **NFR1 (security)**: The issue summary is Backlog content: it shall never appear in logs, error messages or error responses; existing leak tests (`internal/issues/leak_test.go`) shall pass and cover the summary. API keys and tokens stay redacted (project Mandated rule).
- **NFR2 (performance)**: The badge on task rows shall reuse the shared per-workspace links store, so Home > Tasks and the sidebar make at most one `issues.links.list` call per workspace per refresh, not one per row.
- **NFR3 (startup)**: The load-time ON/OFF check (FR3.1) shall be bounded so plugin initialization stays within the host's initialize timeout; on timeout it falls back to FR3.4.
- **NFR4 (accessibility)**: The hover card shall also open on keyboard focus, and the badge shall have an accessible label with the issue key and summary.
- **NFR5 (compatibility)**: The stored link format change shall be backward compatible (new optional field); the minimum Kandev version stays 0.96.0.

## Constraints

- Kandev v0.96.0 plugin SDK: no nav-item visibility flag, no unregister, no registration after `initialize` returns.
- Team practices: TDD, `go test -race`, 80% Go line coverage floor, Vitest for UI, Makefile targets as the CI contract, stdlib-first Go.
- Express scope Minimal test strategy: at least one requirement-driven test per FR, existing tests stay green (updated only where this change supersedes them: `ui/src/index.test.ts` entry test, `sections.test.tsx` order and empty-message assertions).

## Assumptions

- "Show issue in task card in Home > Tasks" means the task rows on the `/tasks` page (Kandev's `task-row-metadata` slot) [assumption].
- "Space (project) selection" means the Projects section (project picker) and "authentication method" means the Connection section [Q4].
- "Per-watcher Add watch button" means the button inside the empty list state [assumption, from code scan].
- A test baseline (Go and Vitest) will be recorded before Code Generation; this environment currently lacks Go, `../kandev` and `ui/node_modules`.

## Out of Scope

- Changing Kandev itself (upstream nav visibility) [Q1].
- Live (no-reload) hiding or showing of the Home > Integrations entry.
- Other settings sections, the issue list page, and the Kanban badge visuals.

## Open Questions

- None blocking. Code Generation decides the exact hover-card component (host UI kit `Tooltip` or `Popover`) and the bound for the load-time check (NFR3).

## Sources

- Initial description: [desc] (project-description.json).
- Workflow-selected scope: [scope] express.
- Answers: [Q1]-[Q5] in requirements-analysis-questions.md.
- Gate feedback: [feedback] Request Changes on the first draft (task top bar issue and PRs), checked against Kandev v0.96.0 source (`~/repo/kandev`, f099a46).
- Code knowledge base: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` and `inception/reverse-engineering/developer-scan.md`.
