# Units Generation — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

This step groups the 6 components in `components.md` into units of work (Units) to build, with the dependencies between units. The build order is chosen at the Delivery Planning step, not here.

Already decided, not asked again:

- The plugin is deployed as **one single package** installed into Kandev, made of a backend and a UI bundle.
- **The first unit is a thin slice that runs end to end**: API key connection, a minimal settings screen, packaging, package verification, install on self-hosted Kandev, and one successful Backlog call (`team-practices`).
- The Git/PR part is optional.

---

## Q1. How to split units

After the thin slice unit, how should the remaining parts be split into units?

- A. By vertical feature slice: each unit contains both the backend and the screens of that feature (for example "Issues and task links" contains the IssueIntegration component and the Issues screen)
- B. By component: each component in `components.md` is one unit (Connection, BacklogGateway, IssueIntegration, GitIntegration, KandevAdapter, PluginUI)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Unit size

How big should each unit be?

- A. Coarse: about 5 units. For example: thin slice; full connection; issues and sync; Git/PR and PR watch; release
- B. Fine: about 8 units. For example: split issue sync from issues, split PR watch from PR, split connection restore, split accessibility and multi-language
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Parallel work

You work alone with AI. Can independent units be built in parallel (for example, letting AI build two units at the same time)?

- A. Allow: independent units can be built in parallel
- B. No: always build one unit at a time
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Placement of the optional Git/PR part

The Git/PR part (repository, PR, PR watch, queries) is optional. How should it be split so it can be dropped or done later without affecting the rest?

- A. Split into its own unit that only depends on the connection part and the thin slice; no Must unit depends on it
- B. Merge it with other units, and mark it optional at story level
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: A. After the thin slice, split by vertical feature slice: each unit contains both the backend and the screens of the feature
- Q2: A. Coarse split, about 5 units
- Q3: A. Allow independent units to be built in parallel
- Q4: A. The Git/PR part is its own unit, and no Must unit depends on it

Planned unit split from the answers above:

- U1 `walking-skeleton` (service): thin slice that runs end to end (US1.1, US1.2, US7.1, US7.2)
- U2 `connection` (service): full connection, OAuth, change space, projects, API call limit (US1.3–US1.9, US8.3, US8.4); depends on U1
- U3 `issues` (service): issues, task links, `#`, status sync, UI quality (US2.x, US3.x, US4.x, US8.1, US8.2, US8.5); depends on U2
- U4 `git-pr` (service, optional): repository, PR, PR watch, queries (US5.x, US6.x); depends on U2
- U5 `ci-release` (packaging): CI quality gate, check on minimum Kandev, release, marketplace (US7.3–US7.6); depends on U1
- U3, U4 and U5 can be built in parallel with each other

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
