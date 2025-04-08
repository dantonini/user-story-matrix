# Implement template variables support
As a  
USM developer (internal),  
I want  
to implement support for variables in prompt templates,  
So that  
users can create reusable prompt files that can be customized for different steps.

## Acceptance Criteria
- The `WorkflowStep` struct is updated to include variables:
  ```go
  type WorkflowStep struct {
      ID          string
      Description string
      Prompt      string            // Path to prompt file
      Variables   map[string]string // Variables for template substitution
  }
  ```
- A template processing system is implemented:
  ```go
  func ApplyTemplateVariables(promptContent string, variables map[string]string) (string, error)
  ```
- The template system uses Go's text/template package
- Template variables use the {{.variable_name}} syntax
- The system supports:
  - Basic variable substitution
  - Default values with the `default` function: {{.variable_name | default "default value"}}
  - Conditional sections with if/else: {{if .variable_name}}...{{else}}...{{end}}
  - Iteration over arrays: {{range .items}}...{{end}}
- Variables are properly escaped to prevent injection vulnerabilities
- The system provides warnings for undefined variables
- The template processing handles errors gracefully
- Documentation is provided for template syntax and available functions

## Priority: MUST HAVE
This functionality is required for creating reusable prompt templates. 
