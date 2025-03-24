## 1. Initial View of Change Request Selection UI

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