# Requirements — Backlog Panel Retouch (GitHub-style UI/UX)

Intent: `261007-backlog-panel-retouch` · Scope: refactor · Depth: Minimal

## Intent Analysis

Initial description (verbatim): "Bắt chước giống UIUX của Github integration bên kandev. Card issue => hiện thị rõ task; click vào task => mở task tương ứng; click vào title của Issue => mở link issue trên browser; phần filter panel: retouch lại"

The user wants the Backlog plugin's lists to look and behave like Kandev's built-in GitHub integration (Kandev v0.96.0, `apps/web/components/github/my-github/` and `apps/web/components/integrations/`), so that a person who switches between GitHub and Backlog in Kandev sees one consistent pattern:

- linked Kandev tasks are visible at a glance and one click away;
- the Backlog issue itself is one click away in the browser;
- the filter panel is the same compact, one-row toolbar GitHub uses.

This is a UI refactor. No new Backlog API calls, actions, or stored data are needed (code scan: `TaskRowIndicator` resolves the task title from the host task store, so the backend's existing `linkedTasks {taskId, taskKey}` is enough).

Request type: refactoring/enhancement of existing UI · Scope: multi-component inside `ui/src` (issues page, PR list, Kanban badge, shared toolbar) · Complexity: simple-to-standard.

## Functional Requirements

### FR1 — Linked tasks on the issue row (source: [desc] "Card issue => hiện thị rõ task", "click vào task => mở task tương ứng"; [Q1]=C, [Q2]=A)

- **FR1.1** The issue row on the `/backlog` page shall render its linked Kandev tasks with Kandev's linked-task component (`host.ui.TaskRowIndicator`), the same way the GitHub issue list does: the task title (resolved by the host; the task key or id as fallback) with the task icon.
- **FR1.2** When an issue has more than one linked task, the row shall collapse them into the component's task menu (GitHub's "Tasks (n)" dropdown) listing every linked task.
- **FR1.3** Clicking a linked task (directly or from the menu) shall open that task in Kandev.
- **FR1.4** When an issue has no linked task, the row shall show no task indicator (or the component's empty label), and the existing "+ Task" quick action stays available.

Acceptance criteria:
- Given an issue linked to one task, When the `/backlog` page renders, Then the row shows that task's title and icon, And clicking it opens the task in Kandev.
- Given an issue linked to three tasks, When the row renders, Then one "Tasks (3)" control is shown, And opening it lists all three, And choosing one opens that task.
- Given a linked task the host cannot find (deleted or not yet loaded), When the row renders, Then the fallback label (task key, else id) is shown instead of a blank.

### FR2 — Backlog badge on Kanban task cards (source: [Q1]=C, [Q7]=A)

- **FR2.1** The Backlog badge on a Kandev Kanban task card shall clearly identify the linked Backlog issue (issue key and status visible on the card).
- **FR2.2** Clicking the badge shall open the linked Backlog issue in the browser in a new tab (`https://<space>/view/<KEY>`, `rel="noopener noreferrer"`), replacing today's toggle of the status detail.

Acceptance criteria:
- Given a task linked to issue `PROJ-12`, When the Kanban board renders, Then the card's Backlog badge shows `PROJ-12` and its status.
- Given that badge, When the user clicks it, Then the Backlog issue page for `PROJ-12` opens in a new browser tab, And the Kanban card itself is not opened.
- Given a linked issue that is unavailable (deleted or access lost), When the badge renders, Then it shows the existing unavailable state and does not open a broken link.

### FR3 — Issue title opens the Backlog issue (source: [desc] "click vào title của Issue => mở link issue trên browser"; [Q6] supersedes [Q3])

- **FR3.1** Clicking the issue title on the `/backlog` issue row shall keep opening the Backlog issue URL in a new browser tab. The user confirmed this already works correctly; no change is required, and the refactor must not break it.

Acceptance criteria:
- Given an issue row, When the user clicks its title, Then `https://<space>/view/<KEY>` opens in a new tab (existing test `ui/src/issues/issues-page.test.tsx` keeps passing).

### FR4 — Filter panel like GitHub (source: [desc] "phần filter panel: retouch lại"; [Q4]=A)

- **FR4.1** The issue list's filter panel shall use Kandev's integration list toolbar (`host.ui.IntegrationListToolbar`) like the GitHub list: one row with the list title and result count, the filter controls, the query input, last-updated time and a refresh button.
- **FR4.2** The keyword search shall run when the user presses Enter (committed query), not automatically while typing.
- **FR4.3** The Project, Status and Assignee filters shall be compact dropdown controls (GitHub's filter combobox style, no separate field labels; each has an accessible name) placed in the toolbar's filter slot on one row on desktop.
- **FR4.4** Changing a dropdown filter shall reload the list immediately (as today) and reset to the first page.
- **FR4.5** Existing saved-query and default-query behaviour (save query, apply saved query, scope bar) shall keep working with the new toolbar.
- **FR4.6** On phone width the toolbar shall stay usable without horizontal page scroll (filters may move behind a compact control, as GitHub does).

Acceptance criteria:
- Given the `/backlog` issue list, When the user types a keyword and does not press Enter, Then the list does not reload; When the user presses Enter, Then the list reloads with that keyword from page 1.
- Given the toolbar, When the user picks a Status in its dropdown, Then the list reloads filtered by that status.
- Given a saved query with Project and Assignee set, When the user applies it, Then the toolbar's dropdowns and query input show those values and the list matches.

### FR5 — Pull-request list kept consistent (source: [Q5]=A)

- **FR5.1** The pull-request list shall render linked tasks with the same linked-task component and behaviour as FR1.1–FR1.4.
- **FR5.2** The pull-request list shall use the same GitHub-style toolbar as FR4 (Enter-to-search, compact dropdown filters, count, last-updated, refresh), keeping its existing filters and saved-query behaviour.

Acceptance criteria:
- Given a PR linked to a task, When the PR list renders, Then the task title and icon are shown, And clicking opens the task.
- Given the PR list toolbar, When the user presses Enter in the query input, Then the PR list reloads with that query.

## Non-Functional Requirements

- **NFR1 — Consistency with Kandev**: Components used for FR1, FR4 and FR5 are the host's own (`host.ui.*`) where the host provides them, so styling follows Kandev's theme; the plugin bundle adds no CSS of its own (existing locked rule).
- **NFR2 — Compatibility**: Works on the declared minimum Kandev version (`min_kandev_version`, v0.96.0); proven by the packaged-host contract test, because jsdom cannot prove host components render inside the plugin route.
- **NFR3 — Accessibility**: Every filter control and the query input has an accessible name; task links and the badge are keyboard reachable; external links use `target="_blank"` with `rel="noopener noreferrer"`.
- **NFR4 — Quality gates**: Existing Go and UI suites stay green (baseline: Go `-race` all pass; Vitest 322/322; tsc, ESLint, Prettier clean). Go coverage floor 80% unchanged. Tests that relied on removed test ids are updated, not deleted, so the behaviour stays covered.
- **NFR5 — Security**: No secrets or Backlog response content reach logs or the UI beyond what is shown today (existing redaction rules unchanged).

## Constraints

- Kandev host API at v0.96.0 only (`host.ui.TaskRowIndicator`, `IntegrationListToolbar`, `Combobox`, `Popover`, `Sheet`/`Drawer`); no host changes.
- Existing action keys, the opt-in default, `settingsHref()` format and the settings card/switch/nav entry behaviour are unchanged (BR5.4/BR7.6/BR7.8).
- Backlog logo shown only unmodified (Nulab brand terms).
- Team practices: TDD ordering, Vitest for UI, testify for Go, `../kandev` linked to the pinned v0.96.0 checkout before Code Generation.

## Assumptions

- [assumption] On the Kanban card (FR2.1), "show the task clearly" is read as "the badge clearly shows the linked Backlog issue", since the card already is the task. Confirm at the gate if a different display is wanted.
- [assumption] `IntegrationListToolbar` accepts one `filter` node; the three dropdowns are placed together in that node. If it cannot hold them on one row at desktop width, the plugin keeps the GitHub layout and moves extra filters behind one compact "Filters" control.
- [assumption] No backend change is needed; passing a task title from Go as `fallbackTitle` is optional and decided in Functional Design.

## Out of Scope

- Changing how the issue URL is built (user confirmed it works).
- New Backlog features, new actions or new stored data.
- Settings screen, watches and connection flows.
- Changes to the Kandev host.

## Open Questions

- Exact mobile layout of the three filters (FR4.6) — settle in Functional Design against GitHub's mobile behaviour.
- Whether the test harness fakes for `TaskRowIndicator` / `IntegrationListToolbar` / `Combobox` mirror host props closely enough — settle in Functional Design.

## Sources

- [desc] Initial description (verbatim above).
- [scope] Workflow-selected scope: refactor.
- [Q1]–[Q7] `requirements-analysis-questions.md` in this directory.
- Code knowledge base: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (business-overview.md, architecture.md § "External Reference: Kandev GitHub Integration (v0.96.0)", code-structure.md, code-quality-assessment.md § "Intent 261007-backlog-panel-retouch Risks").
