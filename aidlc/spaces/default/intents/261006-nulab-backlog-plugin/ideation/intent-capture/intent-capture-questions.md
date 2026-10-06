# Intent Capture — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Context note: the reference Bitbucket plugin is a Kandev runtime plugin. It connects a Kandev workspace to Bitbucket (repository discovery, linking pull requests to tasks, review screen, PR watch). Nulab Backlog is both an issue/project management tool and a Git host of its own, so "similar" can be read in several ways. The questions below help settle the direction.

---

## Q1. Problem to solve

The Bitbucket plugin ties Kandev to a Git platform (repos, pull requests). Nulab Backlog has both issue management and Git/pull requests. Which problem should the new Backlog plugin solve?

- A. Link Backlog issues to Kandev tasks (view/create/attach issues, sync status)
- B. Integrate Backlog Git/pull requests (repo discovery, link PRs to tasks, review), like the Bitbucket plugin
- C. Both: issues and Git/pull requests
- D. Not yet defined
- X. Other (please specify)

[Answer]: C

## Q2. Customers / users

Who will use this plugin, and what inconvenience do they face?

- A. An internal team that uses Nulab Backlog and wants to use Kandev with it
- B. Kandev users in general who use Nulab Backlog (public distribution)
- C. Both the internal team and external Kandev users
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q3. What success looks like (select all that apply)

Which measurable results show that the plugin has met its goal? (select all that apply)

- A. Connects to a real Backlog space and runs the main flow end-to-end from Kandev
- B. Passes checks equivalent to the Bitbucket plugin (build, test, packaging, package verification)
- C. Ships a usable release (an install package for Kandev)
- D. Not yet defined
- X. Other (please specify)

[Answer]: A, B, C

## Q4. Why now

What drives this initiative?

- A. The team uses Nulab Backlog and Kandev does not support it yet
- B. The Bitbucket plugin already exists as a template, so there is a chance to build a similar plugin quickly
- C. Both reasons above
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. Key stakeholders

Who has a stake in this outcome? (select all that apply)

- A. The initiative's proposer (you)
- B. Team members who use Nulab Backlog daily
- C. Maintainers of the Kandev project / the original Bitbucket plugin
- D. Not identified
- X. Other (please specify)

[Answer]: B

## Q6. Who decides scope and priority

Who has the authority to settle the plugin's scope or priority order, and who only gives input?

- A. Only the initiative's proposer decides
- B. The proposer decides, after consulting the Backlog user team
- C. Kandev maintainers must agree before it is settled
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q7. Communication / reporting needs

Is any update frequency or progress-reporting channel needed?

- A. Not needed; report only at the workflow's approval points
- B. Updates through pull requests/commits in this repo
- C. A dedicated channel or frequency (state it under Other)
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q8. Product scope confirmation

The workflow was selected with scope `feature` (workflow-selected). Does this scope match the product boundary you want?

- A. Matches: confirm `feature` as the product boundary (a complete plugin, similar to the Bitbucket plugin)
- B. Different: I want a smaller or different boundary (describe under Other, e.g. a read-only version first)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: C. Both (issues and Git/pull requests)
- Q2: B. Kandev users in general who use Backlog
- Q3: A, B, C (end-to-end flow with a real Backlog space; passes checks equivalent to the Bitbucket plugin; a usable release)
- Q4: A. The team uses Nulab Backlog and Kandev does not support it yet
- Q5: B. The team that uses Backlog daily
- Q6: A. Only the proposer decides
- Q7: B. Updates through pull requests/commits in this repo
- Q8: A. Confirm scope `feature`

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
