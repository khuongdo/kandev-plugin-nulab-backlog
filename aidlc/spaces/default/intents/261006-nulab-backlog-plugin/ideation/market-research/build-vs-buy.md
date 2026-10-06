# Build vs Buy — Kandev Plugin for Nulab Backlog

Input: `intent-statement` (ideation/intent-capture/intent-statement.md) — Kandev does not support Backlog yet, and the Backlog user team needs a plugin.

## Conclusion

**Build a new plugin**, in line with your decision. [Q3] Public research found no solution that does this job, so there is no real "buy" option. [S1] [S2]

## Options considered

| Option | Exists? | Meets the goal? | Source |
|-----------|-------------|-------------------|-------|
| Built-in Backlog support in Kandev | No. Proposal #4215 is still open; the maintainer suggested a plugin | No | [S2] |
| Backlog plugin in the Kandev marketplace | None | No | [S1] |
| Nulab's official Backlog MCP server | Yes (issues, Git, pull requests, wiki…) | Partly: gives agents access to Backlog, but does not link issues/PRs to tasks and is not a Kandev repository provider | [S3] [hypothesis] |
| Other Nulab tools (`bee` CLI, backlog-js library, backlog4j, Slack integration) | Yes | No: not tied to Kandev | [S4] [S5] [S6] [S7] |
| Third-party tools (GitHub Action linking PRs to Backlog issues, Jenkins plugin) | Yes | No: serve GitHub/Jenkins, not Kandev | [S8] [S9] |

## Assessment by criterion

Scores from -2 (leans buy) to +2 (leans build).

| Criterion | Assessment | Score |
|----------|----------|------|
| Mature solution available? | No solution does exactly the job of linking Backlog to Kandev tasks | +2 |
| Time to adoption | There is a reference template (the Bitbucket plugin) and precedent community plugins (Redmine, YouTrack, Forgejo) | +1 |
| Customisation needed | Needs both issues and Git/PR — more than the template | +1 |
| Maintenance burden | Must keep up with changes in Kandev (minimum version requirement) and the Backlog API | -1 |
| Data sensitivity | The plugin holds Backlog credentials; the Bitbucket template already has a way to store encrypted secrets | 0 |
| **Total** | | **+3 → Build** |

## What can be reused (not "buy")

- The structure, check, and release process of the Bitbucket plugin. [S10]
- The official Backlog API documentation. [S11]
- Given the "build anyway" choice, this section is recorded for reference only; it does not replace the decision to build. [Q3]

## Sources

- [desc] Initial description: build a plugin similar to the Bitbucket plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q3] [Q6]: answers in `market-research-questions.md`.
- [S1] Kandev marketplace — https://github.com/kdlbs/kandev/blob/main/plugin-registry/plugins.yaml
- [S2] Kandev issue #4215 — https://github.com/kdlbs/kandev/issues/4215
- [S3] Backlog MCP server — https://github.com/nulab/backlog-mcp-server
- [S4] bee CLI — https://github.com/nulab/bee
- [S5] backlog-js — https://github.com/nulab/backlog-js
- [S6] backlog4j — https://github.com/nulab/backlog4j
- [S7] Slack integration — https://nulab.com/blog/product-updates/backlog/slack-integration-update-add-issues-directly-from-slack/
- [S8] backlog-github-integration-action — https://github.com/kazamori/backlog-github-integration-action
- [S9] Jenkins plugin for Backlog — https://plugins.jenkins.io/backlog
- [S10] Bitbucket plugin — https://github.com/kdlbs/kandev-plugin-bitbucket
- [S11] Backlog API documentation — https://developer.nulab.com/docs/backlog/

## Assumptions & Open Questions

- [hypothesis] Nulab's MCP server cannot replace the plugin because it lacks Kandev-specific extension points (repository provider, task linking); this assessment is based on the public description and has not been tried in practice.
