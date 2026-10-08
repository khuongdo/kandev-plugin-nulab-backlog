# Cross-Unit Traceability — Link Task modal, GitHub-style

## Verdict

**PASS.** Every FR and NFR from `inception/requirements-analysis/requirements.md` is covered. Coverage comes from the stage-level `construction/code-generation/traceability.json` (zero-Unit scope). No User Stories stage ran, so there are no AC IDs to check.

## Coverage

| ID | Status | Owner | Target | Target exists |
|---|---|---|---|---|
| FR1 | OK | code-generation (stage-level) | `ui/src/issues/issue-link.ts` | yes |
| FR1.1 | OK | code-generation | `ui/src/index.ts` | yes |
| FR1.2 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |
| FR1.3 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |
| FR1.4 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |
| FR1.5 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| FR2 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |
| FR2.1 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| FR2.2 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| FR2.3 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| FR3 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |
| FR3.1 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| FR3.2 | N/A | code-generation | Existing Unlink item (`ui/src/issues/task-menu.ts`) unchanged; `task-menu.test.ts` green | yes |
| FR4 | OK | code-generation | `ui/src/issues/link-task-dialog.tsx` | yes |
| FR4.1 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| FR4.2 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| FR4.3 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| FR4.4 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| FR4.5 | OK | code-generation | `ui/src/issues/issues-page.tsx` | yes |
| FR4.6 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| NFR1 | OK | code-generation | `ui/src/issues/issue-link.test.ts` | yes |
| NFR2 | OK | code-generation | `ui/src/issues/link-task-dialog.test.tsx` | yes |
| NFR3 | OK | code-generation | `ui/src/messages/en.ts` | yes |
| NFR4 | OK | code-generation | `ui/src/issues/issue-link.ts` | yes |

## Uncovered Elements

None. FR3.2 is `N/A` by design: it keeps the existing two-step unlink-then-link flow, which needs no new code.
