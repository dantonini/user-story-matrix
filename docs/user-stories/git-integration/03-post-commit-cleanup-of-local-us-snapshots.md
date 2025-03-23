---
file_path: docs/user-stories/git-integration/03-post-commit-cleanup-of-local-us-snapshots.md
created_at: 2025-03-23T12:54:08+01:00
last_updated: 2025-03-23T20:33:40+01:00
_content_hash: 178c0cf76d26966fd548ee09e3e58c96
---

# Post-Commit Cleanup of Local US Snapshots
**Description**  
When the user commits their work to Git, USM should remove or archive the local `.usm/uss` snapshots that are now safely stored in Git history.

**As a** USM user  
**I want** to automatically clean up the temporary US versions that were created prior to my commit  
**so that** my workspace remains uncluttered and consistent with the Git repository

## Acceptance Criteria
- **AC1:** After a successful `git commit`, USM checks which local US snapshots are fully captured in the commit.
- **AC2:** USM deletes or archives those snapshots from `.usm/uss`, ensuring no references are lost.
- **AC3:** A summary of cleaned-up snapshots is presented to the user, showing which versions are now in Git.

## Context
- The `.usm/uss` directory is intended to be a temporary staging area for local versions.
- Once changes are in Git, storing them locally is redundant.
