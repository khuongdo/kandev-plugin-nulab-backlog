# User Stories — Kandev Plugin for Nulab Backlog

Inputs:

- `requirements.md`: FR1–FR7 and NFR1–NFR11.
- `team-practices.md`: TDD, thin slice first, tests against a fake Backlog server.
- `personas.md`: one shared persona, Minh.
- The answers to Q1–Q12 in `user-stories-questions.md`.
- Feedback from the designer, the developer and the quality engineer, in the `contributions/` folder.

## Conventions

- **Priority**: Must, Should or Could.
  - Stories about **issues** are Must.
  - Stories about **Git and pull requests** are Should, because this part is optional and the source code may live elsewhere, such as GitHub (Q8, Q12).
  - All other levels follow `intent-backlog` (IB-…).
- **[Skeleton]**: the story belongs to the thin slice done first. The thin slice covers connecting with an API key, a minimal settings screen, packaging, package verification, installing on a self-hosted Kandev, and calling Backlog once (`team-practices`).
- **Verification type**:
  - No label: automated test against a fake Backlog server.
  - `[CI]`: checked in the pipeline.
  - `[manual]`: manual check; evidence must be recorded.
- **Time**: every AC with "after one cycle", "wait", "expired" is checked with a fake clock, not real waiting.
- **UI strings** in quotes are only illustrative, written in English like `wireframes.md`. The real text follows the Kandev language (NFR10). Tests compare by meaning or by translation key.
- **Bait secret**: every AC about masking secrets uses a fake API key, token or password with the prefix `TESTSECRET-`. Tests assert that this string does not appear in logs, errors, responses sent to the UI, or logged URLs.
- **Terms**:
  - *API key*: an access key the user creates themselves in Backlog.
  - *OAuth*: signing in through Backlog's page without pasting a key.
  - *Pull request (PR)*: a request to merge code on a Backlog Git repository.
  - *PR watch*: a PR filter that runs periodically and automatically creates tasks for matching PRs.
  - *Integrations*: the item in the Kandev sidebar where the plugin shows its pages.

---

## US1. Connect a Backlog space (FR1)

### US1.1 Connect with an API key — Must [Skeleton]
As Minh, I want to connect the Kandev workspace to a Backlog space with the space address and an API key, so that the plugin can read the team's Backlog data.
- **AC1.1.1** Given the workspace is not connected, When I enter `myteam.backlog.com` (the plugin adds `https://` itself) and a valid API key and click Connect, Then the settings page shows "Connected as <Backlog user name> @ myteam.backlog.com".
- **AC1.1.2** Given the fake server returns 401, When I click Connect, Then the error "invalid key" shows below the API key input with instructions for creating a key, and the secret store has no record.
- **AC1.1.3** Given the fake server returns 5xx or times out, When I click Connect, Then "could not reach Backlog" shows with Retry, it does not report a wrong key, and nothing is stored.
- **AC1.1.4** Given it is connected, When the UI reads the settings, Then the response does not contain the key (not even part of it), only a flag like `hasApiKey: true`, and the API key input on the page is empty.
- **AC1.1.5** Given Connect was clicked, When the plugin is waiting for Backlog, Then the button changes to "Connecting..." and is locked; clicking many times does not create two connections.
- **AC1.1.6** Given it is connected, When another user in the same workspace opens a Backlog page, Then that person can use the connection without connecting again (shared connection).
- **AC1.1.7** Given it is connected, When checking where it is stored, Then the API key is stored through Kandev's encrypted secret mechanism, not in plain settings, and is sent to Backlog in a header rather than a URL parameter.
- Depends on: US7.1. Source: FR1.1, FR1.3, NFR3.

### US1.2 Validate the space address — Must [Skeleton]
As Minh, I want the plugin to accept only valid Backlog space addresses, so that credentials are never sent to an unknown host.
- **AC1.2.1** Given I enter `myteam.backlog.jp`, `x.backlogtool.com` or `MyTeam.Backlog.com`, When I click Connect, Then the address is accepted. Upper case is converted to lower case.
- **AC1.2.2** Given one of the following addresses, When I click Connect, Then the plugin rejects it, states the correct format, and the fake server receives 0 requests:
  - `http://myteam.backlog.com`
  - `evil.example.com`
  - `192.168.1.10`
  - `myteam.backlog.com.evil.io`
  - `https://myteam.backlog.com/path`
  - `myteam.backlog.com:8443`
  - `https://a@evil.io`
  - `backlog.com` (missing the space name)
  - `myteam.backlog.com.` (trailing dot)
- **AC1.2.3** Given the fake server returns a 302 redirect to a host not on the allow list, When the plugin calls, Then the plugin does not follow the redirect and does not send secrets to that host.
- Source: FR1.2.

