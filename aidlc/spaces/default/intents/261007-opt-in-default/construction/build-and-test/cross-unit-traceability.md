# Cross-Unit Traceability — Opt-in by default

## Verdict

**PASS with one deferred item.** Zero-Unit run: coverage comes from the stage-level `construction/code-generation/traceability.json`. User Stories did not execute, so there are no AC IDs. 12 of 13 IDs are `OK` with an existing target file; FR5.2 is `Deferred` to the Deployment Pipeline stage, which is scheduled next in this workflow.

## Per-ID Coverage

| ID | Status | Owner | Target | Target exists |
|---|---|---|---|---|
| FR1 | OK | code-generation (stage-level) | internal/connection/store.go | yes |
| FR1.1 | OK | code-generation | internal/connection/store_test.go | yes |
| FR1.2 | OK | code-generation | internal/plugin/opt_in_test.go | yes |
| FR2 | OK | code-generation | internal/plugin/opt_in_test.go | yes |
| FR3 | OK | code-generation | ui/src/switch/integration-switch.tsx | yes |
| FR4 | OK | code-generation | ui/src/settings/settings.test.tsx | yes |
| FR5 | OK | code-generation | README.md | yes |
| FR5.1 | OK | code-generation | README.md | yes |
| FR5.2 | Deferred | deployment-pipeline (next stage) | release notes for the release shipping this fix | n/a |
| FR6 | OK | code-generation | internal/ci/contract.go | yes |
| NFR1 | OK | code-generation | internal/plugin/opt_in_test.go | yes |
| NFR2 | OK | code-generation | internal/connection/fakes_test.go | yes |
| NFR3 | OK | code-generation | internal/ci/contract_test.go | yes |

## Uncovered Elements

- **FR5.2** (release/upgrade notes state the v0.1.x upgrade effect and remedy): not yet `OK`. Owned by the Deployment Pipeline stage; the README "Upgrading from v0.1.x" note is the source text. Also noted by the Code Generation review (R-02): the note must sit under the version that ships this fix.
