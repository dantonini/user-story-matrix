---
file_path: docs/user-stories/basic-commands/01-add-user-story.md
created_at: 2025-03-17T19:47:09+01:00
last_updated: 2025-03-18T05:00:52+01:00
_content_hash: 219341c58d4f1186c239b98ce9289108
---

# Add a user story
The CLI should have a structured way to add a user story.

As a cli user,  
I want a command to add a user story,  
so that I can add new user stories easily.

## Acceptance criteria

- The CLI has a command to add a user story: `usm add user-story`
- The command can optionally accept a directory `--into` as an argument:
  - If the directory is provided, the command saves the user story in the given directory.
  - If no directory is provided, the command saves the user story in the default directory: `docs/user-stories`
- The command ask for the following information:
  - Title
- The command saves the user story in markdown format in the directory specified in the command or in the default one( `docs/user-stories`) if not specified:
  - If the directory does not exist, the command creates it.
  - The file name to use is: `<sequential-number-starting-from-01>-<user-story-title>.md`
- The command creates a template for a a user story with the following content:
  ```
  # <User Story Title>
  Add here a description of the user story
  
  As a ... 
  I want ... 
  so that ...
  
  ## Acceptance criteria
  - Add here your acceptance criteria
  - As many as needed
  ```
- The command prints a success message in the console.