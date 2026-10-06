<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T02:40:42Z — Used the refined-mockups questions to settle two open product decisions from the user-stories review (restore after reconnect, US3.5 priority) because they shape screens; the restore rule (Q4-A) is a new behaviour not yet in stories.md.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T02:40:42Z — Documented only changed or new screens (M2m, M7, M8, M12) with full mockups and kept W1–W10 layouts by reference instead of redrawing all ten.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T02:40:42Z — Marked Kandev-rendered pieces (task menu, # suggestions, PR badge, create-PR flow) as data-only for the plugin rather than designing them, since Kandev owns their UI.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T02:40:42Z — Reviewer: auto-restore (Q4-A) needs a story/AC and edge cases (switch A to B and back, different account); M3/M5/M12 state tables and PR unlink flow still missing.
