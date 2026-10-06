# Requirements Analysis — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Scope, UI and way of working were settled in earlier steps (`scope-document`, `wireframes`, `team-practices`), so they are not asked again. The questions below only settle the open details that the requirements need.

---

## Q1. Who uses the Backlog credentials

When a Kandev workspace is connected to Backlog, does everyone in the workspace share one account, or does each person use their own account? The answer affects who can see which issues, whose name is on a created pull request, and the API rate limits (Backlog counts limits per user).

- A. One shared connection for the whole workspace, like the Bitbucket plugin
- B. Each Kandev user connects with their own Backlog account
- C. One shared default connection; each person can use their own account if they want
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Link between issue and task

When linking, in what ratio are an issue and a task tied together? (The sketching step tentatively assumed "an issue links to at most one task".)

- A. One issue ↔ one task (each side ties to only one)
- B. An issue can tie to many tasks; each task ties to at most one issue
- C. No limit on either side
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q3. Status update frequency

Issue and pull request status is updated by polling Backlog periodically. How often should it poll?

- A. Every 5 minutes, plus a manual refresh button
- B. Every 1 minute, plus a manual refresh button
- C. Configurable (default 5 minutes, no less than 1 minute), plus a manual refresh button
- D. Not yet defined
- X. Other (please specify)

[Answer]: C

## Q4. Which projects are shown

Which Backlog projects should the issue list and repositories come from?

- A. Every project the connected account can view
- B. Only the projects selected on the settings page
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q5. Information carried over when creating a task from an issue (select all that apply)

When you click "Create task" on an issue, what information does the new task get?

- A. The issue title and description
- B. The issue key (for example `PROJ-120`) and the link to the issue on Backlog
- C. Assignee, priority, due date (shown alongside, not synced back)
- D. The issue's attachments and comments
- X. Other (please specify)

[Answer]: A, B, C, D

## Q6. What a pull request watch (PR watch) does when it finds results

Each watch is a filter that runs periodically over pull requests. When a new pull request matches the filter, what should the plugin do?

- A. Automatically create one Kandev task for each new matching pull request (for example, for an agent to review)
- B. Only show it in the list and the dashboard, do not create a task automatically
- C. Ask the user before creating a task
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q7. Response speed

Is there a specific target for display speed (when Backlog responds normally)?

- A. An issue or pull request list (20 rows) displays within 3 seconds for 95% of opens
- B. No specific target, just a clear loading state
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: A. One Backlog connection shared by the whole Kandev workspace, like the Bitbucket plugin
- Q2: B. An issue can tie to many tasks; each task ties to at most one issue
- Q3: C. Configurable update frequency (default 5 minutes, no less than 1 minute), plus a manual refresh button
- Q4: B. Only take issues and repositories from the projects selected on the settings page
- Q5: A, B, C, D. A task created from an issue gets: title and description; issue key and link; assignee, priority, due date (shown alongside, not synced back); attachments and comments (a read-only copy taken when the task is created)
- Q6: A. A PR watch automatically creates one Kandev task for each new pull request that matches the filter
- Q7: A. A 20-row list (issues or pull requests) displays within 3 seconds for 95% of opens, when Backlog responds normally

Does this all look correct before I generate the requirements artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
