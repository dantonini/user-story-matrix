---
name: update metadata command
created-at: 2025-04-04T00:35:19+02:00
user-stories:
  - title: Update User Stories Metadata
    file: docs/user-stories/update-metadata-command/01-update-user-stories-metadata.md
    content-hash: d266bf5244a4433071e267b9919814e1
  - title: Detect and Track Content Changes
    file: docs/user-stories/update-metadata-command/02-detect-and-track-content-changes.md
    content-hash: ee6eced0f69ee22d373845bc6fa3e79eb67cf0bb58b683dcc1121f451fb4e01c
  - title: Preserve Original Creation Date
    file: docs/user-stories/update-metadata-command/03-preserve-original-creation-date.md
    content-hash: 3fb2387aa05443306b3a853e06f3760adbaa25f95905366de7f05f24f1955a42

---

# Blueprint

## Overview

This change request aims to enhance the metadata management capabilities of the USM CLI tool. The proposed changes will improve how user story files track their history and ensure that references to these files in change requests remain accurate over time. 

The key themes across these user stories are:
- Ensuring proper tracking of user story creation and modification dates
- Maintaining the integrity of content-based references between files
- Providing a mechanism to keep change request references up-to-date

The implementation will build upon the existing metadata command, extending it to also update change request references when user story content changes.

## Fundamentals

### Data Structures

#### UserStoryMetadata
```
{
  file_path: string      // Relative path to the file
  created_at: string     // ISO 8601 timestamp of creation
  last_updated: string   // ISO 8601 timestamp of last update
  _content_hash: string  // SHA-256 hash of file content (without metadata)
}
```

#### ChangeRequestUserStoryReference
```
{
  title: string          // Title of the user story
  file: string           // Path to user story file
  content-hash: string   // Content hash of referenced user story
}
```

#### ContentHashMap
```
{
  filePath: string       // Path to user story file
  oldHash: string        // Previous content hash
  newHash: string        // New content hash
  changed: boolean       // Whether content has changed
}
```

### Algorithms

#### Update User Story Metadata Algorithm
```
function updateUserStoryMetadata(userStoryFilePaths):
  updatedFiles = []
  unchangedFiles = []
  updatedHashMap = {} // track file paths and their old/new hashes
  
  foreach file in userStoryFilePaths:
    content = readFile(file)
    existingMetadata = extractMetadata(content)
    contentWithoutMetadata = getContentWithoutMetadata(content)
    newContentHash = calculateHash(contentWithoutMetadata)
    
    // Check if content has changed
    oldContentHash = existingMetadata._content_hash || ""
    contentChanged = (oldContentHash != newContentHash)
    
    // Preserve original creation date if present
    creationDate = existingMetadata.created_at || getCurrentTimestamp()
    
    // Only update last_updated if content changed
    lastUpdated = contentChanged ? getCurrentTimestamp() : existingMetadata.last_updated
    
    // Generate new metadata section
    newMetadata = formatMetadata(file, creationDate, lastUpdated, newContentHash)
    
    // Update file with new metadata if needed
    if metadataChanged(existingMetadata, newMetadata):
      updateFile(file, newMetadata, contentWithoutMetadata)
      updatedFiles.append(file)
      updatedHashMap[file] = {
        oldHash: oldContentHash,
        newHash: newContentHash,
        changed: contentChanged
      }
    else:
      unchangedFiles.append(file)
      
  return { updatedFiles, unchangedFiles, updatedHashMap }
```

#### Update Change Request References Algorithm
```
function updateChangeRequestReferences(updatedHashMap):
  updatedChangeRequests = []
  
  // Find all change request files
  changeRequestFiles = findAllChangeRequestFiles()
  
  foreach crFile in changeRequestFiles:
    content = readFile(crFile)
    changeRequest = parseChangeRequest(content)
    changesMade = false
    
    // Check each user story reference
    foreach ref in changeRequest.userStories:
      if ref.file in updatedHashMap && updatedHashMap[ref.file].changed:
        // Update the reference hash
        oldHash = ref.contentHash
        newHash = updatedHashMap[ref.file].newHash
        content = updateReferenceHash(content, ref.file, oldHash, newHash)
        changesMade = true
    
    if changesMade:
      writeFile(crFile, content)
      updatedChangeRequests.append(crFile)
  
  return updatedChangeRequests
```

