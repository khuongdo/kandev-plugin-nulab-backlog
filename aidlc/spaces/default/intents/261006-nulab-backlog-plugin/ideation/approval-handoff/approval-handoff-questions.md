# Approval & Handoff — Approval Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Two questions do not need to be asked again:

- Stakeholder agreement: you alone decide the scope (Q6 in the Intent Capture step).
- Team plan: you work alone with AI, so the Team Formation step was skipped.

The questions below settle the decision to move on to the specification phase.

---

## Q1. Accepting the main risks

These are the biggest risks in `raid-log.md`, each with a proposed mitigation:

- R1 Unfamiliar Go part: follow the Bitbucket template, keep the required checks.
- R2 Scope larger than the template: boundaries already settled in the scope step.
- R3 API rate limits: set a rate limit inside the plugin.
- R6 Kandev changes the plugin programming interface: test on the minimum Kandev version.

Do you accept these risks with the mitigations above?

- A. Accept
- B. Accept, but one risk needs an extra mitigation (state it under Other)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Preparing resources before building

Two things must be done before end-to-end testing: create a test Backlog space (Free plan or trial) and register an OAuth application with Nulab. When will you do them?

- A. Before the construction phase starts
- B. During the construction phase, when needed
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q3. UI sketches

You approved the sketches with 7 open review comments (for example, missing accessibility notes, missing pull request linking screen). Where are these comments handled?

- A. Move them to the Refined Mockups step in the specification phase
- B. No need to handle them
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Plugin licence

The licence has not been chosen. The Bitbucket plugin uses MIT, and Kandev uses AGPL-3.0. Which licence should the Backlog plugin use?

- A. MIT, like the Bitbucket plugin
- B. AGPL-3.0, like Kandev
- C. Decide later
- X. Other (please specify)

[Answer]: A

## Q5. Go/no-go decision

Based on all results of the ideation phase, what is your decision?

- A. Move on to the specification phase (Inception)
- B. Move on with conditions (state them under Other)
- C. Stop the initiative
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: A. Accept the main risks (R1, R2, R3, R6) with the proposed mitigations
- Q2: B. Create the test Backlog space and register the OAuth application during the construction phase, when needed
- Q3: A. The 7 open comments on the UI sketches move to the Refined Mockups step
- Q4: A. The plugin uses the MIT licence
- Q5: A. Move on to the specification phase (Inception)

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
