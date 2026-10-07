# Requirements Analysis Questions — 261007-uiux-github-style

Context: restyle the plugin UI after Kandev's GitHub integration (four items in the initial request). The code scan found five decisions the request does not settle. Answer each by writing the letter after `[Answer]:`.

## Q1. What does "Issue watcher" mean in item (1)?

Today the backend only has PR watches (they create a review task for each new pull request). GitHub's Issue watch creates a Kandev task for each new issue that matches a filter; the plugin has nothing like it. The only issue-related background setting is the issue status sync interval.

A. Only move the existing issue status sync interval into the Settings page (no new backend feature)
B. Build a real Backlog issue watch (new backend feature: create a task for each new matching issue), like GitHub
C. Move the sync interval now and record a real issue watch as a follow-up, out of scope for this refactor
X. Other (please specify)

[Answer]: B

## Q2. How should the PR list inside the single Backlog page work (item 2)?

Today PRs are only visible through saved queries: one repository, at most 20 PRs, no paging, no search. GitHub shows a full PR list with preset filters and paging.

A. Reuse saved queries as the PR list's presets (pick a saved query, see its PRs); no backend change
B. Add a new backend action that lists PRs of a repository with filters and paging (manifest and contract change), and keep saved queries as presets
C. Show the PR list only for PRs linked to Kandev tasks
X. Other (please specify)

[Answer]: B

## Q3. How should the outline-only logo be made (item 4)?

The current icon is the official filled Nulab Backlog mark. Nulab's brand terms (quoted in `docs/brand/backlog-logo.md`) forbid modified or recoloured versions, so redrawing the official mark as an outline is a brand-terms risk.

A. Draw an original stroke-only `currentColor` icon (for example a ticket/board shape) that is not the Nulab mark, used in nav, settings and page header
B. Use Kandev's built-in outline icon name `ticket` (no custom drawing)
C. Redraw the official Backlog mark as an outline anyway (I accept the brand-terms risk)
D. Keep the filled official mark in Settings, and use an outline icon (A or B) only in the Integrations menu
X. Other (please specify)

[Answer]: A

## Q4. Who may see and edit PR watches once they move into Settings (item 1)?

Today any signed-in member can create and edit PR watches and saved queries. The Settings page shows members a reduced view, and connection settings are admin-only.

A. Admins only: watches become admin settings (change the actions' access to admin)
B. Keep it as today: any signed-in member can see and edit watches in Settings
C. Members can see the watch list, only admins can add, edit or delete
X. Other (please specify)

[Answer]: B

## Q5. What happens to the old pages `/backlog/watches` and `/backlog/dashboard`?

With a single Integrations entry, these two pages and their menu items go away. Old bookmarks and the "Review watches" notice in Settings point to them.

A. Remove the menu items but keep both routes, redirecting to the new location (Settings for watches, the PR list for the dashboard)
B. Remove the menu items and the routes completely; update the Settings notice
C. Remove the menu items, keep the old pages reachable by URL unchanged
X. Other (please specify)

[Answer]: B

## Follow-up Questions

Q1 = B (real issue watch) and Q2 = B (new PR list action) add backend features to a refactor. These follow-ups settle what the new issue watch does and how the work is organised.

## F1. Which conditions should an issue watch filter on?

An existing PR watch stores: name, project, repository, statuses, assignee, creator, and the Kandev workflow/step for the created task. The issue list already filters by project, status and assignee.

A. Same shape as a PR watch, minus repository: name, project, statuses, assignee, creator, plus the Kandev workflow/step for new tasks
B. Option A plus issue type, category and a keyword
C. Project only (every new issue in the project creates a task)
X. Other (please specify)

[Answer]: A

## F2. Which issues should a new issue watch pick up?

A. Only issues created after the watch is saved (the first run records existing issues without creating tasks)
B. Every open issue that matches, including ones that already exist when the watch is saved
X. Other (please specify)

[Answer]: B

## F3. Should the new backend features stay in this piece of work?

The current plan for this refactor runs one Functional Design pass, then code and tests. The issue watch and the PR list action are new features with their own data, background loop and manifest changes.

A. Keep everything in this piece of work; Functional Design covers the new backend features too
B. Do the UI refactor here; move the issue watch and the PR list action into a separate piece of work
C. Keep the PR list action here; move only the issue watch into a separate piece of work
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Issue watcher (Q1 = B): build a real Backlog issue watch that creates a Kandev task for each matching issue, configured in Settings > Integrations > Backlog next to PR watches.
- Issue watch filter (F1 = A): same shape as a PR watch without repository: name, project, statuses, assignee, creator, plus the Kandev workflow/step for new tasks.
- Issue watch pickup (F2 = B): every open issue that matches, including issues that already exist when the watch is saved; an issue already linked to a task or already handled by the watch never creates a second task.
- PR list (Q2 = B): a new backend action lists the PRs of a repository with filters and paging; saved queries stay as presets on the PR list.
- Logo (Q3 = A): an original stroke-only `currentColor` icon that is not the Nulab mark, used in the nav entry, the settings card and the page header.
- Watch permissions (Q4 = B): unchanged; any signed-in member can see and edit PR and issue watches in Settings; connection settings stay admin-only.
- Old pages (Q5 = B): remove the `/backlog/watches` and `/backlog/dashboard` menu items and routes; update the Settings notice that points to them.
- Single entry (request item 2): Home > Integrations shows one Backlog entry; `/backlog` holds an Issue list and a PR list, styled after the GitHub integration.
- Controls (request item 3): textboxes, selects, checkboxes, tables and buttons use `host.ui` components with GitHub-style `variant`/`size`.
- Organisation (F3 = A): all of the above stays in this piece of work; Functional Design covers the new backend features.

Does this all look correct before I generate the requirements artifact?

Looks correct
Request changes

[Answer]: Looks correct
