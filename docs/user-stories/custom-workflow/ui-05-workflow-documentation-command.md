---
file_path: docs/user-stories/custom-workflow/ui-05-workflow-documentation-command.md
created_at: 2025-04-13T09:15:31+02:00
last_updated: 2025-04-13T09:53:56+02:00
_content_hash: c33a188e106e98c1b81c768278eb8edd66eff03cd8a14d9462485d557df60a29
---

# Workflow documentation command
As a  
CLI user,  
I want  
to view documentation about available workflows and their steps,  
So that  
I can understand how each workflow functions without examining the source files.

## Acceptance Criteria
- A new `workflow show` command is available:
  ```
  usm workflow show [name]
  ```
- The command displays:
  - Workflow name and description
  - Number of steps in the workflow
  - List of all steps with their IDs and descriptions
  - A note about any variables used in step prompts
- The output is formatted for readability in the terminal
- The command supports different output formats:
  - `--format=text` (default): Formatted text for terminal display
  - `--format=markdown`: Markdown output for documentation
  - `--format=json`: JSON output for programmatic use
- The command fails with a clear error message if the workflow does not exist

## Priority: COULD HAVE
This improves discoverability and understanding but isn't essential for core functionality. 
