---
file_path: docs/user-stories/basic-commands/06-display-aggregated-user-story-content-from-a-change-request-usm-cat.md
created_at: 2025-04-05T15:19:46+02:00
last_updated: 2025-04-05T15:19:46+02:00
_content_hash: 21210b8649384e6c758734e2a61994ab77c3327f940469d986433ebc45de862d
---

# Display aggregated user story content from a change request (usm cat)
As a CLI user of USM,
I want a usm cat <change-request-file> command that prints the content of all referenced user stories,
so that I can quickly view the full context of a change request in one place without manually opening multip

## Acceptance criteria
- When I run usm cat <file>, it prints the content of each user story referenced in the change request.
- Each story’s content is preceded by a comment with its file path (e.g. "[//]: # (docs/user-stories/path.md)").
- If the file doesn’t exist or is malformed, a helpful error is displayed.
- If no argument is provided, usage instructions are shown.
