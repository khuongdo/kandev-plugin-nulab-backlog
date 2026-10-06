# Requirements — Kandev Plugin for Nulab Backlog

Inputs:

- `intent-statement`: goal and success criteria.
- `scope-document`: in-scope and out-of-scope capabilities; `intent-backlog` with IB-1 to IB-14.
- `team-practices`: how we test, release, and code conventions.
- Constraints in `constraint-register` (C-…) and decisions in `decision-log` (D-…).

## Intent Analysis

- **What the user wants to achieve**: work with Nulab Backlog issues and Git/pull requests right inside Kandev, where agents work on tasks, instead of switching between two tools. (`intent-statement`)
- **Type of change**: new feature, a brand-new (greenfield) project. It is a plugin installed into Kandev, similar to the Bitbucket plugin.
- **Scope**: one plugin, many parts. It covers the backend, the UI, packaging and release.
- **Complexity**: Standard. There is one external system (the Backlog API), two sign-in methods, periodic polling and API rate limits.
- **Success criteria** (`intent-statement`):
  - The main flow runs against a real Backlog space.
  - It passes checks equivalent to the Bitbucket plugin.
  - There is an installable release.

## Functional Requirements

Each requirement has a pass/fail criterion. The "Source" column traces back to an ideation artifact or a question from this step.

### FR1. Connect a Backlog space

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR1.1 | The system allows connecting **one** Backlog space per Kandev workspace, and this connection is shared by all users in the workspace | A second connection in the same workspace is rejected or replaces the old connection, after the user confirms | IB-2; `scope-document`; [Q1] |
| FR1.2 | The system only accepts space addresses that use `https` with a host under `*.backlog.com`, `*.backlog.jp` or `*.backlogtool.com` | An `http://…` address, another host, an IP, or an unexpected path is rejected, with a message stating the correct format | C-T5; hard rule in `project.md` |
| FR1.3 | The system allows connecting with an API key | With a valid API key, the settings page shows the Backlog user name and the space address. With a wrong API key, an error message shows right below the input | IB-2; D-11 |
| FR1.4 | The system allows connecting with OAuth 2.0 (Authorization Code) through the space's sign-in page | After the user grants access, the settings page shows "Connected as …". Cancelling on the OAuth page returns with the message "Sign-in was cancelled" | IB-3; D-11 |
| FR1.5 | The system refreshes the OAuth token automatically before or when it expires (1 hour). If it cannot refresh, the settings page switches to the "Sign in again" state | When the token expires, the next call still succeeds thanks to the refresh. When the refresh token is no longer valid, the state changes to "Sign in again" | C-T6 |
| FR1.6 | The system allows re-checking the connection and disconnecting. Disconnecting deletes all stored credentials | After disconnecting, no secret is stored for the workspace; the Backlog pages switch to the "not connected" state | C-R3; `wireframes` W1 |
| FR1.7 | The system allows choosing the Backlog projects to use (the issue list, repositories and PR watches only take data from these projects) | A project that is not selected does not appear in any list | [Q4] |

### FR2. Browse and find issues

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR2.1 | The system shows a paginated issue list for the selected projects | The list shows key, title, status, assignee, updated time; it has previous and next page buttons, and the total number of issues | IB-9; `wireframes` W2 |
| FR2.2 | The system allows filtering by project, status, assignee, and searching by keyword or issue key (for example `PROJ-123`) | Each filter narrows the results exactly as on Backlog. Typing an exact issue key makes that issue appear | IB-9; `rough-mockups` Q2 |
| FR2.3 | Each issue in the list shows which tasks it is linked to (if any) | A linked issue shows the keys of the linked tasks | IB-10 |

### FR3. Link issues to tasks

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR3.1 | The system allows creating a Kandev task from an issue. The new task is already linked to that issue | Clicking "Create task" adds a new task to the board, labelled with the issue key | IB-10; D-14 |
| FR3.2 | A task created from an issue gets: title and description; issue key and link; assignee, priority, due date (display only); attachments and comments as a read-only copy taken at creation time | The task contains all the fields above. Changing these fields in Kandev changes nothing on Backlog | [Q5] |
| FR3.3 | The system allows linking an issue to an existing task, and unlinking | After linking, the task card shows the issue label and the issue row shows the task key. After unlinking, both disappear | IB-10; `wireframes` W3, W6 |
| FR3.4 | An issue can be linked to many tasks; each task is linked to at most one issue | Linking a second issue to a task that already has an issue is rejected, with a message; linking an issue to a second task is allowed | [Q2] |
| FR3.5 | The system supports typing `#` in a task's composer to pick and insert a reference to a Backlog issue | Typing `#PROJ-12` shows matching suggestions; picking a suggestion inserts a reference to that issue | IB-11; D-14 |

