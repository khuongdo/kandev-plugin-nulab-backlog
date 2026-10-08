# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.4.2`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked.

## Key Functionality

- **Connection**: connect one Backlog space per workspace with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in after install (per-workspace switch on the Settings > Integrations card).
- **Issues**: issue list with filters and saved queries, create tasks from issues, issue links and sync, issue watches that create tasks, quick actions. A linked task shows an issue badge (key + status) on its Kanban card that opens the issue in Backlog.
- **Backlog Git**: Backlog Git repository provider, pull request link/create/status/list, PR watches, Git credential lease.
- **Source control (SCM)**: provider-neutral PR context for GitHub, GitLab and Bitbucket (read-only links, queries, watches).
- **Packaging and release**: one `.tar.gz` package with 4 platform binaries (Windows dropped in v0.4.2) plus a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Kandev workspace members (actions with `access: authenticated`) and admins (connection, project selection and Git credential actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server (manual install; see [architecture.md](architecture.md#interaction-diagrams)).
- The maintainer, who works through pull requests on a protected `main` and releases by pushing a `vX.Y.Z` tag (see [architecture.md](architecture.md#ci-and-release-pipeline)).

## Repository Content

The repository holds two kinds of content: the app (Go server, UI bundle, manifest, build and CI configuration) and AI-DLC process records (`aidlc/`, `.claude/`) plus docs. Many merged pull requests change only records (for example PRs #3, #5, #7, #10, #12). Path classification: [code-structure.md](code-structure.md#top-level-path-classification).

## User-Facing Surfaces in Kandev

Where the plugin appears and how each surface is registered: [architecture.md](architecture.md#ui-surfaces-host-slots).

## Current Intent Context

Two focused scans on 2026-10-08 updated this store:

- Intent `261008-ci-path-filter`: run CI and release only for changes in app-related paths, excluding `aidlc/`, `docs/` and other non-app paths. At scan time every pull request and every push to `main` ran the full CI, including records-only changes. Constraints that shape the design: [code-quality-assessment.md](code-quality-assessment.md#ci-path-filter-constraints).
- Intent `261008-fix-uiux-backlog` (scope express): show the Backlog issue on Home > Tasks rows like GitHub PRs (hover shows the issue summary, click opens the issue), hide the Backlog entry in Home > Integrations while Backlog is OFF, and three Backlog settings fixes (Projects right below the sign-in method, a correct issue-watch empty message, no duplicate Add watch button). Findings and constraints: [code-quality-assessment.md](code-quality-assessment.md#intent-findings-261008-fix-uiux-backlog).

Previous intent `261007-plugin-install-502` (install upload cut by Kandev's 30 s read timeout) was addressed in v0.4.2 by a smaller package; history in [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502).
