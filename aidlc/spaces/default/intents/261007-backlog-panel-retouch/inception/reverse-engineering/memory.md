<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
- 2026-10-07T11:28:45Z — Used the local Kandev checkout at ~/repo/kandev (pinned v0.96.0) as the reference for the GitHub integration UI, since ../kandev is not present in this worktree.
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

## Deviations
- 2026-10-07T11:33:47Z — Took the CodeKB snapshot with directory paths ending in '/' (ui/src/, internal/plugin/, internal/issues/) because codekb-publish treats a path as a directory only with a trailing slash; the first publish without slashes was refused.
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

## Tradeoffs
- 2026-10-07T11:28:45Z — Chose a focused scan (ui/src, internal/plugin, internal/issues, manifest.yaml) over a full rescan of the stale store; the change is UI-only, so a full rescan would cost much more for little gain, at the price of demoting prior deep coverage outside these paths to shallow.
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
- 2026-10-07T11:28:45Z — 'Card issue' may mean the issue list row or the Backlog badge on Kanban task cards; and issue-title click may already open the browser. Confirm in Requirements Analysis.
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
