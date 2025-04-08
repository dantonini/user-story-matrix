# Workflow documentation command
As a  
software engineer using USM,  
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