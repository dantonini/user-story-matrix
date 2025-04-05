---
file_path: docs/user-stories/paste-text/03-progress-indicator-during-llm-processing.md
created_at: 2025-04-05T22:23:40+02:00
last_updated: 2025-04-05T22:23:40+02:00
_content_hash: 363e5b9537009380343c63be566b843c42d1cdfe338d7e3b08ec0f3b1c383b2e
---

# Progress Indicator During LLM Processing
As a a user waiting for LLM-based parsing,
I want a progress or loading indicator,
so that I understand the system is working and not frozen.

## Acceptance criteria
- Loading Animation: A simple spinner or progress message is displayed while waiting for the LLM response.
- Timeout Warning: If parsing takes longer than a predefined threshold (e.g., 5 seconds), display a message indicating potential delays.
- Cancelable Action: The user can cancel the parsing request (e.g., press Esc) if they decide to proceed manually.
