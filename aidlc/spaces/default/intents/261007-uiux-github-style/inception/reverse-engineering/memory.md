<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

- 2026-10-07T02:27:51Z — Ran a full scan with no reuse question because no code knowledge base existed yet for this repo; asked the developer to go deeper on the UI side since the change is a UI/UX refactor.

- 2026-10-07T02:28:37Z — Treated the request to consult github.com/kdlbs/kandev as part of this refactor, not new work; used the local v0.96.0 checkout (same version the plugin pins) as the GitHub integration UI reference, kept outside the scanned paths.

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->

- 2026-10-07T02:27:51Z — The worktree has no .claude folder, so framework files were read from the main checkout; the scan itself runs on the worktree.

- 2026-10-07T02:48:56Z — Translated the knowledge base and the developer scan from Vietnamese to English before the gate, because the project rule says every document is written in English while chat stays in Vietnamese.

- 2026-10-07T02:51:11Z — Installed the Go toolchain and linked ../kandev before the gate to record a Go test baseline, instead of leaving the scan without one (user request).

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

- 2026-10-07T02:38:39Z — Accepted a knowledge base without a Go test baseline (no Go toolchain, ../kandev missing in this worktree) rather than blocking the scan; the UI baseline (229/229 vitest) was measured on a scratch copy.

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-07T02:38:39Z — Whether 'Issue watcher' in the request means a new backend feature (auto-create tasks from issues) or just the issue sync interval setting.
- 2026-10-07T02:38:39Z — Whether redrawing the Backlog logo as an outline is allowed under Nulab's brand rules, or a separate outlined icon should be used.
