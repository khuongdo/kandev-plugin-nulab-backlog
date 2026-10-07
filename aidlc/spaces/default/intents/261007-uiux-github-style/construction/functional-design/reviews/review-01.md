## Review

**Verdict:** NOT-READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T03:22:34Z
**Iteration:** 1

NOT-READY

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | aidlc/spaces/default/intents/261007-uiux-github-style/construction/functional-design/rules.md > BR3.5, BR4.1; functional-spec.md > WF4 step 5, WF5 step 2.2 | The design relies on BacklogGateway capabilities that do not exist and are never listed as a change. internal/backlog/issues.go IssueQuery has no creator filter, no sort by created, no ascending order and no created-since (Values() hard-codes sort=updated, order=desc). There is no PR count call (WF4 "asks the gateway for the PR count"; only Client.PullRequests exists, with no count endpoint). PullRequestQuery has no sort, so BR4.1 "newest updated first" is unbacked by the client. | Add a gateway section to the spec naming the new calls and query fields (issue creator, created ordering, PR count), state the Backlog parameters used, and either drop or justify the PR ordering claim. | New |
| R-02 | Major | rules.md > BR3.5, BR3.8; entities.md > IssueCursor | The cursor is a timestamp plus issueId, but Backlog's created-since filter is date-granular and each run fetches one page of at most 100. The design does not say how "created at or after the cursor, skipping the cursor issue and earlier" is evaluated. If more than 100 matching issues share the cursor day, or handled issues fill the page, the run never advances past them (livelock) because only one task per run is created. | Specify the paging rule: offset/continuation past already-handled issues, client-side filtering on (created, id), and behaviour when a page contains only handled issues. Add a test case to the rule. | New |
| R-03 | Major | frontend-components.md > Backlog Page (ScopeBar, PullRequestsList toolbar) | The design says IntegrationScopeBar's own saved-preset menu "is not used" and that IntegrationListToolbar has no free-text search for PRs. In SDK v0.96.0 neither can be omitted: IntegrationScopeBar always renders SavedMenu (onDeleteSaved and onSaveCurrent are required props, and the menu lists and deletes saved presets), and IntegrationListToolbar always renders the search Input (customQuery, onCustomQueryChange, queryPlaceholder are required). This contradicts BR2.5 (delete only in Settings) and the PR no-search rule, and the page would show duplicate saved-query controls. | Redesign: either use the SavedMenu with delete wired to a confirm dialog (change BR2.5/Q3 outcome with the user), or compose the bar and toolbar from lower-level host pieces. State what the PR toolbar's search box does (hide, or use it as a title filter). | New |
| R-04 | Major | rules.md > BR3.1; functional-spec.md > WF2 step 4; entities.md > IssueWatch.workflowId | BR3.1 and FR3.6 require rejecting a workflow that does not exist, and lastError includes workflow_missing. The existing issues and git host ports only check for a non-empty workflowId (internal/issues/service.go:382, internal/git/service.go:711) and expose no workflow lookup. No design says how existence is verified (host call, SDK API) or when workflow_missing is set at run time. | Define the host-port method used to verify the workflow and step, its failure mapping, and the run-time path that sets workflow_missing; or relax the rule. | New |
| R-05 | Minor | functional-spec.md > SM1; entities.md > IssueWatch | SM1 restores the previous state (active or paused) after reconnect, but IssueWatch has no attribute that stores the state to restore. | Add a field (for example resumeState) or state that not_connected is derived at read time. | New |
| R-06 | Minor | requirements.md > FR1.1, FR3.2, FR3 acceptance vs rules.md > BR1.1, BR3.4 | Requirements list six settings sections and state that each matching issue gets one task per run. The design adds a seventh section (Saved PR queries, Q3) and caps one task per run (Q1 = X). Both are justified by Q&A, but the traceability marks FR1.1 and FR3.2 as plainly OK and the mismatch with the approved acceptance wording is not recorded. | Note the two deliberate deviations in traceability.json or the spec scope table. | New |
| R-07 | Minor | entities.md > IssueWatch.statusIds; requirements.md > Assumptions | statusIds requires min 1, while the requirements assume that no status filter means every non-closed issue. The design has no rule for that case. | Decide: keep min 1 and update the assumption, or define the empty-status behaviour. | New |
| R-08 | Minor | entities.md > IssueWatchLedgerEntry constraints; rules.md > BR3.8 | Pruning the oldest ledger entries (cap 5000) combined with a cursor reset after a filter change, and removal of a task link, can let a pruned issue get a second task. | State the accepted limit or protect created entries from pruning while the issue link is active. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| Manual cross-check of FR/NFR coverage in traceability.json | All FR1-FR6 and NFR1-NFR6 listed; NFR1/NFR5/NFR6 deferred with targets | Coverage is complete on paper; defects are in implementability (R-01 to R-04). |
| Spot-check of internal/backlog, internal/git, internal/issues and SDK v0.96.0 components | R-01, R-03 and R-04 confirmed against code | Design claims diverge from existing code and the host components. |

### Summary

Four Major findings (more than two) make this NOT-READY. The coverage and rule structure are good, but the issue watcher depends on undeclared gateway features and an unspecified cursor paging rule, the page layout assumes host components that cannot be configured that way, and workflow validation has no mechanism.
