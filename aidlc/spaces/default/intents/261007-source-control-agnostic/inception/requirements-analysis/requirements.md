# Requirements — Multi-provider source control (261007-source-control-agnostic)

Initial description: "Backlog integration có thể link với nhiều source control khác nhau, ví dụ: github, backlog, bitbucket.... Trong settings có thể setup được những điều kiện để liên kết với source control service tương ứng (auth, repo, space...)" [desc]

Workflow-selected scope: express, Minimal depth [scope]

Answers: `requirements-analysis-questions.md` (Q1-Q7, F1). Code facts: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (architecture.md § Source Control Coupling, code-quality-assessment.md § Intent 261007-source-control-agnostic Risks).

## Intent Analysis

- **Goal**: Teams that track work in Backlog but keep code on GitHub, GitLab or Bitbucket can see and link those pull requests from the Backlog plugin, the same way they already can for Backlog Git. [desc]
- **Type**: Enhancement of an existing, released plugin (v0.3.0). Brownfield.
- **Today**: Git/PR features are Backlog Git only and reuse the Backlog connection's credentials, space and selected projects (codekb architecture.md § Source Control Coupling).
- **Shape of the change**: Add a "source-control provider" concept to the Git/PR area. Backlog Git stays as it is. GitHub, GitLab and Bitbucket are added as **read-and-link** providers: a PR list, saved PR queries, PR watches, PR status and linking to Kandev tasks / Backlog issues. Each provider is set up in Settings with an admin token and a mapping from Backlog projects to repositories. [Q1] [Q2] [Q4] [Q5] [F1]
- **Complexity**: Standard. One new bounded concept (provider) and three new outbound HTTP clients, all built on existing patterns (`internal/backlog` client, `internal/git` documents, Settings sections).

## Functional Requirements

### FR1 — Supported providers

- **FR1.1** The plugin shall support four source-control providers: Backlog Git (existing), GitHub, GitLab and Bitbucket. [Q1]
- **FR1.2** Only the cloud services are supported for the new providers: GitHub (`github.com`), GitLab (`gitlab.com`) and Bitbucket Cloud (`bitbucket.org`). Self-hosted servers are not accepted. [Q5]
- **FR1.3** Backlog Git keeps every existing capability unchanged: repository provider, clone/push credential lease, PR list, PR status, create PR, PR watches, saved PR queries. [Q2] [desc]

### FR2 — Provider settings (admin)

- **FR2.1** Settings shall show a "Source control" area that lists the four providers with their connection state (not configured / connected / error).
- **FR2.2** For GitHub, GitLab and Bitbucket, a workspace admin shall be able to enter one access token per provider per workspace, test it, replace it and remove it. Accepted credential types: GitHub fine-grained or classic personal access token; GitLab personal access token; Bitbucket API token or app password with its username. [Q4]
- **FR2.3** The token is used for read-only API calls only. Settings shall state the minimum read scopes needed for each provider. [F1]
- **FR2.4** Testing a token shall call the provider's "current user" endpoint and show the account name on success, or a plain error (invalid token, missing scope, rate limited, provider unreachable) on failure.
- **FR2.5** Backlog Git keeps its existing settings ("Git access" credential, selected Backlog projects); it is shown in the same "Source control" area.
- **FR2.6** Only admins can change provider settings and mappings, matching the existing rule that connection-changing actions are `admin`.

### FR3 — Backlog project to repository mapping

- **FR3.1** For each connected new provider, an admin shall be able to map each selected Backlog project to zero or more repositories of that provider. [Q5]
- **FR3.2** The repository picker shall list the repositories the provider token can read, with a search box; the admin may also type `owner/name` (GitHub), `group/project` (GitLab) or `workspace/repo` (Bitbucket) directly, validated against the provider before saving.
- **FR3.3** PR lists, saved PR queries, PR watches and automatic issue linking for a new provider shall only cover mapped repositories, and shall be grouped/filterable by Backlog project. [Q5]
- **FR3.4** Removing a mapping shall disable (not delete) the PR watches and saved queries that use it, with a visible "repository no longer mapped" state, as Backlog Git does today for a deselected project.

### FR4 — Pull requests from new providers

- **FR4.1** The PR list on `/backlog` shall let the user choose the provider and show PRs from the mapped repositories with: number, title, author, state (open / closed / merged / draft), source and target branch, updated time and a link to the PR on the provider. [Q2]
- **FR4.2** Saved PR queries shall work for every provider, with the same limits as today (name required, at most 100 characters, at least one state, at most 50 per workspace), plus a provider field. A default query can be set per provider. [Q2]
- **FR4.3** PR watches shall work for every provider: at most one new Kandev task per watch per run, with the per-watch interval (default 5 minutes) as today. [Q2]
- **FR4.4** The PR status shown on a linked task / issue shall be refreshed from the provider on the same schedule as Backlog Git PRs.
- **FR4.5** Creating PRs, merging, commenting and clone/push credentials are not offered for GitHub, GitLab or Bitbucket by this plugin. [Q2] [F1]

### FR5 — Linking PRs to Backlog issues and Kandev tasks

- **FR5.1** A user shall be able to link a PR from any provider to a Kandev task / Backlog issue manually, by picking it from the PR list or by pasting its URL. The URL must belong to a mapped repository. [Q7]
- **FR5.2** The plugin shall automatically link a PR from a new provider to a Backlog issue when a Backlog issue key of a mapped Backlog project (e.g. `PROJ-123`) appears in the PR's source branch name or title. [Q7]
- **FR5.3** Automatic links are created when the PR is first seen by a list refresh or watch run and shall be shown as "auto-linked"; a user can remove an auto-link, and a removed auto-link is not re-created for that PR.
- **FR5.4** For GitHub and GitLab, PRs that Kandev itself has attached to a task (from its built-in integration) shall be shown on the linked Backlog issue without needing a plugin token. [Q3] [F1]

