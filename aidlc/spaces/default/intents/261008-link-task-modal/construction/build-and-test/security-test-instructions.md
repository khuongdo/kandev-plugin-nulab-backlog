# Security Test Instructions — Link Task modal, GitHub-style

## Scope

The new user input is the task-side "issue key or link" field (FR2). Two security requirements apply (NFR1):

- accept only `https` Backlog hosts (project Mandated rule);
- never show backend response content or secrets in error text.

## Threat Notes (STRIDE, input boundary)

- **Spoofing / tampering via look-alike URLs.** The parser accepts only `https:` with a host ending in `.backlog.com`, `.backlog.jp` or `.backlogtool.com`, and a path of exactly `/view/<KEY>`. `backlog.com.example.com`, `http://` and other hosts are rejected before any backend call. The URL is never fetched; only the extracted key is sent.
- **Information disclosure.** Errors from `issues.link` map to catalogue messages. The backend `detail` field is never shown.
- **Injection.** The key must match `^[A-Z][A-Z0-9_]*-[1-9][0-9]*$` before it is sent. React escapes rendered text.

## How to Run

```bash
cd ui && npx vitest run src/issues/issue-link.test.ts
```

The relevant cases: rejected `http://`, rejected foreign host, rejected look-alike host, rejected garbage or empty, and a check that the error's `detail` never appears in the thrown message.

The existing Go redaction tests (`internal/redact`, and the "no secrets in responses" test) run under `make coverage`.

No dependency vulnerability scan, per the team decision (Testing Posture).
