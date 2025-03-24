## Toggle Between All and Only Unimplemented User Stories

**User Story**  
As a CLI user,  
I want to switch between showing all user stories and only unimplemented ones,  
so that I can focus on pending work or review previously completed features when needed.

**Acceptance Criteria**
- By default, the UI only shows unimplemented user stories
- Pressing a keyboard shortcut (e.g. `CTRL+a`) toggles between:
  - `🔘 Only unimplemented`
  - `🔘 All stories`
- A filter status label is shown near the search bar or in the footer:

	```
	🔍 Search: login        (   CTRL+a to show all)
	───────────────────────────────────────────────
	[✓] [Implemented] Export to CSV
	[ ] [Unimplemented] Add login support
	───────────────────────────────────────────────
	Showing 2 / 30 | ↑↓ ␣ ⏎ Ctrl+C
	```

- When toggled, the story list refreshes immediately with current filter applied
- Search input remains intact when toggling
- Previously selected stories remain selected even if they are hidden by the current filter
- Footer always reflects the current filter and selection state, e.g.:

	```
	✔ 2 selected | 5 visible / 30 total | Filter: All | ↑↓ ␣ ⏎ f toggle | Ctrl+C
	```
