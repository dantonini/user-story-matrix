---
file_path: docs/user-stories/basic-commands/04-recap.md
created_at: 2025-03-18T06:44:02+01:00
last_updated: 2025-03-18T06:44:02+01:00
_content_hash: d4297aa8e8199f3407a189bbe882f0ae
---

# Recap what you have done
The CLI should have a command to instruct the ai-assistant editor to recap what have done.

As a CLI user,  
I want a command to recap what have done,  
so that I can review my work.

Recap what you did in a file in docs/changes-requests/2025-03-18-060000-basic-commands.implementation.md

## Acceptance criteria
- The command should be `recap`
- The command look for an "incomplete" change request in the docs/changes-requests directory and use it as a base to generate the recap: an incomplete change request is a change request that has a blueprint file but no implementation file.
- When no incomplete change request is found, the command should print a fancy congratulation message.
- When more than one incomplete change request is found, the command should allow the user to select one of them.
- When the user select one of the incomplete change requests, the command should print the following output:
  ```
  Recap what you did in a file in docs/changes-requests/<change-request-name>.implementation.md
  ```