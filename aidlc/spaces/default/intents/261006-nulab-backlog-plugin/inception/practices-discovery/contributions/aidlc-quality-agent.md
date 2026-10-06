**Collaborator:** aidlc-quality-agent

## Contribution

Quality perspective: testing posture, coverage tooling, CI quality gates, test and code patterns, and the points the interview needs to settle. I read the template repo `kdlbs/kandev-plugin-bitbucket` directly (`Makefile`, `package.json`, `.github/workflows/ci.yml`, `build.yml`, the directory tree) via `gh api`, so the comments below are based on the repo's real content, not only on the brief.

### 1. What the Bitbucket template actually does (supplements `evidence.md`)

- `make test` = `test-ui` (`npm test` → `build:ui` then `vitest run`) + `test-package-verifier` + `test-release-version` (two shell scripts that self-test the package verifier and the version checker) + `go test ./...`.
- **No `-race` and no coverage measurement** anywhere in the `Makefile` or CI. The template also has **no coverage threshold at all**. So the 80% floor of the `feature` scope has to be built from scratch; it cannot be copied from the template.
- The TypeScript UI has only `tsc --noEmit` (the `lint` script is just `typecheck`). **No ESLint or Prettier.**
- `ci.yml` (runs on pull requests into `main`) has two jobs:
  - `verify`: checks that `go mod tidy` does not change `go.mod`/`go.sum`, `make check-format`, `make vet`, `make test`, `make package-host verify-package-host`, and an optional `make e2e` when the `KANDEV_PLUGIN_E2E_URL` secret is present.
  - `packaged-host-contract`: checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway Kandev server, installs the real package, and runs Playwright on both desktop and mobile UI. **This is exactly how the template satisfies C-T2 automatically in CI.**
- The Kandev SDK source is pinned by commit SHA in the `.kandev-sdk-ref` file; CI checks the SHA format before using it.
- `e2e-live` is an optional step, run against a throwaway Bitbucket with secrets in environment variables. The live Playwright config **turns off traces, screenshots, and video** so credentials do not leak into test artifacts.
- Test layout: `*_test.go` next to the code in `internal/<package>/`, API response sample data in `internal/<package>/testdata/*.json`, a separate `internal/redact` package (secret redaction) with tests, and a test for `server/manifest_test.go`.

### 2. Proposals for the Testing Posture section

- **Methodology**: test-after (keep the org.md default). **Ordering**: as in the lead's draft.
- **Test volume under the `Standard` strategy**: about 5-8 tests per component, unit and integration, roughly 75/20/5. Every user story acceptance criterion (Given/When/Then) has at least one test tied to its ID.
- **Every bug gets a test**: when fixing a bug, write a test that reproduces it first, then fix it (the standard item of `bugfix` in org.md, applied even while doing `feature`).
- **Coverage measurement, concrete suggestion**:
  - Go: `go test -race -coverprofile=coverage.out ./...`, then a small script reads the `total:` line of `go tool cover -func=coverage.out` and fails if below 80%. No extra library needed. Make it a `make coverage` target and call it from CI.
  - `-race` needs CGO and works on the `ubuntu-latest` runner. This is an **addition compared with the template**, worth keeping because the plugin does token refresh and API rate limiting, two places prone to data races between goroutines.
  - UI (if any): Vitest needs `@vitest/coverage-v8` added to measure coverage. This is a new devDependency compared with the template.
- **Coverage scope (needs settling)**: suggest measuring `./internal/...` and `./server/...`. Exclude only the minimal wiring (the `main` function that calls the go-plugin handshake), and list it explicitly in the Makefile. Do not exclude more to reach the threshold; this matches org.md's "do not lower the floor to make it pass" rule.

### 3. Test and code patterns to record as practices

