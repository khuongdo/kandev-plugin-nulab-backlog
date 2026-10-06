# RAID log — Kandev Plugin for Nulab Backlog

Input: `intent-statement`, `competitive-analysis`, `market-trends`, `build-vs-buy`, plus the constraints in `constraint-register.md`.

## Risks

| ID | Risk | Likelihood | Impact | Proposed handling |
|----|--------|----------|-----------|---------------|
| R1 | The backend is written in Go and the builder does not know Go, so quality and speed suffer [Q5] | High | Medium | Follow the Bitbucket plugin structure closely; keep the checks (formatting, static analysis, test) mandatory as in the template |
| R2 | Scope is larger than the template: the issue part is added, and OAuth must be there from the start [Q4]. The workload exceeds the Bitbucket plugin | Medium | Medium | Settle the minimum boundary at the Scope Definition step |
| R3 | Exceeding API rate limits, especially when watching PRs or loading the issue list often. Limits are per user, so they are shared with other tools on the same account [S1] | Medium | High | Set a rate limit inside the plugin; respect code 429 and the limit information in responses |
| R4 | Webhooks cannot be received when Kandev runs on an internal network [S2] | High | Low | Use periodic polling as the default |
| R5 | The marketplace proposal is delayed or rejected, because it depends on the Kandev maintainers [S3] | Low | Medium | The GitHub Release build still works on its own; follow the catalogue's format requirements exactly |
| R6 | Kandev changes its plugin programming interface between versions, because Kandev releases often [S4] | Medium | Medium | Declare the minimum Kandev version; test on that version before every release, as in the template |
| R7 | The Backlog Free plan is not enough for testing pull requests or webhooks (not verified) [S5] | Low | Medium | Use the 30-day trial of a paid plan if needed |

## Assumptions

| ID | Assumption | To be confirmed by |
|----|----------|------------------|
| A1 | Pull requests are available on the Backlog Free plan | Creating the test space |
| A2 | The Backlog OAuth callback accepts a `localhost` address for local testing | Registering the OAuth application |
| A3 | The Go part can be learned or written with AI to a level that passes the checks of the Bitbucket template | The first build |
| A4 | Nulab keeps maintaining Backlog Git and pull requests during the plugin's lifetime | Following Nulab's release notes |

## Issues

| ID | Issue | Resolution |
|----|--------|------------------|
| I1 | No Backlog space for testing yet [Q1] | Create a space (Free or trial) with Git, one repository, and one sample pull request |
| I2 | No OAuth application registered with Nulab yet | Register it with the plugin publisher's Nulab account |
| I3 | No license chosen for the plugin yet | Decide at a later step |

## Dependencies

| ID | Dependency | Type |
|----|-----------|------|
| D1 | Kandev's plugin development kit and packaging tools (Kandev source code) | External |
| D2 | The Backlog public API and Nulab's terms | External |
| D3 | Your self-hosted Kandev server, with admin rights [Q2] | Internal, available |
| D4 | Kandev maintainers approve the marketplace proposal [Q8] | External |
| D5 | Backlog test space (I1) | Internal, not available yet |

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q8]: answers in `feasibility-questions.md`.
- [S1] Backlog API rate limits — https://developer.nulab.com/docs/backlog/rate-limit/
- [S2] Backlog webhooks — https://support.nulab.com/hc/en-us/articles/8840133998489
- [S3] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins-marketplace.md
- [S4] Kandev releases — https://github.com/kdlbs/kandev
- [S5] Backlog pricing — https://nulab.com/pricing/backlog/

## Assumptions & Open Questions

- [assumption] A1–A4 above. None of these assumptions has been confirmed yet.
