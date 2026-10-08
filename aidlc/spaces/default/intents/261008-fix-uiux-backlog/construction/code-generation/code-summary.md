# Code Summary - 261008-fix-uiux-backlog

Zero-Unit express run, TDD (Red, Green, Refactor per step) per the Testing Contract in `code-generation-plan.md`.

## Files Changed

| File | Change | Requirements |
|------|--------|--------------|
| `internal/issues/types.go` | `Link.Summary` (`json:"summary,omitempty"`) | FR2.1, NFR5 |
| `internal/issues/service.go` | `newLink` stores `issue.Summary`; `LinkView.Summary` (`omitempty`), copied in `Links` | FR2.1, FR2.3 |
| `internal/issues/sync.go` | the status refresh writes the current summary next to the status | FR2.2 |
| `internal/issues/links_test.go`, `sync_test.go`, `leak_test.go`, `create_test.go` | new FR2/NFR1 tests; `create_test.go` expected link now carries the summary | FR2, NFR1, NFR5 |
| `internal/git/status_test.go` | regression guard: one and two linked PRs, every summary carries `taskStatus` | FR5.4 |
| `ui/src/issues/issues-state.ts` | `LinkView.summary?`; `badgeHover(link)` returns key, summary (when known), status | FR1.4, FR5.2 |
| `ui/src/issues/issue-badge.tsx` | badge wrapped in host `Tooltip` / `TooltipTrigger` / `TooltipContent` (hover and keyboard focus); accessible label with key and summary; native `title` removed; phone sizing `min-h-11 min-w-11` for `presentation: "mobile"` | FR1.4-FR1.6, FR5.2, FR5.5, NFR4 |
| `ui/src/index.ts` | one badge component registered for `task-card-tags`, `task-row-metadata`, `chat-top-bar`; `initialize` is async: nav item and `/backlog` route only when Backlog is ON somewhere at load, fail open on error, no workspaces, or timeout (`ENTRY_CHECK_TIMEOUT_MS = 3000`); settings card always registered; `connection.get` per workspace is shared with the switch publishing (no extra calls) | FR1.1-FR1.3, FR3.1-FR3.5, FR5.1, FR5.3, FR5.6, NFR2, NFR3 |
| `ui/src/settings/SettingsScreen.tsx` | Projects section right after Connection | FR4.1 |
| `ui/src/settings/issue-watches-section.tsx` | empty state says `issueWatchesEmpty`; no empty-state Add watch button | FR4.2, FR4.3 |
| `ui/src/settings/pr-watches-section.tsx` | no empty-state Add watch button | FR4.3 |
| `ui/src/messages/en.ts` | `issueWatchesEmpty` ("No issue watches yet"), `issueBadgeSrSummary` ("Backlog issue {label}, {summary}") | FR4.2, NFR4 |
| `ui/src/testing/harness.ts` | fake host `Tooltip`, `TooltipTrigger` (asChild, composes hover/focus handlers), `TooltipContent` | test support |
| `ui/src/index.test.ts`, `issues/issue-badge.test.tsx`, `issues/issues-state.test.ts`, `settings/sections.test.tsx` | new and updated tests (see below) | FR1, FR3, FR4, FR5, NFR2-NFR4 |
| `README.md` | settings section order; Integrations entry rule (on somewhere at load, reload after a toggle, fail open after 3 s); badge on task rows, sidebar and task top bar with hover card | FR1, FR3, FR4, FR5 |

## Key Decisions

- **Host awaits an async initialize (Step 0)**: Kandev v0.96.0 `apps/web/lib/plugins/host.ts` runs `Promise.resolve(registeredPlugin.initialize(staged.registry, host))` raced against `DEFAULT_INITIALIZE_TIMEOUT_MS = 10_000`; the staged registry stays open (`generationOpen`) until the promise settles, and a timeout fails the whole plugin generation. So `initialize` awaits the ON/OFF check and registers the nav item and route afterwards, bounded at 3 s, well under the 10 s host limit (NFR3).
- **Fail open covers "no workspaces yet"**: `host.context.getWorkspaceIds()` reads the app store and may be empty at load; an empty list is treated as unknown and the entry is shown (FR3.4). The entry is hidden only when every workspace read OFF.
- **One component for three slots**: the same `createIssueBadge` instance is registered for all three slots, so they share one `LinksStore` and one `issues.links.list` per workspace (NFR2, FR5.6).
- **Hover card replaces the native title**: the badge link no longer has a `title` attribute, so the browser tooltip does not duplicate the hover card; the detail (updated at, may be out of date) stays in the hover card and in the `sr-only` `aria-describedby` text.
- **Not-openable branch kept**: the button branch (unavailable, not connected, non-https URL) keeps its focus/tap detail; it also gets the hover card with key, summary and status.
- **FR5.4 needs no production change**: `internal/git/service.go` already sets `TaskStatus` on every summary (including failed fetches and not-connected links); the new test is a regression guard.
- **FR5.3 (archived / OFF)**: an archived task gets no `chat-top-bar` render from the host; with Backlog OFF the links list is empty or fails, so the badge renders nothing. No extra code.

## Test Results

### Baseline (before changes, recorded by the orchestrator)

- Go: `go test -race ./internal/issues/ ./internal/git/ ./internal/plugin/` all ok, 326 top-level tests.
- UI: `npx vitest run` 33 files, 373 tests pass.

### Final

