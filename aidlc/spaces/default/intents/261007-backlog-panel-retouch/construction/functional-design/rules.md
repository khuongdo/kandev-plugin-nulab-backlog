# Business Rules — 261007-backlog-panel-retouch

Source of truth is the YAML block. Group 1: linked tasks. Group 2: Kanban badge. Group 3: issue title link. Group 4: issue toolbar. Group 5: PR toolbar. Group 6: unchanged locked behaviour.

```yaml
rules:
  # 1 — Linked tasks (issue list and PR list)
  - id: BR1.1
    statement: Linked tasks on an issue row and a PR row are rendered by the host linked-task component, never by plugin-made links.
    category: policy
    applies_to: issue row, PR row (taskIndicator slot)
    trigger: row render
    logic: IF the row has one or more linked tasks THEN pass them to host.ui.TaskRowIndicator as {id, taskId, fallbackTitle}.
    violation: not applicable (checked by UI tests)
    source: [FR1.1, FR5.1]
  - id: BR1.2
    statement: The fallback title of a linked task is its task key, else its task id.
    category: calculation
    applies_to: TaskLink
    trigger: row render
    logic: IF taskKey is present THEN fallbackTitle = taskKey ELSE fallbackTitle = taskId. The host shows the task's real title when it knows the task, else the fallback.
    violation: not applicable
    source: [FR1.1, FR5.1]
  - id: BR1.3
    statement: One linked task shows as one task button; two or more collapse into one "Tasks (n)" menu listing every task.
    category: policy
    applies_to: TaskRowIndicator
    trigger: row render
    logic: IF count = 1 THEN single button ELSE IF count >= 2 THEN one menu trigger labelled with the count, whose items are all linked tasks in link order.
    violation: not applicable
    source: [FR1.2]
  - id: BR1.4
    statement: Choosing a linked task opens that task in Kandev.
    category: policy
    applies_to: TaskRowIndicator
    trigger: click on the task button or a menu item
    logic: The host component sets the active task and routes to the task page; the plugin adds no navigation of its own.
    violation: not applicable
    source: [FR1.3, FR5.1]
  - id: BR1.5
    statement: A row with no linked task shows no task indicator and no empty label.
    category: policy
    applies_to: issue row, PR row
    trigger: row render
    logic: IF linkedTasks is empty THEN pass no emptyLabel, so nothing renders in the task slot; the "+ Task" quick action stays.
    violation: not applicable
    source: [FR1.4]

  # 2 — Backlog badge on Kanban cards
  - id: BR2.1
    statement: The badge text shows the issue key and its status.
    category: policy
    applies_to: Kanban card badge
    trigger: card render with a linked issue
    logic: text = existing badgeText (key and status; unavailable and not-connected variants unchanged). The existing badge detail (last status update, may-be-out-of-date, reconnect hint) moves into the badge's tooltip (title attribute) because click now opens the issue.
    violation: not applicable
    source: [FR2.1]
  - id: BR2.2
    statement: Clicking an available badge opens the Backlog issue in a new browser tab and does not open the Kanban card.
    category: policy
    applies_to: Kanban card badge
    trigger: click or Enter on the badge
    logic: IF the issue is available THEN render the badge as an external link (href = issue url, target _blank, rel noopener noreferrer) and stop event propagation to the card.
    violation: not applicable
    source: [FR2.2, NFR3]
  - id: BR2.3
    statement: A badge whose issue cannot be opened is not a link.
    category: constraint
    applies_to: Kanban card badge
    trigger: card render
    logic: IF unavailable OR state = not_connected OR url is absent THEN render plain text (existing unavailable / not-connected label, detail in the tooltip) with no href; a click only stops propagation.
    violation: not applicable
    source: [FR2.2]

  # 3 — Issue title link
  - id: BR3.1
    statement: The issue title on the issue row stays an external link to the Backlog issue URL built by the backend.
    category: constraint
    applies_to: issue row
    trigger: row render
    logic: ChangeRequestRow href = IssueRow.url (unchanged).
    violation: regression test fails
    source: [FR3.1]

  # 4 — Issue toolbar
  - id: BR4.1
    statement: The issue list toolbar is the host IntegrationListToolbar.
    category: policy
    applies_to: /backlog issue list
    trigger: page render
    logic: Pass title, count, loading, lastFetchedAt, draft and committed query, the filter node and refresh handler to host.ui.IntegrationListToolbar.
    violation: not applicable
    source: [FR4.1, Q1]
  - id: BR4.2
    statement: Typing in the query box does not reload the list; Enter or leaving the box with a changed value commits it.
    category: policy
    applies_to: IssueListQuery
    trigger: keystroke, Enter, blur
    logic: Keystroke updates draftQuery only. Enter, or blur while draftQuery differs from committedQuery, sets committedQuery = trim(draftQuery), sets page = 1 and reloads.
    violation: not applicable
    source: [FR4.2]
  - id: BR4.3
    statement: Committing an empty query clears the keyword filter.
    category: validation
    applies_to: IssueListQuery
    trigger: commit with a blank draft
    logic: IF trim(draftQuery) = "" THEN committedQuery = "" and the list reloads without a keyword (page 1). IF the trimmed draft equals committedQuery THEN no reload.
    violation: not applicable
    source: [FR4.2]
  - id: BR4.4
    statement: Project, Status and Assignee filters are GitHub-style searchable dropdowns without visible field labels, each with an accessible name.
    category: policy
    applies_to: issue toolbar filter slot
    trigger: toolbar render
    logic: Each filter is host.ui.IntegrationRepositoryFilter with ariaLabel = the field name and allLabel = "All <field>". Status adds a "Not closed" choice; Assignee adds a "Me" choice.
    violation: not applicable
    source: [FR4.3, NFR3]
  - id: BR4.5
    statement: Changing a dropdown filter reloads the list at once from page 1.
    category: policy
    applies_to: IssueListQuery
    trigger: filter change
    logic: set the filter value, page = 1, reload.
    violation: not applicable
    source: [FR4.4]
  - id: BR4.6
    statement: Applying a saved or default query sets the dropdowns, the draft and the committed query together, then reloads once.
    category: policy
    applies_to: issue toolbar, saved queries
    trigger: saved-query selection
    logic: projectKey, statusChoice, assignee from the query; draftQuery = committedQuery = query keyword; page = 1.
    violation: not applicable
    source: [FR4.5]
  - id: BR4.7
    statement: "Save query" saves the dropdown values and the committed query (not an uncommitted draft).
    category: policy
    applies_to: save-query action
    trigger: Save query click
    logic: saved keyword = committedQuery.
    violation: not applicable
    source: [FR4.5]
  - id: BR4.8
    statement: On a phone-width viewport the toolbar stacks like GitHub with no horizontal page scroll and no "Filters (n)" toggle.
    category: constraint
    applies_to: issue and PR toolbars
    trigger: viewport below the md breakpoint
    logic: title + count row, then each filter at full width, then the query box (issues only), then count/last-updated/refresh row. All filters are visible.
    violation: not applicable
    source: [FR4.6, Q3]

  # 5 — PR toolbar
  - id: BR5.1
    statement: The PR list toolbar is the plugin's own toolbar with the same layout and classes as IntegrationListToolbar, minus the query box.
    category: policy
    applies_to: PR list
    trigger: page render
    logic: one row on desktop - title + count, filter slot, then last-updated and ghost refresh at the right end; no query input because Backlog's PR API has no keyword search.
    violation: not applicable
    source: [FR5.2, Q1]
  - id: BR5.2
    statement: Repository, Assignee and Creator are GitHub-style searchable dropdowns without visible field labels.
    category: policy
    applies_to: PR toolbar filter slot
    trigger: toolbar render
    logic: host.ui.IntegrationRepositoryFilter each; Assignee and Creator offer "Anyone" (the All choice) and "Me".
    violation: not applicable
    source: [FR5.2, NFR3]
  - id: BR5.3
    statement: The PR status filter is one compact dropdown that keeps multi-select.
    category: policy
    applies_to: PR toolbar filter slot
    trigger: toolbar render and checkbox toggle
    logic: A "Status" trigger button shows "Status (n)" where n = number picked; its popover lists Open, Closed, Merged as checkboxes; toggling one reloads from page 1. At least one status stays picked (the last checked box cannot be cleared).
    violation: not applicable
    source: [FR5.2, Q2]
  - id: BR5.4
    statement: Save query stays in the toolbar after the filters on both lists.
    category: policy
    applies_to: issue and PR toolbars
    trigger: toolbar render
    logic: The Save query button is the last item of the filter slot; disabled while the list has no filters loaded (issues) or no repository (PRs).
    violation: not applicable
    source: [FR4.5, FR5.2]

  # 6 — Locked behaviour kept
  - id: BR6.1
    statement: The plugin bundle adds no CSS and uses host components and host utility classes only.
    category: constraint
    applies_to: UI bundle
    trigger: build
    logic: no stylesheet imports; classes are Tailwind utilities already used by the host.
    violation: package verification fails
    source: [NFR1]
  - id: BR6.2
    statement: Action keys, the opt-in default, settings card/switch/nav behaviour and secret redaction are unchanged.
    category: constraint
    applies_to: whole plugin
    trigger: every release
    logic: no backend or manifest action changes in this refactor.
    violation: existing Go/UI tests fail
    source: [NFR5]
```

