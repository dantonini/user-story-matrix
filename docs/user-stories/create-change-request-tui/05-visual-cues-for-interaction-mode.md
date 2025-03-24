---
file_path: docs/user-stories/create-change-request-tui/05-visual-cues-for-interaction-mode.md
created_at: 2025-03-24T07:23:50+01:00
last_updated: 2025-03-24T07:23:50+01:00
_content_hash: 54af54a9da736693c4b109a7eb447e29
---

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