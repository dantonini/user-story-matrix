---
file_path: docs/user-stories/create-change-request-tui/07-exiting-the-ui-gracefully.md
created_at: 2025-03-24T07:25:38+01:00
last_updated: 2025-03-24T07:25:38+01:00
_content_hash: 14dc62c151c0eb3a454fb012077122ef
---

## 7. Exiting the UI Gracefully

**User Story**  
As a CLI user,  
I want to exit the interface at any time using a known key,  
so that I can safely cancel the operation.

**Acceptance Criteria**
- Pressing `ESC`:
  - Cancels the operation
  - Displays: `Change request creation canceled by user.`
  - Discards all current selections