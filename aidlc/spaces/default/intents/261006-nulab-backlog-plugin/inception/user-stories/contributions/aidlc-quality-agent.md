**Collaborator:** aidlc-quality-agent

## Contribution

Viewpoint: whether the acceptance criteria (ACs) are testable, under the testing approach the team settled on: TDD, an 80% Go line coverage floor, `go test -race`, a fake Backlog server with `httptest`, a clock and wait function injected into the code, the `packaged-host-contract` contract test, and two manual checks against a real space.

General remark: the draft is good. Most ACs have a clear Given/When/Then, concrete numbers (57 issues, 25 PRs, 2 minutes, 79%), and every story has at least one error case. The feedback below aims to turn the still-vague ACs into pass/fail conditions that automated tests can assert, and to add the edge cases that the team's rules (`-race`, secret masking, 429) require.

### 1. General conventions proposed for the "Conventions" section of `stories.md`

- **Q-T1. Time uses a fake clock.** Every AC with "after one cycle", "2 minutes apart", "wait for the time Backlog specifies", "expired" means: the test advances a fake clock, without a real `time.Sleep`. Propose stating this so AC4.1.1, AC4.2.1, AC1.4.1, AC7.2.2, AC8.1.2 are not read as tests that really wait.
- **Q-T2. Displayed strings.** Because of NFR10 (localization), tests should assert the **message key** or the string in the default language, not the translated string. The draft currently mixes Vietnamese strings ("API key không hợp lệ", "Chưa chọn dự án nào") and English ones ("Couldn't load issues…"). Proposal: state that strings in ACs are strings in the default language (English), or replace them with a description of the meaning.
- **Q-T3. Bait secret (canary).** Every AC about masking secrets uses a fake API key/token with an easily recognised prefix (for example `TESTSECRET-…`). Tests assert this string does not appear in logs, errors, responses sent to the UI, or logged URLs. No real key is used anywhere (matches the NEVER rule in `project.md`).
- **Q-T4. Check type of each AC.** Propose short labels for ACs that cannot be automated: `[manual]` (AC7.2.1, AC7.5.3, part of AC8.2.1, AC5.4.2), `[CI]` (AC7.3.x, AC7.4.x, AC7.5.x). The default for the rest is an automated test against the fake server. This way the Build and Test step knows which ACs need manual evidence.

### 2. ACs not yet testable or still vague — proposed rewrites

