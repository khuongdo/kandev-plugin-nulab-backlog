# Risk and Sequencing Rationale — Kandev Plugin for Nulab Backlog

This file explains why the Bolts in `bolt-plan.md` are in that order. A **Bolt** is one build pass over part of the work that ends in something that runs.

Inputs: `unit-of-work-dependency` (the dependency graph), `unit-of-work`, `unit-of-work-story-map`, `requirements`, `stories`, `mockups`, `components`, `contract-summary`, `team-practices`, and answers Q1–Q10 in `delivery-planning-questions.md`.

## Heuristic

Two rules decide the order:

1. **Walking skeleton first.** The first Bolt is the smallest slice that runs end to end (Cockburn). This is required by `team-practices` and is fixed by the graph, because U1 has no dependencies.
2. **WSJF** (Weighted Shortest Job First, Reinertsen / SAFe) [Q4] ranks the rest. Each Bolt gets a score on the Fibonacci scale (1, 2, 3, 5, 8, 13) for:
   - user-business value;
   - time criticality;
   - risk reduction.

   The sum is divided by job size. The three criteria have equal weight [Q10]. Size uses M = 5 and XL = 13, from `unit-of-work`.

## WSJF Scores

| Bolt | Value | Time criticality | Risk reduction | Size | WSJF |
|------|-------|------------------|----------------|------|------|
| B1 walking-skeleton (U1) | 5 | 13 | 13 | 5 (M) | 6.2 |
| B2 ci-release (U5) | 3 | 8 | 8 | 5 (M) | 3.8 |
| B3 connection (U2) | 8 | 8 | 13 | 8 (L) | 3.6 |
| B4 issues (U3) | 13 | 5 | 5 | 13 (XL) | 1.8 |
| B5 git-pr (U4) | 8 | 3 | 5 | 13 (XL) | 1.2 |

Why each Bolt scores as it does:

- **B1**: highest time criticality and risk reduction. Nothing else can start before it, and it tests the two biggest unknowns: Kandev compatibility and reaching the real Backlog API [Q5: A].
- **B2**: high time criticality. Every later pull request should pass through real gates [Q1]. The contract test on `min_kandev_version` guards against Kandev changing its plugin SDK [Q5: E].
- **B3**: highest risk reduction after B1. It tests OAuth on Backlog [Q5: B] and the rate-limit and time-budget rules from `contract-summary`. U3 and U4 both depend on it.
- **B4**: the highest user value (all Must stories), but the largest job.
- **B5**: optional value, a large job, and lower urgency.

## Fit with the Dependency Graph

The WSJF order (B1, B2, B3, B4, B5) is also a valid topological order of `unit-of-work-dependency`, so there is no deviation to justify.

Two choices inside the graph's freedom:

- **B2 before B3.** The graph allows U5 and U2 to run in either order or together. Q1 puts CI first. With parallel batches [Q7], both may be coded together, but B2 is merged first.
- **B4 before B5.** Both depend only on U2. B4 scores higher, and its functional design settles contract findings R-03, R-04 and R-05, which B5 reuses (the `task.deleted` contract and the time budget).

## Risk Register

| ID | Risk | Likelihood | Impact | Mitigation | Bolt |
|----|------|------------|--------|------------|------|
| R-A | The plugin package does not install or run on self-hosted Kandev (manifest, SDK `replace`, capabilities) | Medium | High | Walking skeleton first; install on a real self-hosted server; the skeleton checkpoint needs your approval | B1 |
| R-B | Backlog OAuth does not fit: no PKCE, the callback must go through a public Kandev webhook, `localhost` may be refused | Medium | High | Settle contract findings R-01, R-02 and R-08 in B3's functional design; register the OAuth app early in B3; API-key login stays as the fallback | B3 |
| R-C | Rate limits and the 15-second Kandev action limit break multi-call actions | Medium | Medium | Settle R-03 in B4's functional design; fake-server tests for 429 with `Retry-After` | B3, B4 |
| R-D | U3 is too large for one Bolt | Medium | Medium | Kept as one Bolt [Q3]; the stories inside U3 are ordered in `unit-of-work-story-map`, and functional design can split the work into steps | B4 |
| R-E | Kandev changes its plugin SDK during the build | Medium | High | Pin the SDK commit in `.kandev-sdk-ref`; run the contract test on `min_kandev_version` in CI from B2 on; only `internal/plugin` imports the SDK (ADR-007) | B2 onwards |
| R-F | The first release is delayed because it waits for the optional U4 [Q2, Q9] | Medium | Medium | Accepted by you. U4 starts in parallel with U3 to shorten the wait | B5 |
| R-G | Pull requests are not available on the test space's Backlog plan (assumption A1 in `requirements`) | Low | High | Check while creating the test space before B5 starts | B5 |

You named R-A, R-B and R-E as your top worries [Q5]. All three are tackled in the first three Bolts.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q10]: answers in `delivery-planning-questions.md`.
- `bolt-plan.md`; `unit-of-work.md`; `unit-of-work-dependency.md`; `unit-of-work-story-map.md`; `contract-summary.md`; `components.md` and ADR-007; `requirements.md`; `stories.md`; `mockups.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] The WSJF scores are relative judgements made in this stage, not measured values.
