<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T04:50:06Z — walking-skeleton: returned `validation` on the apiKey field, not `reconnect_required`, when Backlog rejects a key during Connect; nothing is connected yet and AC1.1.2 puts the error on the key field.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T04:50:06Z — walking-skeleton: AC1.1.7 asks for the key in a header, but the user chose the documented `apiKey` query parameter (Q1) with full URL redaction; stories.md was left unchanged and the deviation is recorded in functional-spec.md.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->
- 2026-10-06T04:50:06Z — walking-skeleton: added an all-or-nothing rule (BR2.8) for writing the secret and the connection record, since Kandev state has no transactions; restoring the previous secret on a failed record write keeps replace-after-verify (Q2) safe.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
