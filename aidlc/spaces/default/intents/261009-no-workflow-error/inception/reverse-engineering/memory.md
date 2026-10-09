<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-09T00:00:00Z — Focused the scan on the create-task flow and error display (ui/src/page/, ui/src/issues/, ui/src/messages/, internal/issues/, internal/plugin/) because the bug and UI retouch both live there; user chose a focused scan over a full rescan.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-09T00:00:00Z — The three watch dialogs (issue-watch-dialog, watch-form, scm-watch-form) share the same no-workflow failure but must store a real workflowId; Requirements Analysis must decide whether they are in scope.
- 2026-10-09T00:00:00Z — Passing workflowId: null to Kandev's TaskCreateDialog is only proven against a test double; needs confirmation on real Kandev v0.96.0 (contract test or manual check).
