# Practices Discovery — Interview Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

This is a brand-new project with no code yet. The suggestions below come from the organisation defaults, the release engineer's draft, and the comments of the quality engineer, developer, and security engineer. The last three checked directly against the Bitbucket template plugin repo. Your answers become the project's official way of working.

---

## Q1. How changes get into the main branch

You work alone. How should each change get into the `main` branch?

- A. Open a pull request from a short-lived branch, self squash-merge when CI is green; turn on `main` protection so direct pushes are blocked
- B. Like A but without `main` protection
- C. Push directly to `main`, no pull request
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Build a thin end-to-end slice first

Build a thin end-to-end slice first? A walking skeleton is a minimal version that runs the whole way through, built first to prove the pieces connect before the real features go in. For this plugin, what does the thin slice contain?

- A. Yes: minimal Go part, connect a space with an API key, package, verify the package, install on the self-hosted Kandev server, and call Backlog successfully. OAuth later
- B. Yes, and include OAuth login too
- C. No thin slice first
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Test order

Write tests before or after the code?

- A. Write code piece by piece, then write and run tests for that piece (test-after, the organisation default)
- B. Write tests first, then the code (TDD)
- C. Mixed: tests first for business logic, tests after for the rest
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q4. Quality checks in CI (select all that apply)

The template repo does not measure coverage and runs no security checks. What do you want to add?

- A. An 80% coverage floor for the Go code; CI blocks if lower (organisation default)
- B. Run Go tests with data race detection between threads (`-race`)
- C. Dependency vulnerability scan (`govulncheck` for Go, `npm audit` for the UI) and turn on Dependabot
- D. Automatically install the real package on Kandev at the minimum version and run it, like the template repo
- E. Manual end-to-end check against a real Backlog space before every release
- X. Other (please specify)

[Answer]: A, B, D (E: không)

## Q5. How to release

How should new versions be released?

- A. Tag `vX.Y.Z` on `main` → an automated pipeline creates a GitHub Release with a package provenance attestation → send a marketplace update proposal. Install on the self-hosted server by hand
- B. Like A, and also automatically build the package after each merge into `main` for trial installs
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q6. Code conventions for the TypeScript UI

The template repo uses only TypeScript's strict type check (`tsc --noEmit`), with no ESLint or Prettier. What should the UI use?

- A. Same as the template: strict type check only
- B. Add ESLint and Prettier
- C. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q7. Code conventions for the Go code

The template repo uses `gofmt` and `go vet`. Add any other code checking tools?

- A. Add `golangci-lint`, including the `gosec` security checks
- B. Only `gofmt` and `go vet`, same as the template
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q8. Hard rules (select all that apply)

A hard rule is something that must always be done or must never be done. Which rules do you want to apply?

- A. NEVER put real credentials in the repo, test data, or test artifacts
- B. ALWAYS redact API keys and tokens in logs, error messages, and test output (the Backlog API key travels in the URL, so it is easily printed)
- C. ALWAYS accept only `https` space addresses under `backlog.com`, `backlog.jp`, or `backlogtool.com`, so credentials are never sent to unknown hosts
- D. NEVER delete or overwrite a released tag
- E. ALWAYS run package verification before tagging a release
- X. Other (please specify)

[Answer]: A, B, C, D, E

## Q9. Clarification: verifying against a real Backlog space

In Q4 you did not choose a manual end-to-end check before every release (E). But one success criterion in `intent-statement` is "connects to a real Backlog space and runs the main flow end-to-end". How will this criterion be verified?

- A. Manual check against a real space at two points: when the thin slice is done and before the first release; later versions rely only on automated tests
- B. An automated end-to-end test suite running against a real space (using CI secrets), run on demand or before release
- C. Rely only on tests against the fake Backlog server; the real-space criterion counts as met when the thin slice calls successfully
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers (including the Q9 clarification):

- Q1: A. Every change goes through a pull request on a short-lived branch, self squash-merge when CI is green; `main` protection on
- Q2: A. Build a thin end-to-end slice first: minimal Go part, connect a space with an API key, package, verify the package, install on self-hosted Kandev, and call Backlog. OAuth later
- Q3: B. Write tests first, then the code (TDD)
- Q4: A, B, D. Minimum 80% coverage for Go (CI blocks); Go tests with `-race`; automatically install the real package on the minimum Kandev version and run it. No dependency vulnerability scan (C); no manual check before every release (E)
- Q5: A. Tag `vX.Y.Z` on `main` → automatically create a GitHub Release with a provenance attestation → marketplace update proposal; manual install on the self-hosted server
- Q6: B. The TypeScript UI also uses ESLint and Prettier
- Q7: A. The Go code adds `golangci-lint`, including `gosec`
- Q8: A, B, C, D, E. All five hard rules: no real secrets in the repo/tests; always redact API keys and tokens; accept only `https` space addresses under Backlog domains; never delete/overwrite a released tag; always verify the package before tagging
- Q9: A. Manual check against a real Backlog space twice: when the thin slice is done and before the first release

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
