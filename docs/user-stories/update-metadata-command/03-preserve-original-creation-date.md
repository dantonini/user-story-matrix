---
file_path: docs/user-stories/update-metadata-command/03-preserve-original-creation-date.md
created_at: 2025-04-04T00:33:16+02:00
last_updated: 2025-04-04T00:33:16+02:00
_content_hash: 3fb2387aa05443306b3a853e06f3760adbaa25f95905366de7f05f24f1955a42
---

# Preserve Original Creation Date
As a developer
I want the created_at date in metadata to remain unchanged,
so that I don’t lose track of when the story was first written.

## Acceptance criteria
- If created_at is present in metadata, it remains untouched.
- If absent, a new ISO 8601 timestamp is added based on the file’s creation time (fallback: git log or file stat).
