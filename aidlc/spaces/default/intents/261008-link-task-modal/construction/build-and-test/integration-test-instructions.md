# Integration Test Instructions — Link Task modal, GitHub-style

## Applicability

Test Strategy is Minimal, so no additional integration tests are generated. The change is UI-only and stays on the existing `issues.link` contract. Its integration with the host is covered by the existing suites, listed below.

## Existing Coverage Relied On

- `ui/src/index.test.ts`: checks that the plugin registers "Link Backlog pull request" and "Link Backlog issue", in that order, on the fake host.
- `internal/plugin` Go tests: the `issues.link` action and its error codes (`not_found`, `conflict`, `validation` with `field=issueKey`). Unchanged and green under `make coverage`.
- CI `packaged-host-contract`: installs the packaged plugin on Kandev `v0.96.0`. It runs on the pull request.

## How to Run

```bash
cd ui && npx vitest run src/index.test.ts
export PATH=$HOME/.local/go/bin:$PATH && go test -race ./internal/plugin/... ./internal/issues/...
```
