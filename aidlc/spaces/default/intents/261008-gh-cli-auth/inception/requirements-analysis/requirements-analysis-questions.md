# Requirements Analysis Questions — gh CLI auth for GitHub source control

Initial description: "Hiện tại link với github source control chỉ có qua apikey. Thêm 1 phương thức nữa là bằng gh cli auth"

Context from the code scan: every GitHub call gets its token from one place (`scm.Service.credential`). A token from `gh auth token` works with the existing GitHub client unchanged. The `gh` command runs on the machine that runs the Kandev server (the plugin is a process of that server), not on the browser user's machine.

## Question 1
When GitHub is connected through the gh CLI, how should the plugin get the token?

A. Ask `gh` each time a token is needed (with a short in-memory cache, e.g. 5 minutes); the token is never stored by the plugin, and it follows `gh auth login` / `gh auth refresh` / account switches automatically
B. Ask `gh` once when connecting and store the token in the plugin's secret store like a typed token; it does not follow later gh changes (re-connect needed)
X. Other (please specify)

[Answer]: A

## Question 2
Which providers should get the CLI auth method?

A. GitHub only (`gh`); GitLab and Bitbucket stay token-only
B. GitHub (`gh`) and GitLab (`glab`); Bitbucket stays token-only
X. Other (please specify)

[Answer]: B

## Question 3
If the GitHub connection uses gh CLI and later `gh` stops working on the server (not installed, logged out, token revoked), what should happen?

A. GitHub calls fail with a clear message ("gh CLI is not available or not logged in on the Kandev server"); the card shows the problem; no automatic fallback
B. Fall back to a typed token if one was also saved; otherwise fail as in A
X. Other (please specify)

[Answer]: A

## Question 4
How should the admin choose the method in Settings > Source Control (GitHub card)?

A. Keep the token field and add a "Use gh CLI login" button next to it; using one method replaces the other; the card shows which method is active and the GitHub account name
B. A method selector (Token / gh CLI) first, then only the matching controls
X. Other (please specify)

[Answer]: A

## Interpretation notes

- Q2 = B adds GitLab (`glab`). Q1, Q3 and Q4 were asked about GitHub; they are applied the same way to GitLab with `glab` (live token lookup with a short cache, clear error and no fallback, a "Use glab CLI login" button next to the token field). Bitbucket is unchanged.
- No contradictions found between the answers.
