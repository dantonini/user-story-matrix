---
file_path: docs/user-stories/code-command/07-allow-manual-override-or-reset.md
created_at: 2025-03-26T01:48:54+01:00
last_updated: 2025-03-26T01:49:08+01:00
_content_hash: 6458c3afa0055497e0e0807f2169309c
---

# Allow manual override or reset  
As a  
developer who made a mistake or wants to restart a workflow,  
I want  
to manually reset or override the `.step` file,  
So that  
I can rerun from the beginning or from a specific point.

### Acceptance Criteria
- If the user deletes the `.step` file, the workflow restarts from step 1.
- (Optional enhancement) Support a `--reset` flag to explicitly restart from the beginning.
