# Select workflow for code command
As a  
CLI user,  
I want  
to select which workflow to use when running the `code` command,  
So that  
I can apply different workflows to different projects or change requests without modifying my configuration.

## Acceptance Criteria
- The `code` command accepts a `--workflow` flag to specify which workflow to use:
  ```
  usm code --workflow=my-custom-workflow path/to/blueprint.md
  ```
- If no workflow is specified, the default "standard" workflow is used
- The system provides a clear error message if the specified workflow does not exist
- The system searches for workflows in the following locations (in order):
  1. `.usm/workflows/` in the current directory (project-specific workflows)
  2. `~/.usm/workflows/` in the user's home directory (user-specific workflows)
  3. Built-in workflows provided by the USM application
- The selected workflow is used for all subsequent steps in the current execution of the `code` command
- The workflow selection is stored in the state file, so subsequent executions of the `code` command use the same workflow

## Priority: MUST HAVE
This is required to enable users to actually use custom workflows. 
