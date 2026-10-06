# Evidence — Practices Discovery

## What the participants checked or inferred

### Lead — aidlc-pipeline-deploy-agent

- Checked `aidlc-state.md`: Greenfield project, scope `feature`, Depth `Standard`, Test Strategy `Standard`. The repo has no commits yet (`git rev-parse HEAD` fails), and no source code or CI config.
- Checked `.claude/scopes/aidlc-feature.md`: declares `skeleton: on`.
- Read `memory/org.md` (five default sections), `memory/team.md` (all sections empty, so this is the first run), `memory/project.md` (only `## Corrections` from Ideation).
- Read `ideation/feasibility/constraint-register.md` (C-T1 Go required, C-T2 minimum Kandev version, C-T3/C-T4 API call limits, C-T5 space domains, C-T6 OAuth expires in 1 hour, C-O1 not yet familiar with Go, C-O2 no test space yet, C-O3 has a self-hosted Kandev server, C-O5 one decision maker, C-O6 marketplace conditions, C-R3 secret encryption, C-R4 license) and `ideation/scope-definition/scope-document.md` ("risk first" ordering, end-to-end success criterion with a real space).
- Initial inference: trunk-based + squash per org.md; map "staging" to the self-hosted server and "production" to a tagged GitHub Release; MIT license per the brief.

### aidlc-developer-agent (read the template repo `kdlbs/kandev-plugin-bitbucket` directly)

- Layout: one-line `server/main.go` + `internal/<package>`; only `internal/plugin` imports `pluginsdk`.
- The Kandev SDK is not yet released as its own module: `go.mod` uses `replace` to `../kandev/apps/backend`, the UI uses `file:../kandev/...`, and the SDK commit is pinned in `.kandev-sdk-ref`.
- Error handling through `ProviderHTTPError{Status, RetryAfter}` and `action_errors.go`; an `internal/redact` package; response size limit.
- Backlog-specific point: the API key is sent in the `apiKey` query parameter, and Go's `*url.Error` prints the full URL, so the template's redaction regex is not enough.
- Proposed an `internal/backlog` layout instead of `cloud`/`datacenter`, naming conventions, error-handling conventions — merged into Code Style.

### aidlc-quality-agent (read the template repo directly)

- The template does **not** run `-race`, does **not** measure coverage, and has no coverage threshold.
- `ci.yml` has a `verify` job (`go mod tidy` check, format, vet, test, package-host, verify-package-host) and a `packaged-host-contract` job (builds Kandev at exactly `min_kandev_version`, installs the real package, runs Playwright) — this is how the template automates C-T2.
- `e2e-live` turns off Playwright traces, screenshots, and video.
- Proposed `make coverage`, the measured scope, table-driven tests, clock injection, and redaction tests — merged into Testing Posture.

### aidlc-devsecops-agent (read `.github/workflows/{ci,build,release}.yml`, `Makefile`, `scripts/verify-package.sh` directly via `gh api`)

- The template already pins actions to full SHAs, uses `permissions: contents: read`, does not use `pull_request_target`, pins the SDK, ships `checksums.txt`, releases only from `main`, and refuses existing tags.
- The template has no SAST, no dependency vulnerability scan, no secret scanning in CI, and no provenance attestation.
- Proposed `gosec` via `golangci-lint`, `govulncheck`/`npm audit`/Dependabot, `actions/attest-build-provenance`, a tag protection ruleset, and strict space host checks.

### Earlier assumptions corrected after checking the template repo

- The draft said the template toolchain was "per the brief, not accessed directly" → the three supporting agents read the repo directly; this assumption is replaced by the checked points above.
- The draft said `go test` "has `-race` and coverage, per the template" → wrong: both are **new additions** compared with the template.
- The draft treated Prettier + ESLint as the TypeScript default → the template uses only strict `tsc --noEmit`; adding ESLint + Prettier is your choice in Q6.
- The draft assumed CI builds the package after each merge into `main` → wrong: the template's CI runs only on `pull_request`; and you chose not to add this step (Q5 A).
- The draft missed the SDK dependency mechanism (`replace` + `.kandev-sdk-ref`), the `packaged-host-contract` job, and the `go mod tidy` check → added to the practices.

