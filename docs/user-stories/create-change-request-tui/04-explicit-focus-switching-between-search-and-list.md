---
file_path: docs/user-stories/create-change-request-tui/04-explicit-focus-switching-between-search-and-list.md
created_at: 2025-03-24T07:22:46+01:00
last_updated: 2025-03-24T20:05:42+01:00
_content_hash: c608c0eb94c6a52b3923f96fff8963fc
---

## Explicit Focus Switching Between Search and List

**User Story**  
As a CLI user,  
I want to explicitly switch between search input and list navigation modes using the `Tab` key,  
so that I have clear control over interaction mode.

**Acceptance Criteria**
- `Tab` toggles focus between:
  - Search input
  - Story list
- In search mode:
  - Keyboard input affects the search bar
- In list mode:
  - Arrow keys navigate
  - `space` toggles selection

	Example:
	```
	🔍 Search: login page        ← Focused
	→ List:
	[ ] Add login page
	[✓] Fix profile view
	```