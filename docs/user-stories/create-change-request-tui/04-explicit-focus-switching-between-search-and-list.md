## 4. Explicit Focus Switching Between Search and List

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