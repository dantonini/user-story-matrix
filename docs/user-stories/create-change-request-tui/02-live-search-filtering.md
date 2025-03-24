---
file_path: docs/user-stories/create-change-request-tui/02-live-search-filtering.md
created_at: 2025-03-24T07:21:59+01:00
last_updated: 2025-03-24T07:21:59+01:00
_content_hash: cab84d112716971eed823a3f90b000d0
---

## 2. Live Search Filtering

**User Story**  
As a CLI user,  
I want real-time filtering while typing in the search bar,  
so that I can quickly narrow down to relevant user stories.

**Acceptance Criteria**
- Typing in the search bar filters the list in real-time
- Filtering is case-insensitive and supports partial word matches
- Matches apply to:
  - User story titles
  - Descriptions
  - Acceptance criteria
- If no matches are found:

    ```
	🔍 Search: login  
	────────────────────────────────────────────────────
	⚠️  No matching user stories found.
	────────────────────────────────────────────────────
	Stories shown: 0 / 15 | ↑↓ to move | ␣ select | ⏎ confirm | ESC quit
    ```
