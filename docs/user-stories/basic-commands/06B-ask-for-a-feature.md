---
file_path: docs/user-stories/basic-commands/06B-ask-for-a-feature.md
created_at: 2025-03-18T09:06:34+01:00
last_updated: 2025-03-18T19:02:20+01:00
_content_hash: 88e542ac744a671d176ee588edff67d4
---

# Save and Resume Feature Request Drafts
As a user
I want my partially completed feature request to be saved automatically
So that I can resume it later if I interrupt the process

## Acceptance Criteria
- The command should save user input in a hidden file.
- If the user interrupts the process (e.g., using Ctrl+C), the entered data should not be lost.
- When the user restarts the command, they should be able to continue from where they left off.