## Decisions in the interview

- [Q1] A — pull request on a short-lived branch, self squash-merge when CI is green, `main` protection on.
- [Q2] A — walking skeleton on: minimal Go backend, connect with an API key, package, verify the package, install on self-hosted Kandev, one successful Backlog call. OAuth later.
- [Q3] B — TDD.
- [Q4] A, B, D — 80% Go coverage floor (CI blocks), `go test -race`, automated contract test with the real package on the minimum Kandev version. **You declined** C (dependency vulnerability scan: `govulncheck`, `npm audit`, Dependabot) and E (manual end-to-end check before every release).
- [Q5] A — tag `vX.Y.Z` on `main` → GitHub Release with provenance attestation → PR updating the marketplace registry; manual install on the self-hosted server. **Not chosen**: B (automatic package build after each merge).
- [Q6] B — TypeScript adds ESLint and Prettier (on top of strict `tsc`).
- [Q7] A — Go adds `golangci-lint` including `gosec`.
- [Q8] A, B, C, D, E — all five hard rules, recorded in `discovered-rules.md`.
- [Q9] A — manual check against a real Backlog space exactly twice: when the walking skeleton is done and before the first release; after that, rely only on automated tests. This resolves the conflict between Q4-E (no) and the `intent-statement` success criterion "works with a real space".
- Consolidated summary confirmed: "Looks correct".
- MIT license: taken from the stage brief, recorded in the Deployment section.

## Candidates not adopted as hard rules

Candidates proposed by supporting agents but not in Q8 (for example pinning actions to SHAs, banning `pull_request_target`, mandatory `govulncheck`, writing a reproducing test before each bug fix, banning `time.Sleep` in tests, the error-handling rules) are **not** recorded as hard rules. The parts that do not conflict with your answers are kept as practices in `team-practices.md`; `govulncheck` is dropped entirely because you declined it in Q4.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, based on the Bitbucket plugin template.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q9] Answers in `practices-discovery-questions.md`, with the Consolidated Summary Confirmation section.
- `aidlc/spaces/default/memory/org.md`, `team.md`, `project.md`.
- `aidlc/spaces/default/intents/261006-nulab-backlog-plugin/aidlc-state.md`, `.claude/scopes/aidlc-feature.md`.
- `ideation/feasibility/constraint-register.md`, `ideation/scope-definition/scope-document.md`.
- `contributions/aidlc-developer-agent.md`, `contributions/aidlc-quality-agent.md`, `contributions/aidlc-devsecops-agent.md`.
- https://github.com/kdlbs/kandev-plugin-bitbucket (read directly by the three supporting agents on 2026-10-06).

## Assumptions & Open Questions

- [assumption] The MIT license is your choice per the stage brief; it was not asked again in the interview.
- There is no Backlog test space yet (C-O2); the two manual checks in Q9 depend on creating this space before the walking skeleton is done.
- The specific Construction verification command (for example `make verify-package-host` plus a recorded install and Backlog call step) will be chosen by you when entering Construction.
- The choice between `autonomous` and `gated` after the skeleton is approved will be asked after the skeleton checkpoint, not settled here.
- Not yet settled: Conventional Commits; auto-generated release notes or a hand-written `CHANGELOG.md`; the Go module name (GitHub path or bare name) and the repo owner; whether to split out `internal/domain`; commit `ui/bundle.js` or generate it at packaging time; whether the first release has a TypeScript UI.
- Not yet settled: the GitHub platform enforcement measures suggested by the security agent (`v*` tag protection ruleset, secret scanning + push protection, `SECURITY.md`). The no-delete/no-overwrite tag rule currently relies on discipline and the existing-tag refusal step in `release.yml`.
- The `packaged-host-contract` job takes about 30 minutes per run by the quality agent's estimate; not measured yet.
