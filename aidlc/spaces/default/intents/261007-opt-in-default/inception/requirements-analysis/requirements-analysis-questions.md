# Requirements Analysis Questions — opt-in by default

Initial request (verbatim): "hiện tại đang opt-out by default, chuyển sang opt-in sau khi cài đặt plugin"

Context from the code scan: each workspace has an integration switch. Today a workspace with no switch record is treated as ON (`Store.LoadSwitch`, `internal/connection/store.go:219`), so the plugin is active right after installation. Every action, RPC, webhook and background worker already checks the switch, so changing that default to OFF makes the whole plugin opt-in: after installing, an admin must turn the switch on (then connect) before anything runs.

## Question 1: Workspaces upgrading from v0.1.0 / v0.1.1

A workspace that already connected Backlog on v0.1.x but never touched the switch has no switch record. If the default simply becomes OFF, Backlog turns off for that workspace after the upgrade: the connection is kept, but sync, issue/PR watches and Git credentials pause until an admin turns the switch back on. What should happen to these workspaces?

A. Accept it: they turn off after upgrading; document it in the release/upgrade notes (admin turns the switch on again)
B. Grandfather them: a workspace that already has a saved connection is treated as ON; only workspaces with no connection start OFF
X. Other (please specify)

[Answer]: A. Accept it: they turn off after upgrading; document it in the release/upgrade notes (admin turns the switch on again)

## Question 2: What a freshly installed (off) workspace shows

The existing Off state already tells the admin "Turn it on with the switch above to connect" on the Settings page, and members see that Backlog is off. Is that enough, or do you want a more visible prompt?

A. Keep the existing Off state and wording as-is (smallest change)
B. Also add a more visible prompt for admins (for example on the Backlog page) pointing to the switch
X. Other (please specify)

[Answer]: A. Keep the existing Off state and wording as-is (smallest change)

## Consolidated Summary Confirmation

- Q1: v0.1.x workspaces with no switch record turn OFF after upgrading; this is accepted and documented in the release/upgrade notes (an admin turns the switch on again).
- Q2: keep the existing Off state and wording unchanged; no new prompt.
- Implied by the request: a workspace with no switch record is OFF by default (opt-in); everything already guarded by the switch stays guarded; tests, the contract test and the README are updated to match.

Does this all look correct before I generate the requirements artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
