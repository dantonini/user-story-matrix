# Deprecate StandardWorkflowSteps
As a  
USM developer (internal),  
I want  
to deprecate the direct use of StandardWorkflowSteps while maintaining backward compatibility,  
So that  
the codebase can transition to the new workflow system without breaking existing code.

## Acceptance Criteria
- The `StandardWorkflowSteps` global variable is marked as deprecated:
  ```go
  // StandardWorkflowSteps defines the predefined sequence of steps in the implementation workflow
  // Deprecated: Use WorkflowRegistry.GetWorkflow("standard") instead.
  var StandardWorkflowSteps = []WorkflowStep{...}
  ```
- Static analysis linter rules are added to flag direct usage
- A compatibility layer is implemented to map between:
  - The legacy `StandardWorkflowSteps` array
  - The new workflow system
- Existing code that accesses `StandardWorkflowSteps` continues to work
- New code is guided to use the new workflow API
- Documentation is updated to reflect the deprecation
- Migration guides are provided for updating existing code
- Unit tests verify that both old and new approaches work correctly
- A plan is established for complete removal in a future version

## Priority: MUST HAVE
This deprecation approach ensures a smooth transition to the new system. 