### US1.3 Connect with OAuth — Must
As Minh, I want to sign in to Backlog through OAuth instead of pasting an API key, so that I do not have to create and manage a key myself.
- **AC1.3.1** Given the workspace is not connected, When I choose OAuth, click "Sign in with Nulab" and grant access on the Backlog page, Then I return to the settings page with the status "Connected as …".
- **AC1.3.2** Given I click cancel on the OAuth page, When I return, Then the settings page shows "Sign-in was cancelled" and nothing is stored.
- **AC1.3.3** Given the return request has a `state` that does not match, was already used or has expired, or is missing `code`, When the plugin receives the request, Then the fake server receives no code-for-token exchange request, nothing is stored, and the settings page shows an error.
- Depends on: US1.1. There is one manual task first: register an OAuth application with Nulab to get a client id, secret and callback address (assumption A2). Source: FR1.4.

### US1.4 Refresh the OAuth sign-in automatically — Must
As Minh, I want the OAuth connection to keep itself alive, so that I do not have to sign in again every hour.
- **AC1.4.1** Given the token has less time left than the refresh threshold (set in functional design), When the plugin calls Backlog, Then the token is refreshed before the call.
- **AC1.4.2** Given the token has expired and the refresh token is still valid, When 5 calls run concurrently (tested with `-race`), Then the fake server receives exactly 1 refresh request, all 5 calls succeed, and the new refresh token (if Backlog issues one) is stored in place of the old one.
- **AC1.4.3** Given the refresh token is no longer valid, When the plugin calls Backlog, Then the plugin tries to refresh only once. The settings page switches to "Sign in again" with a sign-in-again button right on the page. Other Backlog pages show a link to the settings page. Later polling cycles do not call Backlog until the user signs in again.
- Depends on: US1.3. Source: FR1.5.

### US1.5 Re-check and disconnect — Must
As Minh, I want to re-check or disconnect the Backlog connection, so that I can handle a revoked key or stop using it.
- **AC1.5.1** Given the fake server returns 200, When I click Test connection, Then during the check the button shows "Testing..." and is locked. When the check finishes, success shows with the user name, and the result is announced to the screen reader.
- **AC1.5.2** Given the fake server returns 401, When I click Test connection, Then "key is invalid or has been revoked" shows.
- **AC1.5.3** Given it is connected, When I click Disconnect, Then the confirmation dialog states clearly that the credentials will be deleted and **everyone in the workspace** will lose the Backlog connection. The default focus is on the Cancel button.
- **AC1.5.4** Given disconnecting was confirmed, Then the secret store no longer has the API key, access token, refresh token or Git password. All links and PR watches switch to the "no longer connected" state. The next polling cycle sends no request to Backlog.
- Source: FR1.6.

### US1.6 Replace credentials in the same space — Must
As Minh, I want to replace the API key or sign in again in the same space, so that I can handle a revoked old key.
- **AC1.6.1** Given space A is connected, When I reconnect space A (same host after lower-casing) with a new key and confirm, Then the old secret is deleted, the new secret is stored, and the next polling cycle updates the existing links with the new key. The fake server only sees the new key.
- **AC1.6.2** Given space A is connected, When I reconnect with a new key but the fake server returns 401, Then the old connection stays as it is and the old secret is not deleted.
- **AC1.6.3** Given the confirmation dialog is open, When I click Cancel, Then the old connection stays as it is.
- Depends on: US1.1. Source: FR1.1, Q5.

### US1.7 Choose projects — Must
As Minh, I want to choose the Backlog projects the plugin uses, so that the issue and repository lists only contain the team's work.
- **AC1.7.1** Given it is connected, When I select project PROJ and save, Then the issue list, repositories and PR watches only take data from PROJ.
- **AC1.7.2** Given no project is selected, When I open the Issues page, Then "No project selected" shows with a link to the settings page.
- **AC1.7.3** Given the account cannot view any project, When I open the project picker, Then "No projects available" shows.
- Depends on: US1.1. Source: FR1.7.

### US1.8 Switch to another space — Must
As Minh, I want to move the workspace to another Backlog space, so that I can use the plugin when the team changes spaces.
- **AC1.8.1** Given space A is connected with 5 links and 2 PR watches, When I connect to space B, Then the confirmation dialog states clearly "5 issue/PR links and 2 PR watches will be turned off".
- **AC1.8.2** Given it was confirmed, Then:
  - The secrets of space A are deleted, including the Git password.
  - The list of selected projects is deleted.
  - All links and PR watches of space A switch to the "no longer connected" state. This state is different from Paused, cannot be resumed with Resume, and shows on the task card as well as the watch row.
  - The fake server for space A receives 0 requests after the switch.
- **AC1.8.3** Given results from space A arrive late after the switch to B, When the plugin receives those results, Then they are ignored and do not overwrite space B's data.
- Depends on: US1.6. Source: FR1.1, Q5.

### US1.9 Unselect a project in use — Must
As Minh, I want to unselect a project without breaking data, so that I can narrow the plugin's scope when needed.
- **AC1.9.1** Given PROJ has 3 links and 1 PR watch, When I unselect PROJ, Then the confirmation dialog states clearly how many links and watches will be turned off.
- **AC1.9.2** Given it was confirmed, Then the links and watches of PROJ switch to "no longer connected", and other projects are not affected.
- Depends on: US1.7. Source: FR1.7, Q11.

