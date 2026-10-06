<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T02:29:14Z — Read the Q8 free-text answer as: all issue-related stories are Must and all Git/PR stories are Should; this raised IB-10, IB-11, IB-12 to Must, but the change list only recorded the Git/PR downgrade.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T02:29:14Z — Kept US3.5 (comments and attachments) at Should although it is issue-related and Q8 says issue features are Must; reviewer flagged the inconsistency.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T02:29:14Z — Put four mob judgment calls (Git credentials, issue detail display, watch re-creation, unselecting a project) to the user instead of resolving them as lead; knowledge disputes (no PR reviewers, rate-limit header, PR status refresh) were resolved from API docs without a second round.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T02:29:14Z — Reviewer: no recovery path from the 'no longer connected' state; Must stories US1.5/US1.8/US1.9 reference optional Git/PR features; requirements.md not yet updated for seven recorded deviations.
