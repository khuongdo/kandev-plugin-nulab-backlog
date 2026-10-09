# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.6.0` at commit `3803248`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked. It also gives Backlog projects pull-request context from GitHub, GitLab and Bitbucket.

## Key Functionality

- **Connection**: one Backlog space per workspace, connected with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in per workspace (switch on the Settings > Integrations card).
- **Issues**: issue list with filters and saved queries, create tasks from issues (Kandev's own task dialog or issue watches), issue links and status sync, quick actions. A linked task shows an issue badge on its Kanban card, task rows and top bar. Linking works from both sides: the `/backlog` issue row ("Link to task") and the task's Link menu ("Link Backlog issue", host dialog).
- **Backlog Git**: Backlog Git repository provider, PR link/create/status/list, PR watches, Git credential lease for `nulab-backlog` repositories. Always available while Backlog is connected; its Git access form sits at the top of the Source Control settings section.
- **Source control (SCM)**: GitHub, GitLab and Bitbucket per workspace — repo search, project-to-repo mappings, PR lists, links, queries, watches. A provider is connected with a pasted access token (kept in Kandev's secret store) or, for GitHub and GitLab, with the `gh` / `glab` CLI login on the Kandev server (`scm.providers.use_cli`; since v0.5.3 the admin can pick which `gh` account via `scm.providers.cli_accounts`). Up to v0.5.3 any number of the three providers could be connected at the same time; **v0.6.0 (PR #26) limits a workspace to one source control service** (Backlog Git, GitHub, GitLab or Bitbucket) chosen by an admin, keeping the others' data disabled (README "0.6.0" notes; the code was not re-read in the `261009-no-workflow-error` run). Pre-0.6.0 model: [architecture.md](architecture.md#scm-provider-model-multi-provider-today).
- **Packaging and release**: one `.tar.gz` with 4 platform binaries and a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Workspace members (`access: authenticated` actions; they see the Source Control section read-only) and admins (connection, project selection, Git credential, SCM credential and mapping actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server. That server's OS account is where the `gh` / `glab` login lives.
- The maintainer, who merges through pull requests on a protected `main` and releases by tagging `vX.Y.Z`.

## Current Intent Context

Intent `261009-no-workflow-error` (scope bugfix, depth Minimal), request summarised: clicking a "+ Task" quick action on `/backlog` shows "Kandev has no workflow for this workspace yet." although the workspace has a workflow; also retouch how such errors are shown on mobile and desktop.

- Root cause and flow: [architecture.md](architecture.md#start-a-task-from-an-issue-row-current-where-no-workflow-arises).
- Findings, sibling callers and error-display issues: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261009-no-workflow-error).

Earlier intents (`261006-*` to `261008-source-control-settings`) are shipped up to v0.6.0; their history lives in their intent records. The previous intent's context (one source control service at a time) is kept below in [architecture.md](architecture.md#source-control-settings-page-layout) and [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-source-control-settings) as pre-0.6.0 analysis.
