---
file_path: docs/user-stories/paste-text/01-pasting-unstructured-text.md
created_at: 2025-04-05T22:14:52+02:00
last_updated: 2025-04-05T22:55:16+02:00
_content_hash: 72103a47ad4818b1ed09f79483da977cc3047a1e11d285fa1fdf9c5e7340cdcf
---

# Pasting Unstructured Text
As a a user with unstructured text,
I want to paste my text into the TUI and have it processed by OpenAI gpt4o-mini using the OpenAI structured output feature,
so that my text is automatically formatted and structured into the form fields.

## Acceptance criteria
- Paste Detection: When the user pastes text, the application recognizes the input and triggers a call to OpenAI gpt4o-mini.
- API Key Configuration:
    - The user can input and store their own OpenAI API key in the settings.
    - If no API key is configured, a small, non-intrusive message is displayed (e.g., “API key not configured. Please set your OpenAI API key in the settings to enable auto-formatting.”).
- Auto-Population: Relevant fields in the TUI are filled with structured data returned by the LLM.
- Fallback: If the LLM cannot parse the text, the application notifies the user and leaves the fields blank for manual entry.
- Performance: The form updates within 2 seconds of pasting (under normal network conditions).
