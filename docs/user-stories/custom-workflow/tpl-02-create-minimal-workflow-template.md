---
file_path: docs/user-stories/custom-workflow/tpl-02-create-minimal-workflow-template.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: 860ed4a5c270c62a11b22da7cd70ffd864c6b8097c4009e479f49c8ec780055d
---

# Create minimal workflow template
As a  
USM developer (internal),  
I want  
to create a "minimal" workflow template with fewer steps,  
So that  
users have an alternative to the standard workflow for simpler change requests.

## Acceptance Criteria
- A new "minimal" workflow template is created with a streamlined structure:
  - Fewer steps than the standard workflow
  - Focus on essential development tasks
  - Appropriate for smaller changes
- The workflow structure is defined in `internal/workflow/templates/minimal/workflow.yaml`
- Prompt files are created in `internal/workflow/templates/minimal/prompts/`
- The workflow includes these essential steps:
  - Planning/foundation
  - Implementation
  - Testing
  - Linting
- Each step has a well-crafted prompt that is:
  - Clear and concise
  - Focused on the specific task
  - Suitable for smaller change requests
- The workflow is registered in the `WorkflowRegistry` as a built-in template
- Documentation is provided explaining when to use this workflow
- The workflow is thoroughly tested

## Priority: SHOULD HAVE
This provides users with a lighter alternative to the standard workflow. 
