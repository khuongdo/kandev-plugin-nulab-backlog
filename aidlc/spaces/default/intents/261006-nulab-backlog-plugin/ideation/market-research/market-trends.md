# Trends and External Constraints — Kandev Plugin for Nulab Backlog

Input: `intent-statement` (ideation/intent-capture/intent-statement.md). No market size estimate, as you chose. [Q4]

## Relevant trends

| Trend | Evidence | Meaning for the initiative |
|----------|-----------|------------------------|
| The Kandev plugin ecosystem is growing | The official marketplace has 15 plugins, including community plugins for YouTrack, Redmine, Forgejo [S1]; Kandev releases often (v0.96 on 2026-09-25, v0.97 on 2026-10-04) [S2] | There is clear precedent for community-built issue/Git tool plugins |
| Kandev steers new integrations toward plugins | The maintainer replied to issue #4215 (proposal to support Backlog): no Backlog account to test with, suggested a plugin [S3] | Built-in support in Kandev is unlikely; a plugin is the recommended path |
| Nulab is pushing AI and developer tools | Backlog added many AI features in 2026; the official MCP server supports multiple organisations since 04/2026 [S4] [S5] | Backlog users are used to connecting Backlog to AI tools — a good fit for Kandev |
| Backlog Git/pull requests are still maintained | Git and pull request endpoints are still in the documentation; the 01/2026 release notes even added a `useGit` field [S4] [S6] | The Git/PR integration shows no sign of being retired [hypothesis — no official commitment] |
| The Backlog API is improving authentication | Since 08/2026 the API key can be sent in a header instead of a URL parameter [S4] | Lowers the risk of leaking the key in logs/URLs |

## External constraints: the Backlog public API

You asked to consider the API's terms and limits. [Q5]

- **Authentication**: API key or OAuth 2.0 (Authorization Code). OAuth tokens expire after 1 hour and must be refreshed with a refresh token. [S7]
- **Rate limits**: counted per **user** (not per API key), per minute, in 4 groups: read, update, search, icon. Limits differ between free and paid plans; the documentation publishes no fixed numbers, and they must be read through the `rateLimit` API. Example in the documentation: read 600, update 150, search 150, icon 60 per minute (not verified which plan this is). Exceeding the limit returns code 429. [S8] [S9]
- **Nulab's recommendation**: avoid concurrent calls; wait at least 1 second between update, search, and icon calls. [S8]
- **Domains**: a Backlog space can be on `backlog.com`, `backlog.jp`, or `backlogtool.com`; the plugin must let users enter their space's exact domain. [S10] (exact meaning of each domain: not verified)
- **Terms**: the API documentation only points to Nulab's general Terms and Privacy Policy; no API-specific terms were found (not verified that none exist). [S11] [S12]

### Product impact (for later steps to consider)

- Per-user limits mean that many tasks/PR watches running under one account share one quota. [hypothesis]
- The "search" group (issue list) has a lower quota than ordinary reads. [S8] [S9]

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog based on the public API.
- [scope] Workflow-selected scope: `feature`.
- [Q4] [Q5] [Q6]: answers in `market-research-questions.md`.
- [S1] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/plugin-registry/plugins.yaml
- [S2] Kandev releases — https://github.com/kdlbs/kandev
- [S3] Kandev issue #4215 — https://github.com/kdlbs/kandev/issues/4215
- [S4] Backlog release notes — https://nulab.com/release-notes/backlog/
- [S5] Backlog MCP server — https://github.com/nulab/backlog-mcp-server
- [S6] Backlog API documentation — https://developer.nulab.com/docs/backlog/
- [S7] Backlog API authentication — https://developer.nulab.com/docs/backlog/auth/
- [S8] Backlog API rate limits — https://developer.nulab.com/docs/backlog/rate-limit/
- [S9] Get rate limit API — https://developer.nulab.com/docs/backlog/api/2/get-rate-limit/
- [S10] Domain examples in backlog-js — https://github.com/nulab/backlog-js
- [S11] Nulab Terms — https://nulab.com/terms/
- [S12] Nulab Privacy Policy — https://nulab.com/privacy/

## Assumptions & Open Questions

- [hypothesis] Backlog Git/PR will stay maintained during the plugin's lifetime; there is no official commitment from Nulab.
- [assumption] The limit numbers in the documentation example are not guaranteed to apply to every plan; the plugin must rely on the real values returned.
