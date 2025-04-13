---
file_path: docs/user-stories/custom-workflow/ui-04-initialize-new-workflow.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 2f5cd322093a9ea6b351504df172963f87ba0ee808b2a1b6bed02f6d75d8bf3e
---

# Initialize new workflow
As a  
CLI user,  
I want  
to initialize a new workflow from a template,  
So that  
I can quickly create a custom workflow without starting from scratch.

## Acceptance Criteria
- A new `workflow init` command is available:
  ```
  usm workflow init [name] [--template=standard]
  ```
- The command creates a new workflow directory with the given name in `.usm/workflows/` 
- If the `--global` flag is provided, the workflow is created in `~/.usm/workflows/` instead
- The command creates a minimal workflow structure:
  - `workflow.yaml` with basic metadata and empty steps array
  - `prompts/` directory for prompt files
  - `README.md` with instructions for customizing the workflow
- If a template is specified, the workflow is initialized with a copy of that template:
  - Standard (default): The standard 4-phase development workflow
  - Minimal: A simplified workflow with fewer steps
  - Extended: A comprehensive workflow with additional testing and quality steps
- The command fails with a clear error message if:
  - A workflow with the given name already exists
  - The specified template does not exist
- After creation, the command displays instructions for next steps

## Priority: SHOULD HAVE
This improves workflow developer experience but isn't required for core functionality. 
