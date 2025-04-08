# List available workflows
As a  
CLI user,  
I want  
to list all available workflows with their descriptions,  
So that  
I can see what workflows are available and choose the appropriate one for my task.

## Acceptance Criteria
- A new `workflow list` command is available:
  ```
  usm workflow list
  ```
- The command displays a table of available workflows with:
  - Workflow name
  - Description
  - Source (built-in, user, or project)
  - Path (for user and project workflows)
- Built-in workflows are clearly marked as such
- User workflows (from ~/.usm/workflows/) are marked as "user"
- Project workflows (from .usm/workflows/) are marked as "project"
- The table is sorted alphabetically by workflow name
- If no workflows are found in a particular source, this is indicated in the output
- The command supports a `--format` flag to output in different formats (text, json)

## Priority: SHOULD HAVE
This is important for usability but not essential for core functionality. 
