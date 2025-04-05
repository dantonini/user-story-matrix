---
file_path: docs/user-stories/update-metadata-command/01-update-user-stories-metadata.md
created_at: 2025-03-17T20:22:58+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 629e68d30703ed913d7cf1f102dbdc4ec1bdef1a92ac2b6004cc09c76534df27
---

# Update User Stories Metadata

The CLI should automatically update metadata in user story files to track their creation, modification, and content changes.

As a developer,  
I want a command to update metadata in user story files,  
so that I can track when files were created, last modified, and detect content changes.

## Acceptance criteria

- The CLI has a command to update metadata in user story files.
- Running `usm update user-stories metadata` scans for all markdown files in the `docs/user-stories` directory and subdirectories.
- The command adds or updates a metadata section at the top of each markdown file with:
  - File path (relative to project root)
  - Creation date
  - Last edited date
  - Content hash (hidden with underscore prefix)
- The metadata section uses the following format:
  ```
  ---
  file_path: docs/user-stories/path/to/file.md
  created_at: 2023-01-01T12:00:00Z
  last_updated: 2023-01-02T12:00:00Z
  _content_hash: abcdef1234567890
  ---
  ```
- The command preserves the original creation date if already present.
- The command only updates the `last_updated` date when the content has actually changed.
- The command skips directories like `node_modules`, `.git`, `dist`, and `build`.
- The command prints a summary of processed files, showing which were updated and which had no changes.
- The command supports a `--debug` flag to show detailed information about the processing.
- The command is idempotent (running it multiple times without content changes won't modify files).