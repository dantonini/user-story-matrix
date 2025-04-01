---
file_path: docs/user-stories/golangci-lint/01-integrate-golangci-lint-in-the-ci-cd-pipeline-with-optional-usage.md
created_at: 2025-04-01T07:08:12+02:00
last_updated: 2025-04-01T07:08:12+02:00
_content_hash: 586a2d13669c351f01ffee13ae3967db
---

# Integrate golangci-lint in the CI/CD pipeline with optional usage
As a developer
I want to have golangci-lint integrated into the build pipeline with three distinct commands
so that I can control when static analysis runs and avoid blocking builds unnecessarily.

## Acceptance criteria
- The Makefile (or build script) supports: make build (no lint), make build-full (includes lint), make lint (lint only)
- CI supports all three modes.
