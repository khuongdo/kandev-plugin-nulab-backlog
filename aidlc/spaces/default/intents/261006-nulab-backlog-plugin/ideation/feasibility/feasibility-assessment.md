# Feasibility Assessment — Kandev Plugin for Nulab Backlog

Input: `intent-statement` (goals and success criteria), `competitive-analysis` (Bitbucket plugin template), `market-trends` (Backlog API constraints), `build-vs-buy` (decision to build).

## Conclusion

**Feasible, with conditions.** No technical barrier blocks the initiative. Three conditions must be handled early:

1. **A Backlog space for testing is required.** There is none yet. The Backlog Free plan includes Git, and there is a 30-day trial of the paid plans. [Q1] [S1]
2. **The plugin's backend must be written in Go.** Kandev supports only Go for this part, while the builder only knows TypeScript. [Q5] [S2]
3. **OAuth 2.0 must be in the first release.** This needs an application registered with Nulab and a fixed callback address. [Q4] [S3]

## Assessment by area

| Area | Assessment | Basis |
|-----|----------|-------|
| Integration with Kandev | Feasible | The Bitbucket plugin template exists with the same install, check, and release approach; community plugins (Redmine, YouTrack, Forgejo) follow the same path [S2] [S4] |
| Integration with the Backlog API | Feasible | The public API has all needed function groups: issues, statuses, Git repositories, pull requests, comments, webhooks. Available on every plan, including Free [S1] [S5] |
| Support for every Backlog domain | Feasible | One registered OAuth application works for `backlog.com`, `backlog.jp`, and `backlogtool.com`. Users must enter their space address [Q3] [S3] |
| Sign-in with API key and OAuth | Feasible, medium difficulty | The Bitbucket template already does both. Backlog OAuth tokens expire after 1 hour and must be refreshed; the callback address must exactly match the registered address [S3] [S6] |
| Real-time updates (webhooks) | Partly feasible | Backlog webhooks can only be sent to a public address. A Kandev running on an internal network cannot receive them, so a periodic polling fallback is needed [S7] [S8] |
| Skills | Risk | Knows TypeScript, not Go. This can be offset with the Bitbucket template and AI tools, but the speed and quality of the Go part need watching [Q5] [hypothesis] |
| Time and cost | Favourable | No hard deadline or cost limit [Q6]. The only possible extra cost is a Backlog space; the Free plan costs 0 [S1] |
| Compliance | Low | No specific requirements [Q7]. Credentials only need protecting at the Bitbucket plugin level (encrypted, not shown again) [S6] |
| Cloud infrastructure | Not applicable | The plugin runs inside the Kandev server you already operate yourself [Q2] [S2]. No AWS account or service needed |
| Release to the marketplace | Feasible | Needs a public GitHub repo, a correctly formatted release package, and a proposal approved by the Kandev maintainers [Q8] [S4] |

## Infrastructure view (platform engineer)

- There is no infrastructure to set up. The plugin is packaged as executables for several operating systems and machine architectures, and Kandev runs it. [S2]
- The only network dependency is the OAuth callback address. It only needs to be reachable from the user's browser, not exposed to the internet. The Bitbucket template requires HTTPS, except when running locally for testing. [S6] [hypothesis: this applies the same way to Backlog]

## Compliance view (compliance specialist)

- Data the plugin handles:
  - Issue and pull request content: the organisation's internal data.
  - Assignee names: low-level personal data.
  - API key and OAuth token: secrets.
- No specific regulations apply. [Q7]
- Minimum controls to keep: encrypt secrets, do not write secrets to logs, delete secrets on disconnect, send the API key in a header instead of the URL. [S6] [S9]

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin, based on the public API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q8]: answers in `feasibility-questions.md`.
- [S1] Backlog pricing — https://nulab.com/pricing/backlog/
- [S2] Kandev plugin documentation — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins.md
- [S3] Backlog API authentication — https://developer.nulab.com/docs/backlog/auth/ ; application registration — https://nulab.com/backlog/developer/applications/
- [S4] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins-marketplace.md
- [S5] Backlog API documentation — https://developer.nulab.com/docs/backlog/
- [S6] Bitbucket plugin README — https://github.com/kdlbs/kandev-plugin-bitbucket/blob/main/README.md
- [S7] Backlog webhooks — https://support.nulab.com/hc/en-us/articles/8840133998489
- [S8] Kandev plugin manifest (webhook address) — https://github.com/kdlbs/kandev/blob/main/docs/public/plugins-manifest.md
- [S9] Backlog release notes (API key in header) — https://nulab.com/release-notes/backlog/

## Assumptions & Open Questions

- [assumption] Pull requests are available on the Backlog Free plan, since they are part of Git. The pricing page does not state this clearly; it must be confirmed when the test space is created.
- [assumption] Whether the Backlog OAuth callback address accepts `localhost` for local testing has not been verified.
