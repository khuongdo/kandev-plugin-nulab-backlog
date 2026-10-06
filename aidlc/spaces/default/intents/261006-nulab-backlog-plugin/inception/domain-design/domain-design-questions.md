# Domain Design — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

This step splits the plugin into logical components, meaning the code that will be written, not yet how it is deployed.

Things already settled, not asked again:

- The background part is written in Go. Only one layer may use the Kandev toolkit (`team-practices`, Code Style).
- The Backlog client is separate, and there is a separate secret-masking part.
- One-way sync.
- The Git/PR part is optional.

---

## Q1. Component Granularity

How many logical components should the plugin be split into?

- A. Split by feature, about 8 components:
  - Connection (space, sign-in, secrets, projects, interval)
  - Backlog gateway (API calls, rate limiting, secret masking)
  - Issues and task links
  - Issue status sync
  - Repositories and pull requests
  - PR watches and saved queries
  - Kandev adapter layer
  - UI
- B. Merged, about 6 components:
  - Connection
  - Backlog gateway
  - Issues (including status sync)
  - Git (including repositories, PRs, PR watches, queries)
  - Kandev adapter layer
  - UI
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q2. Periodic Timers

Both issue status sync and PR watches run on cycles. Both jobs share the account's API quota. Where should the timer go?

- A. One shared timer component. It schedules turns for both jobs, so the two never call Backlog at overlapping times
- B. Each component has its own timer. Spacing out calls is handled only by the Backlog gateway
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q3. Where Issue/PR-to-Task Links Are Stored

Where should the link between an issue (or PR) and a task be stored as the primary source?

- A. In the plugin's own data. The label on the task card is only for display
- B. In the Kandev task itself (labels or task metadata). The plugin reads it back from there
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Notifying Other Components of Connection Changes

On disconnect, space change or project deselection, links and PR watches must switch to "no longer connected". When reconnecting to the same old space they must be restored. How do the other components learn about this?

- A. The Connection component emits a "connection changed" event with a connection version marker. Each component updates its own data, and ignores old results from the previous marker
- B. The Connection component calls each other component directly to change the state of their data
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. UI

How should the TypeScript UI part be organised?

- A. A single UI component (one UI package), split by screen M1–M12 inside
- B. Split into several components by area (settings, issues, Git/PR)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: B. Split into about 6 components: Connection, Backlog gateway, Issues (including status sync), Git (including repositories, PRs, PR watches, queries), Kandev adapter layer, UI
- Q2: B. Each periodic component (Issues, Git) has its own timer; spacing and non-overlapping calls are handled by the Backlog gateway
- Q3: A. Issue/PR-to-task links are stored in the plugin's own data; the label on the task card is only for display
- Q4: A. The Connection component emits a "connection changed" event with a version marker; each component updates its own data and ignores results from an old marker
- Q5: A. A single UI component, split by screen M1–M12 inside

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
