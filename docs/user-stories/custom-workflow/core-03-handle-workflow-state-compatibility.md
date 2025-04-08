# Handle workflow state compatibility
As a  
CLI user,  
I want  
the system to properly handle transitions between different workflows,  
So that  
I can switch workflows without losing progress or encountering errors.

## Acceptance Criteria
- The workflow state file (.step) includes the name of the workflow being used
- When switching to a different workflow:
  - If the new workflow has the same or similar step IDs, progress is maintained
  - If the new workflow has completely different steps, a warning is displayed and progress is reset
  - Users have the option to force reset with `--reset` flag
- The system attempts to map between similar workflows by matching step IDs
- When a workflow is not found (e.g., it was deleted), the system falls back to the built-in workflow
- When a step references a prompt file that no longer exists, a clear error message is displayed
- The system provides clear warnings when workflow compatibility issues are detected

## Priority: SHOULD HAVE
This ensures a smooth user experience when using multiple workflows. 
