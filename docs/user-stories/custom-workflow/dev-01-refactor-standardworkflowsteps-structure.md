# Refactor StandardWorkflowSteps structure
As a  
USM developer (internal),  
I want  
to refactor the current StandardWorkflowSteps implementation into a more modular structure,  
So that  
I can prepare the codebase for supporting custom workflow definitions without breaking existing functionality.

## Acceptance Criteria
- The current `StandardWorkflowSteps` global variable is refactored into a structured format:
  ```go
  // WorkflowDefinition represents a complete workflow
  type WorkflowDefinition struct {
      Name        string
      Description string
      Steps       []WorkflowStep
  }
  ```
- A registry is created to manage available workflows:
  ```go
  // WorkflowRegistry manages available workflows
  type WorkflowRegistry struct {
      builtInWorkflows map[string]*WorkflowDefinition
  }
  ```
- The existing `StandardWorkflowSteps` is converted to a "standard" workflow definition
- The `WorkflowManager` is updated to work with a `WorkflowDefinition` instead of directly with `StandardWorkflowSteps`
- All existing tests that reference `StandardWorkflowSteps` continue to pass
- Backward compatibility is maintained through interface adapters
- The `WorkflowManager` constructor accepts an optional workflow definition parameter
- When no workflow is specified, it falls back to the standard workflow

## Priority: MUST HAVE
This foundational refactoring is required before implementing custom workflow definitions. 
