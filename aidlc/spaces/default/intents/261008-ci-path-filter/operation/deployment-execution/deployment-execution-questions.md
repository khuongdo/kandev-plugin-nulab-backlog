# Deployment Execution Questions — 261008-ci-path-filter

Pre-deployment checks (2026-10-08): Build and Test passed locally (all standard targets, contract test on Kandev v0.96.0, coverage 92.8%). No database migrations, no dependent services, no plugin release. Latest release is still `v0.4.2`; `origin/main` is `3a983ab`. Execution follows the approved `cd-config.md`: commit, rebase onto `origin/main`, push, pull request, wait for CI, add `secret-scan` to ruleset 24580280, squash-merge, then the records-only pull request for `261007-plugin-install-502`.

## Question 1
Which files go into this pull request besides the code change (`.github/workflows/ci.yml`, `.github/workflows/secrets.yml`, `internal/ci/*`, `README.md`)?

A. Also this intent's AI-DLC records (`aidlc/spaces/default/intents/261008-ci-path-filter/`), the updated code knowledge base, and only this intent's new row in `intents.json`, like earlier code pull requests (for example #13); the `261007` records stay for their own pull request
B. Code only; this intent's records go into a later records-only pull request together with the `261007` records
X. Other (please specify)

[Answer]: A

## Question 2
Should I carry out the whole sequence now, including the squash merge to `main` once the three checks are green and the ruleset is updated?

A. Yes, run it all now and stop only on a failure
B. Run up to the green pull request and the ruleset update; I merge myself
X. Other (please specify)

[Answer]: A
