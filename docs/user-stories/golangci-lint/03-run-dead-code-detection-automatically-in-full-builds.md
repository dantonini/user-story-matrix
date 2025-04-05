---
file_path: docs/user-stories/golangci-lint/03-run-dead-code-detection-automatically-in-full-builds.md
created_at: 2025-04-01T07:20:13+02:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 1326e697672d9e2a375b7653c0b7d183bff5b0a754350e23b6bed23a7791718a
---

# Run dead code detection automatically in full builds
As a tech lead
I want to automatically detect and remove dead code using golangci-lint
so that we can keep the codebase clean and reduce maintenance overhead.

## Acceptance criteria
- Dead code reports are part of the build output
- deadcode linter is active in make build-full
- Optional auto-removal instructions or tooling support provided
