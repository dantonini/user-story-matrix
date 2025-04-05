---
file_path: docs/user-stories/create-change-request-tui/07-exiting-the-ui-gracefully.md
created_at: 2025-03-24T07:25:38+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 7a5cdd266ccf8b42780da330b173e9a10741c5224c7560a9b367ab91c0e85889
---

# Exiting the UI Gracefully

**User Story**  
As a CLI user,  
I want to exit the interface at any time using a known key,  
so that I can safely cancel the operation.

**Acceptance Criteria**
- Pressing `ESC`:
  - Cancels the operation
  - Displays: `Change request creation canceled by user.`
  - Discards all current selections