#### Integrated Update Algorithm
```
function updateUserStoriesAndReferences(skipReferences = false):
  // Update all user story metadata first
  result = updateUserStoryMetadata(findUserStoryFiles())
  
  // Print summary of user story updates
  printSummary("User stories", result.updatedFiles.length, result.unchangedFiles.length)
  
  // If not explicitly skipped, update change request references
  if (!skipReferences && !isEmpty(result.updatedHashMap)):
    // Only pass files that had content changes (not just metadata changes)
    changedHashMap = filterContentChanges(result.updatedHashMap)
    
    // Update references in change requests
    updatedChangeRequests = updateChangeRequestReferences(changedHashMap)
    
    // Print summary of change request updates
    printSummary("Change requests", updatedChangeRequests.length, 
                findAllChangeRequestFiles().length - updatedChangeRequests.length)
  
  return {
    updatedUserStories: result.updatedFiles,
    unchangedUserStories: result.unchangedFiles,
    updatedChangeRequests: updatedChangeRequests || []
  }
```

### Refactoring Strategy

1. The existing `updateUserStoriesCmd` implementation in `cmd/update_user_stories.go` already handles most requirements for updating metadata in user story files.

2. We need to:
   - Ensure creation dates are properly preserved (this is already handled)
   - Add functionality to update change request references based on updated content hashes as part of the metadata update process
   - Modify the update process to track which files had their hashes changed and update references accordingly

3. The references update shouldn't be a separate command since it would lead to inconsistencies. Instead, we'll integrate it directly into the existing update flow:
   ```
   usm update user-stories metadata [--skip-references]
   ```
   Where updating references is on by default, with an option to skip if needed.

4. For cleaner organization, we'll extract common utility functions into a shared package that can be used throughout the application.

## How to verify – Detailed User Story Breakdown

### User Story 1: Update User Stories Metadata

#### Acceptance Criteria:
1. CLI has a command to update metadata in user story files.
   - **Testing**: Verify `usm update user-stories metadata` command exists and is documented.

2. Command scans for all markdown files in the `docs/user-stories` directory.
   - **Testing**: Create test files in this directory and run the command; verify they are processed.

3. Command adds/updates metadata section with file path, creation date, last edited date, and content hash.
   - **Testing**: Create a new file without metadata, run command, and verify metadata was added.
   - **Testing**: Modify an existing file with metadata, run command, and verify metadata was updated.

4. The metadata uses the specified format.
   - **Testing**: Verify metadata section formatting matches the specified standard.

5. Command preserves original creation date if present.
   - **Testing**: Update a file with existing metadata and verify creation date remains unchanged.

6. Command only updates last_updated date when content changes.
   - **Testing**: Run command on an unchanged file and verify last_updated remains the same.
   - **Testing**: Modify a file's content, run command, and verify last_updated changes.

7. Command skips specified directories.
   - **Testing**: Add test files in node_modules, .git, etc. and verify they are not processed.

8. Command prints a summary of processed files.
   - **Testing**: Verify console output includes counts of updated and unchanged files.

9. Command supports a --debug flag.
   - **Testing**: Run with --debug and verify additional logging.

10. Command is idempotent.
    - **Testing**: Run command twice and verify no unnecessary changes on second run.

### User Story 2: Detect and Track Content Changes

#### Acceptance Criteria:
1. Change request files are scanned for references to user story files.
   - **Testing**: Create a change request with user story references, verify it's found by scanner.

2. If a referenced user story has a new hash, update the reference.
   - **Testing**: Update a user story file referenced in a change request, run the command, and verify the change request is updated.

3. Updates are only done when hashes differ.
   - **Testing**: Run command on unchanged files and verify no change requests are modified.

4. CLI prints which change requests were updated.
   - **Testing**: Verify console output includes names of updated change request files.