| AC | Problem | Proposed rewrite |
|----|--------|------------------|
| AC1.1.2 | Does not distinguish a wrong key (401) from a network error/5xx. If Backlog returns 503 and "invalid API key" shows, that is wrong. | Split: (a) Given the fake server returns 401, When Connect, Then an invalid key error shows and the secret store has no record. (b) Given the fake server returns 5xx or times out, When Connect, Then the error "could not reach Backlog" shows with Retry, does not say the key is wrong, and nothing is stored. |
| AC1.1.3 | "Never shown again in any form" is hard to assert in the UI. | Given it is connected, When the UI calls the API to get settings, Then the response does not contain the key (not even part of it), only a flag like `hasApiKey: true`; the API key input on the page is empty. |
| AC1.2.2 | Missing the dangerous address forms FR1.2 mentions ("unexpected path"). | Write as a table-driven test, adding: `https://myteam.backlog.com/path`, with a port (`:8443`), with userinfo (`https://a@evil.io`), `backlog.com` without a space name, upper-case host (accepted after normalisation), trailing dot (`myteam.backlog.com.`), and input without a scheme (AC1.1.1 uses `myteam.backlog.com` without `https://` — must state that `https://` is added automatically). Then: the fake server receives **0 requests**. |
| AC1.3.3 | "Rejects the connection" does not state anything observable. | Then: no code-for-token exchange request reaches the fake server, nothing is stored, the settings page shows an error. Add the case of an expired `state` and the case of a missing `code`. |
| AC1.4.1 | FR1.5 says refresh "before or when it expires" but the AC only has "when". Missing the race case — the very reason the team turns on `-race`. | Add: (a) Given the token has less than N seconds left (set in functional design), When calling, Then refresh first. (b) Given 5 concurrent calls when the token has expired, When run with `-race`, Then the fake server receives **exactly 1** refresh request and all 5 calls succeed. (c) The new refresh token (if Backlog rotates it) is stored in place of the old one. |
| AC1.4.2 | Does not say whether the plugin retries the refresh repeatedly. | Add: Then the plugin tries to refresh only once per call, without looping forever; later polling cycles do not call Backlog until the user signs in again. |
| AC1.5.1 | Two outcomes in one AC ("success or states the reason"). | Split into two ACs: fake server returns 200 → success shows with the user name; returns 401 → "key is invalid or has been revoked" shows. |
| AC1.5.2 | "All credentials are deleted" must be specific for a test to assert. | Then: the workspace's secret store no longer has the API key, access token, refresh token; the next polling cycle sends no request to Backlog. |
| AC1.6.1 | "Existing links still work" is vague; how "same space" is determined is unclear. | Then: the next polling cycle updates the labels of the old links with the new key (the fake server sees the new key, not the old one). State clearly: same space means the same host after lower-case normalisation. |
| AC1.6.2 | Does "turned off" mean the same as the Paused state of US6.1? Does it keep calling space A? | Then: space A's watches are in their own state (not Paused, cannot be resumed); the fake server for space A receives **0 requests** after the switch; the list of selected projects is deleted. |
| AC2.1.1 | "The first 20 issues" does not state the sort order. | State the order (for example most recently updated first). Add the last page case: page 3 shows "Showing 41-57 of 57" and the next page button is locked. |
| AC2.1.3 | "With a reason" may leak the Backlog response body (forbidden by the team's error convention). | Then: the reason is one of the fixed types (needs reconnecting / being rate limited / cannot reach Backlog), and does not contain the Backlog response body. |
| AC2.2.3 | "Empty state" does not say which state. | Then: exactly the empty state of AC2.1.2 shows, not the error state of AC2.1.3 (a 404 from Backlog must not be treated as an error). |
| AC3.2.1 + AC3.2.2 | Contradiction: AC3.2.1 says "read-only", while AC3.2.2 says "Given I edit that information". Also, FR3.2 is a copy taken at creation time, but no AC asserts the copy does **not** change when Backlog changes. | Rewrite AC3.2.2: Given a task created from PROJ-120, When every plugin action is run in tests (create task, polling cycle, refresh), Then the fake server receives **0 write requests** (no POST/PATCH/DELETE outside the allowed endpoints). Add an AC: Given the comments on Backlog change after the task is created, When the polling cycle runs, Then the copy in the task does not change. |
| AC3.2.3 | The size limit is not settled yet (assumption). | Keep, but write it parametrically: "exceeds limit N (set in functional design)". Add a count case: an issue with 150 comments (Backlog returns at most 100 per page) → copies exactly per the settled count limit and states clearly which part was dropped. |
| AC3.4.1 | "Match" is not defined; it does not say what is inserted. | State clearly: matches by key prefix or keyword in the title, at most N suggestions; picking a suggestion inserts a reference string in format X (set in functional design). Add: typing many characters quickly sends only one request after a wait (debounce) — related to NFR2. |
| AC4.1.1 | "Within at most one cycle" — needs a fake clock (Q-T1). Only issues, PRs missing. | Given a fake clock, When the clock advances exactly one cycle, Then the label changes. Add an AC for PRs: when PR #42 goes Open → Merged, the PR label on the card changes within one cycle (FR4.1 includes PRs). |
| AC4.1.2 | Merges 404 and 403 into one; does not say whether other links are affected. | Table-driven test with 404 and 403. Add: Then other links in the same cycle are still updated (one error does not stop the whole cycle). |
| AC4.2.1 | Missing the 5-minute default (FR4.2) and invalid values. | Add: Given nothing is configured, Then the interval is 5 minutes. Table-driven test for the values 0, negative, non-numeric, empty → rejected. State clearly that a new value applies from the next cycle. |
| AC4.2.3 | Clicking refresh while a cycle is running may cause bursts of calls (NFR2). | Add: Given a cycle is running, When refresh is clicked, Then no two polls run concurrently (`-race` test, the fake server counts maximum parallel requests = 1 for the update group). |
| AC5.1.2 | Real code fetching and branch pushing cannot run against a fake server. | Split: (a) automated: the repository source returns the correct Git HTTPS URL and credentials for `web-app`; (b) `[manual]`: real code fetch and branch push during the manual check. See also risk R2 in section 4. |
| AC5.1.3 | Missing the place where secrets leak most often with Git. | Add: Then the secret is not in the stored remote URL, not in command-line arguments, not in error messages. |
| AC5.3.1 | Two actions in one AC (prefill the form, then click Create). | Split into two ACs: the form is prefilled correctly; clicking Create makes the fake server receive a create PR request with the correct source/target branches, and the task is linked to the returned PR. |
| AC5.3.2 | How is "branch not pushed" determined? | State the source of truth (for example: the Backlog API does not see that branch) so the test can be set up with the fake server. |
| AC5.4.1 | "1/2 approved" — must verify whether the Backlog API has a concept of PR approvals and reviewers (see R3). No Merged and Closed cases yet. | Table-driven test for Open / Merged / Closed. Keep the reviewer part as dependent on the API verification result. |
| AC5.4.2 | "Readable, not cut off" needs a real browser; Vitest + jsdom do not measure layout. | Either label it `[manual]` (check at 375px width during the manual check), or the team decides to add browser tests. If automated: assert there is no horizontal overflow and the status text is not shortened with "…". |
| AC6.1.x | FR6.1 has "edit" and "run" but no ACs. Missing invalid input cases. | Add an AC for editing a watch (a changed filter takes effect from the next cycle), an AC for "Run now", and cases where an empty name / a repository not in a selected project is rejected. AC6.1.2: assert with a fake clock that a Paused watch sends no request over many cycles. |
| AC6.2.2 | Does not say which PRs are picked first when over 10. | State the order (for example oldest PR first) so the test asserts 10, 10, 5 over three cycles deterministically. |
| AC6.2.3 | Deduplication must hold both when the process restarts and when running in parallel. | Add: (a) after the plugin restarts, no duplicate task is created (the dedup state is stored durably); (b) "Run now" and the scheduled cycle running at the same time with `-race` → exactly one task per PR; (c) task creation fails at the 5th PR out of 10 → the next cycle creates from PR 5 on, and PRs 1–4 are not re-created. |
| AC7.1.1 | "5 executables" does not match NFR7 (linux, darwin, windows × amd64, arm64 = 6 combinations). | List exactly the 5 combinations (probably without windows/arm64 like the Bitbucket template) in both the AC and NFR7, so package verification can compare the list of file names. |
| AC7.2.1 | "The result is recorded" does not say where, or what it contains. | State the record file path and the minimum content: date, Kandev version, plugin commit, space domain (no key), steps, result. Label `[manual]`. |
| AC7.2.2 | 429 behaviour is automatic client behaviour, but it sits in the skeleton's manual check story (US7.2). The NFR2 edge cases are missing. | Move to a reliability story (for example US8.4, see section 3) or split from US7.2. Write as a table-driven test: has `Retry-After` (seconds); has `X-RateLimit-Reset` (epoch time); has neither header (use the default wait time); repeated 429 (stop after N times and report an error); `context` cancelled while waiting (return `context.Canceled`, do not call again). Wait with a fake clock. |
| AC7.3.3 | This is a test about tests ("running the secret masking test makes the test fail"), hard to set up and does not prove behaviour. | Rewrite by behaviour: Given the fake server returns 401/500 with a response body containing the bait key (Q-T3), When the error flows run (connect, list, token refresh, Git), Then the captured logs, error messages and responses sent to the UI do not contain the bait key. |
| AC7.4.2 | Building a package that "uses a feature not in the minimum version" in CI is unstable. | Rewrite: Then the contract test job reads `min_kandev_version` from `manifest.yaml` (not hard-coded), checks out exactly that version, and CI fails if installing or the trial run fails. |
| AC7.5.3 | The release workflow cannot check this by itself. | Label `[manual]`: the record file of the second manual check (same format as AC7.2.1) is on `main` before the first tag is pushed. |
| AC8.1.1 | "Backlog responds normally" is not defined; where the measurement starts and ends is unclear. | Split: (a) automated: fake server with a fixed delay (for example 300 ms per request) → the plugin side returns a 20-row list within the time budget; (b) `[manual]`: measure p95 ≤ 3 seconds on a self-hosted Kandev with a test space, exactly per NFR1. |
| AC8.1.2 | "The UI does not freeze" is hard to assert. | Then: the call returns a timeout error within the limit + a small tolerance (fake clock/context deadline), and another request to the plugin during that time is still served. |
| AC8.2.1 | NFR9 also requires an **automated scanner**; the AC only has keyboard checks. | Add an AC: an automated scan (for example axe-core through Vitest) over the W1–W10 components finds no level A/AA violations. Label the keyboard part `[manual]`. |
| AC8.2.2 | No missing-translation case yet. | Add: a string with no translation shows in the default language, without showing the raw key name. |

### 3. Missing edge and error cases

- **Reconnecting fails midway (US1.6).** Given space A is connected, When reconnecting with a new key but Backlog returns 401, Then the old connection stays as it is and the old secret is not deleted. Without this case, the "delete the old secret" rule could break a working connection.
- **HTTP redirect to another host.** Given the fake server returns 302 to a host outside the allow list, When the plugin calls, Then the plugin does not follow it and does not send secrets to that host. (Go's HTTP client follows redirects by default.)
- **`*url.Error` leaks the URL.** Go network errors contain the full URL. If the API key is in the query string, the error message will contain the key. An AC is needed: a network error when calling Backlog does not contain the key in the message or the logs (see R1).
- **Every call has a time limit (NFR5).** AC8.1.2 only covers the list page. Propose a client-level AC: the fake server hangs → every kind of call (connect, list, polling cycle, create PR) returns a timeout error.
- **No concurrent calls in the update and search groups (NFR2).** No AC covers it. Proposal: the polling cycle, PR watches and manual refresh run at the same time with `-race` → the fake server records that the maximum number of parallel requests per group does not exceed the allowed level.
- **Clicking "Create task" twice in a row (US3.1).** Since AC3.1.2 allows a duplicate after confirmation, an AC is needed: clicking twice quickly does not create two tasks without confirmation (the button is locked while waiting).
- **Task creation fails after reading the issue (US3.1).** Kandev refuses to create the task → no orphan link remains.
- **Linking an issue to a second task (FR3.4, success case).** Only the rejected case exists (AC3.3.2). An AC is needed: PROJ-120 is already linked to T-17, also linking T-18 is allowed, and the issue row shows both.
- **A linked task is deleted in Kandev (US2.3).** The issue row no longer shows that task key.
- **Unselecting a project that has links (US1.7).** What happens to that project's links and watches is not defined (see question Q-QA2).
- **Disconnecting while there are links (US1.5).** Q5 only defines the space switch; what happens to links and watches on disconnect is unclear (see Q-QA2).
- **A PR already linked to another task, or a task linked to many PRs (US5.2).** FR3.4 only defines counts for issues. There is no rule for PRs yet, so a reject or accept AC cannot be written.
- **Search with special characters (US2.2).** Japanese/Vietnamese keywords, `&`, `%`, spaces are encoded correctly in the query.
- **Plugin restart.** Polling continues with the configured interval; the watches' dedup state stays intact.

### 4. Risks to verify before settling the ACs (knowledge disputes)

- **R1 — NFR3 "send the API key in a header rather than the URL".** As far as I know, Backlog API v2 accepts the API key through the `apiKey` parameter in the query string; I am not sure whether Backlog accepts the API key in a header. The Backlog docs must be checked. If only the query string works, NFR3 must change, and the ACs for masking secrets in URLs (network errors, request logs, redirects) become mandatory.
- **R2 — FR5.1 Git credentials.** I am not sure whether Backlog's Git over HTTPS accepts an API key/OAuth token or only a user name and password (or an SSH key). If it does not, AC5.1.2 cannot be met with the current connection method. This should be verified early because US5.1 is Must.
- **R3 — FR5.4 reviewers and "approved".** Must verify whether the Backlog pull request API has reviewer/approval status fields so AC5.4.1 can show "1/2 approved".

### 5. FR/NFR coverage between `requirements.md` and the stories

| ID | Current state | What is missing |
|----|-----------|----------|
| FR1.1 | US1.6 | The "shared by all users in the workspace" part has no AC: a second user in the same workspace sees the connection and can use it without connecting again. |
| FR1.5 | US1.4 | The "refresh before expiry" case (see AC1.4.1). |
| FR2.2 | US2.2 | Filtering by project and assignee has no AC; combining filters (AND) has no AC. |
| FR3.4 | US3.3 | The success case: one issue linked to many tasks. |
| FR4.1 | US4.1 | The PR part has no AC (only issues). |
| FR4.2 | US4.2 | The 5-minute default has no AC. |
| FR4.4 | US4.2 | Only "updated at" on the page; the label on the task card has no AC. |
| FR5.4 | US5.4 | Merged/Closed status and reviewers (see R3). |
| FR6.1 | US6.1 | Edit and run now have no AC. |
| NFR2 | US7.2 (AC7.2.2) | No concurrent calls; the 429 header forms; should move out of the manual check story. |
| NFR3 | US1.1, US5.1, US7.3 | Sending the key in a header (see R1); storing with Kandev's encrypted secret mechanism has no AC (test with a fake host: the key goes through the `pluginsdk` secret API, not plain settings). |
| NFR4 | US7.3 (source noted) | No AC covers the criterion "scanning the repo and artifacts finds no secrets". Choose: add a secret scanning step in CI, or state another way clearly (only use prefixed bait keys, plus a manual check). Note: the team dropped **dependency vulnerability** scanning, not **secret** scanning. |
| NFR5 | US2.1, US8.1 | A time limit at the level of every call (see section 3). |
| NFR6 | US7.4 | Fully covered after rewriting AC7.4.2. |
| NFR7 | US7.1 | The 5 files vs 6 combinations contradiction (see AC7.1.1). |
| NFR8 | US7.3 | Covered. Could add the `./internal/...` and `./server/...` measurement scope and the `go mod tidy` check to match `team.md`. |
| NFR9 | US5.4, US8.2 | Missing the automated scan (see AC8.2.1). |
| NFR11 | US8.3 | "The result of each polling cycle" has no AC. Proposal: each cycle writes one structured log record with the number of items updated, the number of errors, the run time; the test captures logs through a fake handler. |
| FR7.2 | US7.5 | Missing: a tag on a commit not on `main` is rejected; package verification reruns in the release workflow (ALWAYS rule in `project.md`). |

The other FRs/NFRs (FR1.2–FR1.4, FR1.6, FR1.7, FR2.1, FR2.3, FR3.1–FR3.3, FR3.5, FR4.3, FR5.1–FR5.3, FR6.2, FR6.3, FR7.1, FR7.3, NFR1, NFR10) already have stories; they only need the AC edits in section 2.

### 6. Proposed questions for the user (choices where both ways are reasonable)

- **Q-QA1 (AC6.2.3).** The user deletes a task that a PR watch created automatically. Does the next cycle re-create a task for that PR? (A. Never re-create; B. Re-create.) FR6.2 says "never duplicate" but says nothing about deleted tasks; the dedup test depends on this answer.
- **Q-QA2 (US1.5, US1.7).** When disconnecting, or when unselecting a project, how are the related issue/PR links and PR watches handled? (A. Like Q5: turn off and label "no longer connected"; B. Delete completely; C. Keep as is, only stop updating.)

## Positions

- AGREE: One shared persona "Minh" for every story, including the release stories — matches the answers to Q1 and Q7, and does not affect testability.
- AGREE: Every story has at least one success case and one error/edge case — meets the floor of the Construction rule and fits TDD.
- AGREE: AC6.2.1–AC6.2.3 correctly reflect Q6 (the 10-tasks-per-cycle limit) and are specific enough to test with a fake clock, once the PR selection order is added.
- OBJECT: AC3.2.1 and AC3.2.2 contradict each other ("read-only" but then "I edit"); they must be rewritten to assert 0 write requests to Backlog and an unchanged copy.
- OBJECT: AC7.2.2 (429 handling) sits in US7.2, which is a manual check story; this automatic behaviour needs its own story/AC with all the header forms and the `context` cancellation case.
- OBJECT: AC7.3.3 is a test about tests and does not prove the secret masking behaviour; it must be rewritten around error flows with a bait key.
- OBJECT: AC7.1.1 says 5 executables while NFR7 lists 6 combinations; the exact list must be stated so package verification can check it.
- OBJECT: NFR4 and the "no concurrent calls" part of NFR2 have no checkable AC; traceability will report a GAP unless they are added.
- OBJECT: AC5.1.2, AC5.4.1 and the header part of NFR3 rely on Backlog API capabilities not yet verified (R1–R3); they must be verified before these ACs are treated as settled.
