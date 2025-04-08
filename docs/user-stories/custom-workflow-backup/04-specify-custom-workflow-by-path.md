# Specify custom workflow by path
As a  
software engineer using USM,  
I want  
to specify a custom workflow by directly providing its path,  
So that  
I can easily test new workflows or use one-off workflows without installing them in the standard locations.

## Acceptance Criteria
- The `code` command accepts a `--workflow-path` flag to specify the path to a workflow directory:
  ```
  usm code --workflow-path=/path/to/my-workflow path/to/blueprint.md
  ```
- The system validates that the specified path exists and contains a valid workflow definition
- The system provides clear error messages if:
  - The specified path does not exist
  - The path does not contain a valid workflow.yaml file
  - Any referenced prompt files are missing
- The `--workflow-path` flag takes precedence over the `--workflow` flag if both are provided
- Relative paths are resolved relative to the current working directory
- The workflow path is stored in the state file for future executions

## Priority: SHOULD HAVE
This enhances workflow development and testing capabilities but isn't essential for the core feature. 