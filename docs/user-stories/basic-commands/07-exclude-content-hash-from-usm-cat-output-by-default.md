---
file_path: docs/user-stories/basic-commands/07-exclude-content-hash-from-usm-cat-output-by-default.md
created_at: 2025-04-05T15:20:47+02:00
last_updated: 2025-04-05T15:20:47+02:00
_content_hash: e2ce941b5792d4b8cc3f4bc1e1d609050375a8a0fbf02aa9f49fecab67ccf48f
---

# Exclude content-hash from usm cat output by default
As a user of the usm cat command,
I want content-hash field to be omitted from the output by default,
so that the aggregated view is cleaner and focused on the user story content itself.

## Acceptance criteria
- When I run usm cat <cr-file>, the content-hash field is not included in the output.
- Only the contents of the referenced user story files are printed (with optional metadata like the file path).
- This behavior applies whether the stories are printed inline or piped to another command.
