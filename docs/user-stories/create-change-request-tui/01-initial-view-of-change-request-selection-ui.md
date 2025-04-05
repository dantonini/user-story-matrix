---
file_path: docs/user-stories/create-change-request-tui/01-initial-view-of-change-request-selection-ui.md
created_at: 2025-03-24T07:22:10+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: e7896fb05c2c6c218b772146cd753f125d3e666f8bd0288a545f0d5d0ed42ed2
---

# Initial View of Change Request Selection UI

**User Story**  
As a CLI user,  
I want to see an initial screen with a search bar, a list of user stories, and a footer with controls,  
so that I can immediately start filtering or selecting stories efficiently.

**Acceptance Criteria**
- When the user starts the `create-change-request` command, the UI shows:

    ```
	🔍 Search: 
	────────────────────────────────────────────────────
	[ ] [U] Add login functionality (usdir/usfilename)
	[ ] [U] Integrate payment provider (usdir/usfilename)
	[ ] [I] Export user data to CSV (usdir/usfilename)
	────────────────────────────────────────────────────
	Stories shown: 3 / 3 | ↑↓ to move | ␣ select | ⏎ confirm | ESC quit
    ```

- The list displays both title and implementation status
- [U] means unimplemented, [I] means implemented
- Only unimplemented stories are shown by default
- Footer always shows a short help on available keybindings