---
file_path: docs/user-stories/create-change-request-tui/02-live-search-filtering.md
created_at: 2025-03-24T07:21:59+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2
---

# Live Search Filtering

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
