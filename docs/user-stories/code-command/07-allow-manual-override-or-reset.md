---
file_path: docs/user-stories/code-command/07-allow-manual-override-or-reset.md
created_at: 2025-03-26T01:48:54+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: c562379b3bc2afc7ce77963b6f3c7005cfbb0a670c9514b67f1bce044d273895
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
