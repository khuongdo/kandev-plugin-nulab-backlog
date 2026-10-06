<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T01:50:41Z — Read Q5-D (copy attachments and comments) as a read-only snapshot at task creation, not the excluded two-way comment feature, so it does not contradict scope W-7.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T01:50:41Z — Added FR4.4 (last-updated indicator) from a rough-mockups review finding rather than from a user answer; source cited as finding R-05.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T01:50:41Z — Left attachment size limits and first-run PR watch volume as open questions for functional design instead of asking now; reviewer flagged the PR watch one as Major.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T01:50:41Z — Reviewer: MoSCoW/skeleton tagging missing on FRs, FR1.1 has two exclusive criteria, FR6.2 first-run behaviour undefined.
