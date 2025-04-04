---
file_path: docs/user-stories/create-change-request-tui/12-persist-selections-across-searches.md
created_at: 2025-03-24T19:38:44+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: a103c3852b62f1d2b7e0c6ac9ac8a07356db44be2e2bc2870a6203b4c5165112
---

# Persist Selections Across Searches

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