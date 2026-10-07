# Functional Design Questions

Context: Kandev v0.96.0 gives plugins the same pieces the GitHub list uses: `IntegrationListToolbar` (one-row toolbar with an Enter-to-search query box), `IntegrationRepositoryFilter` (GitHub's searchable dropdown filter with an "All" choice) and `TaskRowIndicator` (linked tasks). Two gaps showed up while matching our lists to them.

## Question 1
The pull-request list has no keyword search, because Backlog's pull-request API has none. Kandev's toolbar always shows a query box. How should the PR list toolbar work while still looking like GitHub (requirement FR5.2)?

A. Issues use Kandev's toolbar; the PR list uses our own toolbar built to the exact same layout (title + count, dropdown filters, last-updated, refresh) but without a query box
B. Both lists use Kandev's toolbar; on the PR list the query box filters the rows already loaded by title (client-side only)
C. Both lists use our own lookalike toolbar (we keep full control; issues get our own Enter-to-search box)
X. Other (please specify)

[Answer]: A

## Question 2
The PR list filters by several statuses at once (Open / Closed / Merged checkboxes) and by Assignee and Creator. GitHub's filters are single compact dropdowns. How should the PR status filter look?

A. One compact "Status" dropdown that keeps multi-select (checkboxes inside; the button shows how many are picked)
B. One single-choice dropdown (All / Open / Closed / Merged), dropping multi-select
X. Other (please specify)

[Answer]: A

## Question 3
On phones, GitHub's toolbar simply stacks: title and count, then the filters at full width, then the query box. Ours hides the filters behind a "Filters (n)" button. Which should we use (requirement FR4.6)?

A. Copy GitHub: stack the filters at full width, no "Filters" button
B. Keep our "Filters (n)" button on phones
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Issue list uses Kandev's `IntegrationListToolbar` (Enter-to-search query box); the PR list uses our own toolbar rebuilt to the exact same layout (title + count, dropdown filters, last-updated, refresh) without a query box (Q1 = A).
- PR status filter becomes one compact "Status" dropdown that keeps multi-select (checkboxes inside, button shows how many are picked) (Q2 = A).
- On phones both toolbars stack like GitHub: filters at full width, no "Filters (n)" button (Q3 = A).
- Carried from requirements: Project/Status/Assignee (issues) and Repository/Assignee/Creator (PRs) use GitHub's searchable dropdown filter (`IntegrationRepositoryFilter`); linked tasks use `TaskRowIndicator` in both lists; the Kanban badge opens the Backlog issue in a new tab; the issue-title link is unchanged.

Does this all look correct before I generate the artifact?

Looks correct
Request changes

[Answer]: Looks correct