## US2. Browse and find issues (FR2)

### US2.1 View the issue list — Must
As Minh, I want to see the issue list of the selected projects right in Kandev, so that I can pick work to give to an agent.
- **AC2.1.1** Given a selected project has 57 issues, When I open Integrations > Backlog > Issues, Then the 20 most recently updated issues show (key, title, status, assignee, updated time), with "Showing 1-20 of 57" and a next page button.
- **AC2.1.2** Given I am on page 3, Then "Showing 41-57 of 57" shows and the next page button is locked. When I click Next to a new page, focus moves to the top of the list and the "Showing …" line is announced to the screen reader.
- **AC2.1.3** Given no issue matches the filters, When I open the page, Then "No issues match these filters" shows with a Reset filters button.
- **AC2.1.4** Given Backlog returns an error, When the list loads, Then "Couldn't load issues from Backlog" shows with Retry. The reason is one of three fixed types: needs reconnecting, being rate limited, cannot reach Backlog. The message does not contain the Backlog response body.
- **AC2.1.5** Given it is loading, Then placeholder rows show, the filters stay visible, and the screen reader is told "loading". Given it is not connected, Then "Backlog is not connected" shows with a link to the settings page.
- **AC2.1.6** Given an issue title is 200 characters long or in Japanese, When I view the list, Then the title wraps or is shortened with "…" without breaking the layout, and the full title is still readable.
- Depends on: US1.7. Source: FR2.1, NFR5.

### US2.2 Filter and search issues — Must
As Minh, I want to filter by project, status, assignee, and search by keyword or issue key, so that I can quickly find the right issue.
- **AC2.2.1** Given the fake server has sample issues, When I filter by project, status, assignee in turn, and by all three combined, Then every result row satisfies all active filters (table-driven test).
- **AC2.2.2** Given PROJ-123 exists, When I type `PROJ-123`, Then PROJ-123 is at the top of the results, even when the keyword search does not return it.
- **AC2.2.3** Given I type an issue key that does not exist, When I search, Then the empty state of AC2.1.3 shows, not the error state.
- **AC2.2.4** Given the keyword contains Japanese or Vietnamese characters, `&`, `%` or spaces, When I search, Then the keyword is encoded correctly in the request.
- Depends on: US2.1. Source: FR2.2.

### US2.3 See which tasks an issue is linked to — Must
As Minh, I want to see right in the list which issues already have tasks, so that I do not assign the same work twice.
- **AC2.3.1** Given PROJ-123 is linked to T-12 and T-15, When I view the list, Then the PROJ-123 row shows "T-12, T-15" (each key is a link to the task) and still has the "Create task" and "Link to task" actions.
- **AC2.3.2** Given T-15 is deleted in Kandev, When I view the list, Then the PROJ-123 row only shows T-12.
- Depends on: US3.1. Source: FR2.3, FR3.4.

## US3. Link issues to tasks (FR3)

### US3.1 Create a task from an issue — Must
As Minh, I want to create a Kandev task from an issue with one click, so that I do not have to copy by hand.
- **AC3.1.1** Given issue PROJ-120, When I click "Create task", Then a new task is created without an intermediate dialog. The task has:
  - title and description taken from the issue;
  - a link to the issue;
  - the label PROJ-120;
  - priority mapped from Backlog: High → high, Normal → medium, Low → low.
- **AC3.1.2** Given "Create task" was clicked, When the task is being created, Then the button shows "Creating..." and is locked; clicking twice quickly does not create two tasks. When done, "Created T-18" shows with a link, and the issue row updates with the task key.
- **AC3.1.3** Given PROJ-120 already has task T-17, When I click "Create task", Then a warning dialog offers three choices "Open T-17", "Create another task" and "Cancel"; the default focus is not on "Create another task". Choosing "Create another task" creates a second task.
- **AC3.1.4** Given Backlog returns an error when reading the issue, or Kandev refuses to create the task, When I click "Create task", Then no task is created, no orphan link remains, and an error message shows.
- Depends on: US2.1. Source: FR3.1, FR3.4.

### US3.2 View issue information in the task — Must
As Minh, I want to see the issue's assignee, priority and due date right in the task detail, taken directly from Backlog, so that the agent and I always have the latest information.
- **AC3.2.1** Given a task is linked to PROJ-120, When I open the task detail, Then the Backlog section shows the assignee, priority, due date and status of PROJ-120, taken directly from Backlog, read-only, with "updated at …".
- **AC3.2.2** Given the due date of PROJ-120 changes on Backlog, When I reopen the task detail or refresh, Then the new value shows. The task description does not change.
- **AC3.2.3** Given every plugin action is run in tests (create task, polling cycle, refresh, view detail), Then the fake server receives 0 write requests to the issue: no PATCH/DELETE to the issue and no comment POST.
- Depends on: US3.1. Source: FR3.2, Q9.

