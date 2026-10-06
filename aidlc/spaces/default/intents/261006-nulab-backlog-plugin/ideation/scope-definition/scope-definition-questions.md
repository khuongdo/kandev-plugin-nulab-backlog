# Scope Definition — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Things already settled in earlier steps, so not asked again: the plugin covers both issues and Git/pull requests; space connection, viewing issues/repositories, and linking issues and pull requests to Kandev tasks are required; sign-in with both API key and OAuth; all Backlog domains are supported; no deadline; release via GitHub Release and the Kandev marketplace. The questions below settle the detailed boundaries of the first release.

---

## Q1. Issue features in the first release (select all that apply)

Besides viewing issues and linking issues to tasks (already required), what else does the first release need?

- A. Create a Kandev task from a Backlog issue (copy title and description into the task)
- B. Update Backlog issue status from Kandev (for example, move to "In Progress" when a task starts)
- C. Read and write issue comments from Kandev
- D. Create new Backlog issues from Kandev
- E. Type `#` in Kandev to insert a reference to a Backlog issue
- X. Other (please specify)

[Answer]: A, B, E

## Q2. Git and pull request features in the first release (select all that apply)

The Bitbucket plugin has the features below. What does the first release of the Backlog plugin need (besides linking pull requests to tasks, already required)?

- A. Backlog as a Kandev repository source (pick a Backlog repository when creating a task, Kandev fetches the code)
- B. Create a pull request on Backlog after Kandev pushes the task branch
- C. Show pull request status (open, merged, reviewers) on the task list
- D. Review panel in Kandev (view pull request comments, reply)
- E. Automatic pull request watching (PR watch) and saved dashboard queries
- X. Other (please specify)

[Answer]: A, B, C, E

## Q3. Status sync

When a Kandev task and a Backlog issue are linked, how should status be synced?

- A. No automatic sync; users change status themselves (the plugin only displays it)
- B. One-way: changes in Backlog show up in Kandev
- C. Two-way: changes on either side update the other
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q4. Number of Backlog spaces

Does one Kandev workspace need to connect several Backlog spaces at the same time?

- A. One space per Kandev workspace (like the Bitbucket plugin with one account)
- B. Several spaces in the same workspace
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. What is out of scope (select all that apply)

Backlog has many other features. Which ones are definitely NOT in this release?

- A. Backlog Wiki and file sharing
- B. Subversion (Git only)
- C. Gantt charts, burndown, milestones
- D. Backlog notifications in Kandev
- E. Real-time updates via webhook (use periodic polling instead)
- X. Other (please specify)

[Answer]: A, B, C, D (E: không loại webhook — vẫn trong phạm vi)

## Q6. Order of work

Which part should be done first?

- A. Risk first: build a working plugin skeleton (sign-in, space connection, packaging) to validate the Go part and OAuth early
- B. Value first: build the issue-to-task linking flow first, Git/PR later
- C. By dependency: connection and repository first, then pull requests, then issues
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q7. Clarification: updating issue status from Kandev

In Q1 you chose "Update Backlog issue status from Kandev" (B), and in Q3 you chose one-way sync from Backlog to Kandev (B). These two need to match: how should the Kandev-to-Backlog direction work?

- A. The user actively changes the issue status from Kandev (by clicking); not automatic. The Backlog-to-Kandev direction is automatic
- B. Drop issue status updates from Kandev; only automatic one-way sync from Backlog to Kandev
- C. Switch to automatic two-way sync
- X. Other (please specify)

[Answer]: B

## Q8. Clarification: webhook

In Q5 you kept webhooks in scope. Backlog webhooks can only be sent to an address that the Backlog server can reach over the internet (per `feasibility-assessment`). Does your Kandev server have a public address, and how should the plugin handle this?

- A. Has a public address; build webhooks, with periodic polling when webhooks are not available
- B. No public address; still build webhooks as an optional feature for other users, with periodic polling as the default
- C. Move webhooks out of scope for this release; use periodic polling only
- X. Other (please specify)

[Answer]: C

## Consolidated Summary Confirmation

Summary of the answers (with the Q7 and Q8 clarifications merged in):

- Q1: Issue features in the first release: A. Create a Kandev task from a Backlog issue; E. Type `#` to insert an issue reference. Option B (update issue status from Kandev) is dropped per Q7
- Q2: Git/PR features in the first release: A. Backlog as a repository source; B. Create a pull request after Kandev pushes the branch; C. Pull request status on the task list; E. Automatic pull request watching and saved dashboard queries. No review panel (D)
- Q3: B. Automatic one-way status sync: changes in Backlog show up in Kandev
- Q4: A. One Backlog space per Kandev workspace
- Q5: Out of scope: A. Wiki, file sharing; B. Subversion; C. Gantt, burndown, milestone; D. Backlog notifications; and webhooks (per Q8)
- Q6: A. Risk first: build a working plugin skeleton (sign-in, space connection, packaging) first
- Q7: B. No issue status updates from Kandev; only automatic sync from Backlog to Kandev
- Q8: C. Drop webhooks from this release; use periodic polling only

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
