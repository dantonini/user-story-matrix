---
file_path: docs/user-stories/git-integration/04B-automated-copy-on-write.md
created_at: 2025-03-23T12:55:02+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 744ba56c8a616f776462ea1a83cfcbeaf0b6502b85f2badaaad846faa53219a6
---

# Automated Copy On Write
**Description**  
When a user confirms that the modified US should NOT replace the old version in an active CR, USM automatically create for that US a new user story and restore the old version mentionedin the CR.

**As a** developer  
**I want** USM to take care of maintaining the consistency of the CRs referencing a US  
**so that** I don't have to manually create a new file for the US and restore the one mentioned in the CR

## Acceptance Criteria
- **AC1:** Given a user story referenced by existing change requests, when the user modifies it and chooses not to update references, then USM creates a new user story file with the modified content.
- **AC2:** Given a directory containing user stories, when creating a new user story file, then USM assigns the next available sequential number and uses a slugified version of the original title (e.g., "01-my-story.md" -> "02-my-story.md").
- **AC3:** Given a modified user story referenced by a CR, when the user chooses to preserve the CR reference, then USM restores the original file with its exact content and hash as referenced in the CR.
- **AC4:** Given a successful copy-on-write operation, when the process completes, then USM displays a confirmation message showing:
  - Path to the newly created file
  - Path to the restored original file
  - Status of the operation
- **AC5:** Given a copy-on-write operation, when any step fails (file creation, restoration, or validation), then USM:
  - Rolls back any partial changes
  - Restores the initial state
  - Reports the specific error and failure point
- **AC6:** Given a standard development machine (>= 16GB RAM, SSD), when performing a copy-on-write operation, then USM completes the entire process in less than 0.5 seconds.

## Metrics
- Reduction in time spent manually creating new user story files and restoring the original ones.

## Context
- A single US might be referenced in multiple CRs at once.
- Manual updates are error-prone and time-consuming.
