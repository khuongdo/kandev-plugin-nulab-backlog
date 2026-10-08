# Entities — 261007-backlog-panel-retouch

This refactor adds no stored data and no backend fields. The entities below are the **view models** the UI already receives from the plugin backend and the **UI state** the redesigned lists hold. The refactor scope has no domain design, so the existing code structure (code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`) is the de-facto domain design. Source of truth is the YAML block.

```yaml
entities:
  - name: IssueRow
    description: One Backlog issue as shown on the /backlog issue list (existing `issues.list` item; unchanged).
    attributes:
      - { name: issueKey, type: string, required: true, unique: true, constraints: "Backlog key, e.g. PROJ-12" }
      - { name: title, type: string, required: true }
      - { name: url, type: url, required: true, constraints: "https://<space>/view/<issueKey>; built by the backend; unchanged (FR3.1)" }
      - { name: status, type: string, required: true }
      - { name: linkedTasks, type: "list<TaskLink>", required: true, default: "[]" }
    relationships:
      - { to: TaskLink, cardinality: "1..0..n", direction: "IssueRow -> TaskLink" }

  - name: PullRequestRow
    description: One Backlog pull request as shown on the PR list (existing item; unchanged).
    attributes:
      - { name: number, type: integer, required: true }
      - { name: title, type: string, required: true }
      - { name: url, type: url, required: true }
      - { name: linkedTaskIds, type: "list<string>", required: true, default: "[]", constraints: "Kandev task ids only; no task key is available" }
    relationships:
      - { to: TaskLink, cardinality: "1..0..n", direction: "PullRequestRow -> TaskLink (derived from linkedTaskIds)" }

  - name: TaskLink
    description: A Kandev task linked to an issue or a PR, as passed to the host linked-task component.
    attributes:
      - { name: taskId, type: string, required: true }
      - { name: taskKey, type: string, required: false, constraints: "present for issues, absent for PRs" }
      - { name: fallbackTitle, type: string, required: true, constraints: "derived: taskKey if present, else taskId" }

  - name: IssueBadgeLink
    description: The link shown as the Backlog badge on a Kanban task card (existing `issues.links.list` item).
    attributes:
      - { name: taskId, type: string, required: true }
      - { name: issueKey, type: string, required: true }
      - { name: status, type: string, required: false }
      - { name: url, type: url, required: false, constraints: "absent when the backend cannot build it" }
      - { name: state, type: enum, allowed: [connected, not_connected], required: true, constraints: "UI type LinkView.state" }
      - { name: unavailable, type: boolean, required: true, default: false, constraints: "true when the issue is deleted or access was lost" }
      - { name: statusUpdatedAt, type: timestamp, required: false }
      - { name: stale, type: boolean, required: false, default: false }

  - name: IssueListQuery
    description: UI state of the issue toolbar.
    attributes:
      - { name: draftQuery, type: string, required: true, default: "\"\"", constraints: "what the user is typing; does not trigger a reload" }
      - { name: committedQuery, type: string, required: true, default: "\"\"", constraints: "trimmed draftQuery at the last Enter or blur; sent as keyword" }
      - { name: projectKey, type: string, required: true, default: "\"\" (All)" }
      - { name: statusChoice, type: enum, allowed: [all, notClosed, "<statusId>"], required: true, default: notClosed }
      - { name: assignee, type: string, required: true, default: "me", constraints: "\"\" = All, \"me\", or a user id" }
      - { name: page, type: integer, required: true, default: 1, min: 1 }

  - name: PrListQuery
    description: UI state of the PR toolbar (no free-text query; Backlog's PR API has none).
    attributes:
      - { name: repo, type: string, required: true, constraints: "project/repo; required before the list loads" }
      - { name: statuses, type: "set<enum>", allowed: [open, closed, merged], required: true, default: "[open]" }
      - { name: assignee, type: enum, allowed: [anyone, me], required: true, default: anyone }
      - { name: creator, type: enum, allowed: [anyone, me], required: true, default: anyone }
      - { name: page, type: integer, required: true, default: 1, min: 1 }
```

Summary: three existing view models (`IssueRow`, `PullRequestRow`, `IssueBadgeLink`) are reused unchanged; `TaskLink.fallbackTitle` is derived in the UI. Two UI state models change: the issue toolbar splits the search text into a draft and a committed query, and the PR toolbar keeps its multi-status set but shows it in one dropdown.
