<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T03:23:49Z — Added Q6 (detecting deleted Kandev tasks) to the question set; domain-design left this as an assumption to settle here, and the Kandev source shows a `task.deleted` bus event plus a GetTask RPC, so the choice was concrete enough to ask.
- 2026-10-06T03:23:49Z — Read Q4 (10 s, 3 retries) and Q5 (3 s cap) as two layers; 10 s bounds each HTTP call, while the 3 s cap only limits how long a user-triggered action waits on a rate-limit window before returning `rate_limited` with a retry hint.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T03:23:49Z — Treated the in-process boundaries between units as Go interface and event contracts written as shared-schema blocks, not OpenAPI; all units run in one Go process, so only the UI-to-backend actions and the external APIs need wire formats.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