### FR4. Status sync (one-way)

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR4.1 | The status of linked issues and pull requests is updated from Backlog into Kandev by periodic polling | Changing the status on Backlog changes the label on the task card within one polling cycle | IB-12; D-16 |
| FR4.2 | The polling interval is configurable: default 5 minutes, minimum 1 minute. There is a manual refresh button | A value below 1 minute is rejected. Clicking refresh updates immediately | [Q3] |
| FR4.3 | The system never writes status from Kandev to Backlog | No API call updates an issue status | D-16; `scope-document` |
| FR4.4 | The UI shows when the data was last updated | Every list page and label carries "updated at …" information | Feedback R-05 on the sketches |

### FR5. Repositories and pull requests

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR5.1 | The plugin is a Kandev repository source: it lists the Git repositories of the selected projects to pick when creating a task, and provides Git credentials so Kandev can fetch the code | When creating a task, a Backlog repository can be picked; Kandev can fetch the code and push a branch | IB-5; D-15 |
| FR5.2 | The system allows linking an existing Backlog pull request to a task, and unlinking | After linking, the task card shows the PR label. If the PR is not found, a clear message shows | IB-6 |
| FR5.3 | The system allows creating a pull request on Backlog from a pushed Kandev branch. Title, description and related issue are prefilled from the task | The PR appears on Backlog and is linked to the task. When the branch has not been pushed, the action is locked with a reason | IB-7; D-15 |
| FR5.4 | The task card shows the status of the linked pull request (open, merged, closed) and the reviewers, on both desktop and phone | The label shows the correct status from Backlog, and is readable on a narrow screen | IB-8; `rough-mockups` Q5 |

### FR6. Pull request watches and dashboard

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR6.1 | The system allows creating, editing, deleting, running, pausing and resuming PR watches. Each watch has a name, a repository and a pull request filter | Each action changes the watch state correctly; deleting requires confirmation | IB-13; D-15 |
| FR6.2 | When a new pull request matches the filter of a running watch, the system automatically creates a Kandev task for that PR and links the task to the PR | Each new matching PR produces exactly one task. A PR never produces a duplicate task, even across many cycles | [Q6] |
| FR6.3 | The system allows saving pull request queries and viewing the results on a dashboard, with linked tasks | Choosing a saved query shows the list of matching PRs | IB-13 |

### FR7. Quality and release

| ID | Requirement | Pass criterion | Source |
|----|---------|--------------|-------|
| FR7.1 | The plugin is packaged in the correct Kandev plugin format and passes the package verification step | Package verification succeeds; the package installs on a self-hosted Kandev server | IB-1; `intent-statement` |
| FR7.2 | Each release is a GitHub Release from a `vX.Y.Z` tag, with the package, `checksums.txt` and the build provenance attestation attached | The Release has all the files above; the `gh attestation verify` command succeeds | IB-14; `team-practices` |
| FR7.3 | The plugin is listed in the Kandev marketplace | The proposal to add it to the catalogue is accepted by the Kandev maintainers | IB-14; D-13 |

## Non-Functional Requirements

