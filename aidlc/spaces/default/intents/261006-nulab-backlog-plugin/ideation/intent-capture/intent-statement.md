# Intent Statement — Kandev Plugin for Nulab Backlog

## Problem Statement

- A Kandev plugin is needed for Nulab Backlog, similar to the `kdlbs/kandev-plugin-bitbucket` plugin, based on the Nulab Backlog public API. [desc]
- The plugin must cover both areas: linking Backlog issues to Kandev tasks, and integrating Backlog Git/pull requests. [Q1]
- The team uses Nulab Backlog, and Kandev does not support Backlog yet. [Q4]

## Target Customer

- Kandev users in general who use Nulab Backlog. [Q2]
- Among them, the team that uses Backlog daily is the confirmed stakeholder. [Q5]

## Success Metrics

- Connects to a real Backlog space and runs the main flow end-to-end from Kandev. [Q3]
- Passes checks equivalent to the Bitbucket plugin (build, test, packaging, package verification). [Q3]
- Ships a usable release (an install package for Kandev). [Q3]

## Initiative Trigger

- The team uses Nulab Backlog and Kandev does not support Backlog yet, so this plugin is needed now. [Q4]

## Initial Scope Signal

- Scope selected by the workflow (workflow-selected): `feature`. [scope]
- Product boundary confirmed by the user: `feature` is the product boundary, that is, a complete plugin similar to the Bitbucket plugin. [Q8]
- The integration direction within that boundary covers both issues and Git/pull requests. [Q1]

## Assumptions & Open Questions

None.