### US3.3 Link and unlink an issue with an existing task — Must
As Minh, I want to attach an issue to an existing task, or detach it, so that I can connect ongoing work to the matching issue.
- **AC3.3.1** Given T-17 has no issue, When I choose T-17 in the "Link to task" dialog of PROJ-120, Then the T-17 card shows the PROJ-120 label and the PROJ-120 row shows T-17.
- **AC3.3.2** Given PROJ-120 is already linked to T-17, When I also link T-18, Then it is allowed, and the issue row shows both.
- **AC3.3.3** Given T-19 is already linked to PROJ-118, When I open the "Link to task" dialog of PROJ-120, Then T-19 shows with the text "Linked to PROJ-118" and cannot be selected. If a link request is still sent, the plugin rejects it.
- **AC3.3.4** Given the search box in the dialog matches no task, Then "No tasks found" shows and the Link button is locked. Pressing Esc or Cancel closes the dialog, nothing changes, and focus returns to the "Link to task" button of that same row.
- **AC3.3.5** Given T-17 is linked to PROJ-120, When I choose "Unlink Backlog issue", Then the label on the card and the task key on the issue row both disappear.
- Depends on: US2.1. Source: FR3.3, FR3.4.

### US3.4 Reference an issue with `#` — Must
As Minh, I want to type `#` in a task's composer to insert a reference to a Backlog issue, so that the agent knows the related issue.
- **AC3.4.1** Given it is connected, When I type `#PROJ-12`, Then at most N suggestions show (set in functional design) whose key starts with the typed string or whose title contains the keyword. Picking a suggestion inserts a reference to that issue. Selecting with arrow keys and Enter is handled by Kandev.
- **AC3.4.2** Given I type many characters quickly, Then the plugin sends only one search request after a short wait.
- **AC3.4.3** Given it is not connected or no issue matches, When I type `#PROJ-99999`, Then no Backlog suggestions show and there is no error.
- Depends on: US1.1, US1.7. Source: FR3.5.

### US3.5 View issue comments and attachments in the task — Should
As Minh, I want to see the issue's comments and attachments right in the task detail, so that I have full context without opening Backlog.
- **AC3.5.1** Given PROJ-120 has 150 comments and 2 attachments, When I open the Backlog section in the task detail, Then the newest comments show first, with a load more button (fetched by page, at most 100 per page), and the file list with names and sizes. Everything is read-only.
- **AC3.5.2** Given a file exceeds the size limit N (set in functional design), When I view it, Then that file shows its name and a link to Backlog instead of a preview, with the reason.
- **AC3.5.3** Given Backlog returns an error when fetching comments, Then the comments section shows an error with Retry, and the other parts of the task detail still show.
- Depends on: US3.2. Source: FR3.2, Q9.

## US4. Issue status sync (FR4)

The **pull request** status on the task card is refreshed by Kandev itself (US5.4). The plugin does not run its own poller for PRs, following Kandev's plugin writing guide. The plugin's polling cycle only applies to **issue status** and to **PR watches**.

### US4.1 Update issue status on a cycle — Must
As Minh, I want the status of linked issues to update automatically from Backlog, so that I know the progress without opening Backlog.
- **AC4.1.1** Given T-17 is linked to PROJ-120, which is Open, When on the fake server PROJ-120 changes to Resolved and the fake clock advances exactly one cycle, Then the label on T-17 shows Resolved.
- **AC4.1.2** Given PROJ-120 returns 404 or 403 (table-driven test), When the cycle runs, Then the label shows "Issue unavailable", the link is kept, and other links in the same cycle are still updated.
- **AC4.1.3** Given I change the task status in Kandev, When the cycle runs, Then no request updates the issue status on Backlog.
- **AC4.1.4** Given an issue label on the task card, When keyboard focus moves onto it or I tap the label, Then "updated at …" shows. Given the last few cycles all failed, Then the label still shows the last known status, with text saying the data may be stale.
- **AC4.1.5** Given the plugin restarts, Then polling continues with the configured interval.
- Depends on: US3.3, US8.4. Source: FR4.1, FR4.3, FR4.4.

### US4.2 Configure the interval and refresh manually — Must
As Minh, I want to adjust the update interval and click refresh when needed, so that I can balance data freshness against the API rate limits.
- **AC4.2.1** Given nothing is configured, Then the interval is 5 minutes. Given it is set to 2 minutes, Then from the next cycle on, polls are 2 minutes apart.
- **AC4.2.2** Given I enter 30 seconds, 0, a negative number, letters or leave it empty (table-driven test), Then the plugin rejects it and states the minimum is 1 minute.
- **AC4.2.3** Given I click refresh on the Issues page, Then the data reloads immediately, the button is locked while refreshing, and the "updated at …" line changes accordingly.
- **AC4.2.4** Given a cycle is running, When I click refresh, Then no two polls run concurrently (`-race` test, the fake server counts parallel requests).
- Depends on: US4.1. Source: FR4.2, FR4.4.

