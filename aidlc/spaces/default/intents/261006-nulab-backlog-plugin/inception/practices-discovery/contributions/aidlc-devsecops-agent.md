**Collaborator:** aidlc-devsecops-agent

## Contribution

Perspective: lint/formatting, SAST/DAST, secret scanning, dependency scanning, and supply chain (release integrity, pinned actions, checksums). Everything below is still a **suggestion** to ask in the interview, not a confirmed fact.

### 1. Direct comparison with the `kdlbs/kandev-plugin-bitbucket` template

I read `.github/workflows/{ci,build,release}.yml`, `Makefile`, and `scripts/verify-package.sh` via `gh api` (2026-10-06). This replaces the assumption "the toolchain is exactly as described in the brief" in `evidence.md`:

- **Already present, keep as-is:**
  - Every action is pinned to a **full commit SHA** with a version comment (for example `actions/checkout@08eba0b… # v4.3.0`).
  - `permissions: contents: read` at workflow level; only the release job gets `contents: write`.
  - CI runs on `pull_request` (not `pull_request_target`), so code from forks cannot reach secrets with write access.
  - The Kandev SDK version is pinned in `.kandev-sdk-ref`, and CI checks that it is a 40-character hex SHA.
  - Runs `go mod tidy` then `git diff --exit-code -- go.mod go.sum`; installs the UI with `npm ci` from the lockfile.
  - The package contains a `checksums.txt` (the Kandev server checks it on install). The Release attaches `checksums.txt`, including the SHA-256 of the tarball itself.
  - Releases only from `main`, refuses if the tag already exists locally or on origin, and re-runs all checks (format, vet, test, package, verify) before publishing.
  - The `e2e-live` suite turns off Playwright traces, screenshots, and video so credentials do not leak into artifacts.
- **Not in the template:** no SAST, no dependency vulnerability scan (no `govulncheck`, no `.github/dependabot.yml`), no secret scanning in CI, no provenance/attestation for the package, no `SECURITY.md`.
- **Different from the draft's description:** the `Makefile` also has `package-host`, `verify-package-host`, `e2e`, `e2e-live`. CI runs only on pull requests and **not** on pushes to `main`, so the template does not produce an artifact after each merge, as the draft's Deployment section assumes.

### 2. Lint and formatting (supplements the Code Style section)

- Go: `gofmt` (via `make check-format`) + `go vet`, kept as in the template. If you agree to add `golangci-lint`, suggest also enabling `gosec`, `errcheck` (already in the default set), and `bodyclose`. That way SAST for Go lives inside the lint step, with no separate tool. Pin the `golangci-lint` version in CI for stable results.
- Checking that `go mod tidy` does not change `go.mod`/`go.sum` (as in the template) is a required CI step.
- TypeScript (if there is a UI): ESLint + Prettier. No separate ESLint security plugin is needed for a small UI bundle, unless the UI renders HTML taken from Backlog (for example issue content). In that case, add a rule banning `dangerouslySetInnerHTML`/`innerHTML` with unsanitized data.

### 3. SAST and DAST

