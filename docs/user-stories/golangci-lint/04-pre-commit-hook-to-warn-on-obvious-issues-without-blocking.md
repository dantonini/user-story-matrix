---
file_path: docs/user-stories/golangci-lint/04-pre-commit-hook-to-warn-on-obvious-issues-without-blocking.md
created_at: 2025-04-01T07:21:57+02:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 8bc909f29d1a39e9e6da0a0a89e6c94a25a0968fe1aa1d3f4a3c48f9fd6774b0
---

# Pre-commit hook to warn on obvious issues without blocking
As a developer
I want a pre-commit hook that runs a lightweight subset of linters
so that I get early feedback without slowing  my workflow.

## Acceptance criteria
- Hook uses golangci-lint run --fast with few critical linters
- Warnings shown, but commits not blocked
