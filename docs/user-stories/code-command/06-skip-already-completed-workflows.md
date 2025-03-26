---
file_path: docs/user-stories/code-command/06-skip-already-completed-workflows.md
created_at: 2025-03-26T01:47:34+01:00
last_updated: 2025-03-26T01:48:04+01:00
_content_hash: 6e8cd2abd13cb5ea59b707b823986ab1
---

# Skip already completed workflows  
As a  
developer running a completed workflow,  
I want  
the command to recognize when all steps are finished,  
So that  
it doesn’t repeat or break the flow.

### Acceptance Criteria
- If `.step` contains the final step, the command:
  - Outputs: `"✅ All steps completed successfully for change request: <filename>"`
  - Returns a success code without running anything.