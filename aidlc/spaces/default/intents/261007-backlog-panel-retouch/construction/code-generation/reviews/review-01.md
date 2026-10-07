## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T12:41:46Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | ui/src/issues/issues-page.test.tsx > "commits a typed query together with a dropdown change in one reload (R-04)"; ui/src/testing/harness.ts > fake IntegrationListToolbar / IntegrationRepositoryFilter | The test passes only because the fake filter option is a plain button whose click does not blur the query input. In real Kandev v0.96.0 the host toolbar commits on blur when dirty (integration-list-toolbar.tsx onBlur), and a mouse click on a dropdown blurs the input first. That makes one reload for the blur commit and a second for the filter change. code-summary.md admits this in Key Decisions (R-04), but the test title still claims "one reload", so the lenient fake hides the real two-reload path. | Rename or re-scope the test to say it covers the programmatic path (change without blur). Add a second test that fires blur and then the change and asserts the final keyword and statusIds from the newest `issues.list` call (the latest-request guard). Optionally have the fake option's mousedown blur the focused input. | New |
| R-02 | Minor | ui/src/testing/harness.ts > fake TaskRowIndicator; /home/k_do_webfrontier/repo/kandev/apps/web/components/integrations/task-row-indicator.tsx | The fake multi-task menu is always open and renders items with an invented test id `<prefix>-item-<id>`. The host uses a closed Radix DropdownMenu whose items have no test id and show the task title. Navigation in the fake is `history.pushState`, while the host calls `setActiveTask` and `router.push`. The tests therefore pin ids and behaviour that the host does not have. Single and multi trigger ids (`-single`, `-multi`) do match the host. | Note in the fake's doc comment that the `-item-` ids are fake-only. Keep real-host navigation and menu behaviour in the manual check list for Build and Test (NFR2). | New |
| R-03 | Minor | aidlc/spaces/default/intents/261007-backlog-panel-retouch/construction/code-generation/code-generation-plan.md > Step 22; code-summary.md > Environment Notes; traceability.json > NFR2 -> Makefile | Step 22 is ticked as proving "host components render inside the plugin route (NFR2)". code-summary.md says the contract test proves install and run only, not UI rendering. The host components (TaskRowIndicator uses `useAppStore` and `useRouter`; IntegrationListToolbar uses react-i18next) were not rendered in a real Kandev page. traceability.json maps NFR2 to `Makefile`, which overstates that coverage. | Record the browser render check of the toolbar, filters, task indicator, anchor badge and Status popover as an explicit open item for Build and Test. Do not present NFR2 as covered by the contract test alone. | New |
| R-04 | Minor | ui/src/issues/issues-page.tsx > IntegrationListToolbar props (`loading={refreshing \|\| load.kind === "loading"}`) | The host toolbar shows "…" for the count and hides last-updated while `loading` is true. The page now passes `loading` for every page, filter and query load, not only refresh. The count therefore flickers to "…" on each page change, and the refresh button is disabled for the whole initial load. The fake toolbar only exposes count and loading as data attributes, so no test sees this. | Either accept it and say so, or pass only `refreshing` and keep the previous total while a page loads. | New |
| R-05 | Minor | ui/src/git/status-multi-filter.tsx > Checkbox `disabled={checked && value.length === 1}`; functional-design review-01 R-05 | A disabled checkbox drops out of the keyboard focus order. Review-01 R-05 preferred a non-disabled checkbox with `aria-disabled` that ignores the toggle. The plan chose disabled and the code follows the plan, so this is consistent with the approved plan. It remains a small accessibility compromise. | Accept as a recorded decision, or switch to `aria-disabled` with an ignored toggle. The test at backlog-lists.test.tsx line 219 would change with it. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| vitest (7 scoped files, `--root ui`) | PASS: 7 files, 129 tests | Matches the developer's scoped count in code-summary.md. |
| `npm --prefix ui run typecheck` | PASS | tsc --noEmit is clean, so removing `taskHref`, `filterAll`, `filtersButton`, `refreshing` and `taskLink` left no dangling references. |
| `npm --prefix ui run lint` | PASS | ESLint is clean. |
| `npm --prefix ui run format:check` | PASS | Prettier is clean. |
| repo-root `coverage.out` check | PASS: absent | The project rule on stray coverage profiles holds. |
| `git status` vs source-manifest.json | PASS | The 13 ui/src paths changed or created match the manifest exactly (12 modified plus the new status-multi-filter.tsx). No unclaimed source change and no unrelated claimed change. Other changes are AI-DLC records only. |
| traceability.json targets | PASS | Every target path exists (ui/src files, Makefile, manifest.yaml, ui/package.json). All FR1..FR5.2, NFR1..NFR5 and BR1.1..BR6.2 listed in requirements.md and rules.md are covered with status OK. |

### Summary

The implementation is complete against FR1-FR5 and BR1.1-BR6.2. Review-01 R-01..R-08 are all addressed in code and tests (R-03 as a documented limitation, R-04 only partly, see R-01 above). I checked these against host v0.96.0:

- Host `Button` supports `asChild` (it uses the Radix Slot).
- `IntegrationRepositoryFilter` maps "" to All, and `triggerClassName` and `className` exist.
- `TaskRowIndicator` and `IntegrationListToolbar` props match the plugin's usage.
- The anchor badge keeps an `aria-label` with key and status, plus `aria-describedby` and `title` for the detail.
- Click and pointer-down propagation are both stopped.
- `badgeHref` allows only https on the Backlog domains.

No raw controls, plugin CSS or secrets were found. TDD Red evidence is recorded for the main behaviour slices, and the Step 17 green-first case is disclosed honestly. The findings are all Minor. They concern fakes that are more lenient than the host and a few claims in the records that go beyond what was verified. They do not block READY.
