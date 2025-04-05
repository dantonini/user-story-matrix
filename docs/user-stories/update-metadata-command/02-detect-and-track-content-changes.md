---
file_path: docs/user-stories/update-metadata-command/02-detect-and-track-content-changes.md
created_at: 2025-04-04T00:29:08+02:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: c96a77d6dc823ffef6c58cb07e8af9e20fc39d7005c7081b70924bc8649f8e49
---

# Detect and Track Content Changes
As a developer
I want CLI to update change requests that reference modified user stories
so that the references remain accurate after a metadata date

## Acceptance criteria
- Change request files are scanned for references to user story files
- If a referenced user story has a new _content_hash, the corresponding reference in the change request file is updated.
- Updates are done only when hashes differ.
- CLI prints which change requests were updated.
- Change requests are located in docs/change-request/**
