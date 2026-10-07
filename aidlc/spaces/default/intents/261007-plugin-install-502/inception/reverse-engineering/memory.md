<!-- INVARIANT: examples are single-line HTML comments so a fresh template parses to total=0 (MEMORY_EMPTY). Do NOT un-comment or split across lines. t100 guards this. -->
> This file is kept up to date automatically while the stage runs. Add observations at the review step, not by editing here directly.

## Interpretations
<!-- example: 2026-05-29T10:14:32Z — chose REST over GraphQL; the consuming team only needs CRUD, revisit if subscriptions land -->

- 2026-10-07T22:50:00Z — read the 47-byte 400 response after exactly 30 s as `missing multipart field "package"`: the Kandev 0.97.0 server ReadTimeout (30 s default) cut the upload body, and `tailscale serve` turned the dropped connection into the 502 the UI shows
## Deviations
<!-- example: 2026-05-29T10:14:32Z — skipped the optional caching layer the stage prose suggested; the dataset is small enough that it adds risk -->
- 2026-10-07T22:50:00Z — the conductor read the Kandev server logs and reproduced the 502 (throttled 29.5 MB junk upload through the tailscale URL -> 502 after 33.5 s) while the developer scan ran, because the user asked to read local Kandev logs

## Tradeoffs
<!-- example: 2026-05-29T10:14:32Z — picked TDD over BDD this run; the team is unit-first and the domain is well-understood -->

## Open questions
<!-- example: 2026-05-29T10:14:32Z — confirm the retention window with compliance before the next stage hardens the schema -->
- 2026-10-07T22:50:00Z — the fix lives partly outside this repo (Kandev readTimeout); decide in Requirements Analysis between shrinking the package, install-by-URL guidance, and raising KANDEV_SERVER_READTIMEOUT
