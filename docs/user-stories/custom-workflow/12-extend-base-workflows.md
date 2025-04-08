# Extend base workflows
As a  
software engineer using USM,  
I want  
to extend existing workflows with additional or overridden steps,  
So that  
I can customize workflows without duplicating all their content.

## Acceptance Criteria
- Workflow definitions can include an `extends` field to inherit from a base workflow:
  ```yaml
  name: "my-custom-workflow"
  description: "Extended version of the standard workflow"
  extends: "standard"
  steps:
    - id: "01-laying-the-foundation"  # Override existing step
      description: "Custom foundation step"
      prompt: "prompts/custom-foundation.md"
    - id: "05-additional-step"  # Add new step
      description: "Additional custom step"
      prompt: "prompts/additional-step.md"
  ```
- When extending a workflow:
  - Steps with the same ID override the base workflow's steps
  - Steps with new IDs are added to the workflow
  - Steps not specified are inherited from the base workflow
- The system provides a warning if the base workflow cannot be found
- Multiple levels of inheritance are supported (up to a reasonable limit)
- Circular dependencies are detected and prevented
- The system provides a `--flatten` option for the `workflow show` command to display the fully resolved workflow

## Priority: COULD HAVE
This advanced feature improves maintainability but isn't essential for initial implementation. 