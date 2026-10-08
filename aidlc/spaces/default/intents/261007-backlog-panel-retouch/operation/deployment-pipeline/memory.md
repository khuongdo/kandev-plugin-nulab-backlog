<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
- 2026-10-07T13:24:45Z — Placed the manual real-host UI check (NFR2-UI-IN-HOST, accepted as Unverified at Build and Test) before the v0.4.0 tag, using the pull-request package on the self-hosted Kandev, so the release is not tagged on an unobserved UI.
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

## Deviations
- 2026-10-07T22:07:59Z — v0.4.0 was released on main (PR #9) while this intent ran, after the first version question proposed 0.4.0; the gate's Request Changes led to a rebase onto origin/main, a loop-back to bring the new provider PR list in line, and release 0.4.1. Re-check gh release list and origin/main again right before asking the version question and before tagging.
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
