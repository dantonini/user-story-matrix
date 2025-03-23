---
file_path: docs/user-stories/git-integration/02-pre-commit-consistency-checks.md
created_at: 2025-03-23T16:29:07+01:00
last_updated: 2025-03-23T22:14:12+01:00
_content_hash: 6d64c146852c672b1cb7cf053863bdc8
---

# Pre Commit Consistency Checks

**Description**  
Implement a Git pre-commit hook that detects and manages situations where a developer is about to commit changes to a Change Request (CR) without also committing the corresponding User Stories (US). The system should provide both safety guardrails and the flexibility to override them when needed.

**As a** developer using USM with Git  
**I want** to be warned when I'm about to create inconsistent commits, with the option to override when necessary  
**so that** I can maintain repository consistency by default while retaining the flexibility to handle edge cases

## Acceptance Criteria

1. **Pre-commit Detection**
   - Given staged CR files in Git
   - When the pre-commit hook runs
   - Then USM identifies any referenced US files that are:
     - Modified but not staged
     - Referenced but don't exist
   And displays their paths and status

2. **Warning and Options**
   - Given USM detects unstaged US changes
   - When displaying the warning
   - Then it shows:
     ```
     Warning: The following US files are referenced but not staged:
     - path/to/us1.md (modified)
     - path/to/us2.md (missing)
     
     Options:
     1. Stage missing files (y)
     2. Skip check with --no-verify (n)
     3. Abort commit (q)
     ```

3. **Interactive Staging**
   - Given unstaged US changes
   - When the user selects "Stage missing files"
   - Then Git automatically stages the required files
   - And shows a summary of newly staged files

4. **Override Configuration**
   - Given the repository has a `.usm/config.yaml`
   - Users can configure:
     ```yaml
     git:
       pre-commit:
         enabled: true|false        # Enable/disable checks
         log-level: debug|info|warn # Logging verbosity
         auto-stage: true|false     # Auto-stage without prompting
     ```

5. **Audit Logging**
   - Given a commit proceeds with missing US references
   - When using --no-verify
   - Then USM logs to `.usm/logs/consistency-checks.log`:
     - Timestamp
     - Git commit hash
     - List of missing/unstaged files
     - User ID
   - In a machine-parseable format

## Context

- Git's pre-commit hook provides the ideal point to enforce consistency
- Developers often stage partial changes, which can lead to broken references
- Some valid workflows require temporary inconsistency (e.g., work-in-progress commits)
- The system should encourage good practices without being overly restrictive
