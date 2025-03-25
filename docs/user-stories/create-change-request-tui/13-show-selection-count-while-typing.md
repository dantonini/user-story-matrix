---
file_path: docs/user-stories/create-change-request-tui/13-show-selection-count-while-typing.md
created_at: 2025-03-24T19:39:23+01:00
last_updated: 2025-03-25T07:48:26+01:00
_content_hash: 92697041b479593b2c5c3648196b8c45
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