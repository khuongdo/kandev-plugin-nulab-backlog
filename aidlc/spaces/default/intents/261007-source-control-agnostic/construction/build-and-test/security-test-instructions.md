# Security Test Instructions — Multi-provider source control

## Threats Covered (STRIDE, brief)

| Threat | Control | Test |
|---|---|---|
| Information disclosure: provider token leaks to UI, logs or errors | Token only in Kandev secret store (`backlog.scm.<provider>.<ws>`); state responses never carry it; errors and logs redacted | `internal/scm/service_test.go` (capturing `slog` handler, state listing), `internal/plugin/actions_scm_test.go`, `ui/src/settings/source-control-section.test.tsx` (token input never re-filled) — NFR1 |
| Spoofing / SSRF: requests sent to an attacker-chosen host | Fixed `https` API host per provider; PR URLs parsed only for `github.com`, `gitlab.com`, `bitbucket.org`; userinfo, query and fragment rejected | `internal/scm/httpx_test.go`, `internal/scm/types_test.go` — NFR2 |
| Elevation of privilege: non-admin changes tokens or mappings | `scm.providers.set_token/test/remove`, `scm.repos.search`, `scm.mappings.set` declared `admin` | `internal/plugin/actions_scm_test.go`, `manifest_test.go` — FR2.6 |
| Denial of service: oversized responses, rate-limit storms | `io.LimitReader`, page caps, no retry loop, per-provider isolation | `internal/scm/httpx_test.go`, `watcher_test.go` — NFR3, NFR4 |
| Tampering with Backlog-off state | All `scm.*` actions except `scm.providers.list` refused while Backlog is off | `internal/plugin/actions_scm_test.go` — FR6.1 |

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make lint            # golangci-lint with gosec
make check-secrets   # no credentials committed
go test -race -count=1 -run 'Redact|Token|Secret|Host|Admin|Guard' ./internal/scm/... ./internal/plugin/...
```

## Notes

- Dependency vulnerability scanning is intentionally not used (team decision); no new dependency was added.
- Only fake tokens appear in fixtures (project Forbidden rule); `make check-secrets` enforces it.
