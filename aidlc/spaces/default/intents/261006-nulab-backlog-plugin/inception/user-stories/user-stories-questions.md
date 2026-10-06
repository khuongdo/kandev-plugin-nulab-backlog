# User Stories — Plan and Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

## Plan

- Story format: "As [persona], I want [goal], so that [benefit]". Each story meets the INVEST criteria and has an ID `US{group}.{number}`.
- Acceptance criteria are written as Given/When/Then, with IDs `AC{group}.{number}.{order}`. Each story has at least one success case and one error case.
- Each story has a MoSCoW level taken from `intent-backlog`. Stories in the thin slice done first are marked separately.
- After you answer, I write a draft. The designer, the developer and the quality engineer review it independently, then I consolidate.

The last two questions (Q5, Q6) settle two points the requirements reviewer rated as important: FR1.1 and FR6.2 in `requirements.md`. These two points must be decided to write acceptance criteria.

---

## Q1. User groups (personas)

Which user groups should the stories be written for?

- A. Two groups: the workspace admin (installs the plugin, connects the space, chooses projects) and the task worker (browses issues, creates tasks, creates PRs)
- B. Three groups: as A, plus the plugin releaser and maintainer (packaging, release, marketplace)
- C. One shared group: Kandev users
- D. Not yet defined
- X. Other (please specify)

[Answer]: C

## Q2. How to split stories

How should the stories be grouped?

- A. By feature group, matching FR1–FR7 (connection, issues, linking, sync, PR, PR watch, release)
- B. By workflow: connect → pick issue → create task → work → create PR → follow up
- C. By user group
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Size of each story

How large should each story be?

- A. Small: each story takes 1–3 days (expected about 25–35 stories)
- B. Medium: each story maps to one capability (expected about 15–20 stories)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Stories for packaging and release

Should packaging, checks and release (FR7) be written as stories?

- A. Yes, write them as stories for the plugin releaser
- B. No, keep them only as non-functional requirements and constraints
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. Connecting a space a second time in the same workspace

When the workspace is already connected to a space and someone connects again (for example switching to another space or changing the API key), what should the plugin do? (FR1.1 in `requirements.md` currently states two mutually exclusive ways.)

- A. Allow replacement after confirmation. The old secret is deleted. If switching to another space, the old issue/PR links and PR watches are turned off and marked "no longer connected"
- B. Allow replacement after confirmation, but only for changing credentials within the same space; to change the space, you must disconnect first
- C. Reject; you must disconnect first and then connect again
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q6. A PR watch's first run

When a new PR watch is turned on, which pull requests count as "new" and get tasks created automatically? (FR6.2 in `requirements.md` does not define this yet; if old PRs count, the first run may create a very large number of tasks.)

- A. Only PRs created after the watch was turned on; existing PRs do not create tasks
- B. Open PRs matching the filter at the first run also create tasks, but at most 10 tasks per cycle
- C. The first run shows the list of existing PRs so the user picks which need tasks; after that, only new PRs
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q7. Clarification: persona for the release stories

In Q1 you chose one shared user group (Kandev users). In Q4 you chose to write stories for packaging and release, and the person doing this is the plugin maintainer, not a Kandev user. Who should the release stories be written for?

- A. Keep one shared persona for all feature stories, and add a separate "plugin maintainer" persona only for the release stories
- B. Write the release stories under the Kandev user persona too (for example "so that I can install a new version reliably")
- X. Other (please specify)

[Answer]: B

## Q8. Backlog Git credentials (found during feedback)

The developer checked the Backlog docs. Result: Backlog's Git over HTTPS only accepts a user name with the Backlog password, or a separate Git password if the account has two-factor authentication on. The API key and OAuth token cannot be used to fetch code. So for Kandev to fetch code and push branches from a Backlog repository (US5.1, Must level), the plugin needs one more kind of credential. Which way do you choose?

