**Collaborator:** aidlc-developer-agent

## Contribution

Review scope: naming, layer boundaries, error handling, file organisation, and coding conventions for the Go part and the TypeScript UI. I read the public repo `kdlbs/kandev-plugin-bitbucket` directly (directory tree, `Makefile`, `go.mod`, `package.json`, `.github/workflows/ci.yml`, `server/main.go`, `internal/domain/errors.go`, `internal/plugin/action_errors.go`, `internal/cloud/response_error.go`, `internal/redact/redact.go`, the packages' imports). Everything below is still a **suggestion** for the interview, not an affirmed practice.

### 1. What the Bitbucket template actually does (checked, supplements `evidence.md`)

- **Layout**: `server/main.go` has a single line `pluginsdk.Serve(plugin.NewRuntime())`; all code lives under `internal/` (`plugin`, `domain`, `cloud`, `datacenter`, `auth`, `store`, `redact`, `watches`); the UI is in `ui/src/*.ts`, UI tests in `ui/test/*.test.ts`; helper scripts in `scripts/*.sh`; `manifest.yaml`, `assets/`, `CHANGELOG.md`, `LICENSE` at the root.
- **Layer boundaries**: only the `internal/plugin` package (and `server/`) imports `github.com/kandev/kandev/pkg/pluginsdk`. The `domain`, `auth`, `store`, `watches` packages use only the standard library. The HTTP client packages (`cloud`, `datacenter`) depend only on `domain`.
- **SDK dependency**: the Kandev SDK is not yet released as its own module. `go.mod` uses `replace github.com/kandev/kandev => ../kandev/apps/backend`, `package.json` uses `"@kandev/plugin-sdk": "file:../kandev/apps/packages/plugin-sdk"`, and the SDK commit is pinned in the `.kandev-sdk-ref` file (CI checks out exactly that commit). The lead's draft does not mention this yet.
- **Module name**: `module kandev-plugin-bitbucket` (a bare name, not a GitHub path).
- **Error handling**: the client turns every HTTP error response into `*domain.ProviderHTTPError{Status, RetryAfter, Err}`; error messages never contain response content or secrets. Only `internal/plugin/action_errors.go` maps errors to `pluginsdk.ActionErrorCode` codes (InvalidArgument, NotFound, Conflict, Unavailable, Upstream…), using `errors.Is`/`errors.As`. `context.Canceled`/`DeadlineExceeded` errors are returned unchanged. Each package has `ErrXxx` sentinel errors (for example `watches.ErrWatchNotFound`). Response size is limited (`defaultMaxResponseBytes = 1 << 20`).
- **Secret redaction**: the `internal/redact` package removes tokens from error strings and diagnostics.
- **Go tests**: `testing` + `github.com/stretchr/testify/require` + `net/http/httptest`; JSON sample data in `internal/<package>/testdata/`; test names describe behaviour (`TestListRepositoriesFollowsCloudCursorAndRespectsLimit`).
- **TypeScript**: no ESLint or Prettier. `npm run lint` is just `tsc --noEmit` with `"strict": true`; bundled with `esbuild` into `ui/bundle.js` (this file is committed to the repo); tested with Vitest, e2e with Playwright. TS file names are kebab-case (`connection-view.ts`, `view-model-review.ts`).

### 2. Proposed layout for the Backlog plugin

Backlog has only one API (API v2, the same shape for the `backlog.com` / `backlog.jp` domains), so there is no need for two client packages like Bitbucket's `cloud`/`datacenter`.

```
server/main.go          # one line: pluginsdk.Serve(plugin.NewRuntime())
internal/plugin/        # the only layer that imports pluginsdk: action, view, connection, error mapping
internal/backlog/       # HTTP client for Backlog API v2: models, operations, response_error, testdata/
internal/auth/          # API key + OAuth (refresh tokens that expire after 1 hour, C-T6) — add when the OAuth Unit arrives
internal/store/         # encrypted secret storage (C-R3) — add when credentials need storing
internal/redact/        # redact secrets in errors/logs
ui/src/, ui/test/       # only if the first release has a UI
scripts/, manifest.yaml, assets/, Makefile, .kandev-sdk-ref
```

- The intermediate `internal/domain` package: the Bitbucket template needs it because it has two providers. With a single API, the suggestion is to **skip it at first**: Backlog data types and errors live in `internal/backlog`, and `internal/plugin` maps directly to the SDK. Split out `domain` only when the mapping code grows or when `plugin` starts needing types that are not Backlog's. Alternative: copy the structure with `domain` as-is, to make file-by-file comparison with the template easier. This is a question for the interview.
- No packages, interfaces, or files "for later": each `internal/` package appears only when a Unit needs it. Interfaces are declared only on the consumer side (for example `TokenSource` in the client) when there really are two implementations or a test needs to swap one in.
- The walking skeleton must set up the SDK dependency mechanism (`replace` to a `../kandev` checkout next to the repo + `.kandev-sdk-ref` + a CI step that checks out exactly that commit), because nothing builds without it. Document in `README.md` how to place the two repos side by side.

### 3. Naming conventions

- **Go**: follow Effective Go — package names lowercase, one word, singular (`backlog`, `plugin`, `auth`); exported names start with a capital; acronyms keep their capitals (`IssueID`, `APIKey`, `HTTPClient`, `URL`); do not repeat the package name in type names (`backlog.Client`, not `backlog.BacklogClient`); file names `snake_case.go`, tests `<file>_test.go` in the same package.
- **Backlog terms** stay as-is in code: `Space`, `Project`, `Issue`, `IssueKey` (form `PROJ-123`), `Repository`, `PullRequest`, `Status`. Do not rename them to other terms (for example "ticket"), so the code is easy to match against the Backlog API docs.
- **Errors**: sentinel variables `ErrXxx`, error types `XxxError`.
- **Tests**: `TestXxx` describes behaviour; table-driven tests use `t.Run(tc.name, …)`.
- **TypeScript**: kebab-case file names as in the template; camelCase variables/functions; PascalCase types/interfaces; constants camelCase, or UPPER_SNAKE for configuration constants.
- **Go module name**: suggest `github.com/<owner>/kandev-plugin-nulab-backlog` (the repo will be public on GitHub, C-O6) instead of a bare name as in the template; both work. You need to confirm the repo owner.

### 4. Error handling (rule candidates for you to confirm, not yet written to `discovered-rules.md`)

- `ALWAYS` wrap errors with context using `fmt.Errorf("lấy issue %s: %w", key, err)` and check them with `errors.Is`/`errors.As`, not string comparison.
- `ALWAYS` turn Backlog HTTP error responses into an error type with `Status` and `Retry-After` right in `internal/backlog`; only `internal/plugin` maps them to `pluginsdk` error codes (like the template's `action_errors.go`). 429 → Unavailable with `Retry-After` (C-T3, C-T4); 401 → reconnect / refresh token required (C-T6).
- `ALWAYS` take `context.Context` as the first parameter for every function that does network or I/O, and return `context.Canceled`/`DeadlineExceeded` unchanged.
- `NEVER` let an API key or token appear in error messages, logs, or content returned to the UI. **Backlog-specific point**: the API key is sent in the `?apiKey=…` query parameter, and Go's `*url.Error` prints the full URL. So the Bitbucket `redact` regex (which has no `apiKey`) is not enough; `apiKey` must be added to the redaction list, with a test proving that network errors do not leak the key.
- `NEVER` put Backlog response content into public error messages (it may contain private issue data).
- `NEVER` use `panic` outside `main`/initialization, and `NEVER` discard an error with `_ =` without a comment giving the reason.
- Limit the size of responses read (`io.LimitReader`), like the template's `defaultMaxResponseBytes`.

### 5. Coding conventions and tools

- **Go**: `gofmt` (via `make check-format`) + `go vet`, as the lead proposed. If `golangci-lint` is added, suggest enabling only the default set plus `errorlint` (catches wrong error comparisons, very common for people new to Go); add other rules when there is a concrete reason. Pin the version in CI.
- **Comments**: since the developer is new to Go, suggest that every package has a package comment (`// Package backlog …`) and every exported name has a doc comment, as the Bitbucket template does. No comments needed for self-explanatory code.
- **Dependencies**: prefer the standard library (`net/http`, `encoding/json`, `httptest`). Add only `testify` (already in the template). No third-party HTTP client library or Backlog SDK.
- **TypeScript**: at minimum `tsc --noEmit` with `strict: true` (as in the template), ban `any` without a reason, take Kandev interface types from `@kandev/plugin-sdk`. Prettier only for formatting (one default `.prettierrc` file); ESLint is optional and can wait, because the template does not use it and strict `tsc` already catches most errors.
- **Generated files**: the template commits `ui/bundle.js`. Decision needed: commit it as in the template, or put it in `.gitignore` and generate it only on `make package`. Suggestion: do not commit, unless the Kandev plugin docs require it.

### 6. Additional questions proposed for the interview

1. Split out the `internal/domain` package from the start as in the template, or start with `internal/backlog` + `internal/plugin`?
2. Go module name: GitHub path or bare name as in the template? Who owns the repo?
3. Confirm the error-handling rule candidates in section 4 (especially the `apiKey` redaction rule).
4. TypeScript: strict `tsc` only as in the template, add Prettier, or add ESLint too?
5. Commit `ui/bundle.js` or generate it at packaging time?

## Positions

- AGREE: Trunk-based + squash-merge per org.md — fits a single developer, no reason to change.
- AGREE: `gofmt` + `go vet` via `Makefile` targets as in the template, with CI calling exactly those targets — checked, the template does exactly this.
- AGREE: Use `net/http/httptest` instead of calling the real Backlog API in unit tests — the template does this, with JSON `testdata/` and `testify/require`.
- AGREE: Ask about `golangci-lint` — useful for someone new to Go, but keep the rule set small (default + `errorlint`).
- OBJECT: The draft's Code Style section lists Prettier + ESLint for TypeScript "per org.md" as the default — the Bitbucket template uses neither and uses strict `tsc --noEmit` as lint; the template option should go into a question instead of defaulting to ESLint.
- OBJECT: The draft has no conventions for layer boundaries, error handling, and file organisation — this is what decides the quality of AI-generated code for someone new to Go; add it to the Code Style section (see sections 2-4).
- OBJECT: `evidence.md` says the Bitbucket template was "not accessed directly" and misses the SDK dependency mechanism (`replace` to `../kandev` + `.kandev-sdk-ref`) — this is a precondition for building and must be part of the walking skeleton.
- OBJECT: The `NEVER` commit secrets rule candidate is not enough for Backlog — the API key travels in the `apiKey` query parameter, so it can still leak through `*url.Error` messages; add a rule for redacting secrets in errors/logs.
