---
file_path: docs/user-stories/basic-commands/03-create-change-request.md
created_at: 2025-03-17T20:01:25+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 8c2078c32a059d21ae7a041a0e8276d5e469882a00ef223a9509770c3d66f2b7
---

# Create a change request
The CLI should have a structured way to create a change request.

As a CLI user,  
I want a command to create a change request,  
so that I can start working on a change request.

## Acceptance criteria

- The CLI has a command to create a change request.
- Running `usm create change-request` shows usage instructions.
- The command can optionally accept a directory `--from` as an argument:
  - If the directory is provided, the command reads the user stories from the given directory.
  - If no directory is provided, the command reads the user stories from the default directory: `docs/user-stories`
- The command asks for the change request name.
- The command prints the user stories available in the directory specified in the command or in the default one if not specified.
- The command allows to select one or more user stories (something like gh cli)
- Once the user stories are selected, the command create a change request file in the `docs/change-requests` directory: using the following name: yyyy-mm-dd-HHMMSS-<user-story-title>.blueprint.md
- The change request file is created with:
  - A metadata section which includes all the referenced user stories.
  ```
  ---
  name: <change-request-name>
  created-at: <created-at>
  user-stories:
    - title: <user-story-title-1>
      file: <user-story-path-to-file-1>
      content-hash: <content-hash-1>
    - ...
    - title: <user-story-title-n>
      file: <user-story-path-to-file-n>
      content-hash: <content-hash-n>
  ---
  ```
- The command shows a prompt instruction to the user so that an ai-powered editor can consume the file and continue to work on it.

  Read all the <count> user stories files in the change request <change-request-file-path>, validate them against the codebase, and define a detailed plan for the change. Store the plan in the change request file <change-request-file-path> in markdown format in a section called "Blueprint".
  Ensure to include the steps required to satisfy the acceptance criteria of all mentioned user stories.
