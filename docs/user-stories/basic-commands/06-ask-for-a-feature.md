---
file_path: docs/user-stories/basic-commands/06-ask-for-a-feature.md
created_at: 2025-03-18T09:06:34+01:00
last_updated: 2025-03-18T09:17:12+01:00
_content_hash: 15f634bcb0df76f11241f94d3987865c
---

# Ask for a feature
This command should be used to submit a feature suggestion to the CLI developer

As a user
I want to be able to submit a feature suggestion to the CLI developer
so that I can suggest a new feature for my use case

## Acceptance criteria
- The command should be able to submit a feature suggestion to the CLI developer
- The user should provide a user story about the feature they want to see in the CLI
- The command ask for:
  - The title of the feature
  - A description of the feature
  - The reason why this feature is important to the user
  - The As a ... I want ... so that ... (required)
  - The acceptance criteria of the feature
- The command should save the data inputted by the user in a hidden file so that the user can later edit it
- The command can be interrupted by the user at any time and the data inputted by the user should be kept in the file
- Once the command has collected all the data, it should slack it to the CLI developer 
- The command should have a confirmation step where the user should confirm the feature suggestion before it is sent to the CLI developer