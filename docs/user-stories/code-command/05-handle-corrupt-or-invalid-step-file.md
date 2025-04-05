---
file_path: docs/user-stories/code-command/05-handle-corrupt-or-invalid-step-file.md
created_at: 2025-03-26T01:47:09+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: fde2ee6fc26f0ffc2d8a2054f40a277f696cbd3d6c80126e10b0ec73af56cdf9
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