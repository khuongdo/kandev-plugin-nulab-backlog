**Collaborator:** aidlc-developer-agent

## Contribution

Viewpoint: whether each story can be built, and whether it fits 1–3 days (Q3). I checked the draft against Backlog API v2 and against the Kandev source code next to the repo (`../kandev`: `docs/public/plugins-authoring.md`, `apps/backend/pkg/pluginsdk/`, `apps/backend/internal/plugins/manifest/`).

### 1. API feasibility issues (blocking, or the AC must change)

**F1. Backlog has no pull request reviewers or approvals. AC5.4.1 cannot be built.**
The Backlog pull request object only has `status` (Open/Closed/Merged), `assignee`, `issue`, commits and timestamps. There is no reviewer, approval or review status field (https://developer.nulab.com/docs/backlog/api/2/get-pull-request/, https://developer.nulab.com/docs/backlog/api/2/get-pull-request-list/). So "PR #42 Open – 1/2 approved" has no data to show. Proposal:
- Change AC5.4.1 to "PR #42 Open – assignee: <name>".
- Drop "reviewers" from the US5.4 story sentence. FR5.4 also needs to change accordingly; I note it for the requirements team to fix.
- Leave the `review` field in `taskStatus` of `ReviewItemSummary` empty.

**F2. Git clone and push over HTTPS cannot use an API key or OAuth token. US5.1 is missing part of the credentials.**
Backlog's Git HTTPS accepts a user name with the Backlog password. If the account has two-factor authentication on, a separate Git(https) password must be used, created in the personal settings page (https://support.nulab.com/hc/en-us/articles/8749351512217-Git-overview, https://nulab.com/learn/software-development/git-tutorial/git-commands-settings/troubleshooting/). So AC5.1.2 ("Kandev can fetch the source code and push a branch") cannot be met with only the API key or OAuth connection of US1.1 and US1.3. Propose adding a Must story:
- **US5.0 Store Git credentials**: the user enters the Git(https) user name and password. The plugin stores them with `SetSecret` and returns them through Kandev's `ResolveGitCredential` / `GetGitCredentialBinding`.
- Error AC: a wrong password is reported when testing the connection, and the password does not appear in logs.
- Disconnecting (US1.5) and switching spaces (US1.6) must also delete this secret and change the binding.
- This is a new credential not yet in FR1, so its origin must be recorded per the Traceability rule.

**F3. PR status on the card is refreshed by Kandev, not through the plugin's polling cycle. AC4.1.1 and FR4.2 should only apply to issues.**
Kandev draws the PR status on the card and topbar itself from `registerReviewProvider`, and calls `refresh(taskId)` itself: once when the task row appears, when the user hovers, and every 90 seconds while the topbar is open. The docs say clearly "Do not register chat-top-bar/composer lookalikes or run another status poller in the plugin" (`../kandev/docs/public/plugins-authoring.md`, section "Provider, task, review, and reference registrations"). Proposal:
- US4.1 and US4.2 (default interval 5 minutes, minimum 1 minute) apply only to **issue status** and to **PR watches**.
- PR status belongs to US5.4, on Kandev's refresh rhythm.
- Rewrite AC4.1.1 for issues only. Add an AC to US5.4: "Given PR #42 is merged on Backlog, When Kandev refreshes the card, Then the label shows Merged".

**F4. Kandev has no fields for attachments, the Backlog assignee or the due date. AC3.2.1 must say where the information shows.**
`CreateTaskInput` (`apps/backend/pkg/pluginsdk/data_types.go:1123`) only has `Title`, `Description`, `Priority` (critical/high/medium/low), `Repositories`, `Launch`…, with no attachment or due date field. Labels can be set through `SetLabels`. So:
- Priority can be mapped: High → high, Normal → medium, Low → low.
- Assignee and due date can only be written as text in the description, or shown in a separate plugin panel (the `task-sidebar` slot).
- "Read-only" is not true if copied into the description, because users can edit the description. One of two must be chosen: (a) show in `task-sidebar` from Host state or `KANDEV_PLUGIN_DATA_DIR`, which is truly read-only; (b) copy into the description and drop the word "read-only" from the AC.
- Attachments must be downloaded through `GET /api/v2/issues/:issueIdOrKey/attachments/:attachmentId` (https://developer.nulab.com/docs/backlog/api/2/get-issue-attachment/), then stored in the plugin's data folder. A link to Backlog only opens for viewers who are signed in to Backlog.
- Comments are fetched through `GET /api/v2/issues/:issueIdOrKey/comments`, at most 100 per call, and must be paginated (https://developer.nulab.com/docs/backlog/api/2/get-comment-list/).
- Assumption A4 is basically correct on the Backlog side. The uncertain part is on the Kandev side.

**F5. Backlog reports limits with `X-RateLimit-Reset`, not `Retry-After`. AC7.2.2 and NFR2 need fixing.**
Backlog returns 429 with `X-RateLimit-Limit`, `X-RateLimit-Remaining` and `X-RateLimit-Reset` (UTC epoch seconds). The docs advise waiting one minute or waiting per `X-RateLimit-Reset`, calling sequentially, and pausing at least 1 second between Update/Search calls (https://developer.nulab.com/docs/backlog/rate-limit/). Proposal:
- The AC says "wait until `X-RateLimit-Reset`; if the header is missing, use `Retry-After`; if that is also missing, wait 60 seconds".
- The fake server in tests returns both cases.

**F6. Searching by issue key (AC2.2.2) needs separate handling.**
The `keyword` parameter of `GET /api/v2/issues` does not guarantee an exact issue key match (https://developer.nulab.com/docs/backlog/api/2/get-issue-list/). When the search box matches the pattern `^[A-Z][A-Z0-9_]*-\d+$`, the plugin also calls `GET /api/v2/issues/:issueKey` and puts the result at the top. This works, but it should be written into the AC: "an exact issue key match puts that issue at the top, even when the keyword does not return it".

**F7. The "Showing 1-20 of 57" line (AC2.1.1) needs one more Search call.**
The total comes from `GET /api/v2/issues/count`. This call is in the Search group, like the list call (https://developer.nulab.com/docs/backlog/api/2/count-issue/). Two Search calls must run sequentially with a pause of ≥1 second between them (F5), so each page open takes at least about 1 second. It is still within the 3-second level of NFR1, but with little margin. The Functional Design step needs to know this. The total could be dropped, keeping only previous/next buttons, if more speed is wanted.

**F8. US5.3: the PR creation flow is owned by Kandev.**
`registerRepositoryProvider` has an optional `createChangeRequest` hook. Kandev owns the Create PR UI and the "push the branch before creating the PR" step. So:
- The "locked with the reason Branch not pushed yet" part of AC5.3.2 is Kandev behaviour. The plugin can only test that it is not called when the branch has not been pushed.
- On the Backlog side it uses `POST /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName/pullRequests`, with `summary`, `description`, `base`, `branch`, `issueId` (https://developer.nulab.com/docs/backlog/api/2/add-pull-request/).
- The docs do not say that Backlog rejects a second PR for the same branch (AC5.3.3). This must be tried on the test space. If Backlog does not reject it, AC5.3.3 must change to "the plugin itself checks the list of open PRs with the same `branch` before creating".

**F9. PR watch filters (US6.1) can only use the fields the PR list supports.**
`GET .../pullRequests` can filter by `statusId[]`, `assigneeId[]`, `issueId[]`, `createdUserId[]`, with `offset`/`count` pagination (at most 100) (https://developer.nulab.com/docs/backlog/api/2/get-pull-request-list/). "assignee = me" can be done by getting the id through `GET /api/v2/users/myself`. This filter set should be written into AC6.1.1, so no one promises filtering by label, branch or reviewer.

**F10. Capabilities Kandev already has, confirmed to reduce risk.**
- The `#` source: the manifest has `reference_sources`, along with `SearchEntityReferences` and `AuthorizeEntityReference`. Kandev owns the menu, arrow keys and Enter, so US3.4 can be built. The keyboard part of AC3.4.1 is Kandev behaviour.
- The repository source: the manifest has `repository_providers`, along with the `repositories.inspect` and `repositories.branches` actions.
- The issue label on the card: the `task-card-tags` slot.
- The plugin runs continuously as a process supervised by Kandev, so it can run its own polling timer, with an injected clock for tests.

### 2. Stories too large, should be split

| Story | Reason | Proposed split |
|-------|-------|--------------|
| US3.2 | Has three parts, each with its own API calls, limits and storage: the info fields, comments (paginated), attachments (download, size limit, storage) | US3.2a info fields (Should), US3.2b comments (Should), US3.2c attachments (Could, or Should if kept) |
| US5.1 | Covers the repository list in the picker (TS UI, cursor), the `repositories.branches` action, and the Git credential provider (F2). In total over 3 days | US5.0 Git credentials (new, F2); US5.1a choose repository and branch; US5.1b fetch code and push branches through the credential provider |
| US1.6 | AC1.6.2 (switching space turns off all links and watches) needs the "connection epoch" mechanism the Kandev docs recommend to block stale results. This part is much larger than replacing the key | US1.6a replace the key in the same space (AC1.6.1, AC1.6.3); US1.6b switch to another space (AC1.6.2) |
| US7.2 | AC7.2.2 (429 handling) does not belong to "calling Backlog for the first time". It thickens the skeleton, and it is also the only place covering NFR2 | Keep US7.2 with AC7.2.1. Move AC7.2.2 to a new story **US8.4 Respect API limits — Must**, covering sequential calls per group, waiting per F5, and the "Backlog is limiting requests" message |
| US8.2 | Merges two unrelated things: keyboard use on W1–W10, and localization | US8.2a keyboard and input labels; US8.2b localization |

The other stories fit 1–3 days.

### 3. Missing or incorrect dependencies

- **US4.1** needs US5.2 added if it still keeps PR status in this story. Per F3, PR should be dropped from US4.1, and then nothing needs adding.
- **US6.2** also depends on US5.1 (the task has a repository) and the US4.1 timer. Whoever builds first sets up the shared timer. Deduplication in AC6.2.3 should rely on `IdempotencyKey` or `ExternalID` of `ExactTaskCreate`, keyed by (connection, `repositoryId`, `number`), not the repository name.
- **US4.1, US4.2, US6.2** depend on **US8.4** (newly proposed), because the timer is where 429 is most likely.
- **US3.4** only needs US1.1 and US1.7, not US2.1. It calls its own search through `SearchEntityReferences`.
- **US1.3** has one manual task first: register an OAuth application on Nulab Developer to get a client id, secret and redirect URI. It also needs a plugin `webhook` as the callback (assumption A2). This should be written into the dependencies section.
- **US1.5 and US1.6** must also delete the US5.0 Git secret (F2).
- **US7.1**: AC7.1.2 needs a minimal UI bundle with `registerIntegrationSettings` to show the Backlog item. The skeleton list in `team-practices` does not mention the UI, so it must be stated that the skeleton has a minimal TS settings screen.

### 4. ACs that cannot be built or tested as written

- **AC5.4.1**: no approval data (F1).
- **AC5.4.2**: the 375px layout is drawn by Kandev. The plugin only provides the `taskStatus` data. This AC should be kept as a manual check step, not an AC of plugin code.
- **AC3.2.1**: "read-only" contradicts copying into the description (F4).
- **AC3.2.2**: should be rewritten as "the fake Backlog server receives no POST/PATCH/DELETE call". As it stands, there is no writing sync cycle to check.
- **AC7.2.2**: wrong header (F5). Should move to US8.4.
- **AC7.1.1**: says "5 executables", but linux, darwin, windows times amd64, arm64 is 6 combinations. The exact list per the Bitbucket template must be stated (for example windows/arm64 may be dropped).
- **AC8.1.1**: p95 only makes sense when measured against a real space; a fake server gives no real measurement. It should say measured during the manual check before the release.
- **AC7.6.1**: I have not verified the marketplace catalogue fields (`id`, `repo`, `categories`) in the registry repo. They must be checked at the time.
- **AC5.3.3**: unclear whether Backlog rejects it (F8).

## Positions

- AGREE: The skeleton story set US1.1, US1.2, US7.1, US7.2 (with AC7.2.1 only) — this is indeed the thinnest slice that runs end to end, with OAuth later as in `team-practices`.
- AGREE: Q6-B with the 10-tasks-per-cycle limit — buildable, and deduplication is solid thanks to `IdempotencyKey`/`ExternalID` of `ExactTaskCreate`.
- AGREE: US3.4 and US6.3 are feasible — Kandev already has `reference_sources` and the shared dashboard components.
- OBJECT: AC5.4.1 "1/2 approved" — the Backlog API has no reviewers or approvals (F1).
- OBJECT: US5.1 lacks Git credentials — the API key and OAuth token cannot be used for Git HTTPS; US5.0 must be added (F2).
- OBJECT: US4.1/US4.2 apply the plugin's polling cycle to PR status — Kandev refreshes PR status itself and forbids plugins from running their own poller for this (F3).
- OBJECT: US3.2 as one story — it merges three parts with different API calls and storage, exceeds 3 days, and "read-only" has no feasible place to display yet (F4).
- OBJECT: AC7.2.2 sits in a skeleton story and uses `Retry-After` — Backlog uses `X-RateLimit-Reset`, and NFR2 needs its own Must story (US8.4) (F5).
- OBJECT: US1.6, US5.1, US8.2 as one story each — each merges two independent parts, so they should be split as in the table in section 2.
