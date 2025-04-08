# Create extended workflow template
As a  
USM developer,  
I want  
to create an "extended" workflow template with additional quality and testing steps,  
So that  
users have a more comprehensive option for complex change requests.

## Acceptance Criteria
- A new "extended" workflow template is created with a comprehensive structure:
  - More detailed steps than the standard workflow
  - Additional quality and testing phases
  - Appropriate for complex or critical changes
- The workflow structure is defined in `internal/workflow/templates/extended/workflow.yaml`
- Prompt files are created in `internal/workflow/templates/extended/prompts/`
- The workflow includes additional steps beyond the standard workflow:
  - Architecture review
  - Security considerations
  - Performance testing
  - Documentation requirements
  - Code review preparation
- The workflow reuses appropriate prompt files from the standard workflow through the shared pattern
- Each new step has a well-crafted prompt that is:
  - Detailed and specific
  - Focused on quality and thoroughness
  - Suitable for complex changes
- The workflow is registered in the `WorkflowRegistry` as a built-in template
- Documentation is provided explaining when to use this workflow
- The workflow is thoroughly tested

## Priority: SHOULD HAVE
This provides users with a more comprehensive alternative to the standard workflow. 