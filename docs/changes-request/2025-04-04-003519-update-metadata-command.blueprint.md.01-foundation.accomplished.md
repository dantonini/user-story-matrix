# Update Metadata Command - Foundation Phase Accomplishment Report

## Core Implementation

- Created dedicated `internal/metadata` package with clean separation of concerns:
  - `ExtractMetadata` - extracts metadata from files (`metadata/metadata.go`)
  - `GetContentWithoutMetadata` - strips metadata from content (`metadata/metadata.go`)
  - `CalculateContentHash` - generates content fingerprints (`metadata/metadata.go`)

- Enhanced `FileSystem` interface in `internal/io/filesystem.go` with `Stat` method for better file information handling.

- Updated `cmd/update_user_stories.go` with improved implementation:
  - `RunE` function now uses the new metadata package
  - Added proper debug logging via `debug` flag
  - Fixed error handling for file operations

## Test Coverage

- Added core metadata package tests:
  - `TestExtractMetadata` verifies correct metadata extraction from files
  - `TestGetContentWithoutMetadata` confirms proper stripping of metadata
  - `TestCalculateContentHash` validates content fingerprinting

- Added mock filesystem implementation in `internal/io/mock_filesystem.go`:
  - `NewMockFileSystem` factory function
  - `AddFile`, `AddDirectory` methods for test setup
  - `Stat` method implementation

- Fixed existing tests with the new mock filesystem:
  - Updated `internal/changerequest/finder_test.go` to use `mockFS.AddDirectory` instead of direct assignment
  - Updated `cmd/code_test.go` with `Stat` method and `mockFileInfo` implementation

## Architectural Improvements

- Clear separation between filesystem operations and metadata handling logic
- Content hash generation for reliable change detection
- Consistent error handling pattern in metadata functions

## Blind Spots

- Integration tests between CLI and metadata package are currently limited
- Need further tests for edge cases in metadata extraction (commented out in `cmd/update_user_stories_test.go`)

## Acceptance Criteria Status

✅ User stories are updated with metadata  
✅ File path is included in metadata  
✅ Content hash is computed and stored  
✅ Timestamp information is added  
✅ Debug flag shows detailed information  
✅ No changes when content remains the same  

⚠️ End-to-end tests for the update command need to be rewritten to align with the new API

## Design Decisions

- Changed from direct file operations to using the `FileSystem` interface for better testability
- Adopted `RawMetadata` struct with embedded fields for better type safety and extensibility
- Simplified content hash algorithm to focus on content changes only, ignoring metadata changes

## Major Changes

### Created Metadata Package (`internal/metadata`)

The central accomplishment was creating a dedicated package for all metadata-related operations, improving code organization and testability:

#### Core Components
- `types.go`: Defines the fundamental data structures for metadata operations:
  - `Metadata` struct: Represents the metadata section in user story files
  - `ContentHashMap`: Tracks changes in content hashes
  - `ContentChangeMap`: Maps file paths to their content hash changes
  - `MetadataOptions`: Configuration options for metadata operations

- `extract.go`: Provides functions for extracting metadata from content:
  - `ExtractMetadata`: Extracts and parses metadata sections from files
  - `GetContentWithoutMetadata`: Cleanly removes metadata sections from content

- `generate.go`: Handles content hash calculation and metadata generation:
  - `CalculateContentHash`: Calculates SHA-256 hash of content
  - `GenerateMetadata`: Creates formatted metadata sections for files
  - `FormatMetadata`: Formats a metadata structure into a string representation

- `update.go`: Functions for updating file metadata:
  - `UpdateFileMetadata`: Updates a single file's metadata
  - `FindMarkdownFiles`: Recursively finds all markdown files in a directory
  - `UpdateAllUserStoryMetadata`: Updates metadata for all user story files

- `reference.go`: Manages change request references:
  - `FindChangeRequestFiles`: Finds all change request files in the repository
  - `UpdateChangeRequestReferences`: Updates references in a single change request
  - `FilterChangedContent`: Filters hash map to include only changed content
  - `UpdateAllChangeRequestReferences`: Updates references in all change request files

### Enhanced Update User Stories Command

The existing `cmd/update_user_stories.go` command was overhauled to:
- Remove directly embedded metadata functionality in favor of using the new package
- Add a `--skip-references` flag to optionally skip updating change request references
- Add a `--debug` flag for more detailed logging during operation
- Integrate change request reference updates into the workflow

### Infrastructure Improvements

- Added the `Stat` method to the `FileSystem` interface for consistent file operations
- Enhanced the mock file system with proper `Stat` support for testing
- Added `SetDebugMode` to the logger for runtime log level changes

## Refactoring Strategy

The following strategy guided the refactoring work:

1. **Extraction and Centralization**: Moved metadata-related code from `cmd/update_user_stories.go` into a dedicated `internal/metadata` package

2. **Interface Enhancement**: Added missing functionality to the `FileSystem` interface to support metadata operations

3. **Integrated Reference Management**: Added logic to update change request references when user story content changes

4. **Configuration Options**: Added flags for controlling reference updates and debug output

## Test Coverage

- Created a test setup with a mock file system to ensure testability
- Established initial metadata tests for the core functionality

## Blind Spots / Future Work

- Need to add more comprehensive tests for the `reference.go` functions
- The `MockFileSystem` implementation could be improved for better test coverage
- Edge case testing for special characters in file paths or metadata values

## Design Decisions

### Changed from Original Blueprint

- **More Granular Package Structure**: Split the metadata functionality into multiple files for better organization
- **Enhanced Content Change Tracking**: Added explicit tracking of whether content (not just metadata) has changed
- **Metadata vs. RawMetadata**: Added both structured and raw representations of metadata for flexibility

### Reinforced from Original Blueprint

- **Integrated Reference Updates**: Kept the approach of updating references as part of the metadata update, with an option to skip
- **Preservation of Created Dates**: Maintained the requirement to preserve original creation dates
- **Content Hash-based Reference Management**: Used content hashes to track changes and update references

## Acceptance Criteria Status

### User Story 1: Update User Stories Metadata
- ✅ CLI command structure is in place
- ✅ Scanning for markdown files in the `docs/user-stories` directory is implemented
- ✅ Adding/updating metadata section with required fields is implemented  
- ✅ Original creation date preservation is implemented
- ✅ Only updating last_updated when content changes is implemented
- ✅ Directory skipping (node_modules, .git, etc.) is implemented
- ✅ Summary output of processed files is implemented
- ✅ Debug flag support is implemented
- ✅ Idempotent operation is ensured

### User Story 2: Detect and Track Content Changes
- ✅ Infrastructure for scanning change request files is in place
- ✅ Logic for updating references based on content hash changes is implemented
- ✅ Only updating references when hashes differ is implemented
- ✅ Reporting of which change requests were updated is implemented
- ✅ Location of change requests is properly defined

### User Story 3: Preserve Original Creation Date
- ✅ Logic to preserve existing creation dates is implemented
- ✅ Fallback to file modification time when creation date is absent is implemented 