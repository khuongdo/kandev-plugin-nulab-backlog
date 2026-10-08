## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T13:49:00Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | ui/src/issues/issues-page.tsx > `filterFields` wrapper `onPointerDownCapture` + `commitQuery` | Carried forward, code unchanged in this attempt. The pointerdown skip holds for mouse and pen, but not for touch (FR4.6): the pointer events fire before the compatibility mousedown that moves focus, so the box blur commits and reloads, and the pick reloads again. A cancelled press leaves the flag set. | Narrow the code-summary claim to mouse and pen input, and record the touch case as a known limit. Optionally clear the flag on `pointercancel`. | Unresolved |
| R-02 | Minor | ui/src/testing/harness.ts > `TaskRowIndicator` fake (`<prefix>-item-<id>` test ids) | Carried forward. The fake's per-task menu item ids are not confirmed against the host component. This attempt leans on them more: the SCM tests assert `backlog-scm-pr-task-42-item-t-1` and `-multi`/`-single`, so they prove only the fake. | Confirm the host's real ids and align the fake, or mark those assertions as fake-only. | Unresolved |
| R-03 | Minor | aidlc/spaces/default/intents/261007-backlog-panel-retouch/construction/code-generation/code-generation-plan.md and traceability.json (contract test wording) | Carried forward, unchanged. The plan and traceability still describe the packaged-host contract test as covering UI rendering. The Loop-back 2 summary itself says it proves install and run only. | Reword the plan and traceability to say the contract test covers packaging and install only. | Unresolved |
| R-04 | Minor | ui/src/issues/issues-page.tsx > `shown` state and toolbar `count`, `loading`, `lastFetchedAt` | Carried forward, unchanged. `shown` is never cleared, so after a failed filter change or a workspace switch the toolbar can show the previous total next to the error body. | Reset `shown` when `workspaceId` changes, and on a non-ready result after a filter or query change. | Unresolved |
| R-05 | Minor | ui/src/git/status-multi-filter.tsx > last-status `Checkbox` | Carried forward. Behaviour is correct (`aria-disabled`, toggle ignored). `aria-disabled` triggers no host `disabled:` styling, so sighted users may see no cue. The same component now also serves the provider list, so the gap applies there too. | Add an `aria-disabled:opacity-50` (or equivalent) class, or confirm the host styles it. | Unresolved |
| R-06 | Minor | ui/src/git/scm-pr-list.tsx > repository `dropdown` (`IntegrationRepositoryFilter`) and the load effect `if (!filters?.repo) return` | Inherited from v0.4.0 and not introduced here, but now more visible: the host filter always offers an All choice (`backlog-scm-prs-repo-option-all`; the updated test asserts it). A provider list needs one repository. Picking All sets `repo` to `""`, so no request is sent. The previous page stays on screen under an "All repositories" label, while Refresh and Save query are disabled. | Hide or ignore the All choice for the provider list (keep the current repo when `""` is picked), or clear `load` and show the empty state when `repo` is `""`. Add a test. | New |
| R-07 | Minor | README.md > "### 0.4.1: GitHub-style lists" | The provider-list status change (at least one status must stay picked; before 0.4.1 the user could clear them all) is recorded in code-summary.md, but the user-facing upgrade note does not mention it. | Add one line to the 0.4.1 note, for example that at least one status always stays selected on every pull request list. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| vitest scoped (8 files, incl. `src/git/pr-list.test.tsx`, `src/controls.test.ts`) | PASS: 8 files, 147 tests | Matches the summary (147). |
| npm --prefix ui run typecheck | PASS | No errors. |
| npm --prefix ui run lint | PASS | No ESLint output. |
| npm --prefix ui run format:check | PASS | All files match Prettier. |
| source-manifest vs `git status` | PASS | The 17 claimed paths equal the 16 modified source/doc files plus the new `status-multi-filter.tsx`. No unclaimed source change and no `coverage.out` in the repo root. Changes under `aidlc/` are records. |
| manual diff: `pr-list.tsx` conflict resolution | PASS | v0.4.0's provider selector, the `provider !== "backlog"` early return to `ScmPrList` with `providerControl`, and the provider-failure handling are intact. Only the selector's label is removed and the status control is swapped. |
| manual diff: `pr-list.test.tsx` | PASS | v0.4.0's SCM tests are updated to the new ids (repo options, task indicator, status popover, saved filter). Assertions are kept or strengthened, none deleted. One test is added. |
| manifest and README | PASS | `manifest.yaml` is 0.4.1, nothing else in it changed, and the Makefile reads the version from it. The README note matches the behaviour, apart from R-07. |

### Summary

The merge compiles with v0.4.0's provider and SCM behaviour preserved. The SCM list now matches the Backlog list (host task indicator, label-less searchable dropdowns, "Status (n)" with the last-status guard, provider selector first), and v0.4.0's tests were updated, not weakened. Gates are green and the manifest is exact. What is left is minor: R-01 to R-05 are carried unchanged, the SCM "All repositories" choice leaves a stale list (R-06, inherited), and the status-guard behaviour change is not in the README (R-07).
