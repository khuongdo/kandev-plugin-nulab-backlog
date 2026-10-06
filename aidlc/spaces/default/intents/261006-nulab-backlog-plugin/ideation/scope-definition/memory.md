<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T00:51:51Z — Classified items the user picked for the first release beyond the market-research must-haves as Should, not Must; Must is reserved for the end-to-end flow and Bitbucket-equivalent checks from the success criteria.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T00:51:51Z — Split 5-option questions into an A-D multi-select plus a yes/no follow-up for E, because the question picker allows at most 4 options; the file keeps the full A-E set.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T00:51:51Z — Asked two follow-up questions instead of guessing when Q1-B (update issue status from Kandev) conflicted with Q3 one-way sync, and Q5 kept webhooks despite the public-address limit; user resolved both by dropping them.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T00:51:51Z — Polling frequency for one-way status sync must fit Backlog rate limits; PR watch (IB-13) is the largest item and may slip.