- A. Add a Must story: store the Git user name and password (encrypted like other secrets, shared by the whole workspace), used only for Git operations
- B. Use SSH: the user configures SSH keys on the Kandev server themselves; the plugin only provides the repository's SSH address
- C. Drop code fetching from the first release: the plugin only lists repositories and links PRs, and the user configures code fetching in Kandev themselves
- X. Other (please specify)

[Answer]: X. chức năng git là optional. liên quan đến issue là MUST, có case source code sẽ quản lí ở chỗ khác như github

## Q9. Extra information when creating a task from an issue (found during feedback)

A Kandev task has no dedicated fields for Backlog's assignee, due date or attachments. If this information is copied into the task description, users can edit it, so it is no longer "read-only" as settled in Q5 of the requirements step. Which way do you choose?

- A. Copy it into the task description as a copy taken at creation time; accept that users can edit it, and never sync it back to Backlog
- B. The task description only has the title, description and issue link; the rest is shown directly from Backlog in the plugin's task detail section (always fresh, read-only)
- X. Other (please specify)

[Answer]: B

## Q10. PR watches and deleted tasks

When a task created by a PR watch is deleted by the user, and that PR still matches the filter, does the watch re-create the task?

- A. No; each PR produces a task only once over the watch's lifetime
- B. Yes; if the PR still matches, re-create it in the next cycle
- X. Other (please specify)

[Answer]: A

## Q11. Unselecting a project in use

When a project is unselected on the settings page while it has issue/PR links or PR watches, what should the plugin do?

- A. Allow unselecting after confirmation; that project's links and PR watches are turned off and marked "no longer connected" (same as switching spaces in Q5)
- B. Do not allow unselecting while links or watches remain; they must be removed first
- X. Other (please specify)

[Answer]: A

## Q12. Clarification: the Git part is optional

In Q8 you answered: "chức năng git là optional. liên quan đến issue là MUST, có case source code sẽ quản lí ở chỗ khác như github". I understand this as: every story about Git and pull requests (US5, US6) goes from Must down to Should, while the issue stories stay Must. This reading changes the priorities settled in the scope step (IB-5 and IB-6 are currently Must). When the Git part is built, how does Kandev fetch code from a Backlog repository?

- A. Git/PR is Should. When built, store the Backlog Git user name and password (encrypted) to fetch code over HTTPS
- B. Git/PR is Should. When built, use SSH: the user configures keys on the Kandev machine themselves
- C. Git/PR is Should. The first release does not fetch code from Backlog; it only links PRs, creates PRs and watches PRs. Source code may live elsewhere, such as GitHub
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers (with the Q7 clarification merged in):

- Q1: C. One shared persona: a Kandev user who uses Backlog
- Q2: A. Group stories by feature group, matching FR1–FR7
- Q3: A. Small stories, 1–3 days each (about 25–35 stories)
- Q4: A. Write stories for packaging, checks and release
- Q5: A. Reconnecting may replace the connection after confirmation; the old secret is deleted; if switching to another space, the old issue/PR links and PR watches are turned off and marked "no longer connected"
- Q6: B. A PR watch's first run creates tasks for open PRs matching the filter too, but at most 10 tasks per cycle (this limit applies to every cycle)
- Q7: B. The release stories are also written under the Kandev user persona
- Q8 + Q12: The Git and pull request part (US5, US6) is optional, lowered to Should; the issue stories stay Must; source code may live elsewhere, such as GitHub. When the Git part is built, the plugin stores the Backlog Git user name and password (encrypted, shared by the workspace) to fetch code over HTTPS (A)
- Q9: B. The task description only has the title, description and issue link; assignee, priority, due date, attachments and comments are shown directly from Backlog in the task detail section (always fresh, read-only)
- Q10: A. A PR watch does not re-create a deleted task; each PR produces a task only once
- Q11: A. Unselecting a project in use is allowed after confirmation; that project's links and PR watches are turned off and marked "no longer connected"

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
