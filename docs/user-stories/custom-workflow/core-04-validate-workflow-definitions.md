# Validate workflow definitions
As a  
CLI user,  
I want  
the system to validate workflow definitions and prompt files,  
So that  
I can identify and fix errors in my custom workflows before using them.

## Acceptance Criteria
- A new `workflow validate` command is available:
  ```
  usm workflow validate [name or path]
  ```
- The command performs comprehensive validation of a workflow:
  - Checks that workflow.yaml has all required fields (name, description, steps)
  - Verifies that all step IDs are unique
  - Ensures all referenced prompt files exist
  - Validates that prompt templates have valid syntax
  - Checks that all variables used in templates are provided in step definitions
- The command displays detailed error messages for each validation issue found:
  - File paths where errors were found
  - Line numbers when possible
  - Specific error descriptions
  - Suggested fixes
- The command returns:
  - Exit code 0 if validation passes
  - Exit code 1 if validation fails
- The validation is also performed automatically when:
  - Importing a workflow
  - Selecting a workflow for the code command
  - Initializing a new workflow

## Priority: SHOULD HAVE
Validation improves user experience by catching errors early. 
