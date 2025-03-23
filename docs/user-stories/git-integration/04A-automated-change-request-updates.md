---
file_path: docs/user-stories/git-integration/04A-automated-change-request-updates.md
created_at: 2025-03-23T12:55:02+01:00
last_updated: 2025-03-23T21:29:02+01:00
_content_hash: dbd7b87dd8bbc9537bd8fdac526bfdbd
---

# Automated Change Request Updates
**Description**  
When a user confirms that the modified US should replace the old version in an active CR, USM automatically updates all references and records the action.

**As a** developer  
**I want** USM to streamline the process of updating CR references  
**so that** I don’t have to manually edit every CR that points to the old US hash

## Acceptance Criteria
- **AC1:** USM lists all CRs referencing the old US version when a new version is created.
- **AC2:** A single confirmation action updates all CR references to the new US version.

## Metrics
- Number of CRs successfully updated automatically.
- Reduction in time spent manually editing CR references.

## Context
- A single US might be referenced in multiple CRs at once.
- Manual updates are error-prone and time-consuming.
