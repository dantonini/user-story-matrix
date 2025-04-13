---
file_path: docs/user-stories/custom-workflow/dev-04-update-workflow-state-format.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 720fca05da86821139e0e123c5624b4a5ab74f2e7a95f10e247f8a457b57d3ac
---

# Update workflow state format
As a  
USM developer (internal),  
I want  
to update the workflow state file format to include workflow identification,  
So that  
the system can associate state files with specific workflow definitions.

## Acceptance Criteria
- The `WorkflowState` struct is updated to include workflow identification:
  ```go
  type WorkflowState struct {
      ChangeRequestPath string    // Path to the change request file
      CurrentStepIndex  int       // Index of the current step (0-based)
      LastModified      time.Time // When the state was last updated
      CompletedSteps    []string  // List of completed step IDs
      WorkflowName      string    // Name of the workflow being used
      WorkflowPath      string    // Optional path to the workflow definition
  }
  ```
- The state file format (.step files) is updated to include the new fields
- Backward compatibility is maintained:
  - When reading an existing state file without workflow information, the system assumes the "standard" workflow
  - The state file is upgraded to the new format when saved
- The `WorkflowManager` methods are updated to work with the new state format:
  - `LoadState()` detects and handles both old and new formats
  - `SaveState()` always uses the new format
  - `UpdateState()` preserves workflow identification
- When switching workflows, the system:
  - Validates compatibility between the old and new workflow
  - Attempts to map progress between workflows based on step IDs
  - Provides warnings when compatibility issues are detected

## Priority: MUST HAVE
This update is essential for tracking which workflow is associated with each change request. 
