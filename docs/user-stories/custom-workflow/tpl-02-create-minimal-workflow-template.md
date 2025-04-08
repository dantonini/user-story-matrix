# Create minimal workflow template
As a  
USM developer (internal),  
I want  
to create a "minimal" workflow template with fewer steps,  
So that  
users have an alternative to the standard workflow for simpler change requests.

## Acceptance Criteria
- A new "minimal" workflow template is created with a streamlined structure:
  - Fewer steps than the standard workflow
  - Focus on essential development tasks
  - Appropriate for smaller changes
- The workflow structure is defined in `internal/workflow/templates/minimal/workflow.yaml`
- Prompt files are created in `internal/workflow/templates/minimal/prompts/`
- The workflow includes these essential steps:
  - Planning/foundation
  - Implementation
  - Testing
  - Linting
- Each step has a well-crafted prompt that is:
  - Clear and concise
  - Focused on the specific task
  - Suitable for smaller change requests
- The workflow is registered in the `WorkflowRegistry` as a built-in template
- Documentation is provided explaining when to use this workflow
- The workflow is thoroughly tested

## Priority: SHOULD HAVE
This provides users with a lighter alternative to the standard workflow. 
