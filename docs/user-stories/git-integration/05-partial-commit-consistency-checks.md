---
file_path: docs/user-stories/git-integration/05-partial-commit-consistency-checks.md
created_at: 2025-03-23T16:29:07+01:00
last_updated: 2025-03-23T20:33:44+01:00
_content_hash: 921aed83165fb21010e78e5a39790091
---

# Partial Commit Consistency Checks

**Description**  
Provide a mechanism that detects when a developer is about to commit changes to a Change Request (CR) without also committing the corresponding updated User Stories (US), and issue a warning or prompt to maintain consistency.

**As a** developer using USM in combination with Git  
**I want** USM to automatically detect and warn me when I commit CR references without committing the updated US they point to  
**so that** I don’t end up with CRs referencing a version of a US that isn’t actually present in the same commit

## Acceptance Criteria

- **AC1:** When I stage and commit files related to a CR, USM checks if any CR references a US hash that has been modified but is not being staged.
- **AC2:** If a mismatch is detected (i.e., CR references a new US hash but the US file is not in the commit), USM displays a clear warning with the option to:
  1. Stage and include the missing US file(s).
  2. Proceed with the commit anyway (e.g., using a `--force` or “Confirm” option).
- **AC3:** If I choose to force the commit without the updated US, USM logs a notice (in CLI output or a local .usm log) indicating potential reference mismatch.
- **AC4:** If no mismatch is found, the commit proceeds without interruption.

## Context

- In Git, developers frequently stage only part of their changes before a commit, which can lead to accidental omission of necessary US files.
- A CR may reference multiple USs, and each US might have its own modifications in `.usm/uss/<hash>.md`.
- This feature ensures that partial commits do not result in broken references or incomplete changesets.
