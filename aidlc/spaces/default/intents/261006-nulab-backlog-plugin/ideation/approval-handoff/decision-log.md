# Decision Log — Ideation Phase

A summary of the decisions settled in the ideation phase. The sources are the question files and documents of each step: `intent-statement`, `stakeholder-map`, `competitive-analysis`, `feasibility-assessment`, `constraint-register`, `scope-document`, `intent-backlog`, `wireframes`. There is no `team-assessment` because the Team Formation step was skipped.

## Decisions

| ID | Step | Decision | Source |
|----|------|-----------|-------|
| D-01 | Intent Capture | The plugin covers both areas: issues and Git/pull requests | Q1 `intent-capture-questions.md` |
| D-02 | Intent Capture | The target users are Kandev users in general who use Backlog | Q2 |
| D-03 | Intent Capture | Three success criteria: runs end-to-end with a real space, passes the same checks as the Bitbucket plugin, has a usable release | Q3 |
| D-04 | Intent Capture | Only the proposer decides the scope; progress updates via pull requests and commits | Q6, Q7 |
| D-05 | Intent Capture | The product boundary is `feature`: a complete plugin like the Bitbucket plugin | Q8 |
| D-06 | Market Research | Compare with the Bitbucket plugin only | Q1 `market-research-questions.md` |
| D-07 | Market Research | Both the basic level (connection, viewing issues/repositories) and linking to tasks are required | Q2 |
| D-08 | Market Research | Build it ourselves even though other solutions exist; no market size estimate | Q3, Q4 |
| D-09 | Market Research | Consider the Backlog API terms and limits; research based on cited public documentation | Q5, Q6 |
| D-10 | Feasibility | Support all Backlog domains | Q3 `feasibility-questions.md` |
| D-11 | Feasibility | Both API key and OAuth 2.0 from the first release | Q4 |
| D-12 | Feasibility | No hard deadline or cost limit; no specific compliance requirements | Q6, Q7 |
| D-13 | Feasibility | Release via GitHub Release and submit a proposal to the Kandev marketplace | Q8 |
| D-14 | Scope Definition | Issue features: create a task from an issue, `#` reference (plus viewing and linking, already required) | Q1, Q7 `scope-definition-questions.md` |
| D-15 | Scope Definition | Git/PR features: repository source, create PR, PR status, PR watch and dashboard; no review panel | Q2 |
| D-16 | Scope Definition | One-way status sync from Backlog to Kandev; no issue status updates from Kandev | Q3, Q7 |
| D-17 | Scope Definition | One space per Kandev workspace | Q4 |
| D-18 | Scope Definition | Out of scope: wiki, Subversion, Gantt, notifications, webhook | Q5, Q8 |
| D-19 | Scope Definition | Order: risk first | Q6 |
| D-20 | Team Formation | Skip this step because one person works with AI | Team Formation applicability question (audit) |
| D-21 | Rough Mockups | Entry points like the Bitbucket plugin; add a Backlog entry under "Integrations" (Issues, PR watches, Dashboard) | Q1, Q8 `rough-mockups-questions.md` |
| D-22 | Rough Mockups | Issue list page with filters; single connection form; Kandev style | Q2, Q3, Q4 |
| D-23 | Rough Mockups | Desktop and phone support; text follows Kandev's language; basic WCAG 2.1 AA | Q5, Q6, Q7 |
| D-24 | Approval & Handoff | Accept risks R1, R2, R3, R6 with the proposed mitigations | Q1 `approval-handoff-questions.md` |
| D-25 | Approval & Handoff | Create the test space and register OAuth during the construction phase | Q2 |
| D-26 | Approval & Handoff | Open comments on the sketches move to Refined Mockups | Q3 |
| D-27 | Approval & Handoff | MIT licence | Q4 |
| D-28 | Approval & Handoff | Move on to the specification phase | Q5 |

## Deferred decisions

| Item | Deferred to |
|-----|----------|
| Whether each issue links to at most one task or to several tasks | Requirements Analysis |
| Polling frequency for status sync | Requirements Analysis |
| The product lead's 7 comments on the sketches | Refined Mockups |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- The `*-questions.md` files in this intent's `ideation/`.

## Assumptions & Open Questions

None.
