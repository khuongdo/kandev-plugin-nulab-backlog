# Unit Test Instructions — Link Task modal, GitHub-style

## Framework and setup

- Vitest 5 with jsdom and axe-core, as already configured in `ui/` (`npm test` = `vitest run`). No new configuration.
- Prerequisites, run once before the first Red step:
  - `../kandev` must point to the Kandev checkout pinned at `v0.96.0`, for example `ln -s ~/repo/kandev ../kandev`.
  - Install the UI dependencies with `cd ui && npm ci`.
- The shared fake host is `ui/src/testing/harness.ts`. It stubs `openTaskLinkDialog`, `api.invokeAction` and `toast`.

## Run this change's tests (scoped)

```bash
cd ui && npx vitest run src/issues/issue-link.test.ts src/issues/link-task-dialog.test.tsx src/issues/task-menu.test.ts
```

Add `src/index.test.ts` to the list if Step 7 changes it.

## Expected tests (Minimal strategy plus bugfix regression)

| File | Tests |
|---|---|
| `src/issues/issue-link.test.ts` | `parseIssueReference` table (key, lowercase key, URL on the 3 allowed hosts, URL with hash/query, rejected http, rejected foreign host, rejected garbage/empty); action shape (`placement: "link"`, `singleTaskOnly`); `run` opens the host dialog with catalogue copy and test ids; `onSubmit` links with the parsed key and signal, then refreshes the store; refresh failure after success does not throw; invalid input throws without a backend call; `not_found` / `conflict` / `validation` mapping; `visible` false-and-loads / false-when-linked / true-when-unlinked |
| `src/issues/link-task-dialog.test.tsx` | Existing 6 cases kept (labels updated to Save/Saving...); new: description line and width class; Enter submits with a chosen task and not without; inline `text-xs text-destructive` alert; success toast + links-store refresh + `onLinked` + `onClose` (targeted regression for the stale badge) |

About 15-20 tests in total.

## Coverage

The 80% floor applies to Go (`make coverage`). This change touches no Go code, so the Go coverage stays as it is. The UI has no coverage floor: every new branch in `issue-link.ts` and in the dialog's submit, error and success paths gets a test.

## Mocking and test data

- Backend failures are simulated by rejecting `invokeAction` with the error shapes `readFailure` understands (`code`, `field`), as `pr-link.test.ts` does.
- The links store comes from the existing `LinksStore` factory with a fake host, or a minimal object with `get` / `load` / `refresh` spies, the same way `task-menu.test.ts` does it.
- Use fake keys only (`PROJ-123`) and the space `acme.backlog.com`. No real credentials (project Forbidden rule).

## Full-suite check (Step 12, not per-unit)

```bash
cd ui && npx vitest run && npx tsc --noEmit && npx eslint . && npx prettier --check .
```
