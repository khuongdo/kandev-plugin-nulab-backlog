# Tech Stack Decisions — walking-skeleton (U1)

Inputs: `requirements` (constraint C-T1, NFR6–NFR10), `functional-spec` and `rules` for U1, `contract-summary` (C4, C8), `team-practices` (Code Style, Testing Posture), and answers Q2 and Q6 in `nfr-requirements-questions.md`.

Most choices are fixed by `team-practices` and by Kandev itself. This table records them with their reasons, so U1 sets them up once for every later unit.

| Area | Choice | Reason | Source |
|------|--------|--------|--------|
| Backend language | Go, at the version in Kandev's `go.mod` (1.26) | Plugins are Go processes served through Kandev's SDK; matching Kandev's version avoids toolchain mismatches with the `replace`d SDK | C-T1, BR5.5 |
| Kandev SDK | `pluginsdk` from the `../kandev` checkout via `replace`, pinned by commit in `.kandev-sdk-ref` | Same mechanism as the Bitbucket template; reproducible builds | team-practices (Walking Skeleton), BR5.5 |
| Minimum Kandev | `min_kandev_version: "0.96.0"` | Your choice [Q2]; admin actions need at least 0.91.1 | Q2, NFR6 |
| HTTP client | Go standard library `net/http`, with redirects disabled and a 10-second timeout per call | No third-party HTTP client or Backlog SDK (team-practices, Code Style) | BR1.4, BR4.1 |
| Logging | Go standard library `log/slog` with a JSON handler on standard error | Structured logs without new dependencies | NFR11.1 |
| Go tests | `testing` with `github.com/stretchr/testify/require`, table-driven, `go test -race`, a fake Backlog built with `net/http/httptest`, injected clock | team-practices (Testing Posture) | NFR8 |
| Go quality tools | `gofmt`, `go vet`, `golangci-lint` (default set plus `gosec`, version pinned in CI) | team-practices (Code Style) | BR5.6 |
| UI | TypeScript (`strict`), using Kandev's shared React through `host.jsx` and `host.ui` components; no bundled React | Kandev forbids a second React runtime in plugin bundles | `@kandev/plugin-sdk`, NFR9 |
| UI tests and tools | Vitest, ESLint, Prettier, `tsc --noEmit` | team-practices (Code Style) | BR5.6 |
| UI registrations | `registerIntegrationSettings` (with `icon` and `action`), `registerNavItem` (section `integrations`), `registerRoute('/backlog')`, `host.ui.IntegrationEnabledControl`, `host.setIntegrationEnabled` from Kandev 0.96.0 | All exist in the pinned SDK; no host change needed | BR5.4, BR7.5, BR7.6 |
| Backlog logo | Inline SVG React component built with `host.jsx`, from Nulab's official brand assets; no image file fetched at runtime | No network dependency or CSP exception; works offline on self-hosted servers [Q6] | BR7.8, NFR3.10 |
| UI strings | One English message catalogue keyed by message ids | Your choice in functional design (Q5); translations come in U3 | BR6.2, NFR10.1 |
| Packaging | `.tar.gz` package with `manifest.yaml`, `server/` executables and the UI bundle; checksums as in BR5.2 | Kandev's package format | BR5.1, BR5.2 |
| Build entry point | `Makefile` targets of contract C4 | CI and local runs use the same commands | C4, BR5.6 |

## Requirement IDs Covered Here

| ID | Requirement | Pass/fail criterion | Source |
|----|-------------|---------------------|--------|
| NFR8.1 | Go code is written test-first with `go test -race`. Line coverage over `./internal/...` and `./server/...` is at least 80%, excluding only the wiring in `server/main.go` | `make coverage` passes; CI blocks merges below the floor (U5) | NFR8, team-practices, BR5.6 |
| NFR9.1 | The M1 screen and the `/backlog` page meet these WCAG 2.1 AA basics: every input has a label, field errors are linked to their input, the Connect result is announced, and everything is operable by keyboard | A Vitest test with an accessibility checker finds no A/AA violations on each M1 and P1 state; the logo is `aria-hidden` next to a text label and the switch has an accessible name; a manual keyboard, focus-visibility and contrast pass during each manual check, recorded in the manual check record | NFR9, BR6.4 |
| NFR10.1 | Every UI string comes from a message key. U1 ships English only | A lint or test check finds no literal user-facing text in components | NFR10, BR6.2 |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q2], [Q6]: answers in `nfr-requirements-questions.md`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`; `team-practices.md`; Kandev `go.mod`, `apps/packages/plugin-sdk`, `docs/public/plugins-authoring.md`.

## Assumptions & Open Questions

- [assumption] The Node.js version for building the UI is the current LTS. Infrastructure Design pins the exact version.
