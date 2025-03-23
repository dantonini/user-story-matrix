---
file_path: docs/user-stories/git-integration/01-on-demand-detection-and-versioning-of-us-changes-via-git.md
created_at: 2025-03-23T19:44:32+01:00
last_updated: 2025-03-23T20:33:44+01:00
_content_hash: 9817fdbdc90d097fd524215e83a02c0a
---

# On-Demand Detection and Versioning of US Changes via Git
**Description**  
Allow the USM tool to detect modifications to User Stories (US) by leveraging Git’s change detection mechanisms, eliminating the need for a dedicated edit command or background daemon. When any USM command is executed (e.g., creating a blueprint or CR), the tool checks Git’s status/diff to determine if a US file has been modified since its last known state, then proceeds to handle versioning accordingly.

**As a** developer using USM with Git integration  
**I want** the tool to automatically detect changes to US files using Git’s built-in capabilities  
**so that** I can continue to use my preferred editor while ensuring that any modifications are captured and versioned without extra overhead

## Acceptance Criteria

- **AC1:** Upon execution of any USM command that interacts with US files, USM runs a Git command (e.g., `git status --porcelain` or `git diff`) to identify modified US files.
- **AC2:** If a US file has been modified (i.e., its current hash differs from the last recorded hash), USM automatically computes the new hash and creates a version snapshot of the previous state in the local `.usm/uss/<old_hash>.md`.
- **AC3:** USM then prompts the user to decide if any Change Requests (CRs) referencing the modified US should be updated to point to the new version.
- **AC4:** The detection mechanism must only operate on Git-tracked files, and if a file is untracked, USM should either ignore it or warn the user accordingly.
- **AC5:** The versioning check occurs on-demand when a USM command is invoked, avoiding continuous background monitoring.

## Context

- Developers use any editor of their choice, so enforcing a dedicated USM edit command or continuous monitoring is undesirable.
- Git’s robust file change detection is leveraged to simplify the workflow and maintain high accuracy in detecting modifications.
- Integrating with Git ensures that both committed and uncommitted changes are properly tracked and versioned without interfering with the user’s standard Git workflow.
