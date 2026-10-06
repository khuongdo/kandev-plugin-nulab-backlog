<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T02:53:22Z — Modelled the Kandev host API as an external dependency reached through a port implemented by KandevAdapter, so the catalogue stays acyclic while only KandevAdapter imports the SDK.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T02:53:22Z — Marked US7.4/US7.5 as Deferred to ci-pipeline and US7.6 as N/A in traceability, since CI and release work has no code component.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T02:53:22Z — Made BacklogGateway stateless about credentials (passed per call) to avoid a Connection-Gateway cycle, at the cost of callers fetching credentials each time.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T02:53:22Z — Reviewer: Connection's use of Kandev secrets/state is not declared via the host port; ConnectionChanged durability and restore memory need an owning entity; interactive vs background queue priority in the gateway.
