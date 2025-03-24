## 5. Visual Cues for Interaction Mode

**User Story**  
As a CLI user,  
I want clear visual feedback for whether I'm typing or navigating,  
so that I don't get confused about the function of keys like `space`.

**Acceptance Criteria**
- In search mode:
  - Cursor is visible
  - Label shows: `🔍 Search [typing]:`
- In list mode:
  - Highlighted row is shown
  - Footer updates accordingly:
    - `Typing: ⏎ apply | Esc exit`
    - `Navigating: ↑↓ move | ␣ select | ⏎ confirm | ESC quit`