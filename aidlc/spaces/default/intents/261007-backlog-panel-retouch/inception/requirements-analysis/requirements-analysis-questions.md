# Requirements Analysis Questions

Context: the request is to make the Backlog plugin's UI/UX match Kandev's built-in GitHub integration (Kandev v0.96.0, `apps/web/components/github/my-github/`). The code scan found that Kandev already offers plugins the same linked-task component the GitHub list uses (`TaskRowIndicator`), and that clicking an issue title on the `/backlog` page already opens the Backlog issue in a new browser tab.

## Question 1
"Card issue" can mean two places today. Which one should clearly show the linked Kandev task(s)?

A. The issue row in the issue list on the `/backlog` page
B. The Backlog badge on Kandev's Kanban task cards
C. Both
X. Other (please specify)

[Answer]: C

## Question 2
How should linked tasks look on the issue row? (This decides whether we reuse Kandev's own component or keep ours.)

A. Same as GitHub: reuse Kandev's linked-task component — shows the task title and icon; with several tasks, a menu lists them; clicking a task opens it in Kandev
B. Keep our own task links but show the task title instead of the task key; clicking opens the task
X. Other (please specify)

[Answer]: A

## Question 3
Clicking the issue title should open the Backlog issue in the browser. The code already does this on the `/backlog` page. What should we do about it?

A. No code change on the list; only verify it in the real Kandev app (and fix if it does not work there)
B. A, plus make the Kanban badge (if chosen in Question 1) open the Backlog issue in the browser too
X. Other (please specify)

[Answer]: X. Other: It does not work today — the URL is wrong. Investigate and fix the issue URL.

## Question 4
How far should the filter panel retouch go? GitHub's toolbar searches when you press Enter and shows filters as compact dropdowns in one row; ours searches automatically 400 ms after typing and has three filters (Project, Status, Assignee).

A. Copy GitHub exactly: use Kandev's integration list toolbar (Enter to search, compact dropdown filters in one row)
B. Keep our toolbar and its auto-search, but restyle it to look like GitHub (one compact row, same spacing, dropdown filters)
X. Other (please specify)

[Answer]: A

## Question 5
The pull-request list shares the same toolbar and task-link style. Should it change too?

A. Yes — apply the same task display and filter style to the PR list, so both lists stay consistent
B. No — change only the issue list in this piece of work
X. Other (please specify)

[Answer]: A

## Follow-up Questions

## Question 6
In Question 3 you said the issue URL is wrong today. What address does clicking the issue title open? (The code builds `https://<space host>/view/<ISSUE-KEY>`.)

A. It opens a Kandev page instead of Backlog
B. Nothing happens on click
C. Not sure; investigate during the work
X. Other (please specify)

[Answer]: X. Other: It works correctly; no fix is needed. (This supersedes the Question 3 answer: the issue-title link stays as is.)

## Question 7
What should clicking the Backlog badge on a Kanban task card do? (Today it only toggles a status detail.)

A. Open the Backlog issue in the browser (new tab)
B. Keep the current click behaviour; only make the badge show more clearly
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Show linked Kandev tasks clearly in both places: the issue row on the `/backlog` page and the Backlog badge on Kanban task cards (Q1 = C).
- Issue row linked tasks look exactly like GitHub: reuse Kandev's linked-task component (task title and icon; a menu when there are several tasks); clicking a task opens it in Kandev (Q2 = A).
- Issue-title link on the `/backlog` page already opens the correct Backlog URL; no change needed (Q6 supersedes Q3).
- Clicking the Backlog badge on a Kanban task card opens the Backlog issue in the browser in a new tab (Q7 = A).
- Filter panel copies GitHub exactly: Kandev's integration list toolbar, search on Enter, compact dropdown filters in one row (Q4 = A).
- The pull-request list gets the same linked-task display and filter style so both lists stay consistent (Q5 = A).

Does this all look correct before I generate the requirements artifact?

Looks correct
Request changes

[Answer]: Looks correct
