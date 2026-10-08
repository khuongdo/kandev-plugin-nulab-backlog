# Cross-Unit Traceability — 261008-gh-cli-profile

## Verdict

**PASS** — every requirement, sub-requirement, NFR and acceptance criterion is covered with status `OK` and an existing target file.

- Zero-Unit (express): the only code source is the stage-level `construction/code-generation/traceability.json`.
- The group headings FR1-FR6 are covered by their sub-requirements (FR1.1 … FR6.1).
- FR sub-requirements are covered through their user stories (`inception/user-stories/traceability.json`); every AC of those stories is `OK` in code generation with an existing target file.

## Coverage

| ID | Owning stage | Covered via | Target file(s) | Status |
|---|---|---|---|---|
| FR1.1 | code-generation | US1.1 → AC1.1.1-AC1.1.9 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR1.2 | code-generation | US1.1 → AC1.1.1-AC1.1.9 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR2.1 | code-generation | US1.1 → AC1.1.1-AC1.1.9 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR2.2 | code-generation | US1.1 → AC1.1.1-AC1.1.9 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR2.3 | code-generation | US1.2 → AC1.2.1-AC1.2.4 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR2.4 | code-generation | US2.1 → AC2.1.1-AC2.1.8 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go | OK |
| FR3.1 | code-generation | US2.1 → AC2.1.1-AC2.1.8 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go | OK |
| FR3.2 | code-generation | US1.2, US2.1 → AC1.2.1-AC1.2.4, AC2.1.1-AC2.1.8 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR3.3 | code-generation | US1.1, US2.1 → AC1.1.1-AC1.1.9, AC2.1.1-AC2.1.8 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR3.4 | code-generation | US2.1 → AC2.1.1-AC2.1.8 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go | OK |
| FR4.1 | code-generation | US3.2 → AC3.2.1-AC3.2.4 | internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR5.1 | code-generation | US3.1 → AC3.1.1-AC3.1.5 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR5.2 | code-generation | US3.1 → AC3.1.1-AC3.1.5 | internal/scm/cli_token_test.go, internal/scm/service_cli_account_test.go, ui/src/settings/source-control-section.test.tsx | OK |
| FR6.1 | code-generation | US4.1 → AC4.1.1-AC4.1.2 | README.md, ui/src/settings/source-control-section.test.tsx | OK |
| NFR1 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| NFR2 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| NFR3 | code-generation | direct | internal/scm/service_test.go | OK |
| NFR4 | code-generation | direct | Makefile | OK |
| AC1.1.1 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC1.1.2 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC1.1.3 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC1.1.4 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC1.1.5 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC1.1.6 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC1.1.7 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC1.1.8 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC1.1.9 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC1.2.1 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC1.2.2 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC1.2.3 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC1.2.4 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC2.1.1 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC2.1.2 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC2.1.3 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC2.1.4 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC2.1.5 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC2.1.6 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC2.1.7 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC2.1.8 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC3.1.1 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC3.1.2 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC3.1.3 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC3.1.4 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC3.1.5 | code-generation | direct | internal/scm/cli_token_test.go | OK |
| AC3.2.1 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC3.2.2 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC3.2.3 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC3.2.4 | code-generation | direct | internal/scm/service_cli_account_test.go | OK |
| AC4.1.1 | code-generation | direct | ui/src/settings/source-control-section.test.tsx | OK |
| AC4.1.2 | code-generation | direct | README.md | OK |

## Uncovered Elements

None.
