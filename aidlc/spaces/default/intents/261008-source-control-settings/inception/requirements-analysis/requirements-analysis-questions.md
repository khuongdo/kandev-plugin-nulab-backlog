# Requirements Analysis Questions

Context: the Source Control settings section today shows Backlog Git and three provider blocks (GitHub, GitLab, Bitbucket) as unframed siblings with same-size headings; the repository mapping labels ("Repositories of each selected Backlog project", "Search repositories for {project}", one shared placeholder "owner/name, group/project or workspace/repo") never name the service; and the backend lets all three providers be connected at the same time (each keeps its own token, mappings, queries, watches and links).

## Question 1
Does "only one source control service at a time" include Backlog Git (the Git hosting built into Backlog, always available through the Backlog connection)?

A. No - Backlog Git always stays available; the rule applies to the external services only (at most one of GitHub, GitLab, Bitbucket)
B. Yes - Backlog Git is one of the four choices; choosing it turns off GitHub/GitLab/Bitbucket, and choosing an external service hides Backlog Git
X. Other (please specify)

[Answer]: B

## Question 2
How should the admin choose and switch the one active service? (This decides whether the rule is enforced only on the page or also by the plugin backend, which today accepts connections to all three.)

A. Explicit selector: a "Service" choice (GitHub / GitLab / Bitbucket / none) at the top of the section; only the chosen service's card is shown; switching asks for confirmation; the backend stores the active service and refuses actions for other services
B. First connected wins: all cards are shown, but once one is connected the others are disabled with a hint "Disconnect GitHub first"; the backend refuses connecting a second service
C. Connecting another service automatically disconnects the current one after a confirmation dialog; the backend enforces the same
D. Page-only: hide or disable the other cards in the UI; the backend stays as is
X. Other (please specify)

[Answer]: A

## Question 3
When the admin switches from one service to another, what happens to the old service's data (repository mappings, saved PR queries, PR watches, links between tasks and pull requests)?

A. Keep it but disable it (the same as "Remove token" today); it comes back if the admin switches back
B. Delete it after a confirmation that lists what will be removed
X. Other (please specify)

[Answer]: A

## Question 4
Existing workspaces may already have two or three services connected. After the upgrade, which one becomes the active service?

A. None is chosen automatically: the section shows a notice asking the admin to pick one; until then everything keeps working as today
B. Choose automatically: the service with the most repository mappings (ties: GitHub, then GitLab, then Bitbucket); the others are kept disabled as in Question 3 A
C. Choose automatically in fixed order GitHub, then GitLab, then Bitbucket; the others are kept disabled
X. Other (please specify)

[Answer]: A

## Question 5
Proposal for the repository mapping part ("scope repo"), so it clearly says which service the repositories belong to. Which do you prefer? (select all that apply)

A. Rename the block and labels to name the service, e.g. heading "GitHub repositories linked to Backlog projects", per project "GitHub repository for {project}", search "Search GitHub repositories"
B. Per-service placeholder and format hint (GitHub `owner/name`, GitLab `group/project`, Bitbucket `workspace/repo`) instead of the one shared placeholder
C. Show the service logo/name next to every mapped repository line (e.g. "[GitHub] owner/name")
D. Move repository mapping out of the connection card into its own sub-section "Repository mapping - <service>" below the connection
X. Other (please specify)

[Answer]: A, C

## Question 6
How should each service be visually separated on the page?

A. Each service (including Backlog Git) in its own framed card with the service logo, a larger heading, and a status badge (Connected / Not connected / Disabled)
B. Tabs, one per service
C. Collapsible panels, one per service, only the active one expanded
X. Other (please specify)

[Answer]: A

## Follow-up Questions

Why: Question 1 answer B makes Backlog Git one of the four services, and Backlog Git is always available through the Backlog connection. That leaves three gaps between Questions 1, 2 and 4.

## Question 7
A workspace that today uses Backlog Git plus exactly one external service (for example GitHub) technically has two services after the upgrade. Should the admin also be asked to pick (Question 4 answer A), or is that case clear enough to pick automatically?

A. Pick the one connected external service automatically; ask the admin only when two or three external services are connected
B. Always ask the admin whenever any external service is connected
X. Other (please specify)

[Answer]: A

## Question 8
When GitHub, GitLab or Bitbucket is the active service, what happens to the Backlog Git parts (the Backlog Git pull request list and links, and the "Git access" form whose credentials let task worktrees clone Backlog Git repositories)?

A. Turn off all Backlog Git source control features, including the Git access form (strict: one service only)
B. Turn off Backlog Git pull request features, but keep the Git access form, because cloning Backlog Git repositories is separate from pull requests
X. Other (please specify)

[Answer]: A

## Question 9
Should the service selector offer "None" (no source control service at all)?

A. No - the choices are Backlog Git, GitHub, GitLab, Bitbucket; Backlog Git is the default
B. Yes - "None" turns off every source control feature
X. Other (please specify)

[Answer]: A
