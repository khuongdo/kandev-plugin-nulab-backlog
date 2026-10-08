# Requirements Analysis Questions — 261008-gh-cli-profile

Context: today the plugin's "Use gh CLI login" runs `gh auth token --hostname github.com` without `--user`, so every workspace uses whichever gh account is currently active. The plan is to let each workspace pick one of the gh accounts that are logged in (`gh auth status --json hosts`) and fetch its token with `gh auth token --user <login>`, the same way Kandev's own GitHub integration does.

## Question 1
The agent's shell inside a task worktree gets `GH_TOKEN` from Kandev itself (Kandev's own GitHub connection for the workspace, or the executor profile env). The plugin SDK (v0.96.0) has no way for the plugin to set that env. What should "gh CLI in the task worktree uses the chosen profile" mean for this release?

A. Everything the plugin does for a task created from a Backlog issue (PR list, PR link, PR "mine" filter, watches) uses the workspace's chosen gh account; for the agent's shell, document that Kandev's GitHub integration for that workspace must be set to the same gh account (no code in Kandev)
B. Option A, plus: the plugin settings let the user pick a Kandev executor profile per workspace (whose env carries the right `GH_TOKEN`), and tasks the plugin creates itself pass it as `Launch.ExecutorProfileID` (does not cover tasks created through Kandev's own Start task dialog)
C. Option A, plus a change in Kandev itself (separate repo, `~/repo/kandev`) so plugin-started tasks can inherit the plugin's chosen gh account — larger, separate work
X. Other (please specify)

[Answer]: A  **Mode:** guided

## Question 2
GitLab's "Use glab CLI login" has the same single-account behaviour. Should this change also cover GitLab?

A. No — GitHub (gh) only; GitLab stays as it is
B. Yes — add the same per-workspace account choice for glab
X. Other (please specify)

[Answer]: A  **Mode:** guided

## Question 3
Workspaces already connected with "Use gh CLI login" have no chosen account stored yet (only the account name that was active when they connected). What should happen to them after the upgrade?

A. Keep using the account already recorded for that workspace (its saved GitHub login) — no action needed from the user
B. Keep following whichever gh account is active until the user picks one in Settings
X. Other (please specify)

[Answer]: A  **Mode:** guided

## Question 4
If the chosen account is later logged out of gh (or `gh auth switch` changes the active account), what should the workspace do?

A. Keep the chosen account; if gh no longer has it, show a clear "account not logged in to gh" error and ask to re-pick — never switch silently
B. Fall back to the active gh account and show a warning
X. Other (please specify)

[Answer]: A  **Mode:** guided
