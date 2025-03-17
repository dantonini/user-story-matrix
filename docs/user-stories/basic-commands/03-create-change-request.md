---
file_path: docs/user-stories/basic-commands/03-create-change-request.md
created_at: 2025-03-17T20:01:25+01:00
last_updated: 2025-03-17T20:01:25+01:00
_content_hash: 0953f51df28eec3d0eefbb829ae62fec
---

# Create a change request
The CLI should have a structured way to create a change request.

As a developer,  
I want a command to create a change request,  
so that I can create a change request easily.

## Acceptance criteria

- The CLI has a command to create a change request.
- Running `usm create change-request` shows usage instructions.
- The command prints the user stories available in the `docs/user-stories` directory and subdirectories.
- The command asks for the change request name.
- The command allows to select one or more user stories.
- Once the user stories are selected, the command create a change request file in the `docs/change-requests` directory: using the following name: yyyy-mm-dd-HHMMSS<user-story-title>.blueprint.md
- The change request file is created with:
  - A metadata section which includes all the referenced user stories.
  ```
  ---
  name: <change-request-name>
  user-stories:
    - title: <user-story-title-1>
      file: <user-story-path-to-file-1>
    - ...
    - title: <user-story-title-n>
      file: <user-story-path-to-file-n>
  created-at: <created-at>
  ---

  Read all the user stories files, analyze them according to the codebase, and define a detailed plan for the change.
  Ensure to include the steps required to satisfy the acceptance criteria of all the user stories.
  ```
- The command prints a success message in the console.