- Go: `go test -race ./internal/issues/ ./internal/git/ ./internal/plugin/` all ok, 331 top-level tests (0 failures). `make coverage` (all packages, `-race`): **92.8%** (floor 80%, excluded `server/main.go` only); profile under `build/`, no `coverage.out` at the repo root.
- UI: `npx vitest run` 33 files, **387** tests pass.
- `make check-format vet lint`: gofmt, go vet, golangci-lint v2.14.0 (0 issues), prettier, `tsc --noEmit`, eslint, actionlint, `ci workflows` all clean.

### Red Evidence (TDD)

- Step 1 (Go, FR2/NFR1): `go test -race ./internal/issues/ -run 'FR2|NFR1'` -> build failed: `byTask(r.links(t), "task-17").Summary undefined (type Link has no field or method Summary)`; `got[0].Summary undefined (type LinkView has no field or method Summary)`.
- Step 1 Green side effect: `TestU3_Create_MakesTheTaskAndLink` failed (`Summary: "Fix login timeout"` vs expected `""`); its expected link was updated, as the new behaviour requires.
- Step 2 (issues-state): `TypeError: badgeHover is not a function` (2 failed).
- Step 3 (badge): `AssertionError: expected 'Backlog issue PROJ-12 · In Progress' to be 'Backlog issue PROJ-12 · In Progress, Fix login'` (2 failed). After Green, the existing test asserting `title` failed and was updated to assert no native title.
- Step 4 (index): `AssertionError: expected [ 'task-card-tags' ] to deeply equal [ 'task-card-tags', 'task-row-metadata', 'chat-top-bar' ]`.
- Step 5 (git, check only): passed on first run, as the plan foresaw; kept as the regression guard.
- Step 6 (index, FR3): `AssertionError: expected "vi.fn()" to not be called at all, but actually been called 1 times` (OFF everywhere); `TypeError: actual value must be number or bigint, received "undefined"` (no `ENTRY_CHECK_TIMEOUT_MS`).
- Step 7 (settings): 5 failed, e.g. `expected 'No PR watches yetAdd watch' to contain 'No issue watches yet'`, `expected [ <button>, <button> ] to have a length of 1 but got 2`, section order mismatch.
- Quality gate: `controls.test.ts` (Buttons must carry `${BUTTON}` inline) failed once after Step 3's Green; fixed by inlining `${BUTTON}` in both Button `className`s.

### New or Changed Tests (one per requirement)

| Requirement | Test |
|-------------|------|
| FR2.1, FR2.3 | `internal/issues/links_test.go` `TestFR2_Links_ViewCarriesTheIssueSummary` |
| NFR5 | `internal/issues/links_test.go` `TestFR2_Links_OldStoredLinkWithoutSummaryStillReads` |
| FR2.2 | `internal/issues/sync_test.go` `TestFR2_Sync_RefreshStoresTheCurrentSummary` |
| NFR1 | `internal/issues/leak_test.go` `TestNFR1_Leak_SummaryNeverInErrorsOrLogs` |
| FR5.4 | `internal/git/status_test.go` `TestFR5_Status_EveryLinkedPRCarriesTaskStatus` (1 and 2 PRs) |
| FR1.4, FR5.2 | `ui/src/issues/issues-state.test.ts` `badgeHover` (2 tests) |
| FR1.4, FR1.5, NFR4 | `ui/src/issues/issue-badge.test.tsx` hover card on hover and focus; old link without summary line |
| FR1.1, FR1.2, FR5.1, FR5.3, FR5.5 | `ui/src/issues/issue-badge.test.tsx` task-list, sidebar, desktop and phone top bar (4 tests) |
| FR1.3, FR5.6, NFR2 | `ui/src/index.test.ts` one shared badge for the three slots |
| FR3.1-FR3.5, NFR3 | `ui/src/index.test.ts` OFF everywhere, ON in one, read failure, timeout (4 tests; replaces "registers the entry and the route even when Backlog is off everywhere") |
| FR4.1 | `ui/src/settings/sections.test.tsx` `SECTIONS` order (Connection, Projects, ...) |
| FR4.2, FR4.3 | `ui/src/settings/sections.test.tsx` empty Issue watches and empty PR watches (2 tests) |

## Deviations from the Plan

- Step 1 named `internal/issues/service_test.go`, which does not exist; the link tests went to the existing `internal/issues/links_test.go` (where the `Links` tests live). `create_test.go` was also updated (expected link now has the summary).
- Step 5 named `internal/git/service_test.go`, which does not exist; the test went to `internal/git/status_test.go` (where the `Status` tests live).
- Step 4's mobile sizing branch was written during Step 3's Green (same edit of `issue-badge.tsx`), so Step 4's badge-surface tests passed on first run; Step 4's Red is the index registration test.
- `ui/src/testing/harness.ts` (test support, not in the plan's file list) gained Tooltip stubs so the hover card can be tested.
- `ui/src/issues/issue-badge.test.tsx`: an existing assertion on the native `title` was changed to assert it is absent (the hover card replaces it).
- The hover card uses the host `Tooltip` (Requirements left `Tooltip` or `Popover` open); the bound for NFR3 is 3 s.
- No refactor beyond the Green code was needed at any step.

## Open Concerns

- The hover card and the top-bar badge were verified with the fake host only; a manual look in a real Kandev 0.96.0 (Home > Tasks rows, sidebar rows, task top bar on desktop and phone) is advised in Build and Test, together with `make contract-test`.
- Hiding the Integrations entry depends on `getWorkspaceIds()` being populated at plugin load; when it is empty the entry is shown (fail open), so the entry may still appear with Backlog OFF everywhere if Kandev loads plugins before workspaces.
