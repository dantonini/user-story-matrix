---
file_path: docs/user-stories/golangci-lint/02-provide-a-minimal-non-intrusive-baseline-lint-config.md
created_at: 2025-04-01T07:19:07+02:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: adbb2231ceefc3f7f39493fac3078fd9313a6f1a352c8f6c204949f765b86506
---

# Provide a minimal, non-intrusive baseline lint config
As a developer new to static analysis
I want a default golangci-lint configuration with only high-signal linters enabled
so that I’m not overwhelmed by noise and can start using the tool with confidence.

## Acceptance criteria
- Initial config (.golangci.yml) includes only essential linters: deadcode, errcheck, govet, staticcheck.
- No >10 warnings per file on the initial run.
- Config explained in README or internal doc.
