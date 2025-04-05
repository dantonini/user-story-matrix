---
file_path: docs/user-stories/paste-text/02-llm-parsing-feedback-validation.md
created_at: 2025-04-05T22:22:42+02:00
last_updated: 2025-04-05T22:22:42+02:00
_content_hash: f9cec0c9d2f3bfcbee5e35a921e20afd40b5fcfda1d0553762e3713a4708db25
---

# LLM Parsing Feedback & Validation
As a a user who relies on automated parsing,
I want clear feedback on which fields were auto-populated and the confidence level,
so that  I can quickly review and correct any mistakes.

## Acceptance criteria
- Highlight Changed Fields: Auto-populated fields are visually highlighted until confirmed by the user.
- Manual Override: Users can easily override the suggested values by typing new text into the field.
- LLM Error Handling: If the LLM returns an error, the TUI displays a concise error message and logs the event.
