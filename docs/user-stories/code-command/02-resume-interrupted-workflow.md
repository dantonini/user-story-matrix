---
file_path: docs/user-stories/code-command/02-resume-interrupted-workflow.md
created_at: 2025-03-26T01:45:10+01:00
last_updated: 2025-03-26T01:45:30+01:00
_content_hash: 46b1f857bbc2cf415d9c5e66a71cfb25
---

# Resume interrupted workflow  
As a  
developer using USM,  
I want  
the `code` command to resume from where it left off using the `.step` file,  
So that  
I don’t lose progress if the workflow was interrupted.

### Acceptance Criteria
- If the `.step` file exists, the command uses it to determine the next step.
- If the `.step` file does not exist, the command starts from the first step.
- The `.step` file is updated only after successful completion of a step.

