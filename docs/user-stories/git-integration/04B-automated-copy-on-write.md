---
file_path: docs/user-stories/git-integration/04B-automated-copy-on-write.md
created_at: 2025-03-23T12:55:02+01:00
last_updated: 2025-03-23T21:33:34+01:00
_content_hash: 74e301876fb8aaf27cad8bd8c1ea0b4e
---

# Automated Copy On Write
**Description**  
When a user confirms that the modified US should NOT replace the old version in an active CR, USM automatically create for that US a new user story and restore the old version mentionedin the CR.

**As a** developer  
**I want** USM to take care of maintaining the consistency of the CRs referencing a US  
**so that** I don't have to manually create a new file for the US and restore the one mentioned in the CR

## Acceptance Criteria
- **AC1:** When a user modifies a user story referenced by existing change requests and chooses not to update the references, USM creates a new user story file with the modified content.
- **AC2:** The new user story file uses the next available sequential number in the directory and a slugified version of the same title.
- **AC3:** USM restores the original user story file exactly as it was referenced in the change request, including the original content hash.
- **AC4:** The user receives clear feedback about the action taken, including the path of the newly created file.
- **AC5:** The operation either completes successfully by creating the new file and restoring the original, or fails completely without partial modifications.
- **AC6:** The process takes no more than 2 seconds to complete on a standard development machine.

## Metrics
- Reduction in time spent manually creating new user story files and restoring the original ones.

## Context
- A single US might be referenced in multiple CRs at once.
- Manual updates are error-prone and time-consuming.
