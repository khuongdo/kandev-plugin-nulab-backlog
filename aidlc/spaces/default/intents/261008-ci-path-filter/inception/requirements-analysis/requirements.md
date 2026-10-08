# Requirements — CI path filter

## Intent Analysis

Initial request (verbatim): "chỉ trigger ci và release đổi với những thư mục liên quan đến app, loại trừ aidlc và docs.. ra"

Goal: stop spending the full CI run (Go/UI checks, build, package, packaged-host contract test against a throwaway Kandev server) on pull requests and `main` pushes that change only non-app files, such as AI-DLC records (`aidlc/`) and documentation (`docs/`). Pull requests #3, #5, #7, #10 and #12 were records-only and still ran the full CI. Pull requests must still be mergeable under the `main` ruleset, and the repo-wide credential scan must keep running on every change.

- **Type**: enhancement to CI configuration (`.github/workflows/`), no application code change.
- **Scope**: single component (CI workflows; possibly `Makefile` wiring).
- **Complexity**: simple. Depth: Minimal.

Sources: initial request; developer scan `inception/reverse-engineering/developer-scan.md`; CodeKB `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`; answers Q1–Q6 in `requirements-analysis-questions.md`.

## Functional Requirements

### FR1 — App vs non-app path classification
- **FR1.1** A change is **non-app-only** when every changed file path is under `aidlc/`, `.claude/` or `docs/`, or is exactly `README.md`, `LICENSE` or `.gitignore` (Q4 = A).
- **FR1.2** Every other path is **app**, including `.github/workflows/`, `Makefile`, `server/`, `internal/`, `cmd/`, `ui/`, `manifest.yaml`, `go.mod`, `go.sum`, `.kandev-sdk-ref`, `.nvmrc` and `.golangci.yml`.
- **FR1.3** The classification is defined in exactly one place, which both the change-detection logic and its tests use.

Acceptance:
- Given a pull request that changes only `aidlc/spaces/default/intents/x/aidlc-state.md` and `docs/brand/note.md`, when CI classifies it, then the result is non-app-only.
- Given a pull request that changes `docs/brand/note.md` and `internal/backlog/client.go`, when CI classifies it, then the result is app.
- Given a pull request that changes only `.github/workflows/ci.yml`, when CI classifies it, then the result is app.

### FR2 — App CI runs only for app changes (Q1 → Q5 = A)
- **FR2.1** `ci.yml` keeps triggering on every `pull_request` to `main` and every `push` to `main`, with no `paths`/`paths-ignore` filter on the trigger.
- **FR2.2** A lightweight change-detection job computes FR1 from the changed files, using `git diff --name-only` (no new third-party action): against the pull request base for `pull_request`, and from the previous `main` commit to the pushed commit for `push`.
- **FR2.3** The existing heavy jobs `checks` and `packaged-host-contract` keep their exact job names and run only when the change-detection result is app; otherwise they are skipped through `if:`, which GitHub reports as a passing required check.
- **FR2.4** If the change set cannot be determined (for example, an unknown previous commit on a push, or a failed diff), CI treats the change as app and runs the full jobs (fail-safe).

Acceptance:
- Given a records-only pull request (only `aidlc/` files), when CI runs, then `checks` and `packaged-host-contract` are reported as skipped, and the pull request is mergeable under the ruleset with no manual override.
- Given a pull request that changes `internal/` and `docs/`, when CI runs, then `checks` and `packaged-host-contract` run in full, and a failure blocks merge.
- Given a push to `main` whose diff cannot be computed, when CI runs, then the full jobs run.

### FR3 — Separate credential-scan workflow (Q3, Q6)
- **FR3.1** A new, separate workflow runs the repo-wide credential scan (`make check-secrets`, equivalent to `go run ./cmd/ci secrets -root .`) on every `pull_request` to `main` and every `push` to `main`, with no path filter, so non-app-only changes (including `aidlc/` and `docs/`) are still scanned.
- **FR3.2** The credential-scan job has a stable, unique job name that does not collide with `checks` or `packaged-host-contract`, so it can be added as a required status check.
- **FR3.3** The scan must not run the Kandev build, packaging or contract test.

