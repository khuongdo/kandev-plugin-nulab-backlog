# Requirements: Source Control Settings Refactor

## Intent Analysis

Initial description (verbatim): "Refactor source control settings page - Phân định rạch ròi giữa các source control service, hiện tại chúng đang nằm sát nhau khó thấy quá - phần scope repo/ phải ghi rõ là repo của service nào. hiện tại đang ghi chung chung khó hiểu -> đề xuất phương án - chỉ có thể dùng 1 source control service tại 1 thời điểm" [desc]

Goals:

1. An admin can see at a glance which source control service the workspace uses; services no longer blur together on the Settings page.
2. The repository mapping area always says which service its repositories belong to.
3. A workspace uses exactly one source control service at a time: Backlog Git, GitHub, GitLab or Bitbucket.

Type: enhancement + UI refactor, multi-component (settings UI, `internal/scm`, `internal/git`, `internal/plugin` actions/manifest). Depth: Minimal.

Current state (from the code knowledge base, `codekb/kandev-plugin-nulab-backlog/architecture.md` and `code-quality-assessment.md#intent-findings-261008-source-control-settings`): Backlog Git and the three provider blocks are unframed sibling `<section>`s with `h4 text-sm` headings; mapping labels (`scmMappingsHeading`, `scmSearchLabel`, `scmManualLabel`, `scmManualPlaceholder`) are provider-neutral; the backend has no active-provider concept and accepts connections to all three external providers at once.

## Functional Requirements

### FR1 Active source control service

- **FR1.1** The workspace has exactly one active source control service, one of `backlog_git`, `github`, `gitlab`, `bitbucket`. There is no "none" choice. [Q1=B, Q9=A]
- **FR1.2** A workspace with no external service connected has Backlog Git active by default. [Q9=A, Q4 prompt]
- **FR1.3** The active service is stored by the plugin backend (additive field in `scm.settings`; schema version stays 1) and is the single source of truth for every source control feature. [Q2=A]
- **FR1.4** The backend refuses source control actions for a non-active service (connect, test, use CLI, mapping changes, PR queries, PR watches, PR link actions) with a clear error that names the active service. Removing a stored token of a non-active service stays allowed. [Q2=A]
- **FR1.5** Pull request lists, provider selectors in the PR list and watch forms, link refresh and watchers only use the active service. [Q2=A, Q8=A]
- **FR1.6** When an external service is active, every Backlog Git source control feature is off, including the Backlog Git pull request list/links and the "Git access" form. When Backlog Git is active, GitHub/GitLab/Bitbucket features are off. [Q1=B, Q8=A]

Acceptance:

- Given GitHub is active, When a client calls a GitLab connect action, Then the action fails and the message names GitHub as the active service.
- Given Backlog Git is active, When a member opens the PR list, Then only Backlog Git pull requests are listed and no GitHub/GitLab/Bitbucket selector is offered.

### FR2 Choosing and switching the service

- **FR2.1** The Source Control section shows a "Service" selector at the top (Backlog Git, GitHub, GitLab, Bitbucket) and only the active service's card below it. [Q2=A, Q6=A]
- **FR2.2** Only admins can change the selector; members see the active service read-only (same admin/member split as today).
- **FR2.3** Switching asks for confirmation that names the old and the new service and says the old service's data will be kept but disabled. [Q2=A, Q3=A]
- **FR2.4** Switching keeps the old service's token/credentials, repository mappings, saved PR queries, PR watches and task-PR links, but disabled (the existing "Remove token" precedent). Switching back re-enables them without re-entry. [Q3=A]

Acceptance:

- Given GitHub is active with two repository mappings, When the admin switches to GitLab and confirms, Then only the GitLab card is shown, GitHub mappings no longer drive any feature, and When the admin switches back to GitHub, Then both mappings are active again.
- Given the admin opens the switch confirmation, When they cancel, Then the active service is unchanged.

### FR3 Upgrade of existing workspaces

