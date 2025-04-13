---
file_path: docs/user-stories/custom-workflow/ui-03-list-available-workflows.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 58c632c637aa8a2718a1325afdcf9ed3d83e02ad5441192c8ecac08239331b2f
---

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
