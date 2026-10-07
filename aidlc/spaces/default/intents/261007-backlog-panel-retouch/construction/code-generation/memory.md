<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

## Deviations
- 2026-10-07T12:40:21Z — Kept the non-openable Kanban badge as the existing focusable Button instead of the planned span role=note; same keyboard/touch reach with a native control and a smaller change.
- 2026-10-07T12:40:21Z — Corrected the scoped UI test command after approval to `npm --prefix ui exec -- vitest run --root ui ...`; without --root ui Vitest ignores the UI config and 54 tests fail even on unchanged code.
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
- 2026-10-07T12:40:21Z — ~/go-sdk/bin/go switches to a Go 1.26.0 toolchain without the covdata tool, so make coverage fails locally; the developer used a writable toolchain copy with GOTOOLCHAIN=local. Fix the local toolchain before Build and Test.
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
