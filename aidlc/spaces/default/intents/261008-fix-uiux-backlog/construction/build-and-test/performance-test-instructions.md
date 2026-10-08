# Performance Test Instructions - Fix UIUX

## Scope at Minimal Strategy

No performance suite is generated: the strategy is Minimal and the requirements have no latency or throughput targets. Two bounded behaviours are checked by unit tests:

- NFR2: Home > Tasks rows, the sidebar, the Kanban card and the task top bar share one links store, so at most one `issues.links.list` call per workspace per refresh.
- NFR3: the load-time ON/OFF check is bounded at 3 s (`ENTRY_CHECK_TIMEOUT_MS`), under Kandev's 10 s initialize limit; on timeout the entry is shown.

## How to Check

```bash
(cd ui && npx vitest run src/index.test.ts)
```

## Expected Result

The shared-store test and the timeout test pass.
