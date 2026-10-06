# Refined Mockups — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Things already settled, so not asked again:

- UI follows Kandev.
- Accessibility at WCAG 2.1 AA level.
- Has a phone version, checked at 320px width.
- Text follows Kandev's language.
- Has a Backlog entry under "Integrations", with Issues, PR watches and Dashboard.

The questions below settle the missing screens, and two open comments the reviewer raised at the story step.

---

## Q1. Link Pull Request Dialog

Linking a PR to a task (US5.2) has no screen yet. How does the user choose the PR?

- A. A dialog that searches open PRs by number or title, in the repositories of the selected projects, with status and branch
- B. Only enter a PR number or paste a PR link
- C. Both: a search field that accepts both PR numbers and links
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q2. Position of Issue Info in Task Detail

Issue info (assignee, priority, due date, status, comments, attachments) is shown directly from Backlog in the task detail (US3.2, US3.5). Where should it go?

- A. A collapsible "Backlog" block in the task detail sidebar
- B. A separate "Backlog" tab in the task detail
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Lists on Phones

On narrow screens, how should the multi-column issue, PR watch and dashboard lists be shown?

- A. Switch to vertically stacked cards; filters collapse into a "Filters" button
- B. Keep the table form, with horizontal scrolling
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Restore After "No Longer Connected"

On disconnect, space change or project deselection, links and PR watches switch to "no longer connected". If the user later reconnects to **the same old space** or reselects **the same old project**, what should the plugin do?

- A. Restore those links and PR watches automatically (PR watches return to the Paused state so the user turns them back on)
- B. Ask the user whether to restore, with the number of links and watches that will be restored
- C. Do not restore; the user must link again from scratch
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. Priority of Comments and Attachments

At the story step, you said every issue-related feature is Must. The story for viewing an issue's comments and attachments in the task (US3.5) is currently Should. Which level do you want?

- A. Must, like the other issue features
- B. Keep Should
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q6. Create PR Watch and Save Query Forms

How should the "New watch", "Edit watch" and "Save query" forms be shown?

- A. A dialog (modal) right on the list page
- B. A separate page
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: B. Link a PR by entering a PR number or pasting a PR link
- Q2: A. Issue info is shown in a collapsible "Backlog" block in the task detail sidebar
- Q3: A. On phones, lists switch to vertically stacked cards; filters collapse into a "Filters" button
- Q4: A. Reconnecting to the same old space or reselecting the same old project restores links and PR watches automatically (watches return to Paused)
- Q5: B. US3.5 (comments and attachments) stays at Should
- Q6: A. The New watch, Edit watch and Save query forms are dialogs on the list page

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
