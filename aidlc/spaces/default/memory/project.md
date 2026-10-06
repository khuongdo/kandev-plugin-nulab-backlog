# Project-Level Rules

> Project-specific specialisation and corrections. Loaded after `org.md` and
> `team.md` as strict-additive guidance; contradictions with broader policy
> are rejected. Populated by practices-discovery and the self-learning loop.
>
> Use sparingly: most teams don't need a project layer. Reach for it
> only when this specific project needs stable, durable guidance beyond the
> team practice (for example, package-specific release checks or an additional
> regression suite for a legacy component).

## Way of Working

<!-- Project-specific specialisation. Example: -->
<!-- This monorepo requires package-scoped branch names and a package owner -->
<!-- review in addition to the team's normal merge policy. -->

## Walking Skeleton

<!-- Project-specific specialisation. Example: -->
<!-- The walking skeleton must exercise the legacy service adapter as well -->
<!-- as the new service boundary. -->

## Testing Posture

<!-- Project-specific specialisation. -->

## Guard Policy

<!-- Project-specific. Mode: strict, relaxed, or off. Strict here holds for every intent and cannot be changed from chat. A section under the retired Change Control heading, written by an earlier release, is still read. -->

## Deployment

<!-- Project-specific specialisation. -->

## Code Style

<!-- Project-specific specialisation. -->

## Tech Stack

<!-- Technology choices locked for this project. -->

## Decided

<!-- Decisions made in earlier stages that should not be re-asked. -->
<!-- Format: DECIDED: [decision] (Stage [slug], [date]) -->

## Scope Overrides

<!-- Custom scope rules for this project. -->

## Forbidden

<!-- Populated by practices-discovery affirmation gate. -->
<!-- Format: NEVER [behavior] (affirmed [date]) -->
<!-- Example: NEVER throw exceptions across service layer boundaries (affirmed 2026-05-17) -->

- NEVER put real credentials in the repo, test data, or test artifacts. (affirmed 2026-10-06)

- NEVER delete or overwrite a released tag. (affirmed 2026-10-06)

## Mandated

<!-- Populated by practices-discovery affirmation gate. -->
<!-- Format: ALWAYS [behavior] (affirmed [date]) -->
<!-- Example: ALWAYS use Result<T,E> for fallible operations in service layer (affirmed 2026-05-17) -->

- ALWAYS redact API keys and tokens in logs, error messages, and test output. (affirmed 2026-10-06)

- ALWAYS accept only `https` space addresses under `backlog.com`, `backlog.jp`, or `backlogtool.com`. (affirmed 2026-10-06)

- ALWAYS run package verification before tagging a release. (affirmed 2026-10-06)

## Corrections