## US5. Repositories and pull requests (FR5) — optional

Per Q8 and Q12, this whole group is Should. The source code may live elsewhere, such as GitHub. When this part is built, fetching code over HTTPS uses Backlog's separate Git user name and password. Reason: Backlog's Git does not accept an API key or OAuth token.

### US5.1 Choose a Backlog repository when creating a task — Should
As Minh, I want to choose a Backlog Git repository when creating a task, so that the agent works on the right source code when the code lives on Backlog.
- **AC5.1.1** Given project PROJ has repository `web-app`, When I create a task, Then the repository list has "Backlog: myteam / web-app" and a base branch can be chosen.
- **AC5.1.2** Given repositories cannot be loaded from Backlog, When I open the Repository field, Then other repository sources can still be chosen, and the Backlog entry shows an error with Retry.
- Depends on: US1.7. Source: FR5.1.

### US5.2 Link and unlink an existing pull request — Should
As Minh, I want to attach an existing Backlog pull request to a task, so that I can follow that PR from Kandev.
- **AC5.2.1** Given repository `web-app` has PR #42, When I choose "Link Backlog pull request" on T-17 and choose PR #42, Then the T-17 card shows the label "PR #42".
- **AC5.2.2** Given I enter #999, which does not exist, Then "Pull request #999 not found" shows.
- **AC5.2.3** Given T-17 is linked to PR #42, When I choose "Unlink", Then the label disappears.
- Depends on: US5.1. The PR linking screen will be added in Refined Mockups. Source: FR5.2.

### US5.3 Create a pull request from a task — Should
As Minh, I want to create a pull request on Backlog from a pushed Kandev branch, so that I do not have to open Backlog.
- **AC5.3.1** Given branch `task/export-csv` has been pushed and T-17 is linked to PROJ-120, When I open the create PR form, Then the form is prefilled with:
  - the title taken from the task;
  - a description with a reference to PROJ-120, but no issue auto-close keyword;
  - the source branch;
  - the default target branch.
  If the title is left empty, an error shows right below the field.
- **AC5.3.2** Given the form is valid, When I click Create PR, Then the fake server receives a create PR request with the correct source branch, target branch and related issue. T-17 is linked to the returned PR, and a link to open the PR on Backlog shows.
- **AC5.3.3** Given the task's branch has not been pushed, Then the plugin is not called to create a PR. Locking the action is handled by Kandev.
- **AC5.3.4** Given there is already an open PR for the same branch, When I click Create PR, Then the plugin does not create a second PR and offers to link the existing PR.
- Depends on: US5.1, US5.5. Source: FR5.3.

### US5.4 See the pull request status on the task card — Should
As Minh, I want to see the PR status and assignee right on the task card, on both desktop and phone, so that I can grasp progress quickly.
- **AC5.4.1** Given T-17 is linked to PR #42, When Kandev refreshes the card, Then the label shows the correct status (Open / Merged / Closed, table-driven test) and the assignee, for example "PR #42 Open – assignee: Lan". Backlog has no PR reviewers or approvals, so this information is not shown.
- **AC5.4.2** `[manual]` Given a 320px wide screen, When I view the task board, Then the labels stack, are readable, are not cut off, and buttons have a touch target of at least 44x44px.
- **AC5.4.3** Given the status cannot be fetched, Then the label shows "Status unknown". The status always has text with it, not only color. The screen reader reads the full "Pull request #42, Open, assignee Lan".
- Depends on: US5.2. Source: FR5.4, FR4.1 (PR part), NFR9.

### US5.5 Store Git credentials — Should
As Minh, I want to store the Backlog Git user name and password, so that Kandev can fetch code and push branches when the code lives on Backlog.
- **AC5.5.1** Given it is connected, When I enter the Git user name and password (the Backlog password, or a separate Git password if the account has two-factor authentication on) and save, Then they are stored through Kandev's encrypted secret mechanism and are not shown again.
- **AC5.5.2** Given the password is wrong, When I run Test connection, Then a Git authentication error shows, and the password does not appear in logs or messages.
- Depends on: US1.1. Source: FR5.1, Q12. New requirement found during feedback: FR5.1 needs Git credentials.

### US5.6 Fetch code and push branches with Git credentials — Should
As Minh, I want Kandev to fetch code from a Backlog repository and push the task's branch, so that the agent can work on that code.
- **AC5.6.1** Given Git credentials are stored and `web-app` is chosen, When Kandev asks for credentials, Then the plugin returns the correct HTTPS URL of `web-app` together with the credentials.
- **AC5.6.2** `[manual]` Given the setup in AC5.6.1, When a real task runs, Then Kandev fetches the code and pushes the branch.
- **AC5.6.3** Given the Git credentials are no longer valid, Then the task reports an authentication error with instructions for updating them. The password is not in the stored remote URL, in command-line arguments, or in error messages.
- Depends on: US5.1, US5.5. Source: FR5.1, NFR3.

## US6. Pull request watches and dashboard (FR6) — optional

