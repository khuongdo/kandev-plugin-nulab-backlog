<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->
- 2026-10-06T06:52:14Z — walking-skeleton: in Kandev 0.96.0 registerIntegrationSettings only adds a card on Settings > Integrations; a plugin needs registerNavItem with section integrations to appear in the Integrations menu, and must supply its own per-workspace on/off control via the action field (host.ui.IntegrationEnabledControl + host.setIntegrationEnabled).

## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-06T06:26:50Z — walking-skeleton: renamed the action connection.connectApiKey to connection.connect_api_key because Kandev only accepts keys matching ^[a-z0-9][a-z0-9._-]*$; contract C5 still lists camelCase keys, and U2-U4 action names must follow the same rule.
- 2026-10-06T06:26:50Z — walking-skeleton: pinned ../kandev to the v0.96.0 tag (user choice); Kandev's plugin-package-verify does not exist at that tag, so verify-package uses an in-repo verifier instead.

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-06T06:26:50Z — contract-summary C5 action keys need a snake_case update before U2 functional design starts.
