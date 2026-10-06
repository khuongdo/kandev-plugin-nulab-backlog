# NFR Requirements — walking-skeleton (U1) — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Most non-functional targets for U1 are already fixed:

- **From `requirements`**: security (NFR3, NFR4), time limits and errors (NFR5), platforms (NFR7), coverage and `-race` (NFR8), accessibility (NFR9) and logging without secrets (NFR11).
- **From `team-practices`**: tooling.
- **From `contract-summary` and the U1 functional design**: the 10-second call limit, the 1 MiB body limit and the redaction rules.

The tech stack follows Kandev itself:

- Go at the version in Kandev's `go.mod` (1.26).
- The UI uses Kandev's shared React through `host.jsx`. It does not bundle its own React.
- TypeScript strict, Vitest, ESLint and Prettier.

Only the gaps below remain.

---

## Q1. Response time for the settings screen and Connect

NFR1 (3 seconds at p95) covers lists only. U1 has two other user actions: reading the connection when the settings screen opens (no Backlog call), and Connect (one Backlog call).

- A. Reading the connection: p95 ≤ 500 ms. Connect: p95 ≤ 3 s when Backlog responds normally, and never more than 13 s in total (3 s rate-limit cap plus the 10 s call limit, from `contract-summary`)
- B. Only the hard limits (10-second call limit, 15-second Kandev action limit); no percentile target for U1
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Minimum Kandev version

The manifest must declare `min_kandev_version`. The admin-only Connect action needs at least `0.91.1`. The latest Kandev release is `v0.96.0`. The CI contract test (U5) runs on exactly this version.

- A. `0.91.1`, the lowest version that supports what U1 uses. More self-hosted servers can install it
- B. `0.96.0`, the current release. Less to test against, but older servers cannot install it
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q3. What a successful Connect writes to the log

NFR11 requires structured logs without secrets. When a Connect succeeds, the log can identify the account, but the Backlog display name is personal data.

- A. Log the space host, the numeric Backlog user id, the connection epoch and the duration; never the display name or the key
- B. Also log the display name
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Changes After the U1 Change Request (2026-10-06)

The U1 functional design now adds a per-workspace on/off switch stored and enforced by the backend (`connection.set_enabled`, admin-only, error `integration_disabled`), a Backlog entry in the home Integrations menu opening a `/backlog` page, the official Backlog logo, and a spacing retouch. The design review also found that Connect's worst case (10 s Myself call, store writes, 5 s rollback) can exceed Kandev's 15-second action limit. Three gaps follow.

## Q4. Connect time budget inside Kandev's 15-second action limit

- A. One Connect budget of at most 14 s in total: the Myself call keeps its 10 s limit, the store writes (including the second switch read) share 2 s, and the rollback gets its own 2 s on a fresh context. The 13 s figure in Q1 is replaced by this budget
- B. Keep the 5 s rollback and lower the Myself call limit to 7 s
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. What a switch change writes to the log

- A. Log the workspace id, the previous and new value, and the duration; Kandev's own request log identifies the admin
- B. Also log the acting admin's Kandev user id, if the action context provides it
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q6. How the Backlog logo is shipped

- A. Embedded in the UI bundle as an inline SVG component, with no network request at runtime, and the asset's source and terms recorded in the repository
- B. Loaded from Nulab's website at runtime
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: reading the connection p95 ≤ 500 ms; Connect p95 ≤ 3 s when Backlog responds normally. The total-time bound is now set by Q4 (A).
- Q2: `min_kandev_version` is `0.96.0`, the current Kandev release; the CI contract test runs on it (B).
- Q3: a successful Connect logs the space host, numeric Backlog user id, connection epoch and duration, never the display name or the key (A).
- Q4: one Connect budget of at most 14 s, inside Kandev's 15-second action limit: Myself call 10 s, store writes (including the second switch read) 2 s, rollback 2 s on a fresh context. This replaces the 13 s figure and the 5 s rollback in the functional design (A).
- Q5: a switch change logs the workspace id, the previous and new value and the duration; Kandev's request log identifies the admin (A).
- Q6: the Backlog logo is an inline SVG component embedded in the UI bundle, with no runtime network request; its source and terms are recorded in the repository (A).

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
