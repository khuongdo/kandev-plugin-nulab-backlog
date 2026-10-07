<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

- 2026-10-07T04:21:35Z — Resolved the design's open reviewer points R-09/R-10 in the code plan (restart from day start on overshoot; createdSince = UTC date minus one day) instead of reopening Functional Design.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

- 2026-10-07T04:21:35Z — Left BR1.4 partly done: the sign-in dropdown still offers OAuth when the server has no OAuth configured, because connection.get does not report it; marked Deferred.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

- 2026-10-07T04:21:35Z — Accepted that a 429 waits for Retry-After inside the same run (gateway background class) instead of stopping the run as BR3.11 says; no task is created on 429.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