### US6.1 Manage PR watches — Should
As Minh, I want to create, edit, run, pause, resume and delete PR watches, so that the PRs I care about are followed automatically.
- **AC6.1.1** Given the PR watches page, When I create the watch "My open PRs" for `web-app` with the filter "assignee = me", Then the watch shows in the Active state. The filter only includes fields Backlog supports: status, assignee, issue, creator.
- **AC6.1.2** Given the New watch form, When I click Save with no name, or with a repository not in a selected project, Then an error shows right below the field and no watch is created.
- **AC6.1.3** Given an Active watch, When I click Pause, Then the state changes to Paused, and over many cycles (using a fake clock) the watch sends no request until I click Resume.
- **AC6.1.4** Given a watch, When I edit the filter and save, Then from the next cycle on the watch uses the new filter, and tasks already created stay as they are. When I click Run, Then the watch runs one cycle immediately and the last run time changes accordingly.
- **AC6.1.5** Given a watch, When I click Delete, Then the confirmation dialog states clearly that the created tasks are **not** deleted. Given there is no watch yet, Then "No PR watches yet" shows with a New watch button.
- Depends on: US5.1. Source: FR6.1.

### US6.2 PR watch creates tasks automatically — Should
As Minh, I want each PR that matches a watch to produce a Kandev task automatically, so that an agent reviews it without me creating it by hand.
- **AC6.2.1** Given an Active watch and 3 open PRs matching the filter, When the first cycle runs, Then exactly 3 tasks are created, each linked to one PR.
- **AC6.2.2** Given 25 open PRs match the filter, When the cycles run, Then the cycles create 10, 10, then 5 tasks in turn, oldest PR first. The watch row shows progress, for example "Created 10/25 tasks, the rest in later cycles".
- **AC6.2.3** Given a PR already has a task created by the watch, even if that task has been deleted (Q10), When later cycles run, Then no task is re-created for that PR.
- **AC6.2.4** Given the plugin restarts, or "Run" and the scheduled cycle run at the same time (`-race` test), Then each PR still has exactly one task. Deduplication is based on a durably stored key (connection, repository, PR number).
- **AC6.2.5** Given task creation fails at the 5th PR out of 10, When the next cycle runs, Then the plugin continues from the 5th PR, and PRs 1–4 are not re-created.
- Depends on: US6.1, US8.4. Source: FR6.2, Q6, Q10.

### US6.3 Save queries and view the dashboard — Could
As Minh, I want to save pull request queries and view the results on a dashboard, so that I can quickly see PRs by a fixed criterion.
- **AC6.3.1** Given the query "Open PRs in web-app" is saved, When I choose that query on the Dashboard, Then the list of matching PRs shows with status and linked tasks.
- **AC6.3.2** Given the query has no results, Then an empty state shows with a hint to edit the query. Given Backlog errors or is rate limiting, Then an error shows with Retry.
- Depends on: US5.2. Source: FR6.3.

## US7. Packaging, checks and release (FR7)

### US7.1 Plugin skeleton installs into Kandev — Must [Skeleton]
As Minh, I want the plugin packaged in the correct format and installable on a self-hosted Kandev server, so that I can start using the plugin.
- **AC7.1.1** `[CI]` Given the source code on `main`, When the packaging and package verification step runs, Then the package is created, has all 5 executables per the Bitbucket template's platform list, and verification succeeds.
- **AC7.1.2** `[manual]` Given the newly built package, When I (with Kandev admin rights) install it through Settings > Plugins, Then the plugin runs, and the Backlog item with the minimal settings screen (W1) appears in Integrations.
- **AC7.1.3** `[CI]` Given the package is modified after packaging (wrong checksum), When the verification step runs, Then it fails.
- Source: FR7.1, NFR7.

### US7.2 First Backlog call from a real server — Must [Skeleton]
As Minh, I want the freshly installed plugin to call the team's real Backlog space, so that the whole chain is proven to work end to end.
- **AC7.2.1** `[manual]` Given the plugin is installed and connected with an API key, When I check manually against a real space, Then the plugin fetches at least one piece of data from Backlog (for example the user name). The result is recorded in a manual check record file, with the date, Kandev version, plugin commit, space domain (no key), the steps and the result.
- Depends on: US1.1, US7.1. Source: `team-practices` (manual check no. 1).

### US7.3 Quality gates in CI — Must
As Minh, I want every pull request of the plugin to pass automated checks, so that the plugin build I install carries no known bugs and does not leak the API key.
- **AC7.3.1** `[CI]` Given a pull request, When CI runs, Then the following steps all run:
  - formatting;
  - static checks;
  - lint;
  - tests with `-race`;
  - coverage on `./internal/...` and `./server/...`;
  - the `go mod tidy` check;
  - packaging and package verification.

  The merge is blocked if any step fails. For example, Go coverage of 79% is blocked.
