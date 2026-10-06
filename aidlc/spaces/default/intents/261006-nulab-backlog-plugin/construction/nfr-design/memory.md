<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T05:29:56Z — walking-skeleton: lowered the Backlog call limit for Connect from the contract's 10 s to 8 s inside one 13 s deadline (Q1), so the store calls and rollback fit under Kandev's 15 s action limit; background calls keep 10 s.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T05:29:56Z — walking-skeleton: made the secret write the commit point and reconciled by epoch on read, instead of a cross-store transaction, because Kandev's state and secret stores have no transactions; leftover states are either invisible or reported as error, never used.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
