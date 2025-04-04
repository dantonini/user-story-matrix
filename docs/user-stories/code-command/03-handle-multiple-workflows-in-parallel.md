---
file_path: docs/user-stories/code-command/03-handle-multiple-workflows-in-parallel.md
created_at: 2025-03-26T01:45:38+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: c5a522b39a682a181f2412d88a6adafe59a6e16677114908b4b6547c9d97cc31
---

# Handle multiple workflows in parallel  
As a  
developer working on several change requests,  
I want  
each change request file to maintain its own independent state,  
So that  
I can switch between them without losing track of progress.

### Acceptance Criteria
- The `.step` file is created per input file.
- No cross-contamination occurs between different workflows.
- Invoking the command with different change request files does not interfere with each other’s step progress.
