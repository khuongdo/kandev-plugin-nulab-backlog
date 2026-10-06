## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T12:53:59Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | .github/workflows/release.yml (on.push.tags, verify job); Makefile release-preflight; README.md branch/tag protection paragraph | Verified. The "tag must be on main" rule is enforced only by the workflow file at the tagged commit. A user with push access could tag a non-main commit carrying an edited release.yml and get an attested Release. The README tag rule only says tags cannot be moved or deleted, not that creation is restricted. Needs repo write access, so it is a hardening gap, not a defect in the committed code. | Extend the README ruleset guidance to restrict creation of v* tags to maintainers, and/or add an environment: release on the publish job. | New |
| R-02 | Minor | internal/ci/secrets.go passwordLiteral, apiKeyQuery, inScanScope | Verified by reading the code. passwordLiteral needs a quoted value, so unquoted YAML/env password: value is missed. apiKeyQuery matches only the case-sensitive apiKey=, so api_key= and apikey= are missed. Scope is limited to *_test.go, testdata, ui test files, docs/manual-checks and coverage files, so internal/testutil/*.go and non-test ui fixtures are not scanned. The scan is a heuristic backstop for the NEVER-real-credentials rule. The scope matches the AC7.3.4 wording and the committed tree currently passes (go run ./cmd/ci secrets: OK). | Make the key patterns case-insensitive and cover api_key, apikey and unquoted values. Add internal/testutil and ui fixtures to scope, or record the narrow scope as a known limit in the README. | New |
| R-03 | Minor | .github/workflows/release.yml verify job (make ... TAG="$TAG"); Makefile release-preflight | Verified. TAG is passed as a make variable and expanded textually into shell recipe lines. Git tag names may contain backticks, $ and parentheses, so a crafted tag runs commands in the verify job before the semver check in the Go preflight. Exposure is limited to contents: read and the read-only GH_TOKEN, and exploiting it needs tag-push access. | Validate TAG against ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ in a workflow step before invoking make, or pass TAG to the recipe only as a shell env variable. | New |
| R-04 | Minor | internal/ci/release.go hasRelease; Makefile release-preflight (gh release list) | Accepted with adjustment. The GITHUB_TOKEN with contents: read most likely cannot list draft Releases, so the "draft counts as a Release" branch probably never fires in CI. A stale draft for the tag would then surface only when gh release create fails in publish, after the contract job. The failure is loud and nothing is overwritten, so the rule "never overwrite a released tag" holds. | Document that a draft is caught only at publish, or run the preflight listing with a token that can see drafts. | New |
| R-05 | Minor | .github/workflows/ci.yml:17; release.yml:22,82 (actions/checkout) | Verified. checkout leaves persist-credentials at its default (true), so the token stays in .git/config for later steps. Jobs are read-only and run only repository code, so the impact is low. | Set persist-credentials: false on every checkout. The verify job's full-history git commands do not need the credential. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | No formatting drift |
| go vet ./... | clean | No findings |
| go test -race -cover ./internal/ci/... | ok, 91.0% coverage | Above the 80% floor |
| go run ./cmd/ci secrets -root . | OK | No credential-shaped strings in scope |
| git status --short before and after | identical | Workspace not modified |

### Summary

The team's release rules are implemented as written: Actions pinned to full SHAs, default contents: read, write, id-token and attestations only on publish, a pull_request trigger, an on-main and existing-Release preflight, and package verification before the release. The contract job tests the same artifact that publish attests. All five advisory findings are real but low impact, so none blocks.
