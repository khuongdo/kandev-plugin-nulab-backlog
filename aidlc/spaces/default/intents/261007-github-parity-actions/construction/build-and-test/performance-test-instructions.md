# Performance Test Instructions — github-parity-actions

## Scope

There is no load or latency target. NFR4 in `requirements.md` is a call-count requirement: opening `/backlog` makes no extra Backlog calls beyond the one list call for the default query, and opening the quick-action menu makes no request. Both are checked by unit tests, so no load test applies (Minimal strategy).

## How to Run

```bash
(cd ui && npx vitest run src/page/backlog-lists.test.tsx src/page/start-task.test.tsx)
```

Expected: these tests pass. They assert one `issues.list` or `git.prs.list` call per default query, and that opening the "+ Task" menu makes no `invokeAction` call.
