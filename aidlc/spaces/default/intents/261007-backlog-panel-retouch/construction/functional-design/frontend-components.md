# Frontend Components — 261007-backlog-panel-retouch

All components are plugin components rendered through the Kandev host JSX and `host.ui` kit (no plugin CSS, BR6.1). Host components are used by name from `host.ui`; their props follow Kandev v0.96.0.

## Hierarchy

```
BacklogPage (/backlog, unchanged shell + scope bar)
├── IssuesPage                          (changed)
│   ├── host.ui.IntegrationListToolbar  (new use; replaces createPrToolbar here)
│   │   └── filter slot: IssueFilters   (new, inline)
│   │       ├── host.ui.IntegrationRepositoryFilter  Project
│   │       ├── host.ui.IntegrationRepositoryFilter  Status (+ "Not closed")
│   │       ├── host.ui.IntegrationRepositoryFilter  Assignee (+ "Me")
│   │       └── Save query button
│   └── host.ui.ChangeRequestList
│       └── host.ui.ChangeRequestRow (href = issue url, unchanged)
│           └── taskIndicator: host.ui.TaskRowIndicator   (new use)
└── PrList                              (changed)
    ├── PrToolbar (createPrToolbar, restyled to IntegrationListToolbar layout, no query box)
    │   └── filter slot
    │       ├── host.ui.IntegrationRepositoryFilter  Repository
    │       ├── StatusMultiFilter (new): host.ui.Popover + Checkbox ×3
    │       ├── host.ui.IntegrationRepositoryFilter  Assignee ("Anyone", "Me")
    │       ├── host.ui.IntegrationRepositoryFilter  Creator ("Anyone", "Me")
    │       └── Save query button
    └── host.ui.ChangeRequestRow
        └── taskIndicator: host.ui.TaskRowIndicator   (new use)

Kanban card (task-card-tags slot)
└── IssueBadge                           (changed: anchor when openable)
```

## Props and State

| Component | Inputs | State | Notes |
|---|---|---|---|
| IssuesPage → `IntegrationListToolbar` | `title`, `count` (total, 0 while unknown), `loading`, `lastFetchedAt` (Date or null from `refreshedAt`), `customQuery = draftQuery`, `committedQuery`, `onCustomQueryChange`, `onCommitCustomQuery`, `onRefresh`, `filter`, `queryPlaceholder` ("Search issues"), `titleTestId="backlog-issues-list"`, `queryTestId="backlog-issues-search"`, `refreshTestId="backlog-issues-refresh"` | `draftQuery`, `committedQuery` replace `search`/`keyword` and the 400 ms timer | The pagination focus target moves from the plugin heading ref to the results region (the host title takes no ref) |
| Issue filter (each) | `value` ("" = All), `onValueChange`, `options` [{value,label}], `ariaLabel`, `allLabel`, `testId` = `backlog-issues-project` / `-status` / `-assignee` | — | Status options prepend `{value:"open", label:"Not closed"}`; Assignee prepends `{value:"me", label:"Me"}` |
| `TaskRowIndicator` (issues) | `tasks` = linkedTasks → `{id: taskId, taskId, fallbackTitle: taskKey ?? taskId}`; `testIdPrefix = "backlog-issue-task-<issueKey>"` | — | Empty list → `tasks` empty, no `emptyLabel` (BR1.5) |
| `TaskRowIndicator` (PRs) | `tasks` = linkedTaskIds → `{id, taskId: id, fallbackTitle: id}`; `testIdPrefix = "backlog-pr-task-<number>"` | — | |
| PrToolbar | as today (`title`, `count`, `loading`, `lastFetchedAt`, `onRefresh`, `refreshDisabled`, `children`, `idPrefix`) | — | Classes copied from `IntegrationListToolbar`: `flex shrink-0 flex-col gap-2 border-b px-4 py-2.5 sm:px-6 md:flex-row md:flex-wrap md:items-center md:gap-3`; count hidden on phones and shown in the bottom status row |
| StatusMultiFilter | `value` (set), `onToggle(status)`, `testId="backlog-prs-status"` | popover open | Trigger: outline button "Status (n)", `aria-haspopup`; content: three `Checkbox` + `Label` with ids `backlog-prs-status-<s>`; last checked box is disabled (BR5.3) |
| IssueBadge | link from links store | — (the `open` toggle is removed) | Openable → `<a href target=_blank rel="noopener noreferrer" title=detail data-testid=backlog-issue-badge-<taskId>>`, styled like today's outline button; not openable → `<span title=detail>` same test id; both stop click propagation |

## Interaction Flows

See [functional-spec.md](functional-spec.md) W1–W8.

## Validation

- Query: trimmed; blank clears the keyword (BR4.3). No other input validation (filters are pick-lists).
- PR status: at least one status (BR5.3).

## API Integration Points

Unchanged actions: `issues.filters`, `issues.list`, `issues.refresh`, `issues.link`, `issues.links.list`, saved-query actions, PR list actions. No new actions; no request/response shape change.

## Test Harness Impact

| Fake to add in `ui/src/testing/harness.ts` | Behaviour to mirror |
|---|---|
| `TaskRowIndicator` | Renders nothing for empty `tasks`; `${prefix}-single` button showing `fallbackTitle` for one; `${prefix}-multi` trigger with count and items for 2+; click records navigation to the task id |
| `IntegrationListToolbar` | Title (`titleTestId`), count, query input (`queryTestId`) calling `onCustomQueryChange`, Enter and blur-if-dirty calling `onCommitCustomQuery`, `filter` node, refresh button (`refreshTestId`) |
| `Popover`, `PopoverTrigger`, `PopoverContent` | Open/close on trigger click; content rendered when open |

Existing tests to update (not delete): issue search now commits on Enter (was 400 ms wait); filter test ids now target the `IntegrationRepositoryFilter` fake; plain task anchors `backlog-issue-task-<key>-<id>` become `TaskRowIndicator` test ids; the badge click test now expects an anchor with `href`; the mobile "Filters (n)" toggle tests are replaced by a stacked-layout assertion.
