# Constraint Register — Kandev Plugin for Nulab Backlog

Input: `intent-statement`, `market-trends` (API constraints), `competitive-analysis` (Bitbucket plugin standard), `build-vs-buy`.

## Technical constraints

| ID | Constraint | Source | Level |
|----|-----------|-------|--------|
| C-T1 | The plugin's backend must be written in Go. This is the only option Kandev documents and supports, and the official development kit exists only for Go. The UI is optional and can be written in TypeScript | [S1] | Hard |
| C-T2 | The plugin must match the minimum Kandev version declared in the package and must be tested on that version before release, like the Bitbucket template | [S2] | Hard |
| C-T3 | Backlog API rate limits are counted per user, per minute, in 4 groups: read, update, search, icon. Exceeding the limit returns code 429. Nulab advises avoiding concurrent calls and waiting at least 1 second between update or search calls | [S3] | Hard |
| C-T4 | Limits differ between the Free plan and paid plans, and there are no fixed published numbers. The real values must be read from the API | [S3] | Hard |
| C-T5 | All three domains `backlog.com`, `backlog.jp`, `backlogtool.com` must be supported. Users enter their space address, and every API call and sign-in goes to that address | [Q3] [S4] | Hard (your choice) |
| C-T6 | OAuth tokens expire after 1 hour; the plugin must refresh them itself. The callback address must exactly match the address registered with Nulab | [S4] | Hard |
| C-T7 | Backlog webhooks can only be sent to an address the Backlog server can reach; internal network addresses do not work | [S5] | Hard |

## Organisational constraints

| ID | Constraint | Source | Level |
|----|-----------|-------|--------|
| C-O1 | The builder knows TypeScript but not Go | [Q5] | Soft (offset by the Bitbucket template and AI) |
| C-O2 | There is no Backlog space for testing yet. A space with Git and pull requests must be created before end-to-end testing | [Q1] | Hard for the success criteria |
| C-O3 | A self-hosted Kandev server with admin rights is available | [Q2] | Favourable |
| C-O4 | No hard deadline or cost limit | [Q6] | Favourable |
| C-O5 | Only the proposer decides scope. Progress is reported through pull requests and commits in this repo | Q6, Q7 in `intent-capture-questions.md` | Process |
| C-O6 | Getting into the official marketplace needs a public GitHub repo, a correctly formatted release package, a proposal to add it to the catalogue, and approval by the Kandev maintainers | [Q8] [S6] | External dependency |
| C-O7 | Creating an OAuth application with Nulab needs a Nulab account of the plugin publisher | [S4] | Hard for OAuth |

## Legal and compliance constraints

| ID | Constraint | Source | Level |
|----|-----------|-------|--------|
| C-R1 | No specific regulations | [Q7] | — |
| C-R2 | Must comply with Nulab's general Terms when using the API. No API-specific terms were found | [S7] | Hard |
| C-R3 | Secrets (API key, OAuth token) must be encrypted, not shown again after saving, and deleted on disconnect, at the same level as the Bitbucket plugin | [Q7] [S2] | Hard (your choice) |
| C-R4 | Kandev uses the AGPL-3.0 license, while the Bitbucket plugin uses MIT. The license for the Backlog plugin has not been chosen | [S8] [S2] | Decision needed |

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog based on the public API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q8]: answers in `feasibility-questions.md`.
- [S1] Kandev plugin writing guide — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins.md
- [S2] Bitbucket plugin (manifest, README) — https://github.com/kdlbs/kandev-plugin-bitbucket
- [S3] Backlog API rate limits — https://developer.nulab.com/docs/backlog/rate-limit/
- [S4] Backlog API authentication — https://developer.nulab.com/docs/backlog/auth/
- [S5] Backlog webhooks — https://support.nulab.com/hc/en-us/articles/8840133998489
- [S6] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins-marketplace.md
- [S7] Nulab Terms — https://nulab.com/terms/
- [S8] Kandev — https://github.com/kdlbs/kandev

## Assumptions & Open Questions

- [assumption] The plugin license (C-R4) will be chosen at a later step. The Bitbucket template suggests MIT, but no decision has been made.
