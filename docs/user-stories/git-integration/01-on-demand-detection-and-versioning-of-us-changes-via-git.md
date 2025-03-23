---
file_path: docs/user-stories/git-integration/01-on-demand-detection-and-versioning-of-us-changes-via-git.md
created_at: 2025-03-23T19:44:32+01:00
last_updated: 2025-03-23T21:48:56+01:00
_content_hash: 49c0c05820b4da750af61b80d183f69e
---

# On-Demand Detection and Versioning of US Changes via Git
**Description**  
Allow the USM tool to detect modifications to User Stories (US) by leveraging Git's change detection mechanisms, eliminating the need for a dedicated edit command or background daemon. When any USM command is executed (e.g., creating a blueprint or CR), the tool checks Git's status/diff to determine if a US file has been modified since its last known state, then proceeds to handle versioning accordingly.

**As a** developer using USM with Git integration  
**I want** the tool to automatically detect changes to US files using Git's built-in capabilities  
**so that** I can continue to use my preferred editor while ensuring that any modifications are captured and versioned without extra overhead

## Acceptance Criteria

- **AC1: Git-based Change Detection**
  - Given a US file that has been modified
  - When any USM command is executed
  - Then USM should detect the change using `git status --porcelain`
  - And report the specific files that were modified

- **AC2: Hash-based Version Control**
  - Given a modified US file is detected
  - When USM computes its new content hash
  - Then USM should create a version snapshot at `.usm/uss/<old_hash>.md`
  - And update the metadata in the current US file with the new hash
  - And log the version transition for audit purposes

- **AC3: CR Reference Management**
  - Given a US file has been versioned with a new hash
  - When USM finds existing CRs referencing the old hash
  - Then USM should display a list of affected CRs
  - And prompt the user with clear options to:
    - Update all CR references to the new hash
    - Selectively update specific CR references
    - Skip the update process

- **AC4: Untracked File Handling**
  - Given a US file exists but is not tracked by Git
  - When USM detects this during command execution
  - Then USM should display a warning message with:
    - The file path
    - Instructions to track the file with Git
    - Option to proceed or abort the current operation

- **AC5: Performance Requirements**
  - Given any USM command is executed
  - When checking for US file modifications
  - Then the Git-based detection should complete within 500ms
  - And consume no more than 50MB of memory
  - And not interfere with concurrent Git operations

- **AC6: Error Handling**
  - Given Git operations fail during change detection
  - When USM encounters the error
  - Then USM should:
    - Display a user-friendly error message
    - Log the detailed error for debugging
    - Provide specific recovery steps
    - Maintain data consistency by rolling back any partial changes

## Context

- Developers use any editor of their choice, so enforcing a dedicated USM edit command or continuous monitoring is undesirable.
- Git's robust file change detection is leveraged to simplify the workflow and maintain high accuracy in detecting modifications.
- Integrating with Git ensures that both committed and uncommitted changes are properly tracked and versioned without interfering with the user's standard Git workflow.
