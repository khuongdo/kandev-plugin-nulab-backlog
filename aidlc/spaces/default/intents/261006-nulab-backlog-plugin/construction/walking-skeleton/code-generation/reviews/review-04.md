## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T10:20:12Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-04 | Minor | internal/connection/service.go and internal/plugin (guard wrapper) > integration switch read | The integration switch is read up to three times per Connect (action guard, start of Connect, store step). Behaviour is correct and fail-closed, but it is redundant I/O against the 1 s store limit and the 14 s budget. | Collapse to two reads (guard plus the pre-store re-check required by BR7.3), or record the third read as intentional. | Unresolved |
| R-05 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/code-generation/code-generation-plan.md > Step 1 manifest bullet, Step 9 Red cases, Step 13 `verify-package` bullet | The plan still says `connection.connectApiKey` (Steps 1 and 9) and `plugin-package-verify` (Step 13). The code uses `connection.connect_api_key` and the in-repo `cmd/verifypkg`. The "SDK Facts" section explains this, but the checked-off steps read as if they were built that way. | Rename in Steps 1 and 9, and reword the Step 13 `verify-package` bullet to `cmd/verifypkg`. | Unresolved |
| R-06 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/code-generation/code-generation-plan.md > Testing Contract and "Automated contract test" ownership | Contract-test ownership is resolved: `contract-test` and the `packaged-host-contract` CI job are now built by U5 (Makefile `contract-test`, ci.yml, `cmd/ci contract`), and the U1 plan does not claim them. | None. | Resolved |
| R-07 | Minor | functional-design WF6, entities, and contract-summary C5 > `set_enabled` reply wording | Reply wording for `connection.set_enabled` is inconsistent between WF6, entities and C5. Not re-verified in depth. | Align the three documents on one reply shape (the view returned by the action). | Unresolved |
| R-08 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/contract-design/contract-summary.md > C5 (lines 216, 307, 311) | The shared contract still names `connection.connectApiKey` and does not list `connection.set_enabled` or `integration_disabled`. U1 built `connect_api_key` and `set_enabled`; the U1 functional spec (line 297) and code-summary both say C5 needs amending. Later units (U2 to U4, whose action names are also camelCase) will read C5 and hit the same Kandev key-pattern rejection. | Amend C5 to use snake_case action keys (`^[a-z0-9][a-z0-9._-]*$`), add `connection.set_enabled` and `integration_disabled`, and flag every later camelCase action name for renaming. | New |
| R-09 | Minor | Makefile > `check-format` | `gofmt -l server internal` skips `cmd/` (`cmd/verifypkg` from U1, `cmd/ci` from U5). It is clean today (`gofmt -l server internal cmd` is empty), but unformatted code in `cmd/` would pass the CI format gate. | Use `gofmt -l server internal cmd`. | New |
| R-10 | Minor | README.md > "Build" > "Other targets" | The list omits `check-secrets`, which U5 added to the CI run. Cosmetic. | Add `check-secrets`, or point to the "CI checks" section. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| `gofmt -l server internal cmd` | PASS (no output) | Formatting clean. |
| `go vet ./...` | PASS | No findings. |
| `go test -race -cover ./internal/... ./server/...` | PASS. Package coverage: backlog 96.8%, ci 91.0%, connection 97.5%, pkgverify 93.2%, plugin 91.4%, redact 97.4%, testutil 78.9%. `server` is 0.0%, and `server/main.go` is the only exclusion. | The 80% floor holds per package. The exact total needs `make coverage`, which writes files, so it was not run. |
| `ui`: `tsc --noEmit`, `eslint .`, `prettier --check .` | PASS (only npm config warnings) | UI gates clean. |
| `git status --short` before and after | Identical | Workspace not modified. |

**Coherence after the U5 edits to shared files**

- The Makefile targets `check-format`, `vet`, `lint`, `test`, `coverage`, `build`, `package` and `verify-package` keep their names and behaviour. `COVERAGE_EXCLUDE` is still only `server/main.go`.
- `lint` gained `actionlint` and `cmd/ci workflows`. This is additive and does not weaken the U1 gates.
- ci.yml `checks` still runs the U1 gates (`check-format vet lint test coverage ... build package verify-package`) and adds `check-secrets`. The new job `packaged-host-contract` is additive. All actions are pinned by SHA and permissions stay `contents: read`.
- The README U1 sections (card, switch, `/backlog` page, build, install, connect) match the manifest and the `connect_api_key` and `set_enabled` actions.
- Open question outside U1: the manifest has `categories: ["connector"]`, while U5's `marketplace-entry` passes `-categories integrations`. These may be different vocabularies (plugin manifest versus registry). I did not verify this.

### Summary

U1 is still coherent after the U5 edits: the shared Makefile, ci.yml and README keep U1's targets and gates intact, all U1 checks pass, and the workspace is unchanged. The one significant item is R-08, a shared-contract drift: C5 still names the old camelCase action keys. It is one Major, so it does not block READY, but it should be fixed before U2 to U4 build on it. The remaining findings are Minor.
