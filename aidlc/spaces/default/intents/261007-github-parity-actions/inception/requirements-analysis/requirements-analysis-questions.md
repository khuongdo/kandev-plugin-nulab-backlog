# Requirements Analysis Questions

Context: the code scan (`aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`) found that the plugin has no quick actions, no default queries, saved queries only for pull requests (each one tied to a repository), and a `/backlog` page with no padding. Kandev 0.96.0's GitHub integration is the model.

## Question 1: How should a quick action start a task?
In the GitHub integration, the "+ Task" menu on a row lists the quick actions (Implement, Investigate, ...). Picking one opens Kandev's "create task" dialog, prefilled with a title and the action's prompt (with the issue URL and title filled in). The user can edit it before creating the task. Today the plugin's "Create task" on an issue creates the task directly on the server, with no prompt.

A. Same as GitHub: open Kandev's create-task dialog prefilled with the prompt (editable), then link the new task to the Backlog issue or PR automatically
B. One click: create the task directly on the server with the action's prompt as its description (no dialog), and link it
X. Other (please specify)

[Answer]: A

## Question 2: Which quick actions ship by default, and where are they edited?
GitHub ships Issue actions Implement / Investigate / Reproduce and PR actions Review / Address feedback / Fix CI. They are edited in a Settings "Quick actions" section with Issues and PRs tabs: edit label and prompt, add an action, delete, and "Reset to defaults".

A. Same as GitHub: those 6 defaults, for both issues and PRs, edited in a "Quick actions" section in the plugin settings (add / edit / delete / reset), stored per workspace
B. Issues only (Implement / Investigate / Reproduce), same Settings editing; PR rows get no quick actions
X. Other (please specify)

[Answer]: A

## Question 3: What is the default query for the pull request list?
Today the PR list is empty until the user picks a repository or a saved query, and a saved PR query must name a repository (unknown when the plugin is installed). GitHub shows built-in preset buttons above the list (the first one is selected when the page opens), and lets the user star one saved query as the default.

A. A built-in preset (not stored), e.g. "Open, assigned to me", applied to the first repository of the selected projects and selected automatically when the page opens; plus a star to make one saved PR query the default instead
B. Only the built-in preset selected on open; no star on saved queries
C. Seed one real saved query at install time and relax the "repository required" rule (it then searches every repository)
X. Other (please specify)

[Answer]: A

## Question 4: What is the default query for the issue list?
Issues have no saved queries today, and the issue list cannot filter by "me" (only by numeric user IDs).

A. Same as GitHub: a built-in "Assigned to me, open" preset selected when the page opens, plus saved issue queries (save current filters, rename/delete in Settings, star one as default)
B. Only the built-in "Assigned to me, open" preset selected when the page opens; no saved issue queries
X. Other (please specify)

[Answer]: A

## Question 5: How far should the page layout follow the GitHub page?
GitHub's page has a scope bar on top (Issues / Pull requests switch + preset buttons + Saved menu), then a toolbar with title and count on the left and refresh / last-updated on the right, then the list with the "+ Task" menu at the end of each row. Bars use 16px side padding (24px on wider screens).

A. Full parity: replace the tabs with the scope bar, use the same toolbar layout and paddings, and put the "+ Task" menu at the end of each issue and PR row
B. Only paddings and button positions; keep the current tabs and toolbars
X. Other (please specify)

[Answer]: A