### FR6 — Relationship to the Backlog connection

- **FR6.1** All provider features follow the Backlog on/off switch: when Backlog is off, every provider action except reading settings is refused, as today. [Q6]
- **FR6.2** Links, mappings, queries and watches of GitHub, GitLab and Bitbucket shall keep working across a Backlog space change, Backlog disconnect/reconnect and project deselection; items whose Backlog project is no longer selected are hidden from project-scoped views but not deleted. [Q6]
- **FR6.3** Backlog Git keeps its current coupling to the Backlog connection (disconnect, space change and project deselect disable its items; its credential is deleted on disconnect and space change). [Q6] [FR1.3]

### FR7 — Backward compatibility

- **FR7.1** Data stored by v0.3.0 (`git.links`, `git.watches`, `git.queries`, the `backlog.git.<ws>` secret, the `nulab_backlog_pr` task metadata and the review key format) shall keep working unchanged after upgrade and be read as Backlog Git data.
- **FR7.2** Existing action keys keep their names and inputs; new optional `provider` inputs default to Backlog Git.

## Non-Functional Requirements

- **NFR1 — Secret handling**: Provider tokens are stored only in the Kandev secret store, are never returned to the browser, and are redacted in logs, errors and test output. A test asserts this for each provider. (project.md Mandated; team.md Testing Posture)
- **NFR2 — Outbound host allow-list**: Each new client only sends requests to a fixed `https` host for its provider (`api.github.com`, `gitlab.com`, `api.bitbucket.org`). Hosts taken from PR URLs or typed input are validated against this list before any call.
- **NFR3 — Rate limits**: Each client honours the provider's rate-limit responses (HTTP 429 and GitHub's 403 with `X-RateLimit-Remaining: 0`, with `Retry-After` / reset headers), maps them to the existing "unavailable" error, and never retries more than the existing Backlog client does. A watch run for one provider does not block other providers.
- **NFR4 — Response limits**: Response bodies are read through `io.LimitReader`; list calls page with at most 100 items per page and stop at the same caps as Backlog Git lists.
- **NFR5 — Errors**: Each client has its own error type, mapped in `internal/plugin` to the existing `ActionError` codes (401/403 → reconnect required; 429 → unavailable; 404 → not found). Public messages contain no provider response body and no secret.
- **NFR6 — Dependencies**: Clients use only the Go standard library (no GitHub/GitLab/Bitbucket SDKs). (team.md Code Style)
- **NFR7 — Tests**: TDD. Unit and integration tests use `httptest` fake providers with JSON fixtures, including 401, 403, 404 and 429 cases. The existing suites stay green with `-race`, and the 80% line-coverage floor holds. A regression test loads v0.3.0 fixtures and checks FR7.1. (team.md Testing Posture)
- **NFR8 — Responsiveness**: Opening the PR list for one provider shows the first page within 3 seconds when the provider answers within 1 second, measured against the fake provider in tests.

## Constraints

- **C1** Kandev v0.96.0 reserves the provider ids `github`, `gitlab`, `azure_devops` and does not share its GitHub/GitLab tokens with plugins. This plugin therefore does not register GitHub or GitLab as Kandev repository providers. Provider names used inside the plugin's own data (`backlog`, `github`, `gitlab`, `bitbucket`) are not Kandev provider registrations. (codekb Risks 1-2) [Q3] [F1]
- **C2** Only `internal/plugin` imports the SDK. New clients live in new packages (e.g. `internal/github`, `internal/gitlab`, `internal/bitbucket`), not in `internal/backlog`. (team.md Code Style)
- **C3** Minimum Kandev version stays 0.96.0; no host change is required.
- **C4** The project Mandated rule "ALWAYS accept only `https` space addresses under `backlog.com`, `backlog.jp`, or `backlogtool.com`" continues to govern the **Backlog space address**. New providers follow NFR2's separate fixed allow-list. See Open Question OQ1.

## Assumptions

- **A1** One token per provider per workspace (not per user) is acceptable, so every member sees the same set of PRs. [Q4]
- **A2** Bitbucket is used read-only in this release like GitHub/GitLab (Q2 = B covers link + read; clone/push is not included). [Q2] [F1]
- **A3** GitHub/GitLab "draft" state and Bitbucket "declined" state are shown as their own labels; for saved-query filtering they map to open / closed respectively.
- **A4** An issue key is recognised only for Backlog projects that are currently selected, using the Backlog key pattern `[A-Z][A-Z0-9_]*-[0-9]+`.

## Out of Scope

- Self-hosted GitHub Enterprise Server, GitLab self-managed and Bitbucket Data Center. [Q5]
- OAuth sign-in or per-user tokens for the new providers. [Q4]
- Creating, merging or commenting on PRs, and clone/push credentials for GitHub, GitLab and Bitbucket. [Q2]
- Azure DevOps and other services. [Q1]
- Changing Kandev's built-in GitHub/GitLab integration.

## Open Questions

- **OQ1** The Mandated host rule in `aidlc/spaces/default/memory/project.md` is worded only for Backlog space addresses. Should it be extended with a matching rule for the new providers' fixed hosts (NFR2)? Only the human can change that rule; until then NFR2 is a requirement of this intent.
- **OQ2** Exact minimum token scopes per provider (FR2.3) are confirmed during Code Generation against current provider documentation.
