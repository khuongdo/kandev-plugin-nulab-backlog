# CI/CD Pipeline — walking-skeleton (U1)

Inputs:

- U1 NFR design: `security-design`, `reliability-design`, `logical-components`, `performance-design`, `scalability-design`, `observability-design`.
- `components`.
- `functional-spec` (U1).
- `contract-summary` (C4).
- `team-practices` (Way of Working, Code Style, Testing Posture, Deployment).
- Answers Q1–Q3 in `infrastructure-design-questions.md`.

U1 sets up a minimal CI [Q2]: one workflow, one required check. U5 extends the same workflow with the packaged-host contract test, the secret check and the release workflow (`bolt-plan`, B2).

## Workflow `.github/workflows/ci.yml`

- **Triggers**:
  - `pull_request` to `main`;
  - `push` to `main`.
  - Never `pull_request_target` (`team-practices`, Deployment).
- **Permissions**: `contents: read` at the workflow level. No other permission in U1.
- **Runner**: `ubuntu-latest`.
- **Actions**: every action is pinned to a full commit SHA, with the version in a comment.
- **Job `checks`**: this is the single required check.
  1. Check out this repository into `plugin/`.
  2. Read `plugin/.kandev-sdk-ref`. Check out `kdlbs/kandev` at that exact commit into `kandev/`, so that `plugin/../kandev` resolves for the `go.mod` replace (BR5.5).
  3. Set up Go at the version in `plugin/go.mod` (1.26). Set up Node.js at the version in `plugin/.nvmrc`, with an npm cache.
  4. Install UI dependencies with `npm ci` in `plugin/ui`.
  5. Run `go mod tidy`, then fail if `go.mod` or `go.sum` changed (`team-practices`, Code Style).
  6. Run `make check-format vet lint test coverage build package verify-package` in `plugin/`:
     - `test` and `coverage` use `-race`;
     - `coverage` fails below 80% (NFR8.1);
     - `build` cross-compiles the 5 targets (NFR7.1);
     - `verify-package` reads `ui/bundle.js` from the package and fails if it finds an asset load from a Nulab or Backlog domain: an `http(s)` URL on `nulab.com`, `nulab-inc.com`, `backlog.com`, `backlog.jp` or `backlogtool.com` that ends in `.svg`, `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp` or `.ico`, or any `url(` or `src=` that points to those domains. Plain host names in message text (for example the `myteam.backlog.com` placeholder) do not match (NFR3.10). The check lives in `internal/pkgverify`, with unit tests for a matching and a non-matching bundle.
  7. Upload `dist/` as a workflow artifact. Pull requests from forks also get the artifact, but there are no secrets to expose.

`golangci-lint` is installed at a pinned version inside the `lint` target, so CI and local runs use the same version.

## Stage-to-Gate Mapping

| Stage | Gate | On failure |
|-------|------|------------|
| Dependency tidiness | `go mod tidy` leaves no diff | Check fails, merge blocked |
| Format, vet, lint | `check-format`, `vet`, `lint` (golangci-lint + gosec, tsc, ESLint, Prettier) | Check fails, merge blocked |
| Tests | `test` with `-race`, Vitest | Check fails, merge blocked |
| Coverage | `coverage` ≥ 80% of `./internal/...` and `./server/...` | Check fails, merge blocked |
| Build and package | `build`, `package`, `verify-package` (including the bundle URL check) | Check fails, merge blocked |

## Repository Settings (set by you once, external dependency X4)

- `main` is protected: no direct pushes, no force-push, no deletion.
- Required status check: `checks` from `ci.yml`.
- Merge method: squash only (`team-practices`, Way of Working).

## Deployment and Promotion

- No automatic deployment, following `team-practices` (Deployment).
- After the skeleton's pull request is green, an admin takes the `dist/` package from the workflow artifact or a local build.
- The admin installs it on the existing self-hosted Kandev 0.96.0 server [Q1, Q3] and runs the first manual check.
- Tag-based releases (`vX.Y.Z`, `release.yml`, attestation) come with U5.

## Rollback

There is no in-place deployment to roll back. If an installed skeleton build misbehaves, the admin uninstalls it or installs the previous package through Settings > Plugins. Stored connection and switch data stay compatible because both use `schemaVersion` 1. A build without the switch ignores the `integration` key, so going back to the previous package turns Backlog on everywhere until the newer build is reinstalled.

## Secrets in CI

None. CI never uses a Backlog key or any other secret. Tests run against the fake Backlog with random per-run test keys (`security-design`, NFR4.1). The workflow needs no repository secrets.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q3]: answers in `infrastructure-design-questions.md`.
- U1 NFR design (all six files); `functional-spec.md` (U1); `components.md`; `contract-summary.md`; `team-practices.md`; `bolt-plan.md`; `external-dependency-map.md`.

## Assumptions & Open Questions

- [assumption] Checking out `kdlbs/kandev` at a pinned commit needs no token, because the Kandev repository is public.
