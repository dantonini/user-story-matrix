---
file_path: docs/user-stories/git-integration/03-post-commit-cleanup-of-local-us-snapshots.md
created_at: 2025-03-23T12:54:08+01:00
last_updated: 2025-03-23T21:56:36+01:00
_content_hash: 7227c0f34b4108d255b7795fdd4e04d9
---

# Post-Commit Cleanup of Local US Snapshots
**Description**  
When the user commits their work to Git, USM should remove the local `.usm/uss` snapshots that are now safely stored in Git history.

**As a** USM user  
**I want** to automatically clean up the temporary US versions that were created prior to my commit  
**so that** my workspace remains uncluttered and consistent with the Git repository

## Acceptance Criteria
- **AC1:** After a successful `git commit`:
  - USM identifies all US snapshots in `.usm/uss` that are fully captured in the commit
  - The identification is done by comparing file content hashes and metadata
  - A list of matched snapshots is prepared for cleanup

- **AC2:** For each identified snapshot:
  - The snapshot and its metadata are permanently removed from `.usm/uss`
  - All associated temporary files and references are cleaned up
  - The removal is atomic - either all files for a snapshot are removed or none

- **AC3:** After cleanup completion:
  - A summary is displayed showing: total snapshots processed, removed count, and any skipped items
  - Each removed snapshot entry shows: original path and the commit hash where it was integrated
  - If any snapshots couldn't be removed, clear error messages explain why

- **AC4:** The cleanup process:
  - Can be run manually via `usm cleanup` command
  - Supports a `--dry-run` flag to preview changes without executing them
  - Can be disabled via configuration if automatic cleanup is not desired
  - Provides a `--force` flag to remove snapshots without checking Git history

## Context
- The `.usm/uss` directory is intended to be a temporary staging area for local versions
- Once changes are in Git, storing them locally is redundant as Git becomes the source of truth
- Users can always retrieve previous versions from Git history if needed
