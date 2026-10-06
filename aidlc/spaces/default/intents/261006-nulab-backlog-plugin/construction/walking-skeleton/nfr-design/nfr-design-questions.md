# NFR Design — walking-skeleton (U1) — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

The NFR requirements for U1 fix almost every design choice. The remaining design work covers how the key stays out of Go error text, how the epoch-linked writes behave, and the log format. These follow directly from the requirements and need no input. One conflict needs your decision.

---

## Q1. Time budget for one Connect

Kandev stops any action after 15 seconds. A Connect makes one Backlog call (limit 10 s, from `contract-summary`). It then makes up to three store calls: read the old secret, write the new secret, write the record. If the record write fails, it also rolls back (5 s, from the functional design). These add up to more than 15 s in the worst case. The NFR review flagged this, because the promised 13-second ceiling (NFR1.2) cannot hold. How should Connect be bounded?

- A. One deadline for the whole Connect: 13 s. The Backlog call gets at most 8 s, each store call at most 1 s, and the rollback at most 2 s. The rollback runs on a fresh context so a cancelled action still rolls back. The 10-second call limit stays for background work in later units
- B. Keep the 10-second Backlog call and shorten only the store calls and the rollback (1 s each, 2 s rollback). The worst case is about 15 s, so a Connect can hit Kandev's limit
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Changes After the U1 Change Request (2026-10-06)

The U1 functional design and NFR requirements changed after the first manual check:

- NFR requirements Q4 (answered A) replaces the Q1 budget above: one 12 s deadline at Connect entry covers the pre-call steps (at most 1 s), the Myself call (at most 10 s, never past the deadline minus 2 s) and the store writes including the second switch read (at most 2 s together); the rollback has its own 2 s on a fresh context; total at most 14 s (NFR1.4). Q1 is kept above for history and is superseded.
- New requirements to design: the per-workspace switch (NFR1.3, NFR3.8, NFR3.9, NFR5.9, NFR5.10, NFR11.6), the inline Backlog logo (NFR3.10), the `/backlog` page and the drafted switch control (NFR9.1), and the ambiguous SetSecret rollback (NFR5.4).

These follow directly from the approved requirements and Kandev 0.96.0's APIs, so no new question is needed. The design applies them as listed in the summary below.

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: superseded by NFR requirements Q4. Connect uses one 12 s deadline (pre-call ≤ 1 s, Myself ≤ 10 s and never past the deadline minus 2 s, store writes ≤ 2 s together) plus a 2 s rollback on a fresh context; total ≤ 14 s. The rollback also runs when SetSecret fails in a way that may have stored the value.
- Switch: stored as plugin state (scope workspace, key `integration`); read first in every action except `connection.get` and `connection.set_enabled` and read again before Connect stores anything; a refused action returns `integration_disabled` (HTTP 409) before any secret read or Backlog call.
- UI: the settings card action renders Kandev's drafted `IntegrationEnabledControl` (id `nulab-backlog`); persist calls `connection.set_enabled` on Save and publishes with `setIntegrationEnabled` only after success; at start and on workspace changes the UI loads `connection.get` per workspace and publishes each value.
- Home entry and page: `registerNavItem` (section `integrations`, path `/backlog`) and `registerRoute('/backlog')`; the page reads `connection.get` for the active workspace only.
- Logo: an inline, script-free SVG component in the UI bundle; no runtime request.
- Logging: `integration_switched` and `action_refused_disabled` events with the base fields.

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