<!-- Project-specific corrections from human feedback. -->
<!-- Format: NEVER/ALWAYS [behavior] (learned [date]) -->
- Kept competitive analysis to the Bitbucket plugin only (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:market-research:2efd62d64cf2df1091547906ebc3c6222855856155c120d18a379cc5bf8d71cb -->
- Treated the Q5 answer 'typescript' as: familiar with TypeScript only, not Go (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:feasibility:ea7f57dc436827ca3ba6ae44ee61ccc9508e497cb168ecbd6c7aa35458b805d6 -->
- Classified items the user picked for the first release beyond the market-research must-haves as Should, not Must (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:scope-definition:4e3df191b13bb1d7de3e208fc4d0a5d42c093c7b3cff2aafc8610e42b7011544 -->
- Read Q1-A (no separate page) and Q2-A (issue list page) as conflicting and asked a follow-up (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:rough-mockups:ed1353c7cc48b5b8d535b830d86dfe7a1bf54fa250586cc743cce4426460c046 -->
- Did not re-ask stakeholder agreement or mob staffing (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:approval-handoff:5d03e0c8e5c07788ac72932ec1de3e1eb07473d8ee46875617e42b29c7b3ade6 -->
- Noted team-assessment as absent by design in the brief instead of inventing a team plan, since Team Formation was skipped. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:approval-handoff:abd08b8b53ae2789c6efa8daf686ad55b83abd140bbadae8a7e0e74b3c86a9a9 -->
- Asked Walking Skeleton in plain words with the stage's gloss, and asked a follow-up when declining manual pre-release checks conflicted with the real-space success criterion (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:practices-discovery:57c955d47cfced8a79ccf039e69e6db9fbbd74d3b72242e85b8e8625b3241cb1 -->
- Read the Q8 free-text answer as: all issue-related stories are Must and all Git/PR stories are Should (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:user-stories:d3d07b72f7ec86917cd91da3856ffb6cadcb9da87effaedec8b438c06b212372 -->
- Kept US3.5 (comments and attachments) at Should although it is issue-related and Q8 says issue features are Must (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:user-stories:c87ba85ee999ef6076cc7201c122b09bf741fe5191dec80ebe47d1fa8a5c4f85 -->
- Marked Kandev-rendered pieces (task menu, # suggestions, PR badge, create-PR flow) as data-only for the plugin rather than designing them, since Kandev owns their UI. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:refined-mockups:85682af3eae769aca82fe86a5e7cf4bc70b7524bbee01bfedf08e559729a28b2 -->
- Used the refined-mockups questions to settle two open product decisions from the user-stories review (restore after reconnect, US3.5 priority) because they shape screens (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:refined-mockups:8acb1b47d3e33de6602a8eeccf2fdba9b704d1c27cb50db1ead86c91a5c18332 -->
- Modelled the Kandev host API as an external dependency reached through a port implemented by KandevAdapter, so the catalogue stays acyclic while only KandevAdapter imports the SDK. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:domain-design:53c07162c71d8255a54a98c659408475ddc71af869ef045a9508d6c53babff5c -->
- Made BacklogGateway stateless about credentials (passed per call) to avoid a Connection-Gateway cycle, at the cost of callers fetching credentials each time. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:domain-design:c5bb62ccd55f0536238ccdf4c5ff507a1b1e2099a1938580f34cdb3b69e6858b -->
- Kept U2 (connection) to Must-only stories and moved US5.5 (Git credentials) into the optional U4 so no Must unit depends on Git, even though the secret is stored by the Connection component. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:units-generation:f9a9d56c4cd085facd440099669916b6143c6e9426df7429bdf16177312875a3 -->
- Asked a separate Approve Plan / Revise Plan question after the summary confirmation, because the stage requires plan approval before writing unit artifacts. (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:units-generation:67b960f62366eb6403097cf3a162c3eee4c37889d30395cdbfb27a602363838e -->
- Added Q6 (detecting deleted Kandev tasks) to the question set (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:contract-design:1f19c0af6941085b7723adc320aff7751e0a17b39c763426dc223b3eb48ba670 -->
- Read Q1 (CI before connection) together with Q7 (parallel coding) as a merge-order rule (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:delivery-planning:08eb53681de17efc54afceb508fae5e7bce1d39ccb28af5540450300aa38eeaa -->
- walking-skeleton: in Kandev 0.96.0 registerIntegrationSettings only adds a card on Settings > Integrations (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:code-generation:3ea952e72205a9e9b4892a8d8d8b2c08cb3e7e0d5fc0eec1ba1bf8109e0d8e38 -->
- walking-skeleton: renamed the action connection.connectApiKey to connection.connect_api_key because Kandev only accepts keys matching ^[a-z0-9][a-z0-9._-]*$ (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:code-generation:ac91a8cfcfbb4f10fc037d7876e46690380c2df9f2384d6c8efd76eee37171ae -->
- walking-skeleton: pinned ../kandev to the v0.96.0 tag (user choice) (learned 2026-10-06) <!-- cid:261006-nulab-backlog-plugin:code-generation:426144fb5c6cd4cc5246f27b54366710308fcaf496c2097ffb6719380139fabb -->
