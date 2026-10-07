## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T04:21:00Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | internal/git/service.go > Service.RunQuery (action git.queries.run) | QueryInput now stores a Creator filter (internal/git/types.go, FR2.6) and SaveQuery persists it, but RunQuery builds its PullRequestQuery from Statuses and Assignee only. A saved query with creator=me therefore returns other authors' PRs when run through git.queries.run. The new UI does not call this action (presets go through git.prs.list, which does honour Creator), so the action is a dead but still exposed entry point. | Either apply Creator (Myself plus CreatedUserIDs) in RunQuery, or remove git.queries.run from the manifest and handlers together with its tests. | New |
| R-02 | Minor | internal/issues/watcher.go > watchRun.walk (R-09 overshoot restart) and watchRun.advance | The R-09 restart `continue` spends one of the 5 pages of the run budget. In addition, advance() undercounts DayOffset when the new cursor's createdSince differs from the run's. Together this can lead to a run that starts too early, reads 5 full pages of issues all at or before the cursor and never advances it, because skipped-before-cursor issues do not move the cursor. This needs more than 500 issues inside the one-day window and is not covered by a test. It does not lose or duplicate an issue. | Do not count the restart page against watchMaxPages, or add a test and a documented ceiling (ponytail: note) for a window of more than 500 issues before the cursor. | New |
| R-03 | Minor | internal/issues/watcher.go > runWatch, handle (delete during a run) | DeleteWatch empties the ledger and removes the watch, but an in-flight run (the single worker) still calls addEntry afterwards, which recreates the `issues.watch_ledger.<id>` document as an orphan and can still create one task for a watch already deleted. finishRun correctly ignores the missing watch. | Recheck that the watch still exists before reserving (inside the reserve step), or accept it and record it as a ponytail ceiling. | New |
| R-04 | Minor | aidlc/spaces/default/intents/261007-uiux-github-style/construction/code-generation/code-summary.md > Deviations (BR3.11) | BR3.11 says "never retry inside the same run" and "wait for Retry-After before the next run". The implementation relies on the Background class waiting Retry-After up to 3 times inside the run, then records rate_limited and the next run follows the watch interval, with no stored next-allowed time. The deviation is listed honestly and is safe (no task is created on 429, and waits are counted in the cycle log). | Human to accept the deviation explicitly, or add a nextRunAfter value for a 429. | New |
| R-05 | Minor | aidlc/spaces/default/intents/261007-uiux-github-style/construction/code-generation/traceability.json and code-summary.md > BR1.4 | BR1.4 is only partly delivered: the sign-in method is a Select, but both choices are always offered because ConnectionView has no "OAuth configured" flag. This is a verified, documented, acceptable deferral (grep shows only the oauthNotConfigured fallback message). The plan's checklist ticks BR1.4 as done, which overstates it. | Keep the traceability status Deferred and carry it to a follow-up (backend flag). | New |
| R-06 | Minor | ui/src/controls.test.ts > BR5.2 scan | The `<Button\b[^>]*>` regex stops at the first `>` (for example the one in an `onClick={() => ...}` arrow), so a Button whose className comes after such a prop is not checked for the GitHub cursor. The raw-control scan (BR5.1) is sound, and my grep found no raw select, table, details or button elements. | Match the whole opening tag, or assert BR5.2 variants per screen only. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| `go test -race -count=1 ./internal/...` | PASS (all 9 packages) | Matches the code-summary claims; no coverage.out was written to the repo root. |
| `npx vitest run` (ui) | PASS, 286 tests | Matches the claim. |
| `npx tsc --noEmit` (ui) | No errors | Strict typing is clean. |
| source-manifest.json against git status | 81 claimed paths equal the changed paths outside aidlc/ | No unclaimed or unrelated change. |
| layering grep (pluginsdk imports) | Only internal/plugin and server/main.go import it | Layering rule holds. |
| traceability.json targets | All path targets exist | One free-text note (BR1.4) is not a path; see R-05. |
| manifest.yaml actions | git.prs.list and the six issues.watches.* keys are present as authenticated/workspace/16384; issues.set_poll_interval stays admin | BR1.3 holds; handler keys match the manifest keys. |

### Summary

The implementation follows the approved plan (red outputs for steps 3, 6, 9, 12 and 15, then green and refactor notes). The issue watcher matches BR3.1-BR3.14, including the R-09 overshoot restart and the R-10 UTC-day-minus-one createdSince, the reserve-create-mark ledger, one task per run, state restore, and redacted logging (only the error code is logged). The PR list matches BR4.1 and BR4.2, and the single nav item, route and original outline icon (BR2.1, BR6.1) are in place. Only minor findings remain; the stated deviations (BR1.4 partial, 429 handling, toolbar choices) are acceptable once the human acknowledges them.