- **AC7.3.2** Given the fake server returns 401 or 500 with a response body containing the bait key, When the error flows run (connect, list, token refresh, Git), Then logs, error messages and responses sent to the UI do not contain the bait key.
- **AC7.3.3** Given a network error when calling Backlog, When the error is logged or returned, Then the message does not contain a URL with a secret.
- **AC7.3.4** `[CI]` Given test data and test artifacts, When CI runs the check step, Then there is no string that looks like a real API key or password (without the `TESTSECRET-` prefix). If there is, the step fails.
- Depends on: US7.1. Source: NFR3, NFR4, NFR8.

### US7.4 Test on the minimum Kandev version — Should
As Minh, I want the plugin tested automatically on the declared minimum Kandev version, so that I can install the plugin on the team's Kandev without upgrading Kandev.
- **AC7.4.1** `[CI]` Given the built package, When CI runs the contract test, Then the job reads `min_kandev_version` from `manifest.yaml` (not hard-coded), checks out exactly that version, installs the package and runs a trial successfully.
- **AC7.4.2** `[CI]` Given installing or the trial run on the minimum version fails, Then CI fails.
- Depends on: US7.1. Source: NFR6.

### US7.5 Release a new version — Should
As Minh, I want to release by pushing a tag, so that there is a Release that can be installed and whose provenance can be verified.
- **AC7.5.1** `[CI]` Given `main` has passed CI, When the tag `v0.1.0` is pushed, Then the workflow reruns package verification, then creates a GitHub Release with the package, `checksums.txt` and the provenance attestation. The `gh attestation verify` command succeeds.
- **AC7.5.2** `[CI]` Given the tag already exists, or the tag is on a commit not on `main`, When the release workflow runs, Then the workflow refuses.
- **AC7.5.3** `[manual]` Given it is before the first release, Then the record file of the second manual check against a real space (same format as AC7.2.1) is already on `main` before the tag is pushed.
- Depends on: US7.3. Source: FR7.2, `team-practices`.

### US7.6 Submit the plugin to the marketplace — Should
As Minh, I want the plugin listed in the Kandev marketplace, so that other Kandev users can install it easily.
- **AC7.6.1** Given a GitHub Release exists, When a pull request adding the plugin to the marketplace catalogue is submitted, Then the pull request has all the fields the catalogue requires (checked at the time) and passes the catalogue's automated checks. Acceptance by the Kandev maintainers is outside the team's control.
- **AC7.6.2** Given the `id` in the catalogue differs from the `id` in the manifest, When the check step runs, Then it reports an error.
- Depends on: US7.5. Source: FR7.3.

## US8. Performance, reliability, UI

### US8.1 Lists show within 3 seconds — Must
As Minh, I want the issue and PR lists to show within 3 seconds for 95% of opens, so that I do not have to wait.
- **AC8.1.1** Given the fake server has a fixed 300 ms delay per request, When I open a 20-row list, Then the plugin side returns results within the time budget (set in functional design).
- **AC8.1.2** `[manual]` Given a real test space, When I open a 20-row list 20 times, Then at least 19 times it shows within 3 seconds.
- **AC8.1.3** Given the fake server does not respond, When any kind of call is made (connect, list, polling cycle, create PR), Then the call returns a timeout error within the limit (default 10 seconds, set in functional design), and other requests to the plugin during that time are still served.
- Source: NFR1, NFR5.

### US8.2 Usable with keyboard and screen reader — Must
As Minh, I want to operate every Backlog screen with a keyboard and a screen reader, so that I can use it comfortably.
- **AC8.2.1** `[manual]` Given every Backlog screen and dialog (W1–W10 and the extra dialogs from Refined Mockups), When only Tab, Enter, Esc and arrow keys are used, Then every action can be done. The focus ring is clearly visible, and Tab order follows reading order.
- **AC8.2.2** Given the plugin's UI components, When an automated scanner runs (for example axe), Then there are no level A or AA violations.
- **AC8.2.3** Given a dialog is open, Then focus stays inside the dialog, Esc closes the dialog, and focus returns to the button that opened it. Given a background action finishes, Then the result is announced through a polite live region (`aria-live="polite"`) exactly once.
- **AC8.2.4** Given a colored status label, Then the text meets a contrast of at least 4.5:1 and the label border 3:1. A locked menu item still receives focus so its reason can be read.
- Source: NFR9.

### US8.3 Logs for diagnosing errors — Could
As Minh, I want the plugin to write structured logs, so that I can find the cause when there is a problem.
- **AC8.3.1** Given the fake server returns 500, When the plugin calls, Then there is one log record with the action, error code and time, containing no secrets.
- **AC8.3.2** Given each polling cycle, Then there is one log record with the number of items updated, the number of errors, the run time, and the number of waits due to API limits.
- Source: NFR11.

### US8.4 Respect Backlog's API rate limits — Must
As Minh, I want the plugin not to exceed Backlog's API rate limits, so that the shared account is not blocked and the team's other tools keep working.
- **AC8.4.1** Given the fake server returns 429, a table-driven test with the following cases:
  - has `X-RateLimit-Reset` (UTC epoch seconds): the plugin waits until that time;
  - only has `Retry-After`: the plugin waits by this value;
  - has neither header: the plugin waits 60 seconds.

  Then the plugin waits the correct time for each case and then retries.
