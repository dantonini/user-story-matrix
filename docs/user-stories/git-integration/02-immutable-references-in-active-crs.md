---
file_path: docs/user-stories/git-integration/02-immutable-references-in-active-crs.md
created_at: 2025-03-23T12:53:38+01:00
last_updated: 2025-03-23T20:33:40+01:00
_content_hash: 459f1c2a93b95d6dd231f4f0420a9034
---

# Immutable References in Active CRs
**Description**  
When a US is referenced by an active Change Request (CR), updating the US should enforce immutability of the original referenced version while optionally allowing a new version to be created.

**As a** USM user  
**I want** to preserve the existing US version referenced by an active CR  
**so that** the CR’s plan and implementation remain consistent with the original story

## Acceptance Criteria
- **AC1:** If a referenced US is modified, USM automatically creates a new version (with a new hash) instead of altering the existing one.
- **AC2:** The active CR continues to point to the old hash, and I am prompted to decide whether to link the CR to the new version or keep the old.
- **AC3:** I can view both versions and their hashes in `.usm/uss/<hash>.md`.

## Context
- Active CRs may be partially implemented based on a specific story version.
- Changing the US mid-development can disrupt or invalidate the CR if not handled properly.
