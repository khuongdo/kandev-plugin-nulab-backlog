# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.5.0`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked.

## Key Functionality

- **Connection**: connect one Backlog space per workspace with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in after install (per-workspace switch on the Settings > Integrations card).
- **Issues**: issue list with filters and saved queries, create tasks from issues, issue links and sync, issue watches that create tasks, quick actions. A linked task shows an issue badge (key + status) on its Kanban card that opens the issue in Backlog.
- **Backlog Git**: Backlog Git repository provider, pull request link/create/status/list, PR watches, Git credential lease.
- **Source control (SCM)**: provider-neutral PR context for GitHub, GitLab and Bitbucket (read-only repo search, project-to-repo mappings, PR lists, links, queries, watches). Each provider is connected per workspace **only by pasting an access token** (`scm.providers.set_token`, admin); the token is kept in Kandev's secret store. Details: [architecture.md](architecture.md#scm-provider-connection-and-credentials).
- **Packaging and release**: one `.tar.gz` package with 4 platform binaries (Windows dropped in v0.4.2) plus a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Kandev workspace members (actions with `access: authenticated`) and admins (connection, project selection, Git credential and SCM token actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server (manual install; see [architecture.md](architecture.md#interaction-diagrams)). The operator's host is also where any `gh` CLI login would live (see Current Intent Context).
- The maintainer, who works through pull requests on a protected `main` and releases by pushing a `vX.Y.Z` tag (see [architecture.md](architecture.md#ci-and-release-pipeline)).

## Repository Content

The repository holds two kinds of content: the app (Go server, UI bundle, manifest, build and CI configuration) and AI-DLC process records (`aidlc/`, `.claude/`) plus docs. Many merged pull requests change only records (for example PRs #3, #5, #7, #10, #12). Path classification: [code-structure.md](code-structure.md#top-level-path-classification).

## User-Facing Surfaces in Kandev

Where the plugin appears and how each surface is registered: [architecture.md](architecture.md#ui-surfaces-host-slots).

## Current Intent Context

Intent `261008-gh-cli-auth` (scope express, focused scan 2026-10-08): today the GitHub source-control provider can only be linked with a personal access token. The intent adds a second method: use the GitHub CLI login that already exists on the Kandev server host (`gh auth token`). Key facts for the design:

- The token would be resolved on the **Kandev server host** (the plugin binary runs there as a child process of Kandev), not on the browser user's machine. It works only when `gh` is installed and logged in for the OS account that runs Kandev.
- The GitHub client already accepts any bearer token, so the change sits in credential resolution, one new admin action, the stored provider settings and the GitHub card in Settings > Source control.
- Findings and constraints: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-gh-cli-auth).

Earlier intents recorded in this store: `261008-ci-path-filter` (CI only for app paths; constraints in [code-quality-assessment.md](code-quality-assessment.md#ci-path-filter-constraints)), `261008-fix-uiux-backlog` (issue on Home > Tasks rows and settings fixes, released in v0.5.0; findings in [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-fix-uiux-backlog)) and `261007-plugin-install-502` (addressed in v0.4.2; history in [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502)).
