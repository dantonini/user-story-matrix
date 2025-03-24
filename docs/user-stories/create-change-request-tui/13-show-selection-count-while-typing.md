## 13. Show Selection Count While Typing

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