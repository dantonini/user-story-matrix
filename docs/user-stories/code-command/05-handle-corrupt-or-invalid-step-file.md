---
file_path: docs/user-stories/code-command/05-handle-corrupt-or-invalid-step-file.md
created_at: 2025-03-26T01:47:09+01:00
last_updated: 2025-03-26T01:47:26+01:00
_content_hash: eb14fbbad9df2b19d421a552db0463db
---

# Handle corrupt or invalid `.step` file  
As a  
developer relying on persisted progress,  
I want  
the command to validate and handle malformed `.step` files,  
So that  
the workflow doesn’t crash unexpectedly.

### Acceptance Criteria
- If the `.step` file exists but contains an unrecognized step name:
  - The command outputs a warning: `"⚠️ Unrecognized step in <filename>.step"`
  - It suggests deleting or resetting the file.
- If a `.step` file is empty or unreadable, it starts from step 1 with a warning.