- **Table-driven tests with `t.Run`**: the standard Go way to write tests, very readable for someone new to Go; each row is one scenario (success, 4xx error, 429, expired token…).
- **Fake Backlog server with `net/http/httptest`** plus JSON sample data in `testdata/`, exactly as in the Bitbucket template. Once there is a Backlog test space (C-O2), record real responses once as sample data, then re-check them periodically. This replaces Pact-style contract tests, which are too heavy for a one-person project.
- **Inject the clock and wait functions into the code**: rate limiting (wait at least 1 second between update or search calls, C-T3) and OAuth tokens expiring after 1 hour (C-T6) must be testable without a real `time.Sleep`. If this injection point is not designed in from the start, tests become slow and flaky, or this part ends up untested.
- **Space address is a parameter**: the client takes the base URL from configuration, with no hard-coded domain, so tests can point at `httptest` and all three domains `backlog.com`, `backlog.jp`, `backlogtool.com` can be tested (C-T5).
- **Redaction tests**: a test asserts that API keys and tokens do not appear in logs, error messages, or responses sent to the UI (C-R3), following the template's `internal/redact` package.
- **Independent tests**: no dependence on run order or shared state; use `t.TempDir()` for files and `t.Setenv()` for environment variables.

### 4. CI quality gates (suggested, run on every pull request)

| Gate | Command | Source |
|---|---|---|
| Tidy module | `go mod tidy` + `git diff --exit-code` | Bitbucket template |
| Formatting | `make check-format` | Template |
| Static checks | `make vet` (+ `golangci-lint` if chosen) | Template / new |
| TypeScript types | `npm run typecheck` (if there is a UI) | Template |
| Tests | `make test` (with `-race`) | Template + new |
| Coverage ≥ 80% | `make coverage` | **New** |
| Valid package | `make package-host verify-package-host` | Template |
| Runs on the minimum Kandev version | `packaged-host-contract`-style job | Template (**needs settling**) |
| Dependency vulnerability scan | `govulncheck ./...` | New (**needs settling**) |

Every gate must block the merge on failure (org.md: lint failures block the PR). Gates only truly block when `main` branch protection is on; so the lead's question 2 decides whether these gates are enforced or only advisory.

### 5. Points the interview needs to resolve (supplements the lead's list)

1. **C-T2 automated or manual?** Copy the `packaged-host-contract` job (build Kandev at the minimum version then run Playwright, about 30 minutes per CI run, complex while still new to Go), or a recorded manual check on the self-hosted Kandev server (C-O3) before each Release? Suggestion: manual for the first release, bring in the automated job once the skeleton is stable.
2. **Coverage scope and exclusion list** (section 2).
3. **Use `-race`?** Suggestion: yes.
4. **Add `govulncheck`?** Suggestion: yes. It is Go's official command, no configuration needed.
5. **TypeScript UI: only `tsc --noEmit` as in the template, or add ESLint/Prettier?** Suggestion: follow the template first.
6. **Skeleton verification command**: suggest `make verify-package-host` (builds only for the current platform, faster) instead of `make verify-package` (5 platforms), plus a recorded manual install step.
7. **`e2e-live` with a real Backlog space**: if done, turn off traces, screenshots, and video as in the template, and keep secrets only in GitHub Secrets.

### 6. Hard rule candidates, supplementing section 15 of `evidence.md` (for you to confirm or drop)

- `ALWAYS` write a test that reproduces a bug before fixing it.
- `NEVER` put real credentials in `testdata/` sample data, test logs, or Playwright artifacts.
- `NEVER` use a real `time.Sleep` in tests to check rate limits or token refresh; always inject a fake clock.

## Positions

- AGREE: Methodology test-after and per-layer Ordering — new to Go, no reason to force TDD; still record the "every bug gets a test" rule explicitly.
- AGREE: Use `net/http/httptest`, never call real Backlog in unit tests — matches the Bitbucket template, and there is no test space anyway (C-O2).
- AGREE: The 80% floor must not be lowered to make it pass — matches org.md; only the measured scope and an explicit exclusion list need settling.
- OBJECT: The sentence "`go test` … has `-race` and coverage, per the Bitbucket template" — the template has no `-race` and no coverage; record this as a new addition, with a concrete measurement method (`make coverage`), otherwise the 80% floor has no enforcing tool.
- OBJECT: Code Style lists Prettier + ESLint for TypeScript as the default — the template uses only `tsc --noEmit`; take typecheck as the minimum gate and turn ESLint/Prettier into a question.
- OBJECT: `evidence.md` lists the template's CI but misses the `packaged-host-contract` job and the `go mod tidy` check — this job is how the template automates C-T2, so question 9 and the Deployment section need to know this option exists.
- AGREE: Add `golangci-lint` with the default rule set — catches common mistakes of people new to Go that `go vet` misses.
- AGREE: The skeleton verification command includes package verification plus a recorded manual install — but use `verify-package-host` for a faster loop.
