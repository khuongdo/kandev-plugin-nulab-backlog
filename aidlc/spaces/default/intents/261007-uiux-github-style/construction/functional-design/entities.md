# Entities — 261007-uiux-github-style

New and changed entities only. Existing entities (Connection record, IssueLink, PR Link, PR Watch, PR LedgerEntry, IssueSettings) keep their shape unless listed under "Changed". Source of truth is the YAML block; the summary follows it.

```yaml
entities:
  - name: IssueWatch
    owner: Issues
    description: >
      A saved filter over one Backlog project that creates one Kandev task per
      matching issue, at most one task per run (FR3, Q1, Q2, F1).
    attributes:
      - { name: id, type: identifier, required: true, unique: true, constraints: "generated on first save" }
      - { name: name, type: text, required: true, min: 1, max: 100, constraints: "trimmed; counted in characters" }
      - { name: spaceHost, type: text, required: true, constraints: "space of the connection at save time" }
      - { name: projectKey, type: text, required: true, constraints: "one of the workspace's selected projects" }
      - { name: statusIds, type: list<integer>, required: true, min: 1, max: 20, constraints: "Backlog status ids of the project; duplicates removed" }
      - { name: assignee, type: enum, required: true, allowed: [anyone, me], default: anyone }
      - { name: creator, type: enum, required: true, allowed: [anyone, me], default: anyone }
      - { name: assigneeId, type: integer, required: false, constraints: "resolved Backlog user id when assignee = me" }
      - { name: createdUserId, type: integer, required: false, constraints: "resolved Backlog user id when creator = me" }
      - { name: workflowId, type: identifier, required: true, constraints: "existing Kandev workflow of the workspace" }
      - { name: workflowStepId, type: identifier, required: false, constraints: "step of workflowId" }
      - { name: intervalMinutes, type: integer, required: true, default: 5, min: 1, max: 1440 }
      - { name: state, type: enum, required: true, allowed: [active, paused, not_connected], default: active }
      - { name: stateBeforeDisconnect, type: enum, required: false, allowed: [active, paused], constraints: "set when state becomes not_connected; restored and cleared on reconnect" }
      - { name: cursor, type: IssueCursor, required: false, constraints: "absent until the first issue is passed; cleared when filters change" }
      - { name: createdCount, type: integer, required: true, default: 0, min: 0 }
      - { name: pendingCount, type: integer, required: true, default: 0, min: 0, constraints: "matching issues seen after the cursor but not yet handled in the last run" }
      - { name: lastRunAt, type: timestamp, required: false }
      - { name: lastError, type: enum, required: false, allowed: [unauthorized, rate_limited, unavailable, workflow_missing, ledger_full] }
    constraints:
      - "at most 50 issue watches per workspace"
      - "name unique per workspace, case-insensitive"
    relationships:
      - { to: IssueWatchLedgerEntry, cardinality: "1:N", direction: "IssueWatch owns ledger entries; deleting the watch deletes them" }
      - { to: IssueLink, cardinality: "1:N", direction: "tasks the watch creates are linked through the existing IssueLink" }

  - name: IssueCursor
    owner: Issues
    kind: value object
    description: >
      Position of the last passed issue in creation order (created, then issue
      id), plus where that issue sat in the list for its creation day, so a run
      resumes near the cursor instead of rescanning the day (BR3.12).
    attributes:
      - { name: created, type: timestamp, required: true }
      - { name: issueId, type: integer, required: true }
      - { name: dayOffset, type: integer, required: true, min: 0, constraints: "0-based position of the cursor issue in the list filtered by the watch and createdSince = the cursor's date, sorted by created ascending" }

  - name: IssueWatchLedgerEntry
    owner: Issues
    description: Records that a watch reserved, created or skipped the task of one issue, so it is handled at most once (FR3.4).
    attributes:
      - { name: key, type: text, required: true, constraints: "spaceHost|issueId" }
      - { name: watchId, type: identifier, required: true }
      - { name: outcome, type: enum, required: true, allowed: [reserved, created, skipped_linked] }
      - { name: taskId, type: identifier, required: false, constraints: "set when outcome = created" }
      - { name: at, type: timestamp, required: true }
    constraints:
      - "unique (watchId, key)"
      - "never pruned while the watch exists; at most 5000 entries per watch — when full the watch stops creating tasks and sets lastError ledger_full (BR3.13)"

  - name: PullRequestPage
    owner: Git
    kind: read model
    description: One page of the PR list returned by the new list action (FR4).
    attributes:
      - { name: items, type: list<PullRequestRow>, required: true, max: 20 }
      - { name: page, type: integer, required: true, min: 1 }
      - { name: pageSize, type: integer, required: true, allowed: [20] }
      - { name: total, type: integer, required: true, min: 0 }
      - { name: hasNext, type: boolean, required: true }

  - name: PullRequestRow
    owner: Git
    kind: read model
    attributes:
      - { name: number, type: integer, required: true }
      - { name: title, type: text, required: true }
      - { name: status, type: enum, required: true, allowed: [open, closed, merged] }
      - { name: author, type: text, required: true }
      - { name: assignee, type: text, required: false }
      - { name: updated, type: timestamp, required: true }
      - { name: url, type: text, required: true, constraints: "https page of the PR in the connected space" }
      - { name: linkedTask, type: TaskLink, required: false, constraints: "id and key of the linked Kandev task, if any" }

changed:
  - name: SavedPRQuery
    owner: Git
    description: Existing saved query; gains an optional creator filter so the PR list's current filters can be saved (FR2.6, Q3).
    added_attributes:
      - { name: creator, type: enum, required: false, allowed: [anyone, me], default: anyone, constraints: "absent in stored queries means anyone" }
```

## Summary

| Entity | Owner | Kind | New/changed |
|---|---|---|---|
| IssueWatch | Issues | aggregate root | new |
| IssueCursor | Issues | value object | new |
| IssueWatchLedgerEntry | Issues | entity (inside IssueWatch) | new |
| PullRequestPage / PullRequestRow | Git | read model | new |
| SavedPRQuery | Git | entity | changed: optional `creator` |

## Gateway Query Changes (BacklogGateway)

The design needs these additions to `internal/backlog`; existing callers keep today's behaviour because every new field defaults to the current value.

| Type / call | Addition | Used by |
|---|---|---|
| `IssueQuery` | `CreatedUserIDs` (`createdUserId[]`), `CreatedSince` date (`createdSince`, `yyyy-MM-dd`), `Sort` / `Order` (default `updated` / `desc`; issue watch uses `created` / `asc`) | issue watch run (BR3.5, BR3.12) |
| `PullRequestCount` | new call to `GET /api/v2/projects/{project}/git/repositories/{repo}/pullRequests/count` with the same filters as `PullRequestQuery` | PR list total (BR4.1) |
| `PullRequestQuery` | no change: Backlog returns pull requests newest first and has no sort parameter | PR list (BR4.1) |

PR Watch keeps its fixed 5-minute loop and 10-task cap (summary confirmation). IssueWatch reuses the existing IssueLink for the tasks it creates, so linked-issue badges, status sync and `task.deleted` handling apply unchanged.
