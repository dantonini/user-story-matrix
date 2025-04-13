---
file_path: docs/user-stories/custom-workflow/core-02-reuse-prompt-files-with-variables.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 9fc75b8dc5135b2b0c9cec2ca1c14723d5a5a68a229cbb1f7b3532ff7d98a404
---

# Reuse prompt files with variable substitution
As a  
CLI user,  
I want  
to reuse the same prompt file in multiple workflow steps with different variables,  
So that  
I can maintain consistency across similar steps while customizing their behavior for different contexts.

## Acceptance Criteria
- Workflow steps can specify variables that are substituted into the prompt template:
  ```yaml
  steps:
    - id: "01-test-foundation"
      description: "Test foundation code"
      prompt: "prompts/shared/test.md"
      variables:
        phase: "foundation"
        focus_areas: "architecture, interfaces"
    - id: "02-test-implementation"
      description: "Test implementation code"
      prompt: "prompts/shared/test.md"
      variables:
        phase: "implementation"
        focus_areas: "core functionality, error handling"
  ```
- Prompt files support Go template syntax for variable substitution (`{{.variable_name}}`)
- Variables are properly escaped to prevent injection vulnerabilities
- If a variable is referenced in a template but not provided in the step definition, the system provides a warning
- Default values for variables can be specified using the template syntax: `{{.variable_name | default "default value"}}`
- Variables support both simple strings and structured data (arrays, maps)

## Priority: MUST HAVE
This capability is essential for creating maintainable, DRY workflow definitions. 