5. Change requests are located in docs/change-request/**.
   - **Testing**: Place test change requests in this directory and verify they are processed.

### User Story 3: Preserve Original Creation Date

#### Acceptance Criteria:
1. If created_at is present in metadata, it remains untouched.
   - **Testing**: Run command on a file with created_at metadata and verify it remains unchanged.

2. If absent, a new ISO 8601 timestamp is added based on file's creation time.
   - **Testing**: Create a new file without metadata, run command, and verify created_at is set to the file's creation time.

## What is the Plan – Detailed Action Items

### 1. Enhance User Story Metadata Updates with Integrated Reference Management

The existing implementation in `cmd/update_user_stories.go` covers most requirements for updating metadata. We will enhance it to also handle reference updates within change requests automatically:

1. Modify `updateUserStoriesCmd` to enable reference tracking by default:
   - Add a `--skip-references` flag to allow skipping reference updates if needed
   - Otherwise, reference updates should always happen when content hashes change
   - Make reference updating the default behavior for consistency and data integrity

2. Enhance `updateFileMetadata` to collect and return information about changed files:
   - Track original content hash and new content hash for each modified file
   - Return a data structure (`ContentHashMap`) mapping file paths to hash information
   - Flag files whose actual content (not just metadata) has changed

3. Add utility functions to process change request references:
   - Create `findChangeRequestFiles` to locate all change request files
   - Create `updateChangeRequestReferences` to update references in change requests based on updated content hashes
   - Integrate these directly into the metadata update workflow

4. Add debugging support:
   - Implement the `--debug` flag for detailed processing information
   - Add verbose logging of both metadata and reference update operations

5. Ensure comprehensive handling of creation dates:
   - Verify that the existing code correctly preserves original creation dates
   - Add more robust tests for creation date preservation edge cases

### 2. Extract Common Metadata Utilities

Create a shared metadata package to improve code organization:

1. Create an `internal/metadata` package:
   - Move common metadata functions from `cmd/update_user_stories.go`
   - Extract regex patterns for metadata handling
   - Create shared utilities for content hash calculation and metadata extraction

2. Implement core metadata functionality in this package:
   - `ExtractMetadata` - Extract metadata from file content
   - `GenerateMetadata` - Generate new metadata based on file information
   - `UpdateFileMetadata` - Update metadata in a file
   - `CalculateContentHash` - Calculate content hash for a file

3. Create specific functions for change request reference handling:
   - `FindReferences` - Find user story references in change request files
   - `UpdateReference` - Update a specific reference in a change request
   - `WriteUpdatedContent` - Write updated content back to change request files

### 3. Integration and Tests

1. Develop comprehensive tests for the integrated workflow:
   - Create test scenarios with user stories and change requests that reference them
   - Test the complete update flow to verify both metadata and references are updated correctly
   - Verify edge cases like missing files, corrupted metadata, etc.

2. Add unit tests for the new functions:
   - Test the new utilities in the metadata package
   - Test reference finding and updating
   - Test the updated command with different flag combinations

3. Update existing tests to reflect the integrated workflow:
   - Modify tests that assume separate metadata and reference updates
   - Add tests for the `--skip-references` flag

### 4. Code Organization and API Design

1. Design a clean and intuitive command interface:
   - Keep the existing command name: `usm update user-stories metadata`
   - Add the `--skip-references` flag with clear documentation
   - Ensure consistent terminology throughout command help text

2. Implement proper error handling:
   - Provide helpful error messages for common failure scenarios
   - Distinguish between warnings (non-fatal issues) and errors
   - Add context to error messages to help users fix problems

3. Ensure consistent logging:
   - Use the same logging style for both metadata and reference updates
   - Log summary information about what was updated
   - Clearly identify which change requests were updated and why

### 5. Documentation and User Experience

1. Update the CLI help text:
   - Provide clear instructions on how to use the enhanced command
   - Explain the purpose of the `--skip-references` flag
   - Document the default behavior (updating references automatically)

2. Enhance progress reporting:
   - Provide clear status updates during processing
   - Show counts of changed user stories and change requests
   - Make it clear which changes were content changes vs. metadata-only changes

3. Add warning messages for potential issues:
   - Warn about referenced files that don't exist
   - Notify users about corrupted or inconsistent metadata
   - Alert when change requests contain references to non-existent user stories
