---
file_path: docs/user-stories/create-change-request-tui/03-entering-search-mode-separates-typing-from-selection.md
created_at: 2025-03-24T07:21:48+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 1f9a34087be1e027edf4ef0b979b3a846a9c17fb722f176bbb6439561279c663
---

# Entering Search Mode Separates Typing from Selection

**User Story**  
As a CLI user,  
I want the interface to distinguish between typing in the search bar and selecting from the list,  
so that I can type full queries (including spaces) without triggering selection.

**Acceptance Criteria**
- Typing enters "search input mode"
- In search mode:
  - `space` inserts a space character
  - `Enter` or `Esc` exits typing mode and returns focus to list
- Only when focus is in the list does `space` toggle item selection
- Visual cue: cursor visible in search bar, e.g.

	```
	🔍 Search [typing]: login page _
	```