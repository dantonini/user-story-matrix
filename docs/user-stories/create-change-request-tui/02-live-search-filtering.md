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
