---
file_path: docs/user-stories/custom-workflow/dev-02-extract-prompt-files-from-standard-workflow.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 8b7e57f181e7ccc95988df82a6fa37bbebd4ed05dfdcefa3f46c4c9a558521fe
---

# Extract prompt files from standard workflow
As a  
USM developer (internal),  
I want  
to extract the long prompt strings from StandardWorkflowSteps into separate Markdown files,  
So that  
they can serve as examples for the new directory-based workflow structure and improve code maintainability.

## Acceptance Criteria
- All prompts from the current `StandardWorkflowSteps` are extracted into separate Markdown files
- The files are organized in a directory structure that follows the new convention:
  ```
  internal/workflow/templates/standard/
  ├── workflow.yaml      # Generated from StandardWorkflowSteps metadata
  └── prompts/           # Directory for prompt files
      ├── 01-laying-the-foundation.md
      ├── 02-mvi.md
      └── ...
  ```
- The content of each prompt file is identical to the current prompt text
- A utility is created to generate the `workflow.yaml` file from the current `StandardWorkflowSteps`
- The `workflow.yaml` references the extracted prompt files:
  ```yaml
  name: "standard"
  description: "Standard development workflow"
  steps:
    - id: "01-laying-the-foundation"
      description: "Laying the foundation - Setting up the architecture and structure"
      prompt: "prompts/01-laying-the-foundation.md"
    # ...
  ```
- The existing `StandardWorkflowSteps` continues to work as before during the transition
- A mechanism is implemented to load prompts from files when available, falling back to embedded prompts

## Priority: MUST HAVE
This extraction is necessary to move toward the new file-based workflow structure. 
