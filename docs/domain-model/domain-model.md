# USM Domain Model

This document describes the key entities and relationships in the User Story Matrix (USM) domain model.

## Core Entities

### UserStory

The `UserStory` represents a single user story with its metadata and content.

```go
type UserStory struct {
    Title            string    // The user story title
    FilePath         string    // The file path where the user story is stored
    ContentHash      string    // MD5 hash of the content for integrity verification
    SequentialNumber string    // The sequential number in the file name
    CreatedAt        time.Time // When the user story was created
    LastUpdated      time.Time // When the user story was last updated
    Content          string    // The full content of the user story
}
```

### ChangeRequest

The `ChangeRequest` represents a planned or implemented change that fulfills one or more user stories.

```go
type ChangeRequest struct {
    Name        string               // Human-readable name of the change request
    CreatedAt   time.Time            // When the change request was created
    UserStories []UserStoryReference // References to the user stories included
    FilePath    string               // The file path where the change request is stored
}
```

### UserStoryReference

The `UserStoryReference` represents a reference to a user story within a change request.

```go
type UserStoryReference struct {
    Title       string // Title of the referenced user story
    FilePath    string // File path to the referenced user story
    ContentHash string // Hash of the user story content at reference time
}
```

## Relationships

```
┌────────────┐      references      ┌────────────────────┐
│            │ 1..*            1..* │                    │
│ UserStory  ├─────────────────────►│ ChangeRequest      │
│            │                      │                    │
└────────────┘                      └────────────────────┘
                                            │
                                            │ contains
                                            │ 1..*
                                            ▼
                                    ┌────────────────────┐
                                    │                    │
                                    │ UserStoryReference │
                                    │                    │
                                    └────────────────────┘
```

## Filesystem Structure

USM typically organizes files in the following structure:

```
docs/
├── user-stories/
│   ├── feature-area-1/
│   │   ├── 01-first-user-story.md
│   │   └── 02-second-user-story.md
│   └── feature-area-2/
│       ├── 01-another-user-story.md
│       └── 02-yet-another-user-story.md
└── changes-request/
    ├── 2025-03-18-060000-feature-implementation.blueprint.md
    └── 2025-03-18-060000-feature-implementation.implementation.md
```

- User stories are organized in directories by feature area or category
- Change requests are stored in a flat structure with datetime-prefixed filenames
- Each change request has a blueprint version and may have an implementation version

## Implementation Details

The domain model is implemented in:

- `internal/models/user_story.go`: Contains the `UserStory` model and related utility functions
- `internal/models/change_request.go`: Contains the `ChangeRequest` model and related utility functions

Both models include functionality for serialization, deserialization, validation, and various utility operations needed to manage the lifecycle of these entities. 