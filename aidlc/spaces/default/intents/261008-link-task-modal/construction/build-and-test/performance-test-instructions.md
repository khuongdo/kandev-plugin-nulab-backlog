# Performance Test Instructions — Link Task modal, GitHub-style

## Applicability

No performance test is generated: the strategy is Minimal and the change adds no measurable performance target. The only related requirement is NFR4, "no new polling". It is a structural property, checked by unit tests, not by load testing.

## Structural Check

- `ui/src/issues/issue-link.test.ts`: `visible` reads the shared links store synchronously and only starts `store.load` when the workspace is not loaded yet. No timers are added.
- The links store keeps its existing single `issues.links.list` call per workspace (unchanged `links-store.ts`).

```bash
cd ui && npx vitest run src/issues/issue-link.test.ts
```
