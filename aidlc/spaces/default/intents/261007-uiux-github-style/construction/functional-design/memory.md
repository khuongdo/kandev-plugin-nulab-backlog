<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

- 2026-10-07T03:20:32Z — Read Q1 '1 task' as at most one new task per issue watch per run, and Q2 plus F1 as a per-watch interval (default 5 minutes) set in the watch dialog.

- 2026-10-07T03:20:32Z — Treated the existing code structure as the domain design because refactor scope has no units or domain design; placed IssueWatch in the Issues component so created tasks reuse the existing issue link.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

- 2026-10-07T03:20:32Z — Did not use the saved-preset menu of the host scope bar, because it deletes presets and Q3 keeps deletion in Settings; saved queries appear as a Select in the PR toolbar instead.

- 2026-10-07T03:20:32Z — Issue watch walks issues in creation order with a cursor, so issues that enter a matching status later are not picked up; a filter change resets the cursor.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->

- 2026-10-07T03:25:58Z — Reviewer R-09/R-10 left open: resume offset may overshoot when issues before the cursor close, and createdSince timezone is unspecified.
