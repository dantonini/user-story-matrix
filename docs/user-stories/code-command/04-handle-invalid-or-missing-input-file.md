---
file_path: docs/user-stories/code-command/04-handle-invalid-or-missing-input-file.md
created_at: 2025-03-26T01:46:08+01:00
last_updated: 2025-03-26T01:46:39+01:00
_content_hash: eb6491cb4c3794a778818c34907a4187
---

# Handle invalid or missing input file  
As a  
developer running the `code` command,  
I want  
to receive a clear error message if the change request file does not exist,  
So that  
I’m not confused by cryptic errors.

### Acceptance Criteria
- If the specified file path does not exist, the command exits with a message like:  
  `"❌ Error: File <filename> not found."`
- The command returns a non-zero exit code.