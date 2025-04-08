# Provide built-in workflow templates
As a  
USM developer (internal),  
I want  
the system to provide a set of built-in workflow templates,  
So that  
I can quickly start with a workflow tailored to my specific needs.

## Acceptance Criteria
- The system includes at least three built-in workflow templates:
  - `standard`: The default 4-phase workflow with foundation, MVI, extend, and refine
  - `minimal`: A streamlined workflow with fewer steps for smaller changes
  - `extended`: A comprehensive workflow with additional quality and testing steps
- Built-in templates are available without any configuration
- Built-in templates can be selected with the `--workflow` flag
- Built-in templates can be used as a base for creating custom workflows
- Each built-in template includes:
  - Comprehensive prompt files
  - Clear descriptions
  - Appropriate validation for each step
- The system documents the purpose and structure of each built-in template
- Templates are designed for different project sizes and complexity levels

## Priority: SHOULD HAVE
This provides valuable starting points for users but isn't essential for custom workflow functionality. 
