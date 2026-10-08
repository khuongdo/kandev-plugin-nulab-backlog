# Code Summary — 261007-backlog-panel-retouch

Zero-Unit refactor, UI only. All 34 plan steps done (Steps 24-28 are Loop-back 1 and Steps 29-34 are Loop-back 2, see the last two sections; Loop-back 1 supersedes the R-04 and R-05 decisions below). Methodology: TDD (Testing Contract `sha256:c832d7302104c2278f4539d9d4082d5cff074338c2d095de61b5f5091ea1e44a`). No Go or action change (BR6.2). The only manifest change is the release version `0.4.1` (Loop-back 2, Step 33).

## Files Created and Modified

| File | Change |
|---|---|
| `ui/src/issues/issues-state.ts` | Added `TaskRowLink`, `taskRowLinks`, `prTaskRowLinks` (BR1.2) and `badgeHref` (BR2.2, BR2.3, R-07). Removed `taskHref`, which is no longer used. |
| `ui/src/issues/issues-page.tsx` | The host `IntegrationListToolbar` replaces `createPrToolbar`, the `Input` and the 400 ms timer. `draftQuery`/`committedQuery` commit on Enter or blur. Project/Status/Assignee are `IntegrationRepositoryFilter`s with GitHub classes in a `FILTERS` wrapper, with Save query last. Rows use `TaskRowIndicator`. Removed the `isMobile` "Filters (n)" toggle, the field `Label`s, `Select`, `SEARCH_WAIT` and `ALL`. Pagination now focuses the results region. |
| `ui/src/git/pr-toolbar.tsx` | Restyled to the `IntegrationListToolbar` layout without a query box. The count sits beside the title on desktop and in the bottom status row on phones. Dropped the unused `idPrefix`, `refreshLabel` and `headingRef` props. |
| `ui/src/git/pr-list.tsx` | Rows use `TaskRowIndicator` (`prTaskRowLinks`). Repository/Assignee/Creator are label-less `IntegrationRepositoryFilter`s (Anyone = `""` ↔ `anyone`, R-06). Status uses `StatusMultiFilter`. Filters sit in the `FILTERS` wrapper. |
| `ui/src/git/status-multi-filter.tsx` (new) | "Status (n)" outline trigger (`aria-haspopup="dialog"`) and a host `Popover` with three `Checkbox` + `Label`. The last checked box is `disabled` (BR5.3, R-05). |
| `ui/src/issues/issue-badge.tsx` | When openable: host `Button asChild` around `<a href target=_blank rel="noopener noreferrer">`, with `title` set to the detail, `aria-describedby` pointing to an sr-only detail, and click/pointer-down propagation stopped (R-01, R-08). When not openable: the existing focusable Button with the focus/tap detail, now also stopping pointer-down. |
| `ui/src/layout.ts` | `FILTERS`, `FILTER_TRIGGER` (GitHub's `w-full … md:w-[220px]`), `FILTER_POPOVER` (`md:min-w-[360px]`). |
| `ui/src/messages/en.ts` | Added `filterAllProjects`, `filterAllStatuses`, `filterAllAssignees`, `prsStatusCount`. Removed the unused `filterAll`, `filtersButton`, `refreshing`, `taskLink`. |
| `ui/src/testing/harness.ts` | Fakes for `TaskRowIndicator`, `IntegrationListToolbar` and `Popover`/`PopoverTrigger`/`PopoverContent`. The `IntegrationRepositoryFilter` fake now has an "All" option (value `""`) and passes `triggerClassName`/`className` through. `Button` supports `asChild`. |
| Tests | `issues-state.test.ts`, `issues-page.test.tsx`, `issue-badge.test.tsx`, `backlog-lists.test.tsx` were updated or extended. No test was deleted without a replacement; the old 400 ms and "Filters (n)" tests were replaced by Enter/blur and stacked-layout tests. |

## Key Decisions

- **R-03 (query accessible name):** the host `IntegrationListToolbar` renders its `<Input>` with only `placeholder` and forwards no aria-label. I did not wrap the toolbar, because a wrapper cannot name an input it does not own. The placeholder "Search issues" is the accessible name (HTML-AAM fallback), and axe accepts it. A test asserts the placeholder, and the page's axe test passes.
- **R-04 (one combined reload):** `setFilter` also commits the trimmed draft. A dropdown change with an uncommitted draft therefore makes exactly one `issues.list` call carrying both values. A real mouse click on another control blurs the box first, so the host commits the draft at blur. That is a separate, deliberate user action, and the latest-request guard keeps only the newest result.
- **R-05:** I followed the approved plan: the last picked status checkbox is `disabled`, the same behaviour everywhere.
- **R-07:** `badgeHref` requires `https:` **and** a hostname under `backlog.com`, `backlog.jp` or `backlogtool.com`. This follows the project rule "ALWAYS accept only https space addresses under …". It goes one step further than the plan's https-only check, at no extra cost, and a test covers `http:`, `javascript:` and a foreign host.
- **Saved multi-status queries:** a saved issue query can pick several statuses that are not "Not closed". It now shows as an extra "n statuses" choice instead of looking like "All statuses".
- **`lastFetchedAt`:** an invalid `refreshedAt` becomes `null`, so the host's `toISOString()` cannot throw.
- **PR toolbar refresh:** one refresh button, kept last in the bottom/right status row. This differs from the host, which renders two refresh buttons with the same test id. The layout test checks that refresh is last.

## Test Coverage Summary

Red outputs (each recorded before the matching Green):

| Step | Failing command output |
|---|---|
| 3 | `vitest run src/issues/issues-state.test.ts` → `TypeError: taskRowLinks is not a function`, `TypeError: badgeHref is not a function` — Tests 2 failed / 7 passed |
| 7 | `vitest run src/issues/issues-page.test.tsx` → `shows linked tasks with the host task indicator…`: `Cannot read properties of null (reading 'getAttribute')`; `+ Task menu…`: `reading 'textContent'` — 2 failed / 9 passed |
| 9 | `lists 20 rows…`: `expected <h2 tabindex="-1">… to be <div …>` (focus target); `uses the host list toolbar…`: `reading 'getAttribute'` — 2 failed / 9 passed |
| 11 | `filters with label-less searchable host dropdowns…`: `expected 'SelectTrigger' to be 'IntegrationRepositoryFilter'`; `commits a typed query together with a dropdown change… (R-04)`: `reading 'click'`; backlog-lists `opens Issues on the starred issue query`: `expected null to be 'PROJ'` — 3 failed |
| 13 | backlog-lists `opens on the first repository…`: `reading 'getAttribute'` (no `backlog-pr-task-45-single`); `filters with label-less dropdowns and a Status (n)…`: `expected <label …> to be null` — 2 failed / 18 passed |
| 15 | `vitest run src/issues/issue-badge.test.tsx` → `opens the Backlog issue in a new tab…`: `expected 'BUTTON' to be 'A'` — 1 failed / 6 passed |
| 17 | After Step 12 removed the toggle, the old `keeps the rows and adds a Filters (n) drawer on mobile (M2m)` failed: `Cannot read properties of null (reading 'textContent')`. Its replacement test (all filters full width, no toggle, stacked slot) passed at once, because Step 12 had already built the layout. The `controls.test.ts` guard held. It did catch one regression in Step 16 (`<Button asChild … className={badgeClass}>` without the cursor class), which I fixed in the code, not in the test. |

Final results:

- Scoped command: 7 files, **129 tests passed**.
- Full Vitest: 31 files, **329 tests passed** (baseline 322; +7 net).
- `tsc --noEmit`, ESLint and Prettier `--check` are clean (`npm run typecheck && npm run lint && npm run format:check && npm test`).
- Go: `go test -race ./internal/... ./server/...` all `ok`, with no Go change. `make coverage` reports **92.8%** (floor 80%, exclusion `server/main.go` only). The profile is under `build/`, and the repository root has no `coverage.out`.
- `make package` built `dist/nulab-backlog-0.3.0.tar.gz`. `make verify-package` returned `verifypkg: OK`.
- `make contract-test KANDEV_MIN_DIR=../kandev` on v0.96.0 returned **10/10** `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Deviations from Plan

- **Badge, not-openable state:** the plan (Step 15) says `<span tabIndex=0 role="note">`. I kept the existing focusable host `Button` with the focus/tap `role=status` detail (the review's R-01 option). It is a native control, it meets the same need (keyboard and touch can reach the detail and the reconnect hint), and the change is smaller. It also stops pointer-down now.
- **StatusMultiFilter** lives in its own file, `ui/src/git/status-multi-filter.tsx`, instead of inside `pr-list.tsx`. This keeps `pr-list.tsx` smaller; the behaviour is as planned.
- **Unit-test command:** `npm --prefix ui exec -- vitest run src/…` as written in `unit-test-instructions.md` runs from the repository root, without `ui/vitest.config.ts` (no jsdom). It fails 54 tests with `document is not defined` even on the untouched baseline. The working form is `npm --prefix ui exec -- vitest run --root ui src/…`, or `cd ui && npx vitest run …`. Build and Test should use the `--root ui` form.

## Environment Notes

- I linked `../kandev` (outside the repository: `/home/k_do_webfrontier/.kandev/tasks/refactor-backlog-pan_mx1nvmhx/kandev`) to `/home/k_do_webfrontier/repo/kandev`, which is at tag `v0.96.0` (`f099a46`, equal to `.kandev-sdk-ref`).
- `~/go-sdk/bin/go` switches to the `go1.26.0` toolchain in the module cache, which lacks the `covdata` tool. `make coverage` therefore failed with `go: no such tool "covdata"`. This is an environment problem, not a code problem. I used a writable copy of that toolchain with `covdata` built from its own sources (scratchpad, `GOTOOLCHAIN=local`) for `make coverage`, `make package` and the contract test. Nothing in the repository changed for this.
- The contract test builds Kandev's `bin/kandev` and `bin/agentctl` inside `/home/k_do_webfrontier/repo/kandev/apps/backend` (build output only; no tracked file changed). It proves install and run on v0.96.0. It does not render the plugin UI in a browser. The visual check of the host toolbar, filters, task indicator and badge in a real Kandev page is still for the manual check (NFR2).

## Loop-back 1

Fixes for code review review-01 R-01, R-04 and R-05 (R-02 and R-03 are out of scope). TDD: Red, then Green, then refactor while green.

### Files changed

| File | Change |
|---|---|
| `ui/src/testing/harness.ts` | The `IntegrationRepositoryFilter` fake has a focusable trigger (`<testId>-trigger`, fake-only id; the host puts `testId` on its trigger `Button`). New helper `openFilter(c, testId)` dispatches `pointerdown` on the trigger and then focuses it, so a focused query box blurs (and commits when dirty) before an option is picked. |
| `ui/src/issues/issues-page.tsx` | R-01: the three dropdowns sit in a `display: contents` wrapper whose `onPointerDownCapture` marks "picking a filter" until the next document `pointerup`. A blur commit during that press is skipped, and the pick's `setFilter` takes the trimmed draft along, so typing and then picking makes one `issues.list` call. Enter and every other blur still commit at once. Save query is outside the wrapper, so its click still commits the draft first (BR4.7). R-04: a new `shown` state keeps the last ready page; the toolbar gets `count={shown?.total ?? 0}`, `lastFetchedAt` from `shown`, and `loading` only while refreshing or during the first load. |
| `ui/src/git/status-multi-filter.tsx` | R-05: the last checked status uses `aria-disabled="true"` instead of `disabled`, stays in the focus order, and its toggle is ignored. |
| `ui/src/issues/issues-page.test.tsx` | The R-04 test (functional design) is rewritten to the host order and renamed. It now also checks that after the press is released, leaving the box commits again. New test: a page change keeps the count and Refresh, and only the first load shows loading. The skeleton test checks `data-loading="true"` on the first load. |
| `ui/src/page/backlog-lists.test.tsx` | The last status checkbox is not `disabled`, has `aria-disabled="true"`, can take focus, stays checked when clicked and does not reload. |

### Host check (Kandev v0.96.0)

- `integration-list-toolbar.tsx`: the `Input` `onBlur` calls `onCommitCustomQuery()` when the draft differs from the committed query. The callback gets no event, so the plugin cannot tell Enter from blur from the call alone.
- `integration-repository-filter.tsx` uses `components/combobox.tsx`: a `PopoverTrigger asChild` around a `Button`. A mouse press on it moves focus off the query box (blur) before the popover opens. The option is picked later, in a separate interaction, after the popover opens and focuses its search input.
- So the order assumed in Step 24 (blur first, pick later) matches the host. One difference: blur and pick are separate user interactions, not one render cycle, so the microtask coalescing suggested in Step 25 could not merge them. The fix keeps the draft for the pick instead, based on the pointer press on a dropdown.

### Key decisions

- **R-01 known limits (recorded, not hidden):** (1) If the user presses a dropdown and then closes it without picking, the draft stays uncommitted. The host shows its "Press Enter" hint from `sm` width up; Enter, the next pick, or leaving the box again commits it. (2) Moving from the box to a dropdown with the keyboard (Shift+Tab) still commits at blur and then reloads again on the pick. The latest-request guard applies only the newest result. (3) In Safari a click does not focus the button, but the box still blurs on the press, so the same skip applies.
- **R-04:** while another page or filter loads, the toolbar keeps the previous total and last-fetched time, and Refresh stays enabled. After a failed load the toolbar keeps the last good total.
- **R-05:** `aria-disabled` plus an ignored toggle replaces `disabled`. This follows the preferred option of functional-design review-01 R-05.

### Red outputs

| Step | Failing output |
|---|---|
| 24 | `vitest run src/issues/issues-page.test.tsx` → `commits a typed query with the dropdown pick that follows it in one reload, although opening the dropdown blurs the box first (R-01, BR4.2, BR4.5)`: `AssertionError: expected [ [ 'issues.list', { …(2) } ], …(1) ] to have a length of 1 but got 2` (the blur committed and reloaded before the pick) — 1 failed / 12 passed |
| 26 | `keeps the toolbar count and Refresh while another page loads; only the first load and Refresh show loading (R-04, FR4.1)`: `AssertionError: expected '0' to be '57'` — 1 failed / 13 passed. ESLint `prefer-const` then flagged the test's `let pending`; the test was rewritten to answer page 1 and hold page 2, with the same assertions. |
| 27 | `vitest run src/page/backlog-lists.test.tsx` → `filters with label-less dropdowns and a Status (n) multi-select…`: `AssertionError: expected true to be false` (`open.disabled`) — 1 failed / 19 passed |

### Final results

- Scoped command (`npm --prefix ui exec -- vitest run --root ui …`, 7 files): **130 tests passed** (was 129; +1 new test).
- Full Vitest: 31 files, **330 tests passed** (was 329).
- `npm --prefix ui run typecheck`, `lint` and `format:check`: clean.
- No Go, manifest or action file changed in this loop-back, so the Go suite, coverage, package and contract-test results above still apply; they were not re-run. No `coverage.out` at the repository root.
- `source-manifest.json`: all paths changed in Loop-back 1 were already listed; no change. `traceability.json`: targets are unchanged and still exist (BR4.2 and BR4.5 → `issues-page.test.tsx`, FR4.1 → `issues-page.tsx`, BR5.3 → `status-multi-filter.tsx`).

## Loop-back 2

The branch was rebased onto `origin/main` `5724d88` (v0.4.0, PR #9), which added the GitHub, GitLab and Bitbucket pull request list. By the human's decision, FR5 now also covers that list, and this intent ships as release 0.4.1. TDD: Red, then Green, then refactor while green.

### Files changed

| File | Change |
|---|---|
| `ui/src/git/pr-toolbar.tsx` | Step 29: the `idPrefix` prop is back (default `backlog-prs`). It prefixes every test id (`-toolbar`, `-list`, `-count`, `-count-mobile`, `-updated`, `-refresh`), so the Backlog list keeps `backlog-prs-*` and the provider list keeps `backlog-scm-prs-*`. |
| `ui/src/git/status-multi-filter.tsx` | New `idPrefix` prop (default `backlog-prs`) for the trigger and checkbox ids, so the provider list can reuse the same "Status (n)" popover with the same last-status guard. |
| `ui/src/git/pr-list.tsx` | The provider selector is a host `Select` without a field label. Its trigger has `aria-label` "Provider" and the GitHub trigger classes (`FILTER_TRIGGER`). It is the first item in the filter slot of both lists. `FIELD` and `Label` are no longer used. |
| `ui/src/git/scm-pr-list.tsx` | Rows use the host `TaskRowIndicator` with `prTaskRowLinks` (`backlog-scm-pr-task-<number>-single`/`-multi`, the task id as fallback title). Repository, Saved query and Author are label-less `IntegrationRepositoryFilter`s with accessible names and GitHub classes. `""` is the All choice: "Choose a saved query" for Saved query, and "Anyone" for Author (`""` maps to `anyone`). Status uses `StatusMultiFilter` with `idPrefix="backlog-scm-prs"`, so the last checked status is `aria-disabled`. Everything sits in a `FILTERS` slot (`backlog-scm-prs-filters`) with Save query last. Removed: the old task anchors, the `Select`/`Label` fields and the inline checkbox group. |
| `ui/src/git/pr-list.test.tsx` | v0.4.0's SCM tests now use the new ids (updated, none deleted): repository options via the filter's option buttons, including All; the task indicator `-single` (click goes to `/t/task-9`) and `-multi` "Tasks 2" after linking; status picked through the popover; author through `-option-me`; the saved query through its filter value and options. New test: the provider-list layout (provider first and label-less in both lists, no query box or `<label>`, three named dropdowns with GitHub classes, Status (n) with the last status `aria-disabled` and ignored, Anyone = `""`, axe clean). |
| `manifest.yaml` | `version: "0.4.0"` → `"0.4.1"`. No test or CI check pins the manifest version: `make package`/`verify-package` read it from the manifest. |
| `README.md` | `## Upgrade notes` gains "### 0.4.1: GitHub-style lists": search on Enter, the GitHub-style toolbar and searchable dropdown filters on the issue list and every pull request list, linked tasks with title and a "Tasks (n)" menu, the Kanban badge opening the Backlog issue, and no data or setting change. |

### Red and intermediate outputs

| Step | Output |
|---|---|
| 29 (baseline after merge) | `tsc --noEmit`: 15 errors (`pr-list.tsx`: `FIELD`, `Label`, `Select*` not in scope; `scm-pr-list.tsx`: no `taskHref` export, no `messages.taskLink`, `idPrefix` not in `PrToolbarProps`). Vitest: 21 failures in `src/git/pr-list.test.tsx` and `src/page/backlog-lists.test.tsx` (the plan's recorded baseline; at run time the same missing names fail every render of the pull request list). |
| 29 (after the compile fix) | `tsc` clean. Full Vitest: **2 failed / 370 passed**. (1) `pr-list.test.tsx` "lists a provider's mapped repositories…": `expected [ Array(4) ] to deeply equal [ …(2) ]`. The test counted every `button` in the repository filter, and the filter fake gained a trigger and an All option in this intent. (2) `backlog-lists.test.tsx` "filters with label-less dropdowns…": `expected <label …> to be null`. v0.4.0's provider selector still had a field `Label`. To compile without doing Step 31's work early, the provider `Select` kept its label for now, and the SCM task anchor used a local `/t/<id>` href in place of the removed `taskHref`/`taskLink`. |
| 30 | `vitest run src/git/pr-list.test.tsx` → **5 failed / 4 passed**: "lists a provider's mapped repositories…": `TypeError: Cannot read properties of null (reading 'getAttribute')` (no `backlog-scm-pr-task-42-single`). "filters by status and author and pages": `reading 'click'` (no `backlog-scm-prs-status` trigger). "opens on the provider's default saved query…": `AssertionError: expected null to be 'q1'` (Saved query was a `Select`). "links a task created from a row…": `reading 'textContent'` (no `-multi`). "lays the provider list out like GitHub…": `expected <label …> to be null`. |
| 31 (Green) | The first Green run failed to compile: `TS2874: This JSX tag requires 'Fragment' to be in scope`. The esbuild/tsc JSX setup has no `Fragment` in scope, so I replaced the `<>…</>` with two branches (provider only while the filters load, else the full slot). Then `pr-list.test.tsx` + `backlog-lists.test.tsx`: **29 passed**. |

### Key decisions

- **Provider selector:** a host `Select`, not an `IntegrationRepositoryFilter`. The provider list has no "All" choice, and the searchable filter always adds one (`""`). The label is gone and the trigger has `aria-label` and the GitHub width, so it looks and is named like the other filters.
- **Saved query "All":** the filter's `""` choice ("Choose a saved query") clears the saved-query highlight and keeps the current filters. It does not reset them.
- **Status on the provider list:** it now follows BR5.3 like the Backlog list: at least one status stays picked. Before this change, the provider list could clear every status.
- **Step 32 (refactor): not done, on purpose.** The two `dropdown` helpers in `pr-list.tsx` and `scm-pr-list.tsx` differ only in the test-id prefix. They are closures over `h` and the host `IntegrationRepositoryFilter`. A shared factory would add about as many lines as it removes, plus a file, so the plan's condition ("only if it removes duplication") is not met. Both stay local. The shared parts are already shared: `FILTER_TRIGGER`, `FILTER_POPOVER`, `FILTERS`, `StatusMultiFilter` and `prTaskRowLinks`.
- **BR6.2:** the manifest `version` changes for the 0.4.1 release (human decision, Loop-back 2). No other manifest field, Go code or action changed.

### Final results

- Scoped command (`npm --prefix ui exec -- vitest run --root ui …`, 8 files including `src/git/pr-list.test.tsx`): **147 tests passed**.
- Full Vitest: 33 files, **373 tests passed**.
- `npm run typecheck`, `npm run lint`, `npm run format:check`: clean.
- `make check-format vet lint coverage package verify-package` with the complete Go 1.26.0 toolchain (`GOTOOLCHAIN=local`): all passed. golangci-lint reported `0 issues`, `ci workflows: OK`, and `go test -race` reported coverage **92.8%** (floor 80%, exclusion `server/main.go` only; profile at `build/coverage.out`, none at the repository root). `plugin-pack` wrote `dist/nulab-backlog-0.4.1.tar.gz`, and `verifypkg: OK dist/nulab-backlog-0.4.1.tar.gz (nulab-backlog@0.4.1)`.
- `make contract-test KANDEV_MIN_DIR=../kandev`, with `../kandev` at `v0.96.0` (`f099a46dc`): **10/10** `ci contract: OK nulab-backlog on Kandev v0.96.0`.
- `source-manifest.json` now also lists `ui/src/git/scm-pr-list.tsx`, `ui/src/git/pr-list.test.tsx`, `manifest.yaml` and `README.md`. In `traceability.json`, FR5.1 now points to `ui/src/git/pr-list.test.tsx` and BR5.2 to `ui/src/git/scm-pr-list.tsx`.
- Still open (NFR2, as before): the contract test proves install and run on v0.96.0, not the rendered UI. The visual check of the provider list in a real Kandev page belongs to the manual check.