- **SAST:** `gosec` via `golangci-lint` is enough for the first release. GitHub CodeQL can be added as a second layer (free for public repos, with a default setup for Go/TypeScript). Suggestion: turn on "default setup" in the repo settings, no workflow to write. Blocking level: High/Critical findings block the merge; Medium only warns. Every exception is written in the code as `//nolint:gosec // reason`.
- **DAST:** **no separate DAST tool** (ZAP, Burp). The plugin has no public server of its own. The dynamic-testing equivalent is the contract test with the packaged plugin on the minimum Kandev version (like the template's `packaged-host-contract` job) plus unit tests using `httptest` for the security cases. Minimum cases:
  - OAuth: a wrong or missing `state` parameter is rejected; `redirect_uri` exactly matches the registered address (C-T6); an expired token is refreshed, and a failed refresh does not leak the token.
  - The space address entered by the user (C-T5): accept only `https` with a host under `*.backlog.com`, `*.backlog.jp`, `*.backlogtool.com`. Reject IP addresses, `localhost`, unusual ports, and redirects to another host. Reason: the API key/token is sent to this exact address, so a loose check would cause credential leaks or SSRF.
  - 429 and 401 errors do not put the API key/token into error messages or logs.

### 4. Secret scanning

- Turn on **GitHub secret scanning + push protection** (free for public repos, just a setting). This is the main blocking layer: it blocks at push time, with nothing to install.
- Optional: add `gitleaks` to CI (action pinned by SHA) as a second layer. It can also be a local pre-commit hook, because the AI may accidentally copy a sample key into code or fixtures.
- Test fixtures use only obviously fake values (for example `test-api-key-not-real`), never partially masked real keys.
- `.gitignore` must include `.env*`, `*.tar.gz`, and the Playwright output directories (`test-results/`, `playwright-report/`).
- CI secrets: only what the template already uses (the E2E server URL). If there is later an E2E against real Backlog, the key goes in a GitHub Environment with manual approval and is never given to workflows that run on PRs.

### 5. Dependency scanning

- `govulncheck ./...` in CI, blocking the merge on **reachable** vulnerabilities. It is Go's official tool, with few false positives.
- UI: `npm audit --omit=dev --audit-level=high`. The bundle contains only runtime dependencies, so vulnerabilities in devDependencies do not block.
- Add `.github/dependabot.yml` for the three ecosystems `gomod`, `npm`, `github-actions`, running weekly. The `github-actions` entry is what keeps the pinned SHAs from going stale. Dependabot alerts and security updates are turned on in the repo settings.
- New dependencies need your approval. The AI does not add libraries on its own when Go's standard library already does the job (for example `net/http` and `golang.org/x/oauth2` are enough for the Backlog client and OAuth).

### 6. Supply chain and release integrity

- Pin every action to a full SHA, as in the template. Pin the Go version via `go-version-file` and Node by major version.
- Pin the Kandev SDK with `.kandev-sdk-ref` (SHA), with a format check in CI, as in the template.
- The package is **not signed**, so the only integrity layer today is `checksums.txt` (inside the package and attached to the Release). Suggest adding `actions/attest-build-provenance` (pinned by SHA) to the publish job. It creates a SLSA provenance attestation tied to the workflow and commit, which users verify with `gh attestation verify`, with no signing keys to manage. It needs `id-token: write` and `attestations: write` on that job only.
- Turn on a **tag protection ruleset** for `v*` (no deletion, no update) and GitHub **immutable releases** if available. That way the "no delete/overwrite tag" rule is enforced by the platform, not just by discipline.
- Protect `main`: require the CI checks (lint, test, govulncheck, verify-package) to be green, ban force-push, ban branch deletion. For a solo developer: no required reviewer, but the checks are still required.
- A release with a vulnerability: mark the Release, publish a GitHub Security Advisory, and ship a fix-forward patch. Add a `SECURITY.md` pointing reporters to GitHub's "Private vulnerability reporting" (free, just needs turning on).
- SBOM: not needed for the first release (YAGNI). Add it when the Kandev marketplace requires it or when there are enterprise users.

### 7. Additional hard rule candidates (for you to confirm or drop)

- `NEVER` write API keys, OAuth tokens, refresh tokens, or client secrets to logs, error messages, test artifacts, or Playwright traces. Source: C-R3, the template's `e2e-live`.
- `ALWAYS` check that the space address host is under the three Backlog domains (C-T5) and uses `https` before sending credentials.
- `ALWAYS` pin GitHub Actions to a full commit SHA.
- `ALWAYS` have CI run `govulncheck` and the `go mod tidy` check; a failure blocks the merge.
- `NEVER` use `pull_request_target` or give secrets to workflows that run code from PRs.
- `ALWAYS` attach `checksums.txt` (SHA-256 of the package) to every GitHub Release.

## Positions

- AGREE: Trunk-based + squash-merge through pull requests even when working alone — the PR is the only place where required security checks run before reaching `main`.
- AGREE: Turn on `main` branch protection (Question 2) — this should be the default; it is what turns lint/govulncheck/verify-package into real gates instead of advice.
- AGREE: Add `golangci-lint` (Question 13) — with `gosec` it also covers SAST for Go with no extra tool.
- OBJECT: The Code Style and Testing Posture sections have no security checks at all (secret scanning, `govulncheck`, Dependabot, SAST) — the Bitbucket template also lacks them, so "follow the template" is not enough for a plugin that stores API keys and OAuth tokens (C-R3); add them as interview questions.
- OBJECT: `evidence.md` says the Bitbucket template was "not accessed directly" and assumes the toolchain is as in the brief — checked directly (section 1): the template also pins actions by SHA, pins the SDK, checks `go mod tidy`, and ships checksums. Record these points as evidence instead of assumptions.
- OBJECT: The Deployment section says "on every merge into `main`, CI builds and creates the package as an artifact" as if it follows the template — the template's CI runs only on `pull_request`; getting an artifact after merge means adding a `push` trigger on `main`, and that is a new decision, not something inherited from the template.
- AGREE: The `NEVER` commit API key/token/client secret candidate — extend it to logs, error messages, and test artifacts (section 7), and enforce it with GitHub push protection.
- AGREE: The `ALWAYS` run `make verify-package` before tagging candidate — the template's release workflow already re-runs this step before publishing, so keep it as an automated step in `release.yml`, not just something to remember to do by hand.
- AGREE: The `NEVER` delete or overwrite a released tag candidate — enforce it with a GitHub tag protection ruleset, because the package is unsigned and checksums are only meaningful when tags cannot be replaced.
- AGREE: Unit tests use `httptest` instead of the real Backlog API — this is also where the security cases live (OAuth `state`, space host check, no token leak on errors), so no separate DAST tool is needed.
