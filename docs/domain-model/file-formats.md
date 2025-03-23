# USM File Formats

This document describes the file formats used by the User Story Matrix (USM) tool.

## User Story Format

User stories are stored as markdown files with a specific structure containing metadata and content. The metadata is stored in a YAML frontmatter at the beginning of the file.

### File Naming Convention

User story files follow this naming convention:
```
<sequential-number>-<slugified-title>.md
```

For example: `01-add-user-story.md`

- Sequential numbers start from 01 and increment for each new user story in a directory
- The title is slugified (lowercase, spaces replaced with hyphens, special characters removed)

### File Structure

```markdown
---
file_path: <relative file path>
created_at: <ISO 8601 timestamp>
last_updated: <ISO 8601 timestamp>
_content_hash: <MD5 hash of content>
---

# <User Story Title>
<Description of the user story>

As a <role>,  
I want <feature/capability>,  
so that <benefit>.

## Acceptance criteria
- <Criterion 1>
- <Criterion 2>
- ...
```

### Metadata Fields

| Field          | Description                                                 |
|----------------|-------------------------------------------------------------|
| `file_path`    | Relative path to the file in the repository                 |
| `created_at`   | ISO 8601 timestamp of when the user story was created       |
| `last_updated` | ISO 8601 timestamp of when the user story was last updated  |
| `_content_hash`| MD5 hash of the document content for integrity verification |

## Change Request Format

Change requests are stored as markdown files that link to one or more user stories. They come in two variants:
- Blueprint (`.blueprint.md`): The initial change request with a plan
- Implementation (`.implementation.md`): The executed change request with implementation details

### File Naming Convention

Change request files follow this naming convention:
```
<yyyy-mm-dd>-<HHMMSS>-<slugified-name>.<type>.md
```

For example: `2025-03-18-060000-basic-commands.blueprint.md`

- The datetime component ensures uniqueness
- The slugified name describes the change
- The type is either `blueprint` or `implementation`

### File Structure

```markdown
---
name: <Change Request Name>
created-at: <ISO 8601 timestamp>
user-stories:
  - title: <User Story Title 1>
    file: <Relative Path to User Story 1>
    content-hash: <MD5 Hash of User Story 1>
  - title: <User Story Title 2>
    file: <Relative Path to User Story 2>
    content-hash: <MD5 Hash of User Story 2>
---

# Blueprint

## Overview

This is a change request for implementing the following user stories:
1. <User Story Title 1>
2. <User Story Title 2>

<Implementation Plan, Design, and Technical Details>
```

### Metadata Fields

| Field          | Description                                               |
|----------------|-----------------------------------------------------------|
| `name`         | Human-readable name of the change request                 |
| `created-at`   | ISO 8601 timestamp of when the change request was created |
| `user-stories` | List of user story references that this change implements |

### User Story References

Each user story reference contains:
- `title`: The title of the user story
- `file`: The relative path to the user story file
- `content-hash`: The MD5 hash of the user story content at the time of reference

This allows tracking if a user story has changed since the change request was created.

## Implementation Details

The file formats and related functionality are implemented in:

- `internal/models/user_story.go`: Contains the `UserStory` model and related functions:
  - `ExtractTitleFromContent`: Extracts the title from a markdown file
  - `ExtractMetadataFromContent`: Extracts YAML frontmatter metadata
  - `GenerateContentHash`: Calculates the MD5 hash of content
  - `SlugifyTitle`: Converts a title to a slug for filenames
  - `GenerateFilename`: Creates a filename for a user story
  - `GenerateUserStoryTemplate`: Creates the template content for a new user story

- `internal/models/change_request.go`: Contains the `ChangeRequest` model and related functions:
  - `GenerateChangeRequestTemplate`: Creates the template content for a new change request
  - `GenerateChangeRequestFilename`: Creates a filename for a change request
  - `LoadChangeRequestFromContent`: Parses a change request from file content
  - `GetPromptInstruction`: Generates AI instructions for a change request 