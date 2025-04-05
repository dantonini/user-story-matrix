---
file_path: docs/user-stories/git-integration/04A-automated-change-request-updates.md
created_at: 2025-03-23T12:55:02+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 397de1034f0a2e9ce7ba9aa328628650bf284cfd697d9e113c023d3d0245a9d3
---

# Automated Change Request Updates
**Description**  
When a user confirms that the modified US should replace the old version in an active CR, USM automatically updates all references and records the action.

**As a** developer  
**I want** USM to streamline the process of updating CR references  
**so that** I don't have to manually edit every CR that points to the old US hash

## Acceptance Criteria
- **AC1:** When a user story is modified, USM automatically detects and displays a list containing:
  - The hash of the old version
  - The hash of the new version
  - All change requests that reference the old version
  - The status of each affected change request

- **AC2:** When the user confirms the update:
  - USM updates all references in the affected change requests from old hash to new hash
  - USM records the update action in the change request history with timestamp and user info
  - USM generates a summary report showing successful and failed updates
  - USM maintains the original creation date of the change request

- **AC3:** The update process provides clear feedback:
  - Success/failure status for each change request update
  - Detailed error messages for any failed updates
  - Option to retry failed updates
  - Confirmation when all updates are complete

## Metrics
- Number of CRs successfully updated automatically.
- Reduction in time spent manually editing CR references.

## Context
- A single US might be referenced in multiple CRs at once.
- Manual updates are error-prone and time-consuming.
