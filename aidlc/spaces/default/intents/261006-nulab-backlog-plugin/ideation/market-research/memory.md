<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T00:23:21Z — Kept competitive analysis to the Bitbucket plugin only; Q1 chose A, so other Backlog integrations appear only as context in market-trends and build-vs-buy.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T00:23:21Z — Did not write a separate differentiation-strategy brief; the stage prose mentions one but it is not a declared output, so differentiation lives inside competitive-analysis.md.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T00:23:21Z — Listed the Nulab MCP server as a partial alternative instead of ignoring it; it is the closest existing option, but it lacks Kandev task/repository hooks, so build still wins.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T00:23:21Z — Whether PR creation and PR watches (as in the Bitbucket plugin) are in scope for the Backlog plugin; settle at Scope Definition. Exact rate-limit numbers per Backlog plan are not published.
