<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T00:41:03Z — Treated the Q5 answer 'typescript' as: familiar with TypeScript only, not Go; recorded Go-only plugin backend as the main skill risk.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T00:41:03Z — Marked AWS landscape assessment as not applicable; the plugin runs inside the user's self-hosted Kandev, so no cloud account or service is involved.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T00:41:03Z — Recommended polling as the default over Backlog webhooks; webhooks need a publicly reachable Kandev host, which self-hosted internal installs usually lack.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T00:41:03Z — Whether Backlog Free plan includes pull requests, and whether Backlog OAuth accepts a localhost callback; both to confirm when the test space and OAuth app are created. Plugin license still undecided.
