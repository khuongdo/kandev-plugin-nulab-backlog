# Unit Test Instructions — 261007-backlog-panel-retouch

Zero-Unit refactor; UI-only change. Methodology: TDD (Testing Contract in `code-generation-plan.md`). Strategy: Minimal — one verifiable test per requirement/rule at the narrowest level, a happy-path test per changed component; the existing suite stays green.

## Framework and Configuration

- Vitest 5 with jsdom, existing config in `ui/` (no config change). TypeScript strict, ESLint, Prettier as configured in `ui/`.
- Host components are faked in `ui/src/testing/harness.ts`; tests render plugin components through that harness.
- Go suite is unaffected; it runs unchanged as a regression gate only.

## Exact Commands (scoped to this change)

Run from the repository root. The runner-readiness check (plan Step 2) uses the same command before the first Red.

```bash
npm --prefix ui exec -- vitest run --root ui \
  src/issues/issues-state.test.ts \
  src/issues/issues-page.test.tsx \
  src/issues/issue-badge.test.tsx \
  src/issues/links-store.test.ts \
  src/page/backlog-lists.test.tsx \
  src/page/backlog-page.test.tsx \
  src/git/pr-list.test.tsx \
  src/controls.test.ts
```

Single slice while iterating, for example:

```bash
npm --prefix ui exec -- vitest run --root ui src/issues/issue-badge.test.tsx
```

Static checks for the changed UI files:

```bash
npm --prefix ui exec -- tsc --noEmit -p .
npm --prefix ui exec -- eslint src/issues src/git src/page src/testing
npm --prefix ui exec -- prettier --check src/issues src/git src/page src/testing src/messages
```

## Tests per Requirement (Minimal)

| Test file | Covers |
|---|---|
| `src/issues/issues-state.test.ts` | BR1.2 task link mapping (issues, PRs); BR2.2/BR2.3 badge openability incl. https check (R-07) |
| `src/issues/issues-page.test.tsx` | FR1 / BR1.1–BR1.5 task indicator; FR3 / BR3.1 title link (existing test kept); FR4 / BR4.1–BR4.7 toolbar, query commit, filters, saved queries; BR4.8 no Filters toggle; R-02, R-03, R-04, R-06 |
| `src/issues/issue-badge.test.tsx` | FR2 / BR2.1–BR2.3 anchor vs focusable text, propagation (R-08), accessible detail (R-01) |
| `src/page/backlog-lists.test.tsx` | FR5 / BR5.1–BR5.4 PR task indicator, lookalike toolbar without query box, label-less dropdowns, Status (n) multi-select popover (R-05, R-06) |
| `src/git/pr-list.test.tsx` (from v0.4.0) | Loop-back 2: SCM pull-request list task indicator, lookalike toolbar, label-less dropdowns, Status (n), provider selector |
| `src/page/backlog-page.test.tsx`, `src/issues/links-store.test.ts` | Regression only (test ids that change are updated) |
| `src/controls.test.ts` | BR6.1 / NFR1 no raw controls, no plugin CSS |

## Mocking and Stubbing

- Harness fakes mirror Kandev v0.96.0 props and test ids exactly: `TaskRowIndicator` (`${testIdPrefix}-single`, `-multi`, nothing when empty), `IntegrationListToolbar` (`titleTestId`, `queryTestId`, `refreshTestId`; Enter and blur-if-dirty call `onCommitCustomQuery`), `IntegrationRepositoryFilter` (value "" = All; passes `triggerClassName`/`className` through as attributes for assertions), `Popover*`.
- Host actions (`issues.list`, `issues.filters`, PR list actions) use the existing harness action stub; assert call counts to prove "no reload while typing" and "one reload per commit".
- No timers are needed after the 400 ms debounce is removed; do not use real `setTimeout` waits.

## Test Data

Reuse the existing fixtures in the test files (issue pages, link views, PR pages). Add: an issue with 0, 1 and 3 linked tasks; a PR with linked ids only; link views that are available, unavailable, not connected, without url, and with a non-https url. No real credentials (project rule).

## Coverage Targets

- UI: every new helper and branch exercised by the tests above; full Vitest suite green (baseline 322 passing, count may change only by added/replaced tests).
- Go: unchanged; the 80% floor for `./internal/...` and `./server/...` still holds (`make coverage`), with `coverage.out` removed from the repo root afterwards.
