# Infrastructure Design — walking-skeleton (U1) — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

The plugin has no cloud infrastructure of its own. It runs inside a self-hosted Kandev server, and it stores data in Kandev's state and secret stores. So the infrastructure for U1 is:

- the place where the skeleton is installed and checked;
- the build and check pipeline.

The full CI and release workflows belong to U5 (`bolt-plan`). Two decisions are left.

---

## Q1. Where the skeleton is installed for the first manual check

The skeleton is done only after the package is installed on a self-hosted Kandev at `min_kandev_version` (0.96.0) and connected to a real Backlog space (`team-practices`). Kandev publishes Docker images (`ghcr.io/kdlbs/kandev:<version>`).

- A. A throwaway local Kandev 0.96.0 started with Docker (`docker run ... ghcr.io/kdlbs/kandev:0.96.0`), with its data on a named volume. The manual check record notes the image tag
- B. An existing self-hosted Kandev server you already run (it must be on 0.96.0 or later)
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q2. Checks on U1's pull request before U5 adds CI

`team-practices` says `main` is protected and every required CI check must pass before merging. U1 is the first pull request, and CI arrives with U5. U1's pull request would therefore have no CI to pass.

- A. U1 adds a minimal GitHub Actions workflow that runs `make check-format vet lint test coverage build package verify-package` on `pull_request`. Branch protection requires that one check from the start. U5 extends the same workflow (contract test, release, secret check)
- B. U1 is merged after the same `make` targets pass locally. Branch protection and CI are turned on in U5
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

---

## Follow-up Questions

## Q3. Kandev version of your existing server

In Q1 you chose your existing self-hosted server. The NFR requirements (NFR6.1) say the first manual check runs on exactly Kandev 0.96.0, the declared minimum. If your server runs a newer version, the manual check does not prove the minimum.

- A. The server runs exactly 0.96.0, so NFR6.1 holds as written
- B. The server runs a newer version. The manual check runs there, and proof of the 0.96.0 minimum moves to the automated contract test in U5. NFR6.1's pass criterion is read that way
- C. The server will be set to 0.96.0 for the check
- X. Other (please specify)

[Answer]: A

## Changes After the U1 Change Request (2026-10-06)

The U1 design now adds a per-workspace switch (`connection.set_enabled`, state key `integration`), a home Integrations entry with a `/backlog` page, an inline Backlog logo with a brand record, two new log events, and a new Connect budget (12 s deadline plus 2 s rollback). These follow from approved answers and need no new question. The infrastructure files are updated as listed in the summary below.

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: the first manual check runs on your existing self-hosted Kandev server (B).
- Q2: U1 adds a minimal GitHub Actions workflow that runs the standard `make` targets on every pull request, and branch protection requires that check from the start; U5 extends the same workflow (A).
- Q3: that server runs exactly Kandev 0.96.0, so NFR6.1 holds as written (A).
- Change request: the manifest declares `connection.get` (authenticated), `connection.connect_api_key` and `connection.set_enabled` (admin), all `scope: workspace`; the state store also holds key `integration`; the Backlog call limit is 10 s inside a 12 s Connect deadline, with a 2 s store budget and a 2 s rollback; the UI bundle carries the inline logo and `docs/brand/backlog-logo.md` records its source; the CI `checks` job adds a bundle check that no Nulab URL is fetched at runtime; monitoring lists `integration_switched` and `action_refused_disabled`; the second manual check (before the first release) also covers the switch, the home entry and the `/backlog` page.

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
