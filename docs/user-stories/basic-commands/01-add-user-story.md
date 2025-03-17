---
file_path: docs/user-stories/basic-commands/01-add-user-story.md
created_at: 2025-03-17T19:47:09+01:00
last_updated: 2025-03-17T19:47:09+01:00
_content_hash: afd7615331ce267a7c54c5c71fe854fe
---

# Add a user story
The CLI should have a structured way to add a user story.

As a developer,  
I want a command to add a user story,  
so that I can add new user stories easily.

## Acceptance criteria

- The CLI has a command to add a user story.
- Running `usm add user-story` shows usage instructions.
- The command ask for the following information:
  - Title
  - Description
  - As a ... I want ... so that ...
  - Acceptance criteria
- The command saves the user story in markdown format in the `docs/user-stories` directory.
- The command creates the file with the next format:
  ```
  # User Story Title
  Description
  
  As a ... I want ... so that ...
  
  ## Acceptance criteria
  - ...
  ```
- The command prints a success message in the console.