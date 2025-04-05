---
file_path: docs/user-stories/basic-commands/02-list-user-stories.md
created_at: 2025-03-17T19:48:52+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 0f4dc91b4fcb02919df0f7f06833ae6f47d2286a716938d39476e2cbc3850879
---

# List user stories
The CLI should have a structured way to list user stories.

As a cli user,  
I want a command to list user stories,  
so that I can list all user stories easily.

## Acceptance criteria

- The CLI has a command to list user stories.
- Running `usm list user-stories` shows usage instructions.
- The command list all user stories in the `docs/user-stories` directory and subdirectories.
- The command prints the user stories in the console
- The command can optionally accept a directory `--from` as an argument:
  - If the directory is provided, the command list all user stories in the given directory and subdirectories.
  - If no directory is provided, the command list all user stories in the `docs/user-stories` directory and subdirectories.
