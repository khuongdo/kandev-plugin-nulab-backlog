# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.5.2` at commit `ca8146c`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked. It also gives Backlog projects pull-request context from GitHub, GitLab and Bitbucket.

## Key Functionality

- **Connection**: one Backlog space per workspace, connected with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in per workspace (switch on the Settings > Integrations card).
- **Issues**: issue list with filters and saved queries, create tasks from issues (Kandev's own task dialog or issue watches), issue links and status sync, quick actions. A linked task shows an issue badge on its Kanban card, task rows and top bar. Linking works from both sides: the `/backlog` issue row ("Link to task") and the task's Link menu ("Link Backlog issue", host dialog).
- **Backlog Git**: Backlog Git repository provider, PR link/create/status/list, PR watches, Git credential lease for `nulab-backlog` repositories.
- **Source control (SCM)**: GitHub, GitLab and Bitbucket per workspace — repo search, project-to-repo mappings, PR lists, links, queries, watches. A provider is connected with a pasted access token (kept in Kandev's secret store) or, for GitHub and GitLab, with the `gh` / `glab` CLI login on the Kandev server (`scm.providers.use_cli`, token never stored). Details: [architecture.md](architecture.md#scm-provider-connection-and-credentials).
- **Packaging and release**: one `.tar.gz` with 4 platform binaries and a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Workspace members (`access: authenticated` actions) and admins (connection, project selection, Git credential, SCM credential actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server. That server's OS account is where the `gh` / `glab` login lives.
- The maintainer, who merges through pull requests on a protected `main` and releases by tagging `vX.Y.Z`.

## Current Intent Context

Intent `261008-gh-cli-profile` (scope express, depth Minimal): let the admin choose **which gh CLI account** a workspace uses for the GitHub connection (today `gh auth token` returns only the gh *active* account), and make `gh` inside a task worktree created from a Backlog task use that account.

- The first half sits inside the plugin: the CLI command, its cache key, the per-workspace `scm.Settings`, one new action and the GitHub card. Change points: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-gh-cli-profile).
- The second half is **outside the plugin's reach** with pluginsdk v0.96.0: Kandev's executor sets `GH_TOKEN` in the worktree from Kandev's own GitHub connection or executor profile, and the plugin cannot inject environment. See [architecture.md](architecture.md#task-creation-and-the-worktree-gh-environment).

Earlier intents in this repo (`261007-*`, `261008-ci-path-filter`, `261008-fix-uiux-backlog`, `261008-gh-cli-auth`, `261008-link-task-modal`) are shipped up to v0.5.2; their history lives in their intent records.
