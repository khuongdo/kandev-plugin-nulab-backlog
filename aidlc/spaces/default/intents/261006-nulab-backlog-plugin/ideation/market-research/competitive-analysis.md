# Competitive Analysis — Kandev Plugin for Nulab Backlog

Input: `intent-statement` (ideation/intent-capture/intent-statement.md) — a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin, integrating both issues and Git/pull requests, aimed at Kandev users who use Backlog.

## Comparison scope

- As you chose, compare only with the **Kandev Bitbucket plugin** (`kdlbs/kandev-plugin-bitbucket`) as the reference template. [Q1]
- No comparison with other tools; existing Backlog integrations are mentioned only in `market-trends.md` and `build-vs-buy.md` as context. [Q1] [Q6]

## Reference template: the Bitbucket plugin

| Aspect | Bitbucket plugin (actual) | Source |
|-----------|----------------------------|-------|
| Nature | A plugin package installed into Kandev, not a separate service | [S1] |
| Supported platforms | Bitbucket Cloud and Data Center | [S1] |
| Repository | A native Kandev repository provider (repository discovery) | [S1] [S2] |
| Pull request | Creates a PR after Kandev pushes the task's branch; PR indicators (CI, review status) on the task list; review panel | [S1] |
| Watching | PR watch (create, filter, run, pause, resume, delete); saved dashboard queries | [S1] |
| References | Type `#` to insert a PR reference, re-checked for permission before use | [S1] |
| Authentication | API token or OAuth 2.0; secrets are encrypted and not shown again after saving | [S1] |
| Configuration | Per Kandev workspace; disconnecting deletes the credentials | [S1] |
| Quality checks | Formatting, static analysis, test, packaging, package verification (checksum), version check on release | [S3] [S4] |
| Distribution | GitHub Release, listed in the official Kandev marketplace | [S4] [S6] |
| License / maturity | MIT; latest release v0.4.0 (2026-10-05); created 2026-07 | [S5] |

## Comparison with the planned Backlog plugin

Ratings: Strong / Adequate / Weak / Absent (for Backlog this is the *expected level*, not the current state).

| Capability | Bitbucket | Backlog (expected) | Notes |
|----------|-----------|-------------------|---------|
| Connect space/account, view repositories | Strong | Required | The "basic" level you confirmed as required [Q2] |
| View issue list | Absent (the plugin has no built-in Bitbucket issues) | Required | The biggest difference: Backlog is both an issue tool and a Git host [S7] [Q2] |
| Link issues to Kandev tasks | Absent | Required | [Q2] |
| Link pull requests to Kandev tasks | Strong | Required | The Backlog API has a group of pull request endpoints [S8] [Q2] |
| Create PR from task, watch PRs | Strong | Should have (to be "similar to the Bitbucket plugin") | To be confirmed at the scope step [assumption] |
| Build/test/packaging/package verification checks | Strong | Required at the same level | Per the success criteria in `intent-statement` [Q3 intent] |
| Release through the marketplace | Strong | Goal | A usable release is a success criterion |

## Possible differentiators

- **Issues + Git in one plugin**: Backlog combines issue management and Git hosting, so the plugin can connect issue ↔ task ↔ pull request in a single flow — something the Bitbucket plugin does not do. [S7] [S8]
- **Fill an ecosystem gap**: there is no Backlog plugin in the Kandev marketplace yet; the Kandev maintainer suggested building it as a plugin. [S6] [S9]

## Risks compared with the template

- Larger scope than the template (adds the issue part) → more work than the Bitbucket plugin. [assumption]
- Backlog limits the number of API calls per user (see `market-trends.md`), which is tighter than the API call habits of a "PR watch" feature. [S10] [hypothesis]

## Sources

- [desc] Initial description: asks for a plugin similar to `kdlbs/kandev-plugin-bitbucket` for Nulab Backlog, based on the public API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q6]: answers in `market-research-questions.md`; [Q3 intent]: question Q3 in `intent-capture-questions.md`.
- [S1] Bitbucket plugin README — https://github.com/kdlbs/kandev-plugin-bitbucket/blob/main/README.md
- [S2] Bitbucket plugin manifest — https://github.com/kdlbs/kandev-plugin-bitbucket/blob/main/manifest.yaml
- [S3] Bitbucket plugin Makefile — https://github.com/kdlbs/kandev-plugin-bitbucket/blob/main/Makefile
- [S4] CI/release workflows — https://github.com/kdlbs/kandev-plugin-bitbucket/tree/main/.github/workflows
- [S5] Releases — https://github.com/kdlbs/kandev-plugin-bitbucket/releases
- [S6] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/plugin-registry/plugins.yaml
- [S7] Backlog API documentation — https://developer.nulab.com/docs/backlog/
- [S8] Pull request list API — https://developer.nulab.com/docs/backlog/api/2/get-pull-request-list/
- [S9] Kandev issue #4215 "Nulab Backlog issue provider" — https://github.com/kdlbs/kandev/issues/4215
- [S10] Backlog API rate limits — https://developer.nulab.com/docs/backlog/rate-limit/

## Assumptions & Open Questions

- [assumption] Whether creating PRs from tasks and watching PRs are in scope will be settled at the Scope Definition step.
- [assumption] The issue part adds work compared with the Bitbucket plugin; the exact amount will be assessed at the Feasibility step.
