# Delivery Planning — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

This step chooses the build order. Each **Bolt** is one build pass over part of the work (one or a few units). It ends with something that runs. The Bolt order must follow the dependency graph in `unit-of-work-dependency`:

- `walking-skeleton` (U1) comes before everything.
- `connection` (U2) and `ci-release` (U5) can only start after U1.
- `issues` (U3) and `git-pr` (U4) can only start after U2.

Some things were settled in earlier steps, so they are not asked again:

- The first Bolt is the **walking skeleton**: the smallest thin slice that runs end to end. It packages the plugin, installs it on self-hosted Kandev, connects with an API key, and calls the real Backlog (`team-practices`).
- `git-pr` is optional (Q4 in Units Generation).
- Only you work with the AI; the Team Formation step was skipped.
- The test Backlog space and the OAuth application are created during construction, when they are needed (Q2 in Approval & Handoff).

---

## Q1. Where CI and release (U5) go

After the walking skeleton, CI can be done at several points. Doing it early means every later pull request passes real quality gates. Doing it late delivers user value sooner.

- A. Do it right after the walking skeleton, before `connection`. From U2 on, every pull request passes real CI (format, lint, `-race` tests, 80% coverage, package verification)
- B. Do it alongside `connection`, finished before `issues` starts
- C. Do it after `issues`, just before the first release
- X. Other (please specify)

[Answer]: A

## Q2. The first release

The Git/PR part (U4) is optional. What should the first release include?

- A. Release `v0.1.0` when U1, U2, U3 and U5 are done (all the Must work), with the second manual check against a real space. Do U4 later and release it in the next version
- B. Wait until U4 is also done before the first release
- C. Not decided yet; decide after U3 is done
- X. Other (please specify)

[Answer]: B

## Q3. Bolt size

`issues` (U3) is size XL, with 13 stories. How should Bolts be split?

- A. Each unit is one Bolt, including U3
- B. Each unit is one Bolt, but U3 is split into two Bolts: (1) list, search, create task, link; (2) sidebar, `#` suggestions, status sync, accessibility, localization. This split exists only in the plan: at run time, U3 is still one unit
- C. Split U3 into two separate units (requires going back to the Units Generation step)
- X. Other (please specify)

[Answer]: A

## Q4. How to order the work

The order is almost fully decided by the dependency graph. Only the positions of U4 and U5 remain open. Should we score with a model such as WSJF (value plus urgency plus risk reduction, divided by size)?

- A. No need. Walking skeleton first, then prioritise risk reduction (connection, OAuth, API rate limits), then value (issues), with the optional part last. Record the reasons in words
- B. Yes, score each Bolt with WSJF
- X. Other (please specify)

[Answer]: B

## Q5. What worries you most (select all that apply)

Which risks should be verified earliest?

- A. The plugin is not compatible with Kandev: manifest, toolkit, installing the package on a self-hosted server (verified by the walking skeleton)
- B. Backlog OAuth: PKCE, callback through a public webhook, `localhost` during trial runs
- C. Backlog API rate limits, and Kandev's 15-second limit per action when one action needs several calls
- D. The U3 scope is too big for one person
- E. Kandev changes its plugin toolkit midway
- X. Other (please specify)

[Answer]: A, B, E

## Q6. Open points from the contract

You approved Contract Design and accepted 10 open review comments (R-01 to R-10). They include admin permission for OAuth, where to store the OAuth client id and secret, the time budget for multi-call actions, and moving the `task.deleted` event to blocks. When should they be handled?

- A. In the functional design step of the owning unit (U2 for R-01, R-02, R-06, R-08; U3 for R-03, R-04, R-05). They become conditions for that Bolt to count as done
- B. Go back and fix the contract documents now
- C. Handle them while writing code
- X. Other (please specify)

[Answer]: A

## Q7. Sequential or parallel during construction

In Units Generation (Q3), you allowed parallel work on units that do not depend on each other. The workflow is currently set to do one unit at a time from start to finish: design, code, verify, then move to the next unit. Do you want to keep this or change it?

- A. Keep one unit at a time. It is simple and easy to follow with only one reviewer. Parallel work is only allowed, not required
- B. Switch to working step by step across all units, and allow parallel code writing for independent units (for example U2 and U5) after the walking skeleton is approved
- X. Other (please specify)

[Answer]: B

## Q8. Who does the construction work

I can do all construction work in this session, using the mode chosen in Q7. The alternative is that each of your teams takes one unit and approves its own part.

- A. Do it in this session. Only you approve
- B. Each team takes one unit
- X. Other (please specify)

[Answer]: A

---

## Follow-up Questions

## Q9. If the Git/PR part (U4) is late or dropped

In Q2, you chose to wait until U4 is done before the first release. But in Units Generation (Q4), U4 is optional and can be dropped or done later. If U4 is very late or you decide to drop it, how should the first release be handled?

- A. Still wait for U4. Release only when the full Git/PR part is in
- B. Wait for U4 by default. If U4 is dropped or late, you decide at the U3 checkpoint whether to release the Must work first
- C. Switch: release the Must work first, with U4 in a later version (as in Q2-A)
- X. Other (please specify)

[Answer]: A

## Q10. WSJF weights

In Q4, you chose WSJF scoring. Each Bolt is scored on the Fibonacci scale (1, 2, 3, 5, 8, 13) for three criteria: user value, urgency, and risk reduction. The total is divided by size. How should the three criteria be weighted?

- A. Equal (standard SAFe WSJF)
- B. Double risk reduction, because you worry most about Kandev compatibility and OAuth (Q5)
- C. Double user value
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: do CI and release (U5) right after the walking skeleton, before `connection` (A).
- Q2: the first release waits until U4 is also done (B).
- Q3: each unit is one Bolt, including U3 (A).
- Q4: score the order with WSJF (B).
- Q5: risks to verify early: Kandev compatibility, Backlog OAuth, Kandev changing its toolkit (A, B, E).
- Q6: the contract's open review comments are handled in the functional design step of the owning unit, and are conditions for that Bolt to be done (A).
- Q7: switch to working step by step across all units, with parallel code writing for independent units after the walking skeleton is approved (B).
- Q8: do all construction work in this session, with only you approving (A).
- Q9: even if U4 is late or being considered for dropping, the first release still waits for U4 (A).
- Q10: the three WSJF criteria have equal weights (A).

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
