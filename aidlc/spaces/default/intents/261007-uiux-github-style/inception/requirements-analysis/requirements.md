# Requirements — 261007-uiux-github-style

## Intent Analysis

- **Request** (verbatim): "Sửa lại UIUX tương tự như cách Github integration làm: (1) PR, Issue watcher setting đặt trong Settings > Integrations > Backlog; (2) ở Home > Menu Integrations chỉ hiện duy nhất 1 mục cho Backlog, bên trong có giao diện Issue list / PR list (tương tự Github integration); (3) hiệu chỉnh textbox, buttons theo style của Github integration; (4) chỉnh logo kiểu outlined only để đồng bộ với icon hiện hữu"
- **Goal**: the Backlog plugin should look and behave like Kandev's first-party GitHub integration, so a Kandev user who knows GitHub's screens finds Backlog's in the same places with the same controls.
- **Type**: UI/UX refactor plus two new backend features the user chose to include (issue watch, PR list action) — Q1 = B, Q2 = B, F3 = A.
- **Scope**: multi-component — UI bundle (`ui/src/`), manifest, KandevAdapter, Git and Issues domain services, BacklogGateway.
- **Complexity**: standard. The GitHub model is known (Kandev `v0.96.0`), but SDK v0.96.0 cannot reproduce every GitHub component, so screens are composed from `host.ui`.

## Functional Requirements

### FR1 — Watch settings inside Settings > Integrations > Backlog (request item 1, Q4, Q5)

- **FR1.1** The plugin settings screen shall show stacked, framed sections in this order: Connection, PR watches, Issue watches, Issue sync, Git access, Projects.
- **FR1.2** The PR watches section shall list existing PR watches in a table (name, project, repository, filters, workflow, state) with header actions to add a watch, and per-row actions to edit, run, pause/resume and delete; add and edit open in a dialog. It uses the existing `git.watches.*` actions.
- **FR1.3** The Issue watches section shall offer the same table, header actions and dialog for issue watches (FR3).
- **FR1.4** Any signed-in member who can open the settings screen can see and edit PR and issue watches; access of `git.watches.*` stays `authenticated` and new issue-watch actions are `authenticated` (Q4 = B). Connection-changing actions and `issues.set_poll_interval` stay `admin`.
- **FR1.5** A member who is not an admin shall still see the watch sections when the connection-management parts are hidden by the member view.
- **FR1.6** The notice in Settings that today links to "Review watches" (`/backlog/watches`) shall point to the PR watches section in Settings (Q5 = B).

Acceptance:
- Given a connected workspace, When a member opens Settings > Integrations > Backlog, Then the PR watches and Issue watches sections are visible and a watch can be added, edited and deleted there.
- Given a member without admin rights, When they open the screen, Then connection-changing controls are not shown but the watch sections are.

### FR2 — One Backlog entry under Home > Integrations (request item 2, Q2, Q5)

- **FR2.1** The plugin shall register exactly one nav item in the `integrations` section, pointing to `/backlog`.
- **FR2.2** The routes `/backlog/watches` and `/backlog/dashboard` and their nav items shall be removed (Q5 = B).
- **FR2.3** `/backlog` shall show a scope switch between **Issues** and **Pull requests**, a list toolbar (title, result count, search where supported, refresh), the list, and pagination, following the GitHub integration layout composed from `host.ui` (`IntegrationScopeBar`, `IntegrationListToolbar`, `Table*`, `ChangeRequestList`/`ChangeRequestRow`, pagination, `Empty*`, `Alert*`).
- **FR2.4** The Issues scope keeps today's issue list behaviour: search, project/status/assignee filters, 20 rows per page, create task from issue, link to task.
- **FR2.5** The Pull requests scope shall list the PRs of a selected repository through the new action (FR4), with status, assignee and creator filters and paging; each row shows PR number, title, status, author, updated time and the linked Kandev task if any.
- **FR2.6** Saved PR queries shall appear as presets in the Pull requests scope; choosing one applies its repository and filters to the list; the current filters can be saved as a query.
- **FR2.7** When the workspace is not connected or the integration is disabled, `/backlog` shall show an alert with a link to the settings page instead of the lists.

Acceptance:
- Given the plugin is installed, When the Integrations menu renders, Then it shows one Backlog entry and no Watches or Dashboard entries.
- Given a connected workspace, When the user switches to Pull requests and picks a repository, Then PRs are listed 20 per page with working next/previous paging.
- Given a saved query, When the user selects it as a preset, Then the list shows the same PRs the query defines.

### FR3 — Issue watch (Q1 = B, F1 = A, F2 = B)

- **FR3.1** A member shall be able to create an issue watch with: name, project (one of the selected projects), statuses, assignee, creator, and the Kandev workflow and optional workflow step for created tasks.
- **FR3.2** The plugin shall run issue watches in the background and, for each Backlog issue that matches a watch, create one Kandev task in the configured workflow/step and link it to the issue (same link as "create task from issue").
- **FR3.3** On the first run after a watch is saved, the watch shall also pick up matching issues that already exist (F2 = B).
- **FR3.4** An issue already linked to any Kandev task, or already handled by the same watch, shall never create another task (deduplication ledger per watch).
- **FR3.5** Issue watches shall support list, save, delete, run now, pause and resume, mirroring PR watches.
- **FR3.6** Issue watch validation shall reject a project that is not among the selected projects and a workflow that does not exist, with a user-facing error message.

Acceptance:
- Given an issue watch on project P with status Open, When the watcher runs, Then each open issue in P without a linked task gets exactly one new task linked to it.
- Given the same watch runs again, When no new issue matches, Then no task is created.
- Given an issue already linked to a task, When it matches a watch, Then no new task is created.
- Given a paused watch, When the watcher runs, Then it creates no tasks.

