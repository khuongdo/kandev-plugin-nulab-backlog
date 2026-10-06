## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T11:09:25Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | internal/backlog/client.go > `do` (sets `Backlog-API-Key`); aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/contract-design/contract-summary.md > open question C7 | The API key is now sent only in a `Backlog-API-Key` header. Nothing in the repo, README or a manual-check record shows that Backlog accepts it. C7 says Backlog documents only the `apiKey` URL parameter, and C7 is assigned to U1. The fake server in `client_test.go` accepts any header, so the tests cannot catch a wrong header name. If Backlog ignores the header, API-key connect fails with 401 on a real space. `docs/manual-checks/` holds only TEMPLATE.md, so there is no real-space evidence yet (AC7.2.1 is Deferred). | Before the skeleton checkpoint is approved, record the result of the real-space check for the header, or cite the Backlog documentation for it. Close C7. If Backlog rejects the header, fall back to the query parameter, with the existing URL redaction as defence. | New |
| R-02 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/functional-design/rules.md > BR3.4, BR2.10; nfr-design/performance-design.md > NFR2.1; code-generation-plan.md > Step 7 ("`apiKey` query parameter") | Three U1 statements are superseded by U2 but not marked. BR3.4 still says query parameter. BR2.10 and NFR2.1 still say "429: no retry, return at once". The code retries up to 3 times within a 3 s Interactive budget. traceability.json keeps BR3.4, BR2.10 and NFR2.1 as `OK`, pointing at `client_test.go` and `service_test.go`. Those tests now prove the new behaviour (header at `client_test.go:74`; retry and budget in `limiter_test.go`), so the evidence is valid for the new rules, not the stated ones. The 14 s total and the key-never-in-URL intent (AC1.1.7, NFR3.4) still hold. | Add a "superseded by U2 plan" note to BR3.4, BR2.10, NFR2.1 and plan Step 7, or amend the traceability rows so the evidence matches the rule text. | New |
| R-03 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/contract-design/contract-summary.md > C5 | Carried open item. C5 still lists camelCase action keys and omits `connection.set_enabled` and the `integration_disabled` code. manifest.yaml and runtime.go use `connection.connect_api_key` and `connection.set_enabled`. | Update C5 to the shipped action keys and error codes. | Unresolved |
| R-04 | Minor | internal/plugin/runtime.go > HandleAction (guard) and internal/connection/service.go > Connect | Carried open item. A guarded Connect reads the integration switch in the guard, in Connect, and again before Save. This is cheap and inside the 12 s deadline (`TestGuardRunsInsideTheConnectDeadline`), but it is redundant. | Accept the redundancy or drop one read. Record the choice. | Accepted risk |
| R-05 | Minor | Makefile > check-format; aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/code-generation/code-generation-plan.md > Steps 1, 9, 13 | Carried open items. `check-format` runs `gofmt -l server internal`, so `cmd/` is not checked (`gofmt -l .` is clean today). Plan Steps 1, 9 and 13 still use the old action names. The README "omits check-secrets" item is resolved: README line 68 lists `check-secrets`. | Add `cmd` to the gofmt scope. Refresh the old names in the plan steps. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l . | PASS, no output | Formatting clean, including cmd/ |
| go vet ./... | PASS | No findings |
| go test -race -cover ./internal/... ./server/... | PASS, all packages ok | Coverage: backlog 94.7%, connection 95.3%, plugin 95.9%, redact 97.4%. U1 behaviour is tested after the U2 edits. |
| ui: npx tsc --noEmit, npx eslint . | PASS, only npm config warnings | Clean |
| ui: npx prettier --check . | PASS | Clean |
| ui: npx vitest run (read-only) | PASS, 102 tests | Settings, switch, page and logo tests are green |
| traceability.json targets | PASS | All coverage targets exist and every upstream ID is covered. All 69 source-manifest paths exist. |
| git status before/after | identical | Workspace not modified |

### Summary

U1's behaviour (connect, address validation, redaction, 14 s budget, switch and guard, UI, packaging) still works and is tested after the U2 edits, and every traceability target exists. The one real risk is that nothing shows Backlog accepts the `Backlog-API-Key` header (open question C7). Close that with the real-space check at the skeleton checkpoint (R-01). The rest is stale spec text and carried minor items.
