---
file_path: docs/user-stories/git-integration/06-forced-partial-commit-override.md
created_at: 2025-03-23T16:29:46+01:00
last_updated: 2025-03-23T20:33:40+01:00
_content_hash: 63eb7f00e2e9c4091a64670872c7837e
---

# Forced Partial Commit Override

**Description**  
Allow developers to bypass the consistency checks with an explicit override if they intentionally want to commit only part of the CR/US changes, acknowledging the mismatch risk.

**As a** power user  
**I want** to confirm my decision to commit CR changes without including all referenced US files  
**so that** I can manage my work in small increments, even if that temporarily introduces inconsistency in my local repository

## Acceptance Criteria

- **AC1:** When USM detects an inconsistency, it offers a `--force` (or similar) option to proceed.
- **AC2:** A short explanatory message must be shown, clarifying the potential issues of committing partial references.
- **AC3:** The USM CLI logs the override decision for auditability (e.g., in `.usm/logs` or in the commit message).
- **AC4:** Users can disable or enable force-commit behavior at a global configuration level if desired.

## Context

- Advanced developers may have valid reasons for committing CR changes incrementally (e.g., stubbing out references to future work).
- This override prevents the USM tool from being too restrictive while still warning about best practices.
