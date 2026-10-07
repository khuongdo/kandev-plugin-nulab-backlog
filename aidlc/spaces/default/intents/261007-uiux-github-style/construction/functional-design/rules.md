# Business Rules — 261007-uiux-github-style

Source of truth is the YAML block. Group 1: settings layout and permissions. Group 2: Integrations entry and lists. Group 3: issue watch. Group 4: PR list. Group 5: controls. Group 6: icon. Group 7: errors and empty states.

```yaml
rules:
  # 1 — Settings > Integrations > Backlog
  - id: BR1.1
    statement: The settings screen shows, in order, Connection, PR watches, Issue watches, Saved PR queries, Issue sync, Git access and Projects as framed sections.
    category: policy
    applies_to: settings screen
    trigger: screen render while connected
    logic: IF the workspace is connected THEN render all seven sections in that order ELSE render only Connection.
    violation: not applicable (layout rule, checked by UI tests)
    source: [FR1.1, Q3]
  - id: BR1.2
    statement: Watch and saved-query sections are shown to every signed-in member while connected, whether or not the member view is active.
    category: authorization
    applies_to: PR watches, Issue watches, Saved PR queries sections
    trigger: screen render
    logic: IF connected THEN show these sections; the member view hides only admin-only controls (connect, disconnect, change space, projects, enable switch, issue sync interval).
    violation: not applicable
    source: [FR1.4, FR1.5, Q4]
  - id: BR1.3
    statement: Watch, issue-watch, saved-query and PR-list actions are open to any signed-in member; connection-changing actions and the issue sync interval stay admin-only.
    category: authorization
    applies_to: plugin actions
    trigger: every action call
    logic: IF the action is a connection change or issues.set_poll_interval THEN require admin ELSE require an authenticated member.
    violation: the host refuses with forbidden; the UI shows the admin-only message
    source: [FR1.4, Q4, F2]
  - id: BR1.4
    statement: The sign-in method is chosen with a dropdown offering "API key" and "Sign in with Nulab".
    category: policy
    applies_to: Connection section
    trigger: admin opens the connect form
    logic: IF OAuth is not configured THEN the dropdown offers only "API key".
    violation: not applicable
    source: [FR5.1, Q4]
  - id: BR1.5
    statement: The Settings notice that used to link to the watches page links to the PR watches section of the same screen.
    category: policy
    applies_to: restore notice in Connection section
    trigger: restore notice shown after reconnect
    logic: IF the notice is shown THEN its action scrolls to the PR watches section; no navigation to removed routes.
    violation: not applicable
    source: [FR1.6, Q5]

  # 2 — One Integrations entry
  - id: BR2.1
    statement: The plugin registers exactly one nav item in the integrations section, pointing to /backlog, and no routes under /backlog/watches or /backlog/dashboard.
    category: constraint
    applies_to: UI registration
    trigger: bundle load
    logic: register one nav item and one route; the nav item and route do not depend on the enabled state.
    violation: registration test fails
    source: [FR2.1, FR2.2, Q5]
  - id: BR2.2
    statement: /backlog shows an Issues scope and a Pull requests scope; the Issues scope is the default.
    category: policy
    applies_to: /backlog page
    trigger: page open
    logic: IF the URL has scope=prs THEN open Pull requests ELSE open Issues; switching scope keeps each scope's own filters for the session.
    violation: not applicable
    source: [FR2.3]
  - id: BR2.3
    statement: When the workspace is not connected or the integration is disabled, /backlog shows an alert with a link to the settings page and no lists.
    category: policy
    applies_to: /backlog page
    trigger: page open or state change
    logic: IF not connected OR disabled THEN show the alert with settingsHref() ELSE show the scopes.
    violation: not applicable
    source: [FR2.7]
  - id: BR2.4
    statement: Selecting a saved query as a preset applies its repository and filters to the PR list and resets the page to 1.
    category: calculation
    applies_to: Pull requests scope
    trigger: preset selected
    logic: set repository, statuses, assignee, creator from the query; page = 1.
    violation: a query whose project is no longer selected is shown disabled with a "project not selected" hint
    source: [FR2.6]
  - id: BR2.5
    statement: The current PR list filters can be saved as a new saved query from the list; editing and deleting saved queries happens only in Settings.
    category: policy
    applies_to: Pull requests scope, Saved PR queries section
    trigger: save as query; edit; delete
    logic: the list offers "Save query" only; Settings offers edit and delete per row.
    violation: not applicable
    source: [FR2.6, Q3]

  # 3 — Issue watch
  - id: BR3.1
    statement: An issue watch is valid only when its name has 1–100 characters, its project is selected, it has 1–20 status ids, assignee and creator are anyone or me, a workflow is given, and its interval is a whole number of minutes from 1 to 1440.
    category: validation
    applies_to: IssueWatch save
    trigger: issues.watches.save
    logic: >
      IF any field fails THEN reject with a field error naming that field. An
      empty status list is rejected: the watch always names the statuses it
      follows (the requirements' "no status filter" assumption does not apply).
      Workflow existence: the dialog offers only existing workflows from the
      host's workflow picker (as the PR watch dialog does); the plugin has no
      workflow-read capability, so the service checks only that a workflow id
      is given and a removed workflow is caught at run time (BR3.14).
    violation: validation error on the named field; nothing is stored
    source: [FR3.1, FR3.6, F1, F2]
  - id: BR3.2
    statement: A watch whose interval was not given gets 5 minutes.
    category: calculation
    applies_to: IssueWatch save
    trigger: issues.watches.save without intervalMinutes
    logic: intervalMinutes = 5.
    violation: not applicable
    source: [Q2]
  - id: BR3.3
    statement: An active issue watch is due when no run happened yet or its interval has passed since lastRunAt.
    category: calculation
    applies_to: issue watcher loop
    trigger: watcher tick (every minute)
    logic: IF state = active AND (lastRunAt absent OR now - lastRunAt >= intervalMinutes) THEN run the watch.
    violation: not applicable
    source: [Q2, F1]
  - id: BR3.4
    statement: One run creates at most one task.
    category: constraint
    applies_to: issue watch run
    trigger: each run
    logic: walk matching issues oldest first; stop after the first task is created.
    violation: not applicable
    source: [FR3.2, Q1]
  - id: BR3.5
    statement: A run handles matching issues in creation order after the cursor, including issues that existed before the watch was saved.
    category: policy
    applies_to: issue watch run
    trigger: each run
    logic: query issues of the project matching statusIds, assignee and creator, sorted by creation ascending, with createdSince = the cursor's date (none when there is no cursor); skip every issue whose (created, id) is not after the cursor. Paging follows BR3.12.
    violation: not applicable
    source: [FR3.3, F2]
  - id: BR3.6
    statement: An issue that already has an active task link, or that the watch already handled, never gets another task from the watch.
    category: constraint
    applies_to: issue watch run
    trigger: candidate issue found
    logic: IF an active IssueLink exists for the issue THEN record skipped_linked and advance the cursor; IF a ledger entry exists for (watch, issue) THEN advance the cursor without action.
    violation: not applicable
    source: [FR3.4]
  - id: BR3.7
    statement: The ledger entry is reserved before the task is created, then marked created with the task id; a reservation left without a task (crash) is resolved on the next run.
    category: constraint
    applies_to: issue watch run
    trigger: creating a task
    logic: reserve -> create task in workflowId/workflowStepId -> link issue -> mark created; IF task creation fails THEN remove the reservation and keep the cursor; IF a run finds a reservation without taskId THEN mark it created when the issue now has an active link, otherwise retry that issue.
    violation: the run stops with lastError set; the issue is retried next run
    source: [FR3.2, FR3.4, NFR3]
  - id: BR3.8
    statement: Changing a watch's project, statuses, assignee or creator clears its cursor; the ledger is kept.
    category: policy
    applies_to: IssueWatch edit
    trigger: issues.watches.save on an existing watch
    logic: IF any filter field changed THEN cursor = absent.
    violation: not applicable
    source: [FR3.3, FR3.4]
  - id: BR3.9
    statement: Paused watches and watches of a disconnected or disabled workspace do not run; run now is refused for them.
    category: policy
    applies_to: issue watcher loop, issues.watches.run
    trigger: tick or run now
    logic: IF state != active OR integration disabled THEN skip (loop) or reject with conflict (run now). Losing the connection stores the current state in stateBeforeDisconnect and sets not_connected; reconnecting restores stateBeforeDisconnect.
    violation: conflict error "the watch is not active"
    source: [FR3.5]
  - id: BR3.10
    statement: Deleting a watch deletes its ledger entries; tasks and links it created stay.
    category: policy
    applies_to: issues.watches.delete
    trigger: delete
    logic: remove the watch and all ledger entries with its watchId.
    violation: not applicable
    source: [FR3.5]
  - id: BR3.11
    statement: A run that hits 401 sets lastError unauthorized and stops; 429 sets rate_limited and waits for Retry-After before the next run; neither creates a task.
    category: policy
    applies_to: issue watch run
    trigger: Backlog error
    logic: map the gateway error to lastError; never retry inside the same run.
    violation: not applicable
    source: [NFR3]
  - id: BR3.12
    statement: A run reads at most 5 pages of 100 issues, starts near the cursor, and moves the cursor past every issue it passes, so linked or handled issues never block progress.
    category: constraint
    applies_to: issue watch run
    trigger: each run
    logic: >
      offset starts at max(0, cursor.dayOffset - 5) (0 without a cursor);
      read pages of 100 until a task is created, a page is short, or 5 pages
      were read. Every issue passed (linked, already in the ledger, or the one
      that got a task) becomes the new cursor, with dayOffset = its position
      in the list for its creation date. A run that reads 5 full pages without
      creating a task keeps the advanced cursor, so the next run continues
      from there.
    violation: not applicable
    source: [FR3.3, FR3.4, NFR3]
  - id: BR3.13
    statement: A watch whose ledger holds 5000 entries stops creating tasks until it is deleted and created again.
    category: constraint
    applies_to: issue watch run
    trigger: reserve with a full ledger
    logic: IF the watch has 5000 ledger entries THEN set lastError ledger_full and create nothing; entries are never pruned while the watch exists, so a handled issue can never get a second task.
    violation: not applicable
    source: [FR3.4]
  - id: BR3.14
    statement: A task creation refused because the watch's workflow or step no longer exists sets lastError workflow_missing and keeps the watch active.
    category: policy
    applies_to: issue watch run
    trigger: host task creation returns not found or invalid argument for the workflow
    logic: remove the reservation, keep the cursor, set lastError workflow_missing; the Settings row shows the error until the watch is edited with an existing workflow.
    violation: not applicable
    source: [FR3.6]

  # 4 — PR list
  - id: BR4.1
    statement: The PR list returns 20 PRs per page of one repository of a selected project, filtered by statuses, assignee and creator, in Backlog's own order (newest first).
    category: calculation
    applies_to: git.prs.list
    trigger: action call
    logic: offset = (page - 1) * 20; count = 20; total from the new PR count call (entities.md, Gateway Query Changes); hasNext = page * 20 < total.
    violation: page < 1 or repository of an unselected project -> validation error
    source: [FR4.1]
  - id: BR4.2
    statement: Each PR row carries its linked Kandev task when a PR link exists for it.
    category: calculation
    applies_to: git.prs.list
    trigger: action call
    logic: join the page's PRs with stored PR links by spaceHost|repositoryId|number.
    violation: not applicable
    source: [FR4.2]

  # 5 — Controls
  - id: BR5.1
    statement: Plugin components use host controls only; raw select, radio, checkbox, table, details and button elements are not used.
    category: constraint
    applies_to: every plugin component
    trigger: code review and tests
    logic: replace with Select, Checkbox, Table, Collapsible or Dialog, Button, Input from the host.
    violation: test fails
    source: [FR5.1, FR5.3]
  - id: BR5.2
    statement: Buttons follow the GitHub style; primary actions default variant; secondary outline; section header actions small; icon-only ghost with an accessible name; destructive actions destructive.
    category: policy
    applies_to: every Button
    trigger: render
    logic: as stated.
    violation: test fails
    source: [FR5.2, NFR4]

  # 6 — Icon
  - id: BR6.1
    statement: The plugin icon is an original outline drawing in the current text colour, not derived from the Nulab mark, and is used for the nav item, the settings card and the page header.
    category: constraint
    applies_to: PLUGIN_ICON
    trigger: render
    logic: stroke uses currentColor, fill none, size from className.
    violation: test fails; package verification still refuses Nulab asset URLs
    source: [FR6.1, FR6.2, FR6.3, Q3]

  # 7 — Errors and empty states
  - id: BR7.1
    statement: Lists and tables show an empty state with a next step when they have no rows.
    category: policy
    applies_to: Issues list, PR list, watch tables, saved queries table
    trigger: zero rows
    logic: Issues/PR list -> "No results" with clear-filters action; PR list without repository -> "Choose a repository"; repository without PRs -> "No pull requests"; watch tables -> "No watches yet" with Add watch; saved queries -> "Save a query from the Pull requests list".
    violation: not applicable
    source: [FR2.3, R-05]
  - id: BR7.2
    statement: A failed list or action shows an inline error with a retry; 401 asks to reconnect, 429 says to retry later; no Backlog response text or secret is shown.
    category: policy
    applies_to: Issues list, PR list, watch and query actions
    trigger: action error
    logic: map unauthenticated -> reconnect message with settings link; unavailable -> retry-later message; other -> generic error with retry.
    violation: not applicable
    source: [NFR2, R-05]
```

