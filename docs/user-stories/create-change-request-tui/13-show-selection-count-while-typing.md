---
file_path: docs/user-stories/create-change-request-tui/13-show-selection-count-while-typing.md
created_at: 2025-03-24T19:39:23+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 4e9b33c83a2e5dd0816cdef1a8b1c3665eb0f50fbcdce8406be7b3571fce29e6
---

# Show Selection Count While Typing

**User Story**  
As a CLI user,  
I want to see how many stories I've selected even while searching,  
so that I can keep track of my selection state.

**Acceptance Criteria**
- Footer always includes:
	```
	✔ 3 selected | 5 visible / 30 total | ↑↓ ␣ ⏎ Ctrl+C
	```
- Count updates as items are selected/deselected