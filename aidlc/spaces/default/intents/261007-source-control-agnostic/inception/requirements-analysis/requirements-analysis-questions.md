# Requirements Analysis Questions — source-control-agnostic

Initial description: "Backlog integration có thể link với nhiều source control khác nhau, ví dụ: github, backlog, bitbucket.... Trong settings có thể setup được những điều kiện để liên kết với source control service tương ứng (auth, repo, space...)"

Context: today the plugin's Git/PR features work only with Backlog Git, and they reuse the Backlog connection's credentials, space and selected projects (see `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/architecture.md` § Source Control Coupling).

## Q1. Which source-control services must the first release support?

Every extra service adds its own API client, auth and tests.

A. Backlog Git (existing) + GitHub
B. Backlog Git (existing) + GitHub + Bitbucket
C. Backlog Git (existing) + GitHub + Bitbucket + GitLab
D. A provider framework plus GitHub only now; Bitbucket and others later
X. Other (please specify)

[Answer]: C

## Q2. How deep should the integration with a new service (e.g. GitHub) go?

Backlog Git today supports clone/push credentials, PR list, PR status, create PR, PR watches and saved PR queries.

A. Link only: link a PR (by URL) to a Kandev task / Backlog issue and show its status
B. Link + read: A plus a PR list with saved queries and PR watches
C. Full parity with Backlog Git: B plus clone/push credentials and create PR from Kandev
X. Other (please specify)

[Answer]: B

## Q3. Kandev already has its own built-in GitHub integration and reserves the provider name `github`. How should GitHub work in this plugin?

A. Reuse Kandev's built-in GitHub integration (its auth and repos); this plugin only links GitHub PRs to Backlog issues
B. This plugin owns its own GitHub connection under a separate name (e.g. `nulab-backlog-github`), with its own token in Backlog settings
X. Other (please specify)

[Answer]: A

## Q4. How does an admin authenticate each service in Settings?

A. One access token per service per workspace (GitHub personal access token / fine-grained token, Bitbucket app password or API token), entered by an admin
B. OAuth sign-in per service
C. Each user enters their own token
X. Other (please specify)

[Answer]: A

## Q5. What "linking conditions" should Settings let you configure per service? (select all that apply)

A. Host (cloud only, e.g. github.com / bitbucket.org)
B. Host including self-hosted (GitHub Enterprise Server, Bitbucket Data Center) — admin enters an `https` address
C. Owner / organization / Bitbucket workspace
D. An allow-list of repositories that may be linked
E. A mapping from each Backlog project to its repositories (so PR lists and linking are scoped per Backlog project)
X. Other (please specify)

[Answer]: A, E

## Q6. Should a GitHub/Bitbucket link depend on the Backlog connection?

Today all Git items are disabled when Backlog is turned off, disconnected, the space changes, or a project is deselected.

A. Follow the Backlog on/off switch only; keep working across Backlog space or project changes
B. Fully tied, like Backlog Git today (switch, disconnect, space change, project deselect all disable them)
C. Fully independent of Backlog
X. Other (please specify)

[Answer]: A

## Q7. How should a PR from GitHub/Bitbucket be linked to a Backlog issue?

Backlog Git PRs carry a native related-issue field; external services do not.

A. Automatically when the Backlog issue key (e.g. `PROJ-123`) appears in the branch name or PR title, plus manual linking
B. Manual linking only
C. Automatically only, by issue key
X. Other (please specify)

[Answer]: A

## Follow-up F1. Reusing Kandev's GitHub (Q3 = A) vs. a PR list and watches for GitHub (Q2 = B) and admin tokens (Q4 = A)

Kandev v0.96.0 does not give plugins its GitHub/GitLab token. A plugin can only read the GitHub/GitLab PRs that Kandev has already attached to a task (number, URL, title, state, checks, review state). A PR list, saved queries and PR watches need direct API access, which needs a token. GitLab is also a built-in Kandev provider, like GitHub.

A. Kandev's built-in GitHub/GitLab keeps clone/push and the PRs on tasks; this plugin additionally stores a read-only token per service (entered by an admin in Backlog settings) for the PR list, saved queries and watches. Bitbucket uses a plugin token for everything.
B. For GitHub/GitLab, use only what Kandev already provides (link + status of PRs on tasks); PR list, saved queries and watches only for Backlog Git and Bitbucket.
X. Other (please specify)

[Answer]: A
