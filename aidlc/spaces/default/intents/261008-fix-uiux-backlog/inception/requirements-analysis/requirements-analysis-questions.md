# Requirements Analysis Questions - 261008-fix-uiux-backlog

Context: the code scan (codekb `kandev-plugin-nulab-backlog`, developer-scan.md) found that most of the request is a local UI change, but two items need a decision first. Depth: Minimal.

## Question 1
Kandev v0.96.0 gives a plugin no way to hide its Home > Integrations menu entry later (no visibility flag, no unregister, no late registration). Today the entry is always shown on purpose (earlier rules BR5.4, BR7.6, BR7.8). How should "hide Backlog in Home > Integrations when OFF" work?

A. At page load, add the menu entry only when Backlog is ON in at least one workspace; turning Backlog on or off later takes effect after a page reload (no Kandev change)
B. Remove the Home > Integrations entry completely; the Backlog page is reached from the Settings > Integrations card
C. Change Kandev itself (upstream) so plugin menu entries follow the integration switch per workspace, and raise the plugin's minimum Kandev version
D. Keep the entry always visible (drop this item from this change)
X. Other (please specify)

[Answer]: A

## Question 2
Home > Tasks rows show plugin content through a slot that the sidebar task list also uses. Where should the Backlog issue badge appear?

A. Home > Tasks list only (not the sidebar task list); the Kanban card keeps its badge as today
B. Home > Tasks list and the sidebar task list; the Kanban card keeps its badge as today
X. Other (please specify)

[Answer]: B

## Question 3
Hovering the badge should show the issue summary. Links stored today have no summary, so the backend must start saving it (it is filled for older links at the next sync; until then the key is shown). What should the hover card show?

A. Issue key and summary only (for example "PROJ-12: Fix login")
B. Issue key, summary and current status
C. Issue key, summary, status and last updated time
X. Other (please specify)

[Answer]: B

## Question 4
"Space (project) selection" in Backlog settings is the Projects section (project picker), which is currently the last section. The sign-in method dropdown is inside the Connection section. Is this the right change?

A. Move the whole Projects section to right after the Connection section
X. Other (please specify)

[Answer]: A

## Question 5
The duplicate "Add watch" button (shown inside the empty list in addition to the button at the top right) exists in both the Issue watches list and the PR watches list. Which lists should lose it?

A. Both Issue watches and PR watches
B. Issue watches only
X. Other (please specify)

[Answer]: A
