---
file_path: docs/user-stories/basic-commands/02-list-user-stories.md
created_at: 2025-03-17T19:48:52+01:00
last_updated: 2025-03-17T19:48:52+01:00
_content_hash: bc2adc9822f01669af109a95bdcd047b
---

# List user stories
The CLI should have a structured way to list user stories.

As a developer,  
I want a command to list user stories,  
so that I can list all user stories easily.

## Acceptance criteria

- The CLI has a command to list user stories.
- Running `usm list user-stories` shows usage instructions.
- The command list all user stories in the `docs/user-stories` directory and subdirectories.
- The command prints the user stories in the console.
- The command can optionally accept a directory as an argument:
  - If the directory is provided, the command list all user stories in the given directory and subdirectories.
  - If no directory is provided, the command list all user stories in the `docs/user-stories` directory and subdirectories.
