---
file_path: docs/user-stories/setup/03-basic-cli-command.md
created_at: 2025-03-17T08:33:32+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 1400e7da01de2ed7da1b6672b715893fb69f2c87f9f5f7d096509fd72b83be60
---

# Create Basic CLI Command Structure
The CLI should have a structured way to handle commands.

As a developer,  
I want a command structure for usm-cli,  
so that I can add new commands easily.

## Acceptance criteria

- The CLI uses the `cobra` package for command management.
- A root command (`usm`) is available.
- Running `usm --help` shows usage instructions.
- The CLI has a placeholder command (`usm example`) that prints "Hello, USM!".