### FR4 — PR list action (Q2 = B)

- **FR4.1** A new `authenticated` action shall return the pull requests of one repository of a selected project, with optional status, assignee and creator filters, and offset paging (page size 20) plus a total count or a "has more" flag.
- **FR4.2** Each returned PR shall include number, title, status, author, assignee, updated time, Backlog URL and the linked Kandev task if any.
- **FR4.3** The action key shall match `^[a-z0-9][a-z0-9._-]*$` and be declared in `manifest.yaml`.

Acceptance:
- Given a repository with 45 PRs, When pages 1, 2 and 3 are requested, Then they return 20, 20 and 5 PRs and the last page reports no more results.

### FR5 — Controls in the GitHub style (request item 3)

- **FR5.1** Raw `<select>`, `<input type="radio|checkbox">`, `<table>`, `<details>` and `<button>` elements in the plugin UI shall be replaced by `host.ui` components (`Select*`, `Checkbox`, radio via `Select`/`Tabs` or a host equivalent, `Table*`, `Button`, `Input`).
- **FR5.2** Buttons shall use GitHub-style `variant`/`size`: primary actions default; secondary actions `outline`; section header actions `size="sm"`; icon-only actions `variant="ghost"` with an icon size and an accessible label; destructive actions `destructive`.
- **FR5.3** Text inputs shall use `host.ui.Input` with the host's control size and labels.

Acceptance:
- Given the plugin UI source, When it is searched, Then no raw `<select>`, `<table>`, checkbox/radio `<input>` or `<button>` element remains in plugin components.

### FR6 — Outline icon (request item 4, Q3 = A)

- **FR6.1** The plugin icon shall be an original stroke-only SVG that uses `stroke="currentColor"`, `fill="none"`, sizes from `className`, and is not derived from the Nulab Backlog mark.
- **FR6.2** The same icon shall be used for the nav entry, the settings card and the page header, through the single `PLUGIN_ICON` selection point.
- **FR6.3** The filled official Nulab mark shall no longer be used in the plugin UI.

Acceptance:
- Given the nav, settings card and page header, When they render, Then each shows the outline icon in the surrounding text colour.

## Non-Functional Requirements

- **NFR1 Testing**: TDD per `team.md`; Go code keeps the 80% line-coverage floor (`make coverage`, current 92.9%), runs with `-race`; UI tests stay green (current 229/229) and every changed or new component has a Vitest test. Backlog is faked with `httptest` for the issue watcher and PR list.
- **NFR2 Security**: no secret, API key or token appears in logs, errors or UI responses (existing redaction test extended to new actions); public error messages carry no Backlog response content.
- **NFR3 Rate limits**: the issue watcher uses the gateway's background call class and existing per-group rate limiter; a 429 with `Retry-After` delays the run and does not create duplicate tasks.
- **NFR4 Accessibility**: icon-only buttons have accessible names; tables have headers; existing axe-core checks pass on changed screens.
- **NFR5 Compatibility**: the plugin keeps working on `min_kandev_version` 0.96.0; the packaged-host contract test passes; only `host.ui` components available in SDK v0.96.0 are used.
- **NFR6 Startup**: Host-dependent paths of the issue watcher wait (bounded) for the host to be injected, as with the PR watcher.

## Constraints

- Kandev SDK `v0.96.0` (`../kandev` at `f099a46`): no `PageShell`, GitHub-only scope bar/toolbar, nested nav, settings sub-routes, `AlertDialog`, or `@tabler/icons-react`.
- Only `internal/plugin` and `server/` import `pluginsdk`; Backlog calls only through `internal/backlog`.
- Settings card, switch and nav entry stay independent of the enabled state (BR5.4/BR7.6/BR7.8); `settingsHref()` format unchanged; existing action keys unchanged.
- Only `https` space addresses under `backlog.com`, `backlog.jp`, `backlogtool.com`.
- Nulab brand terms: no modified or recoloured Nulab logo.

## Assumptions

- The issue watcher reuses the PR watcher's pattern: a fixed background interval (5 minutes), a per-watch ledger, and pause/resume state in the plugin state store. Owner: Functional Design.
- "Open issue" in F2 means an issue matching the watch's status filter; with no status filter, every issue not in a closed status.
- The PR list action uses Backlog's pull request list with offset/count and the PR count endpoint for totals; there is no keyword search on PRs because Backlog's PR API has none.
- Removing `/backlog/watches` and `/backlog/dashboard` breaking old bookmarks is accepted (Q5 = B).

## Out of Scope

- Issue watch notifications, analytics or charts like GitHub's analytics section.
- Changing who can edit watches (Q4 = B).
- Webhook-based (push) watching; watches stay polling-based.
- Redrawing or recolouring the official Nulab logo.

## Open Questions

- Exact `host.ui` replacement for radio groups (SDK v0.96.0 has no `RadioGroup` in the scanned surface) — resolve in Functional Design.
- Whether the issue watch interval should be configurable or fixed like the PR watcher — Functional Design proposes; default fixed.

## Sources

- Initial description: `project-description.json` (verbatim above).
- Answers: `requirements-analysis-questions.md` Q1–Q5, F1–F3, Consolidated Summary Confirmation.
- Code knowledge base: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/business-overview.md`, `architecture.md` (GitHub integration reference table), `code-structure.md`, `api-documentation.md`, `code-quality-assessment.md` (Intent 261007 risks).
- Rules: `team.md` Testing Posture and Code Style; `project.md` Mandated/Forbidden.
