---
file_path: docs/user-stories/custom-workflow/dev-07-migrate-legacy-workflows.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: d81329cb7346731a47953760696e80817f6be08a52e58da23823ca713ce1e29a
---

# Migrate legacy workflows
As a  
USM developer (internal),  
I want  
the system to migrate from the legacy hardcoded workflow to the new custom workflow system,  
So that  
existing projects continue to work without disruption while gaining the benefits of the new system.

## Acceptance Criteria
- The system seamlessly migrates from the legacy hardcoded workflow to the new system:
  - The current `StandardWorkflowSteps` is converted to a built-in workflow template
  - Existing state files (.step) are compatible with the new workflow system
  - No manual migration is required for users
- When an existing state file is detected without workflow information:
  - The system assumes the "standard" workflow
  - A warning is logged for debugging purposes
  - The state file is updated to include workflow information
- The system provides backward compatibility for direct code references to `StandardWorkflowSteps`:
  - A deprecation warning is logged
  - The code continues to function by referencing the "standard" workflow
- A migration command is available for advanced users:
  ```
  usm workflow migrate [path-to-state-file] --to=[workflow-name]
  ```
- The migration process is well-documented

## Priority: MUST HAVE
This ensures backwards compatibility with existing projects. 
