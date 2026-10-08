<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-08T04:00:00Z — Read "restyle the issue-side dialog" (Q1=D) as keeping its task search and list and changing only the shell and feedback (description, width, inline error, Save/Saving..., Enter submits, toast, links-store refresh); a single URL field cannot pick a task from the issue side.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-08T04:10:00Z — Review findings to settle in Code Generation: Link item hidden on first menu open before the links store loads (R-01), lowercase key handling (R-02), issue-specific error messages (R-03), refresh inside onSubmit (R-04), singleTaskOnly (R-05), tests using the old Link label (R-06).
