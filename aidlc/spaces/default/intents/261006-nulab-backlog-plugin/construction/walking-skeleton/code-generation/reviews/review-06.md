## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T12:46:38Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | internal/backlog/client.go > Client.send (host check) | The transport layer only checks that SpaceHost is non-empty and has no port. The allowed-host rule (https only; backlog.com, backlog.jp, backlogtool.com) is enforced only at the two entry points that call connection.ParseSpaceAddress (service.go connect, oauth.go StartOAuth); stored hosts are trusted afterwards. No bypass was found, because every stored host comes from those two paths, but U2 and U4 added new Backlog call paths (OAuth token, Git smart-HTTP with a Basic password) that depend on that invariant holding. | Optional defense in depth: re-check the host suffix in send, or record in code that the stored host is already validated. | New |
| R-02 | Minor | internal/plugin/runtime.go > NewRuntime and Start | NewRuntime now calls r.Start(), which starts the PR watcher and the git.Service listener before the SDK injects the Host (SetHost runs after NewRuntime). hostStores.get returns an error until then, so this looks safe, and the U1 actions and tests are unaffected. I did not trace the watcher's first tick; the packaged-host contract test is the real check. | Confirm in Build and Test that the watcher's first tick before SetHost only logs and never panics or blocks. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | PASS (no output) | Formatting is clean. The `.` root was blocked by the review scope hook, so only these three directories were checked. |
| go vet ./... | PASS | No findings. |
| go test -race -cover ./internal/... ./server/... | PASS | Per-package coverage: backlog 95.8%, ci 91.0%, connection 94.5%, git 90.7%, pkgverify 93.2%, plugin 93.1%, redact 97.4%, testutil 87.0%. `server` has 0% (main wiring, the only exclusion). Every package is well above 80%, so the 80% floor holds without a coverprofile run. |
| go run ./cmd/ci secrets -root . | PASS (OK) | No credential-shaped strings in test data or artifacts. |
| ui: tsc --noEmit, eslint src, prettier --check src | PASS | No errors. |
| ui: vitest run | PASS (19 files, 161 tests) | `git status --short` was identical before and after all runs. |

### Differential Verification of U1 Claims

- **Steps 1-22 are still implemented.**
  - The manifest keeps `connection.get`, `connection.connect_api_key` and `connection.set_enabled`, all with scope workspace and `max_body_bytes` 8192. `connect_api_key` and `set_enabled` are admin.
  - `min_kandev_version` is "0.96.0", and all 5 executables are present.
  - `guarded()` in runtime.go still exempts only `connection.get` and `connection.set_enabled`, and the switch is read fail-closed (BR7.3, NFR3.9).
  - Connect keeps its flow: validation, then try-lock, then Myself, then a second switch read, then store. Later units added the workspace lock and event emit around the store step without changing the order.
- **Security rules hold.**
  - The https-only transport, the no-redirect `CheckRedirect`, the 1 MiB LimitReader, error bodies never echoed, and `*url.Error` mapped to class-only Unreachable are all unchanged.
  - The credential goes in a header and never in the URL. All secrets, including the OAuth and Git passwords, go through `redact.WithSecrets`.
  - `connection.ParseSpaceAddress` still gates both Connect and StartOAuth.
- **Later edits to shared files are compatible.**
  - The new U2/U4 manifest actions are admin where they change the shared connection.
  - A U1 record and a U1 secret without `authMethod` still load; `record_compat_test.go` covers this.
  - The `manifest_test` loosening (4 capabilities, U4 actions handled in their own test) is explicit and justified.
  - `BacklogPage`, the switch and `brand` files are untouched.
  - In `index.ts` the settings card, nav item and route are registered first and unconditionally, and `setIntegrationEnabled` is called only after a successful `connection.get`, so BR7.5, BR7.6 and BR7.8 hold.
  - The U4 registration methods exist in the Kandev v0.96.0 checkout.
  - The manifest fields `repository_providers`, `api_read`, `webhooks` and `config_schema` exist in the v0.96.0 manifest schema, and the min-version policy only gates the messages capability.
- **Team rules hold.**
  - CI is SHA-pinned, uses `permissions: contents: read`, and runs the standard make targets.
  - The `coverage` target has the single exclusion `server/main.go`.
  - `check-secrets` was added without weakening any gate.
- **Known and deferred, not re-raised:** the 429 retry now happens for Interactive calls within a 3 s budget (BR2.10 staleness); the other deferred items from the brief.

### Summary

U1's claimed code is still correctly implemented on the current main. The later units extended its shared files additively, and the https and allowed-host rules, secret redaction and the fail-closed switch guard are intact. The checks all pass and the coverage floor holds. Only two non-blocking Minor observations remain.