## Rules Summary

| ID | Rule (short) | Category | Source |
|---|---|---|---|
| BR1.1 | Seven settings sections in fixed order | policy | FR1.1, Q3 |
| BR1.2 | Watch/query sections visible to members | authorization | FR1.4, FR1.5 |
| BR1.3 | Watch/query/PR list open to members; connection admin-only | authorization | FR1.4, F2 |
| BR1.4 | Sign-in method dropdown | policy | FR5.1, Q4 |
| BR1.5 | Restore notice points to PR watches section | policy | FR1.6 |
| BR2.1 | One nav item, no old routes | constraint | FR2.1, FR2.2 |
| BR2.2 | Issues / Pull requests scopes | policy | FR2.3 |
| BR2.3 | Not connected or disabled -> alert | policy | FR2.7 |
| BR2.4 | Preset applies query filters | calculation | FR2.6 |
| BR2.5 | Save from list; edit/delete in Settings | policy | FR2.6, Q3 |
| BR3.1 | Issue watch validation | validation | FR3.1, FR3.6 |
| BR3.2 | Default interval 5 min | calculation | Q2 |
| BR3.3 | Due when interval passed | calculation | Q2, F1 |
| BR3.4 | One task per run | constraint | Q1 |
| BR3.5 | Oldest first after cursor, includes existing | policy | FR3.3, F2 |
| BR3.6 | No task for linked or handled issues | constraint | FR3.4 |
| BR3.7 | Reserve, create, mark | constraint | FR3.2, NFR3 |
| BR3.8 | Filter change clears cursor | policy | FR3.3 |
| BR3.9 | Paused/disconnected do not run | policy | FR3.5 |
| BR3.10 | Delete removes ledger, keeps tasks | policy | FR3.5 |
| BR3.11 | 401/429 handling | policy | NFR3 |
| BR3.12 | At most 5 pages per run, resume near cursor, cursor passes skipped issues | constraint | FR3.3, FR3.4 |
| BR3.13 | Ledger never pruned; full ledger stops the watch | constraint | FR3.4 |
| BR3.14 | Removed workflow -> workflow_missing at run time | policy | FR3.6 |
| BR4.1 | 20 per page, total, hasNext | calculation | FR4.1 |
| BR4.2 | Linked task per PR | calculation | FR4.2 |
| BR5.1 | Host controls only | constraint | FR5.1, FR5.3 |
| BR5.2 | GitHub button style | policy | FR5.2 |
| BR6.1 | Original outline icon everywhere | constraint | FR6 |
| BR7.1 | Empty states | policy | FR2.3 |
| BR7.2 | Error states, no leaks | policy | NFR2 |