- **FR3.1** On first load after the upgrade, a workspace with no external service connected gets Backlog Git active. [Q4 prompt]
- **FR3.2** A workspace with exactly one external service connected gets that external service active automatically. [Q7=A]
- **FR3.3** A workspace with two or three external services connected gets no automatic choice: the Source Control section shows a notice asking the admin to pick one service; until the admin picks, existing features keep working as before the upgrade. After the pick, FR1 and FR2.4 apply. [Q4=A]

Acceptance:

- Given a stored `scm.settings` with GitHub and Bitbucket tokens and no active field, When the section loads, Then the "pick one service" notice is shown and both PR lists still work; When the admin picks Bitbucket, Then GitHub data is kept disabled.
- Given a stored `scm.settings` with only a GitLab token, When the plugin loads it, Then GitLab is active and no notice is shown.

### FR4 Visual separation

- **FR4.1** The active service is shown in a framed card with the service logo, a heading larger than its inner sub-headings, and a status badge (Connected / Not connected). [Q6=A]
- **FR4.2** The selector and the card are visually distinct from the rest of the Settings page sections.

### FR5 Repository mapping labels name the service

- **FR5.1** Every heading, label, search field and empty-state text in the repository mapping area names the active service, e.g. "GitHub repositories linked to Backlog projects", "GitHub repository for {project}", "Search GitHub repositories". [Q5=A]
- **FR5.2** Every mapped repository line shows the service logo/name next to the repository, e.g. "[GitHub] owner/name". [Q5=C]
- **FR5.3** New copy does not use the word "scope" for repository mapping (it stays reserved for token permission scopes). [codekb finding 3]

## Non-Functional Requirements

- **NFR1** Secrets: tokens and Git access credentials are never shown, returned to the UI, or logged, including in switch/upgrade paths; existing redaction tests keep passing. [project.md Mandated]
- **NFR2** Compatibility: works on Kandev `min_kandev_version` 0.96.0; existing `scm.settings` documents load without a schema version bump; no data loss on upgrade. [team.md Deployment]
- **NFR3** Accessibility: the selector, confirmation dialog and card pass the existing axe checks; the selector is keyboard operable and labelled.
- **NFR4** All UI text comes from the message catalogue (`ui/src/messages/en.ts`), as today.
- **NFR5** Tests: team TDD posture; Go tests with `-race`, 80% line-coverage floor on `./internal/...` and `./server/...`; Vitest for UI. Express Minimal strategy: at least one test per requirement above. [team.md Testing Posture]

## Constraints

- Keep existing test ids (`backlog-section-source-control`, `backlog-scm-<provider>-*`, `backlog-scm-<provider>-map-<project>`, `backlog-scm-backlog`) unless a test is deliberately updated.
- A new action key (e.g. setting the active service) updates `manifest.yaml`, the SCM handler table, `TestSCM_Manifest_Actions` and the `guarded()` allow-list; `internal/plugin/testdata/v030/manifest.yaml` stays frozen.
- Only `internal/plugin` imports `pluginsdk`; stdlib only, no new dependencies.

## Assumptions

- Members (non-admins) keep their current read-only view; only admins choose the service. [assumption]
- "Disabled" data for a non-active service means it is stored but ignored by every feature and hidden from lists, not shown greyed out. [assumption]
- Backlog Git counts as "connected" whenever the Backlog connection exists, so it is never a reason to show the upgrade notice by itself. [assumption, from Q7=A]

## Out of Scope

- Using two services at once or per-project service choice.
- Deleting old service data on switch (Q3=B was not chosen).
- Self-hosted GitHub Enterprise / GitLab instances.
- Changes to OAuth for Backlog or to issue features.

## Open Questions

- Exact visual style of the card and logos within the Kandev UI kit is settled during Code Generation.

## Sources

- `[desc]` Initial description: project-description.json.
- `[Q1]`-`[Q9]` requirements-analysis-questions.md answers.
- Code knowledge base: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (business-overview, architecture, code-structure, code-quality-assessment).
- Rules: team.md, project.md.
