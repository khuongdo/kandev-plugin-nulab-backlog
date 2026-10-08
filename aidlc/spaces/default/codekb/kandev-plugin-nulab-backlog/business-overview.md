# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.5.3` at commit `3d6a080`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked. It also gives Backlog projects pull-request context from GitHub, GitLab and Bitbucket.

## Key Functionality

- **Connection**: one Backlog space per workspace, connected with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in per workspace (switch on the Settings > Integrations card).
- **Issues**: issue list with filters and saved queries, create tasks from issues (Kandev's own task dialog or issue watches), issue links and status sync, quick actions. A linked task shows an issue badge on its Kanban card, task rows and top bar. Linking works from both sides: the `/backlog` issue row ("Link to task") and the task's Link menu ("Link Backlog issue", host dialog).
- **Backlog Git**: Backlog Git repository provider, PR link/create/status/list, PR watches, Git credential lease for `nulab-backlog` repositories. Always available while Backlog is connected; its Git access form sits at the top of the Source Control settings section.
- **Source control (SCM)**: GitHub, GitLab and Bitbucket per workspace — repo search, project-to-repo mappings, PR lists, links, queries, watches. A provider is connected with a pasted access token (kept in Kandev's secret store) or, for GitHub and GitLab, with the `gh` / `glab` CLI login on the Kandev server (`scm.providers.use_cli`; since v0.5.3 the admin can pick which `gh` account via `scm.providers.cli_accounts`). **Today any number of the three providers can be connected at the same time**; every downstream feature works per provider. Details: [architecture.md](architecture.md#scm-provider-model-multi-provider-today).
- **Packaging and release**: one `.tar.gz` with 4 platform binaries and a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Workspace members (`access: authenticated` actions; they see the Source Control section read-only) and admins (connection, project selection, Git credential, SCM credential and mapping actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server. That server's OS account is where the `gh` / `glab` login lives.
- The maintainer, who merges through pull requests on a protected `main` and releases by tagging `vX.Y.Z`.

## Current Intent Context

Intent `261008-source-control-settings` (scope express, depth Minimal), verbatim request summarised: refactor the Source Control settings page so that (1) the source control services are clearly separated (today they sit next to each other and are hard to tell apart), (2) the repo-scope (project-to-repository mapping) area says which service the repositories belong to, and (3) only **one** source control service can be used at a time.

- (1) and (2) are UI and message-catalogue changes: [architecture.md](architecture.md#source-control-settings-page-layout).
- (3) is a behaviour change with no support in the code today and needs product decisions first: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-source-control-settings).

Earlier intents in this repo (`261006-*`, `261007-*`, `261008-*` up to `261008-gh-cli-profile`) are shipped up to v0.5.3; their history lives in their intent records.
