<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-07T05:56:14Z — read 'opt-out by default' as the integration switch defaulting to ON when no switch record exists (Store.LoadSwitch); one root cause behind RequireEnabled

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-07T05:56:14Z — focused scan into a STALE store: prior deep coverage outside the switch area demoted to shallow (compare NARROWER), accepted to keep the bugfix scan short

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-07T05:56:14Z — upgrade path: v0.1.x workspaces already connected but with no switch record would turn OFF after the default flips; needs a product decision in Requirements Analysis