- **AC8.4.2** Given 429 repeats N times in a row (set in functional design), Then the plugin stops retrying and reports an error. Given the `context` is cancelled while waiting, Then the plugin returns `context.Canceled` and does not call again.
- **AC8.4.3** Given the polling cycle, PR watches and manual refresh run at the same time (`-race` test), Then requests in the update group and the search group never run in parallel, and each request in these two groups is at least 1 second after the previous one.
- **AC8.4.4** Given it is waiting due to a limit, Then the UI shows "Backlog is limiting requests. Retrying in N seconds" and announces it to the screen reader only once.
- Depends on: US1.1. Source: NFR2, C-T3.

### US8.5 Text follows the Kandev language — Should
As Minh, I want the plugin text to follow the language Kandev uses, so that the UI is consistent.
- **AC8.5.1** Given Kandev switches to a language the plugin has a translation for, When I open the Backlog pages, Then the text shows in that language.
- **AC8.5.2** Given Kandev is set to a language the plugin has no translation for yet, Then the text shows in English, without showing raw translation key names.
- Source: NFR10.

---

## Summary

| Group | Stories | Must | Should | Could |
|------|-------|------|--------|-------|
| US1 Connection | 9 | 9 | 0 | 0 |
| US2 Issues | 3 | 3 | 0 | 0 |
| US3 Linking | 5 | 4 | 1 | 0 |
| US4 Issue sync | 2 | 2 | 0 | 0 |
| US5 Repo/PR (optional) | 6 | 0 | 6 | 0 |
| US6 PR watch (optional) | 3 | 0 | 2 | 1 |
| US7 Release | 6 | 3 | 3 | 0 |
| US8 Quality | 5 | 3 | 1 | 1 |
| **Total** | **39** | **24** | **13** | **2** |

**Stories in the thin slice done first:** US1.1, US1.2, US7.1, US7.2.

**Changes from `requirements.md` and `intent-backlog`.** The points below have their sources stated. The Delivery Planning and Units Generation steps need to use this version:

- **Git/PR priority lowered.** IB-5 (repository source) and IB-6 (PR linking) go from Must to Should. The whole Git/PR group becomes optional. Source: Q8.
- **PR reviewers removed.** FR5.4 drops the "reviewers" part, because the Backlog API has no PR reviewers. Source: developer feedback, citing the API docs.
- **Changed how issue information is shown.** FR3.2 changes from "a copy in the task" to "shown directly from Backlog in the task detail". Source: Q9.
- **Git credentials added.** FR5.1 needs an additional Git user name and password (US5.5), because Backlog's Git does not accept an API key or token. Source: developer feedback; Q12.
- **Polling scope changed.** FR4.1/FR4.2: the plugin's polling cycle only applies to issues and PR watches, while PR status is refreshed by Kandev. Source: Kandev's plugin writing guide.
- **API limit header corrected.** NFR2 uses `X-RateLimit-Reset` instead of `Retry-After`. Source: Backlog's API rate limit docs.
- **Executable count corrected.** NFR7 has 5 executables per the Bitbucket template, not 6 combinations.

**INVEST notes:**

- Every story has its own value, with at least one success case and one error or edge case.
- Stories are small enough: the stories that were too large have been split, including US1.6 → US1.6/US1.8, US3.2 → US3.2/US3.5, US5.1 → US5.1/US5.5/US5.6, US8.2 → US8.2/US8.5. The API limit part was moved into its own story, US8.4.
- Stories that depend on other stories can still be tested independently thanks to the fake Backlog server.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q12]: answers in `user-stories-questions.md`.
- FR…, NFR…: `inception/requirements-analysis/requirements.md`. IB-…: `ideation/scope-definition/intent-backlog.md`. C-…: `ideation/feasibility/constraint-register.md`.
- Feedback: `contributions/aidlc-design-agent.md`, `contributions/aidlc-developer-agent.md`, `contributions/aidlc-quality-agent.md`.
- Backlog API v2 docs (pull requests, API limits, Git): https://developer.nulab.com/docs/backlog/

## Assumptions & Open Questions

- [assumption] There are 5 executables per the Bitbucket template's platform list, probably without windows/arm64. To be checked when building the skeleton.
- [assumption] Backlog accepts the API key in a header (release notes 08/2026). If it does not, AC1.1.7 changes to using a URL parameter, and AC7.3.3 becomes a mandatory criterion.
- [assumption] Backlog may not reject a second PR for the same branch by itself, so AC5.3.4 has the plugin check on its own. To be tried on the test space.
- The functional design step needs to set these values: the token refresh threshold (AC1.4.1), the number of `#` suggestions (AC3.4.1), the file size limit (AC3.5.2), the maximum wait time (AC8.1.3), the number of retries on 429 (AC8.4.2), and the plugin-side time budget (AC8.1.1).
