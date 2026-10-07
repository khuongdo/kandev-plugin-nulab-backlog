# Performance Test Instructions — 261007-uiux-github-style

## Scope

Not applicable at the Minimal test strategy: this intent has no NFR Requirements or NFR Design stage and its requirements set no latency or throughput target. Load-related behaviour is bounded by design and covered by unit tests instead:

- the issue watcher reads at most 5 pages of 100 issues per run and creates at most one task per run (BR3.4, BR3.12);
- the PR list fetches one page of 20 plus one count call per request (BR4.1);
- Backlog calls go through the existing per-group rate limiter and honour `Retry-After` (NFR3).

## How to Run

Covered by the unit suites in `../code-generation/unit-test-instructions.md`; no separate performance command exists.

## Expected Coverage

No performance target; no performance evidence is required for this intent.
