---
file_path: docs/user-stories/basic-commands/05-implement-the-change-request.md
created_at: 2025-03-18T07:20:09+01:00
last_updated: 2025-03-18T07:20:09+01:00
_content_hash: ecc0149e53cd4719c59e60624d4ca26b
---

# Implement the change request
The CLI should have a command to instruct the ai-assistant editor to implement the change request.

As a CLI user,  
I want a command to implement the change request,  
so that the ai-assistant editor can implement the change request.

## Acceptance criteria
- The command should be `implement`
- The command look for an "incomplete" change request in the docs/changes-requests directory and use it as a base to generate the implementation: an incomplete change request is a change request that has a blueprint file but no implementation file.
- When no incomplete change request is found, the command should print a sad message saying that there is no change request to implement.
- When more than one incomplete change request is found, the command should allow the user to select one of them.
- When the user select one of the incomplete change requests, the command should print the following output:
  ```
  Read the blueprint file in [<change-request-name-blueprint>.md](mdc:full/file/path/to/<change-request-name-blueprint>.md)
  Read all the mentioned user stories, validate the blueprint against the codebase and proceed with the implementation.
  ```