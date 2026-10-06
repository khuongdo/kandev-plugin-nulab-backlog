# Team-Level Rules

> This team's affirmed practices and corrections. Loaded after `org.md` as
> strict-additive guidance; contradictions with broader policy are rejected.
> Populated by the practices-discovery affirmation gate. Edit at the gate,
> not directly.

## Way of Working

- We work **trunk-based**: every change goes into `main` through a short-lived branch (resolved within 1-2 days). We keep no long-lived branches, including release branches.
- Every change goes through a GitHub **pull request**, even with only one person working. The pull request is where CI runs and where the change history is kept.
- We **self-merge with squash** when CI is green; a second reviewer is not required.
- The `main` branch is **protected**: no direct pushes, no force-push, no branch deletion, and every required CI check must be green before merge. This makes the quality gates real gates, not just advice.
- In Construction, each Bolt runs in its own worktree: base branch `main`, merge target `main`, squash merge — each Bolt becomes exactly one commit on `main`, named by the Bolt slug.
- The repo is on GitHub and public, because the Kandev marketplace requires a public GitHub repo.

## Walking Skeleton

- Walking skeleton: **on**. The first Unit is the smallest slice that runs end to end. It goes through every per-Unit stage (including Code Generation) before the other Units.
- The skeleton contains exactly these parts:
  1. A minimal Go backend (`server/main.go` calls `pluginsdk.Serve`), with the Kandev SDK dependency mechanism as in the template: `replace` to a `../kandev` checkout next to the repo, and the SDK commit pinned in `.kandev-sdk-ref`.
  2. Connect a Backlog space with an **API key**.
  3. Package the plugin and run package verification.
  4. Install the package on the self-hosted Kandev server.
  5. Make at least one successful call to the Backlog API.
- **OAuth comes later**, in a separate Unit after the skeleton is approved.
- The skeleton counts as done only when the Construction verification command (chosen and authorized by you when entering Construction) proves the real result, including the first manual check against a real Backlog space (see Testing Posture), and you then approve the skeleton checkpoint. A design review does not replace a working skeleton.

## Testing Posture

- We treat tests as a deliverable of every Bolt, not extra work.
- **Methodology**: tdd
- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.
- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.
- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.
- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.
- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.
- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.
- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.
- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.
- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4.

## Guard Policy

<!-- Affirmed by the team. Mode: strict, relaxed, or off. Strict here holds for every intent and cannot be changed from chat. A section under the retired Change Control heading, written by an earlier release, is still read. -->

## Deployment

- We release by tagging **`vX.Y.Z`** (semver) on `main`. Deliberately creating the tag is the manual approval step for "production".
- The tag triggers the release workflow (`release.yml`): it runs only from `main`, refuses if the tag already exists, re-runs all checks (format, vet, lint, test, package, **package verification**), then creates a **GitHub Release** with the package and `checksums.txt` attached.
- The publish job creates a **build provenance attestation** (`actions/attest-build-provenance`); only this job is granted `id-token: write` and `attestations: write`; users verify it with `gh attestation verify`.
- After the Release, we send a **pull request updating the Kandev marketplace registry**; the Kandev maintainers review it.
- **Installation on the self-hosted Kandev server is manual.** There is no automatic build or install after each merge into `main`.
- The package declares a minimum Kandev version and is tested on that version through the automated contract test before release.
- As in the template: every GitHub Action is pinned to a full commit SHA, workflows default to `permissions: contents: read`, and CI runs on `pull_request` (not `pull_request_target`).
- Rollback: the plugin has no in-place deployment to roll back. If a release is broken, we mark that Release, reinstall the previous version on the self-hosted server, and fix forward with a new patch release. A released tag is never deleted or overwritten.
- License: **MIT**.

## Code Style

- We use the project's own configuration; every tool runs in CI before merge and a failure blocks the pull request. The standard commands live in the `Makefile` (`check-format`, `vet`, `lint`, `test`, `coverage`, `build`, `package`, `verify-package`) and CI calls exactly these targets, so local and CI results match.
- **Go**: format with `gofmt`, static checks with `go vet`, and **`golangci-lint`** with the default rule set plus **`gosec`** (version pinned in CI). Exceptions are written in the code as `//nolint:<rule> // reason`. CI checks that `go mod tidy` does not change `go.mod`/`go.sum`.
- **TypeScript**: `tsc --noEmit` with `"strict": true`, plus **ESLint** and **Prettier** (you chose to add these two tools beyond the template). Config lives at the root of the UI directory. No `any` without a reason.
- **Layout**: `server/main.go` only calls `pluginsdk.Serve(plugin.NewRuntime())`; code lives under `internal/`. Only `internal/plugin` (and `server/`) may import `pluginsdk`; the Backlog client lives in `internal/backlog`, secret redaction in `internal/redact`. Other packages (`auth`, `store`, …) appear only when a Unit needs them. Interfaces are declared on the consumer side only when there really are two implementations or a test needs to swap one in.
- **Go naming**: follow Effective Go — package names lowercase, one word; acronyms keep their capitals (`IssueID`, `APIKey`, `URL`); do not repeat the package name in type names (`backlog.Client`); file names `snake_case.go`; sentinel errors `ErrXxx`, error types `XxxError`. Backlog terms stay as-is in code (`Space`, `Project`, `Issue`, `IssueKey`, `PullRequest`, `Status`).
- **TypeScript naming**: kebab-case file names, camelCase variables/functions, PascalCase types.
- **Error handling**: wrap errors with context using `fmt.Errorf("...: %w", err)` and check them with `errors.Is`/`errors.As`. `internal/backlog` turns HTTP error responses into an error type with `Status` and `Retry-After`; only `internal/plugin` maps them to `pluginsdk` error codes (429 → Unavailable, 401 → reconnect required). Every function that does network or I/O takes `context.Context` as its first parameter and returns `context.Canceled`/`DeadlineExceeded` unchanged. Public error messages contain no Backlog response content and no secrets. Response size is limited with `io.LimitReader`. No `panic` outside initialization.
- **Dependencies**: prefer the standard library; no third-party HTTP client library or Backlog SDK. Packages and exported names have doc comments.
## Forbidden

<!-- Team-specific forbidden patterns -->

## Mandated

<!-- Team-specific mandates -->

## Corrections

<!-- Self-learning loop appends here. -->
