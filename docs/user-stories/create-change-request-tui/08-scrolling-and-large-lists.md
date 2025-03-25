---
file_path: docs/user-stories/create-change-request-tui/08-scrolling-and-large-lists.md
created_at: 2025-03-24T07:26:27+01:00
last_updated: 2025-03-25T07:48:12+01:00
_content_hash: 9e6d0ca1739819d5c8a040ddafb8be07
---

# Scrolling and Large Lists

**User Story**  
As a CLI user,  
I want to scroll or paginate through large story lists,  
so that I can view and access all items even if they don't fit on screen.

**Acceptance Criteria**
- List supports:
  - Arrow key scroll
  - PageUp/PageDown for faster navigation
- UI shows current visible range:
	```
	Stories 21–40 / 100 | ↑↓ scroll | PgUp/PgDn fast scroll
	```