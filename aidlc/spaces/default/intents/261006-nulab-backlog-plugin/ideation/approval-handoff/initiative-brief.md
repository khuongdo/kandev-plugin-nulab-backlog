# Initiative Brief — Kandev Plugin for Nulab Backlog

## Initiative and problem

- Kandev does not support Nulab Backlog yet. A team that uses Backlog every day needs a Kandev plugin for Backlog, similar to the Bitbucket plugin and based on the Backlog public API. (`intent-statement`)
- The plugin has two areas: linking Backlog issues to Kandev tasks, and Backlog Git/pull request integration. (`intent-statement`)
- The target users are Kandev users in general who use Backlog. The confirmed stakeholders are the team that uses Backlog every day. Only the proposer decides the scope, and progress is reported through pull requests and commits in the repo. (`stakeholder-map`)

## Market validation

- Kandev has no Backlog plugin yet. The Kandev maintainers suggested building a plugin for this feature (issue #4215). (`competitive-analysis`)
- The difference from the Bitbucket template: one plugin covers both issues and Git/pull requests. (`competitive-analysis`)
- The decision is to build it ourselves, because no existing solution does this job. The detailed assessment is in `build-vs-buy.md`.

## Feasibility and risks

- **Feasible, with conditions.** No cloud infrastructure needed and no specific compliance requirements. (`feasibility-assessment`)
- Main constraints (`constraint-register`):
  - the plugin backend must be written in Go;
  - Backlog API rate limits are per user;
  - all Backlog domains must be supported;
  - OAuth tokens expire after 1 hour;
  - webhooks can only be sent to a public address.
- Accepted risks, with mitigations (R1, R2, R3, R6 in `raid-log.md`):
  - unfamiliar Go part;
  - scope larger than the template;
  - API rate limits;
  - Kandev changes the plugin programming interface.
- Preparation needed during the construction phase: create a test Backlog space and register an OAuth application with Nulab. [Q2]

## Scope boundaries

- **In scope** (`scope-document`, `intent-backlog`; 7 Must, 7 Should):
  - connect one space with an API key or OAuth;
  - browse issues, link issues to tasks, create tasks from issues, `#` reference;
  - one-way status sync from Backlog to Kandev;
  - Backlog as a repository source; link and create pull requests; show PR status; PR watch and dashboard;
  - the same checks as the Bitbucket plugin; release via GitHub Release and the marketplace.
- **Out of scope:**
  - wiki, Subversion, Gantt, Backlog notifications, webhook;
  - updating issue status from Kandev, creating new issues, commenting on issues;
  - review panel, multiple spaces in one workspace.
- **Order:** risk first. The plugin skeleton, sign-in and packaging come first.

## UI vision

- 10 sketches in `wireframes.md`:
  - a settings page to connect the space;
  - a Backlog entry under "Integrations" with Issues, PR watches, Dashboard;
  - actions in the task menu;
  - issue and PR labels on the task card, on both desktop and phone.
- The product lead's 7 open comments move to the Refined Mockups step. [Q3]

## Team plan

One person works with AI, so no team plan is needed. The Team Formation step was skipped, so there is no `team-assessment`.

## Licence

MIT, like the Bitbucket plugin. [Q4]

## Recommendation

**Go** to the specification phase (Inception). This is also the proposer's decision. [Q5]

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin, based on the Backlog public API.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q5]: answers in `approval-handoff-questions.md`.
- `intent-statement`, `stakeholder-map`: `ideation/intent-capture/`.
- `competitive-analysis`, `build-vs-buy`: `ideation/market-research/`.
- `feasibility-assessment`, `constraint-register`, `raid-log`: `ideation/feasibility/`.
- `scope-document`, `intent-backlog`: `ideation/scope-definition/`.
- `wireframes`: `ideation/rough-mockups/`.

## Assumptions & Open Questions

- [assumption] Assumptions A1–A4 in `raid-log.md` are not yet confirmed: pull requests are available on the Free plan, OAuth accepts `localhost`, the Go part can be written, Nulab still maintains Git.
- [assumption] Whether each issue links to only one task, and the polling frequency, will be settled in the Requirements Analysis step.