| ID | Type | Requirement | Pass criterion | Source |
|----|------|---------|--------------|-------|
| NFR1 | Performance | A 20-row list (issues or pull requests) displays within 3 seconds for 95% of opens, when Backlog responds normally | Measured on a self-hosted Kandev server with a test space; p95 ≤ 3 seconds | [Q7] |
| NFR2 | API limits | The plugin respects Backlog's API rate limits: it does not call the update and search groups concurrently; on a 429 response it waits per `X-RateLimit-Reset` or `Retry-After` and then retries; the UI reports "Backlog is limiting requests" | A test with a fake server returning 429 shows the plugin waits the right time and does not burst calls | C-T3, C-T4 |
| NFR3 | Security | API keys and tokens are stored with Kandev's encrypted secret mechanism, are never shown again after saving, the API key is sent in a header rather than the URL, and they are always masked in logs, error messages, test results and responses sent to the UI | A test asserts that no secret appears in logs, errors or responses | C-R3; hard rule in `project.md` |
| NFR4 | Security | No real credentials go into the repo, test data or test artifacts | Scanning the repo and artifacts finds no secrets | Hard rule in `project.md` |
| NFR5 | Reliability | Every Backlog call has a time limit. Network errors or Backlog errors are reported to the user with an easy-to-understand message and a retry button; no error is silently ignored | Tests with a fake server returning 5xx or timing out all produce a message and do not freeze the UI | Construction phase rule; `team-practices` |
| NFR6 | Compatibility | The plugin declares a minimum Kandev version and is tested automatically on that version before each release | The contract test with the real package on the minimum version succeeds in CI | C-T2; `team-practices` |
| NFR7 | Platform | The package includes executables for linux, darwin, windows on amd64 and arm64, like the Bitbucket plugin | The package has all 5 platform executables | C-T1 |
| NFR8 | Testability | Line coverage of the Go code is at least 80%; tests run with `-race`; tests are written first (TDD) | CI blocks the merge when coverage is below 80% or the `-race` tests fail | `team-practices` |
| NFR9 | Accessibility | The UI meets basic WCAG 2.1 AA: keyboard operable, every input has a label, information is not conveyed by color alone | Keyboard checks and an automated scanner find no level A/AA issues on screens W1–W10 | `rough-mockups` Q7 |
| NFR10 | Localization | UI text follows the language Kandev is displaying | Changing the Kandev language changes the plugin text (for languages the plugin has translations for) | `rough-mockups` Q6 |
| NFR11 | Observability | The plugin writes structured logs for errors when calling Backlog, waits due to API limits, and the result of each polling cycle, without secrets | The logs contain the events above, and contain no secrets | Operation phase rule |

## Constraints

- The backend is written in Go with Kandev's plugin development kit; the UI is written in TypeScript. (C-T1)
- Only Backlog's public API is used, under Nulab's Terms. (C-R2)
- Webhooks cannot be used because they need a Kandev server with a public address, so periodic polling is used. (C-T7; D-18)
- One person working with AI; no hard deadline or cost limit. (C-O1, C-O4)
- Way of working, testing and release follow `team-practices`: pull requests, protected `main`, TDD, release by tag. MIT license. (D-27)

## Assumptions

| ID | Assumption | Reason | Confirmed when |
|----|----------|-------|--------------|
| A1 | Pull requests are available on the Backlog Free plan | Pull requests are part of the Git feature; the pricing page does not say clearly | Creating the test space |
| A2 | Backlog's OAuth callback accepts `localhost` for trial runs | The Bitbucket plugin allows `localhost` with HTTP; not yet verified for Backlog | Registering the OAuth application |
| A3 | Kandev provides enough UI mount points for screens W1–W10 | The Bitbucket plugin already used similar mount points | Refined Mockups |
| A4 | Issue attachments and comments can be read through the public API to copy into the task | The Backlog API has comment and file function groups; details not yet checked | Functional design |

## Out of Scope

Wiki, file sharing, Subversion, Gantt, burndown, milestones, Backlog notifications, webhooks. Updating issue status from Kandev. Creating new issues and writing issue comments from Kandev; copying comments read-only when creating a task (FR3.2) is in scope. Pull request review panel. Multiple spaces in one workspace. Per-user Backlog connections. Dependency vulnerability scanning. (`scope-document`; [Q1]; `team-practices`)

## Open Questions

- The maximum size and number of attachments copied when creating a task (FR3.2), to stay within the API rate limits and Kandev's storage limits. To be decided in functional design.
- When a PR watch matches many pull requests at once (for example on the first run), is there a limit on the number of tasks created per cycle (FR6.2). To be decided in functional design.
- Which specific languages have translations in the first release (NFR10).

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin, based on Backlog's public API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q7]: answers in `requirements-analysis-questions.md`.
- IB-…: `ideation/scope-definition/intent-backlog.md`. C-…: `ideation/feasibility/constraint-register.md`. D-…: `ideation/approval-handoff/decision-log.md`.
- `intent-statement`, `scope-document`, `wireframes`, `team-practices`: the corresponding artifacts in this intent's record.

## Assumptions & Open Questions

- [assumption] A1–A4 in the Assumptions table.
- The three open questions in the section above.
