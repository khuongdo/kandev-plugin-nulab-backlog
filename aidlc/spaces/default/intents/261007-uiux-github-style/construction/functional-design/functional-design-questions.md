# Functional Design Questions — 261007-uiux-github-style

Context: the requirements (FR1–FR6) are approved. These questions settle design choices the requirements and the product-lead review (R-01 to R-07) left open. Answer each by writing the letter after the answer tag.

## Q1. How many tasks may an issue watch create at once?

A new issue watch also picks up existing matching issues (FR3.3), so a first run on a busy project could create many tasks. The existing PR watcher already caps each watch at 10 new tasks per cycle (every 5 minutes) and leaves the rest for later cycles.

A. Same cap as PR watches: at most 10 new tasks per watch per cycle; the rest follow in later cycles
B. Option A, plus the add/edit dialog shows how many issues currently match before saving
C. No cap: create a task for every matching issue in one run
X. Other (please specify)

[Answer]: X. Other: 1 task (at most one new task per issue watch per cycle)

## Q2. When should issue watches run?

A. In the same 5-minute background loop as PR watches, with the same "run now", pause and resume behaviour
B. On the issue status sync interval set by the admin (1–1440 minutes)
X. Other (please specify)

[Answer]: X. Other: default 5 minutes, adjustable in Settings

## Q3. Where are saved PR queries edited and deleted once the dashboard page is removed?

Today saved queries are created, run and deleted on `/backlog/dashboard`, which FR2.2 removes. FR2.6 makes them presets of the PR list.

A. In a "Saved PR queries" section in Settings (table with edit and delete); the PR list only selects presets and saves the current filters
B. In the PR list itself: the preset menu offers rename and delete next to each saved query
C. Both places
X. Other (please specify)

[Answer]: A

## Q4. What replaces the sign-in method radio buttons (API key / Sign in with Nulab) in Settings?

The host UI has no radio group component; FR5.1 removes raw radio inputs.

A. Two tabs: "API key" and "Sign in with Nulab"
B. A dropdown with the two methods
X. Other (please specify)

[Answer]: B

## Follow-up Questions

Q2 says issue watches run every 5 minutes by default and the interval can be changed in Settings. Two details decide where that setting lives.

## F1. What does the adjustable interval apply to?

A. One interval for all issue watches of the workspace, set in the Issue watches section of Settings
B. Each issue watch has its own interval, set in its add/edit dialog
C. One interval for all watches (PR and issue) of the workspace, set in Settings
X. Other (please specify)

[Answer]: B

## F2. Who may change that interval, and in what range?

A. Admins only, 1–1440 minutes (same as the issue status sync interval)
B. Any signed-in member (like editing watches), 1–1440 minutes
C. Admins only, 5–1440 minutes (never more often than today's 5 minutes)
X. Other (please specify)

[Answer]: B

## Consolidated Summary Confirmation

- Task cap (Q1): an issue watch creates at most one new task per run; remaining matching issues are handled one per later run.
- Schedule (Q2, F1): each issue watch has its own run interval, default 5 minutes, set in its add/edit dialog in Settings; "run now", pause and resume work as for PR watches.
- Interval permissions (F2): any signed-in member who can edit watches may set the interval, from 1 to 1440 minutes.
- Saved PR queries (Q3): a "Saved PR queries" section in Settings lists them with edit and delete; the PR list only selects a saved query as a preset and saves the current filters as a new query.
- Sign-in method (Q4): a dropdown with "API key" and "Sign in with Nulab" replaces the radio buttons.
- PR watches keep their current fixed 5-minute loop and 10-task cap; only issue watches get the per-watch interval and the one-task cap.

Does this all look correct before I generate the functional design artifacts?

Looks correct
Request changes

[Answer]: Looks correct
