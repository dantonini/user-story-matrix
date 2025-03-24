## 12. Persist Selections Across Searches

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