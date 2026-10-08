# Security Test Instructions — 261007-backlog-panel-retouch

## Applicability

No new security surface: no new action, no backend change, no new data. Minimal strategy generates no separate security suite; the existing checks below are run as regression gates, plus one targeted check for the only new outbound link (the Kanban badge).

## Checks

| Check | Command | Expected |
|---|---|---|
| Static security lint (gosec via golangci-lint) | `make lint` | clean |
| Secret redaction tests (keys/tokens never in logs, errors or UI responses) | `go test -race ./internal/... ./server/...` (part of `make coverage`) | pass |
| Badge link only to Backlog over https (project rule: only `https` space addresses under `backlog.com`, `backlog.jp`, `backlogtool.com`) | `npm --prefix ui exec -- vitest run --root ui src/issues/issues-state.test.ts src/issues/issue-badge.test.tsx` | pass; non-https and foreign hosts render as plain badge |
| External links use `target="_blank"` with `rel="noopener noreferrer"` | same as above | pass |
| No plugin CSS / raw HTML controls | `npm --prefix ui exec -- vitest run --root ui src/controls.test.ts` | pass |

Dependency vulnerability scanning is intentionally not added (team decision, Testing Posture).
