# Performance Test Instructions — Plugin install failed: 502

## Scope at Minimal Strategy

No performance test suite is generated: the strategy is Minimal and there are no NFR performance requirements for runtime behaviour. The one size-related target is the package size cap (NFR1), measured as part of the build.

## How to Check

```bash
make package && ls -l dist/nulab-backlog-*.tar.gz
```

## Expected Result

Package size at most 25 MB (25,000,000 bytes). Context: Kandev cuts an upload after 30 s (`server.readTimeout`), so a 23.4 MB package uploads only on links faster than about 0.78 MB/s; slower links use install From URL (README).