Acceptance:
- Given a records-only pull request that adds a file under `aidlc/` containing a test credential pattern the scanner flags, when the credential-scan workflow runs, then it fails.
- Given a records-only pull request without credentials, when the credential-scan workflow runs, then it passes, while `checks` and `packaged-host-contract` are skipped.

### FR4 — Release unchanged (Q2 = A)
- **FR4.1** `release.yml` keeps its trigger (`push` of `v*` tags) and all its jobs and checks unchanged. A release runs only when a tag is created deliberately; no path filter is added, since GitHub ignores path filters on tag pushes.

Acceptance:
- Given the change is merged, when `.github/workflows/release.yml` is compared with `main` before the change, then it is unchanged.

### FR5 — Required-check update (manual, outside the repo) (Q6 = A)
- **FR5.1** After the credential-scan workflow has reported at least once, the `main` ruleset (id 24580280) adds its check to the required status checks, keeping `checks` and `packaged-host-contract`. You make this change in GitHub settings (or with a `gh api` command provided in the release notes or the pull request description); the repo cannot apply it.

Acceptance:
- Given the ruleset is updated, when `gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280` is read, then its required status checks are `checks`, `packaged-host-contract` and the credential-scan check name.

## Non-Functional Requirements

- **NFR1 (security / workflow policy)**: Every workflow, including the new one, passes `make lint` unchanged: every `uses:` pinned to a full 40-character commit SHA, top-level `permissions: contents: read`, no `pull_request_target`, no write permissions outside `release.yml` job `publish`, and actionlint clean.
- **NFR2 (efficiency)**: On a non-app-only pull request, no job checks out or builds Kandev, builds the plugin, or runs the contract test; only change detection and the credential scan run.
- **NFR3 (no new dependencies)**: No new third-party GitHub Action and no new Go or npm dependency.
- **NFR4 (testability)**: The FR1 classification has an automated test covering at least the three FR1 acceptance cases, run by `make test` (Go `-race`, testify, table-driven, following team Testing Posture with TDD ordering), and the classification code counts toward the 80% coverage floor if it lives under `internal/`.
- **NFR5 (no secrets)**: No real credential is added to the repo or test data; scanner test inputs use obviously fake values.

## Constraints

- `main` ruleset 24580280 requires `checks` and `packaged-host-contract` (squash-only, `strict_required_status_checks_policy: false`); their job names must not change.
- Repo workflow policy `internal/ci/workflows.go` (run by `make lint`) applies to every file in `.github/workflows/`.
- CI calls the standard `Makefile` targets so local and CI results match (team Code Style).
- Trunk-based, short-lived branch, pull request with self-merge on green CI (team Way of Working).

## Assumptions

- [assumption] `check-secrets` is removed from the `checks` job's `make` list once the separate workflow runs it on every change, to avoid scanning twice. If you prefer to keep it in both, that is a one-line difference.
- [assumption] The credential-scan workflow may still need the Kandev SDK checkout at `.kandev-sdk-ref` because `check-secrets` depends on `check-sdk` (the module uses a `replace` to `../kandev`). That checkout is a git clone, not a Kandev build, and is allowed by FR3.3. Code Generation decides whether to call the scanner without the `check-sdk` prerequisite.
- [assumption] Applying the same classification to `push` to `main` is intended (the request says "trigger CI"), so records-only merges to `main` also skip the heavy jobs.

## Out of Scope

- Any change to `release.yml` or release preflight behaviour (Q2 = A).
- De-duplicating `ci.yml` and `release.yml` setup steps, adding a `concurrency` group to `ci.yml`.
- Changing which checks the release workflow runs.

## Open Questions

- Exact name of the credential-scan workflow and job (for example workflow `secrets`, job `secret-scan`); settled in Code Generation and given to you for the ruleset update (FR5).
