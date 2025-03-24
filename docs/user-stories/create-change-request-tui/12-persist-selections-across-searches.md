---
file_path: docs/user-stories/create-change-request-tui/12-persist-selections-across-searches.md
created_at: 2025-03-24T19:38:44+01:00
last_updated: 2025-03-24T20:06:14+01:00
_content_hash: ab182ee280a630dc5fc74f0b4f1b273d
---

## Persist Selections Across Searches

**User Story**  
As a CLI user,  
I want my selected stories to remain selected even if I change the search filter,  
so that I don’t lose progress while refining my query.

**Acceptance Criteria**
- Selected stories remain selected even if hidden by current filter
- If search is cleared or changed, selection state is preserved
- Footer shows:
	```
	✔ 2 stories selected (including hidden)
	```