## Rules Summary

| ID | Rule (short) | Source |
|---|---|---|
| BR1.1 | Linked tasks use host `TaskRowIndicator` | FR1.1, FR5.1 |
| BR1.2 | Fallback title = task key, else id | FR1.1, FR5.1 |
| BR1.3 | 1 task = button; 2+ = "Tasks (n)" menu | FR1.2 |
| BR1.4 | Choosing a task opens it in Kandev | FR1.3, FR5.1 |
| BR1.5 | No linked task = nothing in the slot | FR1.4 |
| BR2.1 | Badge shows key · status; last-checked in tooltip | FR2.1 |
| BR2.2 | Badge click opens issue in new tab, not the card | FR2.2, NFR3 |
| BR2.3 | Badge that cannot be opened is not a link | FR2.2 |
| BR3.1 | Issue title link unchanged | FR3.1 |
| BR4.1 | Issue toolbar = host `IntegrationListToolbar` | FR4.1 |
| BR4.2 | Enter/blur commits query; typing does not reload | FR4.2 |
| BR4.3 | Empty commit clears keyword; same value = no reload | FR4.2 |
| BR4.4 | Issue filters = searchable dropdowns, aria labels | FR4.3, NFR3 |
| BR4.5 | Filter change reloads from page 1 | FR4.4 |
| BR4.6 | Saved query sets dropdowns + query together | FR4.5 |
| BR4.7 | Save query stores committed query | FR4.5 |
| BR4.8 | Phone: stacked, full-width, no Filters toggle | FR4.6 |
| BR5.1 | PR toolbar = lookalike without query box | FR5.2 |
| BR5.2 | PR filters = searchable dropdowns | FR5.2, NFR3 |
| BR5.3 | PR status = one multi-select dropdown | FR5.2 |
| BR5.4 | Save query last in filter slot | FR4.5, FR5.2 |
| BR6.1 | No plugin CSS | NFR1 |
| BR6.2 | Locked behaviour unchanged | NFR5 |
