---
file_path: docs/user-stories/basic-commands/06B-ask-for-a-feature.md
created_at: 2025-03-18T09:06:34+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: fdfd4c477fb5b8ddaf321972444982c07d42e8e13e3dca9e54d119fc0b0f7c35
---

# Save and Resume Feature Request Drafts
As a user
I want my partially completed feature request to be saved automatically
So that I can resume it later if I interrupt the process

## Acceptance Criteria
- The command should save user input in a hidden file.
- If the user interrupts the process (e.g., using Ctrl+C), the entered data should not be lost.
- When the user restarts the command, they should be able to continue from where they left off.