# User Stories Questions — 261008-gh-cli-profile

## Story Plan

- **Personas**: derived from requirements and the codekb business overview — (1) a developer who uses several GitHub accounts across Kandev workspaces (primary), (2) an existing gh CLI user upgrading from v0.5.2 (secondary).
- **Format**: "As a [persona], I want [goal], so that [benefit]", INVEST, Given/When/Then acceptance criteria.
- **Priority**: MoSCoW from requirements; all FR1-FR5 behaviour is Must, the worktree-shell note (FR6) is Should.
- **Breakdown**: by workflow (connect and pick, use, change, upgrade, failure, worktree note).

## Question 1
How fine-grained should the stories be? (About 6 stories are planned.)

A. One story per workflow step (about 6 stories: pick on connect, change later, plugin calls use chosen account, upgrade keeps account, logged-out error, worktree note)
B. Fewer, bigger stories (about 3: connect and change, use and failure, upgrade and note)
X. Other (please specify)

[Answer]: A  **Mode:** guided

## Question 2
Should the worktree note (the agent's shell gets its GitHub credential from Kandev's own GitHub integration) be a user story, or only a documentation task?

A. A Should-priority user story with acceptance criteria (shown on the settings card and in the README)
B. Only a documentation task, no user story
X. Other (please specify)

[Answer]: A  **Mode:** guided
