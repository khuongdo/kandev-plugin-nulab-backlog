# Business Overview — kandev-plugin-nulab-backlog

## Purpose

A Kandev plugin (id `nulab-backlog`, version `0.4.1`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Developers see Backlog issues and pull requests inside Kandev, create Kandev tasks from issues, and keep both sides linked.

## Key Functionality

- **Connection**: connect one Backlog space per workspace with an API key or OAuth; only `https` hosts under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted. Backlog is opt-in after install.
- **Issues**: issue list with filters and saved queries, create tasks from issues, issue links and sync, issue watches that create tasks, quick actions.
- **Backlog Git**: Backlog Git repository provider, pull request link/create/status/list, PR watches, Git credential lease.
- **Source control (SCM)**: provider-neutral PR context for GitHub, GitLab and Bitbucket (read-only links, queries, watches).
- **Packaging and release**: one `.tar.gz` package with 5 platform binaries plus a UI bundle, verified offline, published as a GitHub Release and listed in the Kandev marketplace registry.

## Users

- Kandev workspace members (actions with `access: authenticated`) and admins (connection and Git credential actions with `access: admin`).
- The self-hosted operator who installs the package on the Kandev server (manual install; see [architecture.md](architecture.md#interaction-diagrams)).

## Current Intent Context

Intent `261007-plugin-install-502`: installing the v0.4.1 package through the Kandev web UI fails with `Plugin install failed: 502`. Root cause and evidence: [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502).
