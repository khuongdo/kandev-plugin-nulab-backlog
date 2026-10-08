# Requirements Analysis Questions

Intent: run CI and release only for changes in app-related directories, excluding `aidlc/`, `docs/` and similar non-app paths.

## Question 1
The `main` branch ruleset requires the status checks `checks` and `packaged-host-contract`. If `ci.yml` only gains a `paths-ignore` filter, a pull request that changes only `aidlc/` or `docs/` never starts CI, those two required checks stay "waiting" forever, and the pull request cannot be merged. How should CI skip non-app changes while still letting such pull requests merge?

A. Keep CI triggering on every pull request and push to `main`; add a small change-detection job (plain `git diff --name-only`, no new third-party action) and skip the heavy jobs with `if:` when only non-app files changed (GitHub counts a job skipped by `if:` as passing a required check)
B. Add `paths-ignore` to `ci.yml` and a second "stub" workflow with the inverse filter whose jobs have the exact same names and always succeed
C. Add `paths-ignore` to `ci.yml` and remove the two required checks from the `main` ruleset
X. Other (please specify)

[Answer]: B

## Question 2
Release runs only when you push a `vX.Y.Z` tag, and GitHub ignores path filters on tag pushes, so a path filter on `release.yml` would have no effect. What should "release only for app changes" mean?

A. Leave the release trigger unchanged: release already runs only when you deliberately create a tag, so only CI changes
B. Make the release preflight refuse a tag when no app path changed since the previous release tag
X. Other (please specify)

[Answer]: A

## Question 3
The credential scan (`check-secrets`) currently runs inside CI on every pull request and scans the whole repo, including `aidlc/` and `docs/`. The project forbids real credentials anywhere in the repo. When CI skips the heavy jobs for a non-app-only change, what should happen to the credential scan?

A. Keep running the credential scan on every pull request and push, including non-app-only changes (fast, no Kandev build)
B. Skip it too; non-app-only changes get no checks at all
X. Other (please specify)

[Answer]: X. Other: "tạo 1 CI để quét riêng, ci chạy check và test app chỉ chạy khi có thay đổi liên quan tới app" (create a separate CI workflow for the credential scan; the CI that runs app checks and tests only runs when app-related files change)

## Question 4
Which paths count as "not app" (changes only in these skip the heavy CI jobs)? Everything else, including `.github/workflows/` and `Makefile`, stays "app".

A. `aidlc/`, `.claude/`, `docs/`, `README.md`, `LICENSE`, `.gitignore`
B. Only `aidlc/` and `docs/`, exactly as in the request
C. Option A without `README.md` (keep README changes running full CI)
X. Other (please specify)

[Answer]: A

## Follow-up Question 5
With the stub-workflow approach chosen in Question 1 (B), GitHub path filters can only express "run when at least one changed file matches", not "run only when every changed file is non-app". A pull request that changes both app code and `docs/` therefore starts both workflows: `ci.yml` runs the real checks while the stub reports a same-named `checks` success immediately. With two same-named checks, a failing real check might be masked by the passing stub. How should we proceed?

A. Switch to the change-detection job (Question 1 option A): every pull request starts `ci.yml`, a small job decides whether app files changed, and the heavy jobs skip with `if:`; the credential scan lives in its own separate workflow as you asked in Question 3
B. Keep the stub workflow and accept the mixed-change risk
X. Other (please specify)

[Answer]: A

## Follow-up Question 6
The new separate credential-scan workflow (Question 3) will report its own check. Should it become a required check on `main` (a ruleset change you make in GitHub settings, since it is outside the repo)?

A. Yes, make it required so a pull request with a credential cannot merge
B. No, keep it advisory
X. Other (please specify)

[Answer]: A
