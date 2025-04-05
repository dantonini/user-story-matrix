// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// setupReferenceTestFiles creates a mock filesystem with test files for reference testing
func setupReferenceTestFiles() io.FileSystem {
	fs := io.NewMockFileSystem()
	
	// Create directories
	fs.AddDirectory("docs")
	fs.AddDirectory("docs/user-stories")
	fs.AddDirectory("docs/changes-request")
	
	// Create user story files
	fs.AddFile("docs/user-stories/story1.md", []byte("# Story 1\n\nThis is story 1."))
	fs.AddFile("docs/user-stories/story2.md", []byte("# Story 2\n\nThis is story 2."))
	
	// Create change request files with references to user stories
	cr1Content := `---
name: Change Request 1
created-at: 2023-01-05T12:00:00Z
user-stories:
  - title: Story 1
    file: docs/user-stories/story1.md
    content-hash: old-hash-1
  - title: Story 2
    file: docs/user-stories/story2.md
    content-hash: old-hash-2
---

# Blueprint
This is change request 1.
`
	cr2Content := `---
name: Change Request 2
created-at: 2023-01-06T12:00:00Z
user-stories:
  - title: Story 1
    file: docs/user-stories/story1.md
    content-hash: old-hash-1
---

# Blueprint
This is change request 2.
`
	nonBlueprintContent := `---
name: Not a Blueprint
created-at: 2023-01-07T12:00:00Z
---

# Not a Blueprint
This is not a blueprint file.
`
	
	// Add change request files to the filesystem
	fs.AddFile("docs/changes-request/cr1.blueprint.md", []byte(cr1Content))
	fs.AddFile("docs/changes-request/cr2.blueprint.md", []byte(cr2Content))
	fs.AddFile("docs/changes-request/not-a-blueprint.md", []byte(nonBlueprintContent))
	
	return fs
}

func TestFindChangeRequestFiles(t *testing.T) {
	fs := setupReferenceTestFiles()

	// Test finding change request files
	files, err := FindChangeRequestFiles("", fs)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))
	assert.Contains(t, files, "docs/changes-request/cr1.blueprint.md")
	assert.Contains(t, files, "docs/changes-request/cr2.blueprint.md")
	assert.Contains(t, files, "docs/changes-request/not-a-blueprint.md")
}

func TestUpdateChangeRequestReferences(t *testing.T) {
	// Setup
	mockFS := io.NewMockFileSystem()
	fileContent := `
---
title: Test Change Request
description: This is a test change request
---

User stories:
- title: Test User Story
  file: docs/user-stories/test.md
  content-hash: oldhash123
`
	
	mockFS.AddFile("test_change_request.md", []byte(fileContent))
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/test.md": ContentHashMap{
			FilePath: "docs/user-stories/test.md",
			OldHash: "oldhash123",
			NewHash: "newhash456",
			Changed: true,
		},
	}
	
	// Call the function
	updated, count, mismatches, err := UpdateChangeRequestReferences("test_change_request.md", hashMap, mockFS)
	
	// Assertions
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, 1, count)
	assert.Equal(t, 0, len(mismatches))
	
	// Verify the content was updated
	updatedContent, err := mockFS.ReadFile("test_change_request.md")
	assert.NoError(t, err)
	assert.Contains(t, string(updatedContent), "content-hash: newhash456")
}

func TestUpdateChangeRequestReferences_NoChanges(t *testing.T) {
	fs := setupReferenceTestFiles()

	// Create a hash map with no changes
	hashMap := make(ContentChangeMap)
	hashMap["docs/user-stories/story1.md"] = ContentHashMap{
		FilePath: "docs/user-stories/story1.md",
		OldHash:  "old-hash-1",
		NewHash:  "old-hash-1",
		Changed:  false,
	}

	// Test updating references in a change request
	updated, refCount, mismatches, err := UpdateChangeRequestReferences("docs/changes-request/cr2.blueprint.md", hashMap, fs)
	assert.NoError(t, err)
	assert.False(t, updated)
	assert.Equal(t, 0, refCount)
	assert.Equal(t, 0, len(mismatches))

	// Verify the change request was not updated
	content, err := fs.ReadFile("docs/changes-request/cr2.blueprint.md")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "content-hash: old-hash-1")
}

func TestFilterChangedContent(t *testing.T) {
	// Create a hash map with both changed and unchanged content
	hashMap := make(ContentChangeMap)
	hashMap["docs/user-stories/story1.md"] = ContentHashMap{
		FilePath: "docs/user-stories/story1.md",
		OldHash:  "old-hash-1",
		NewHash:  "new-hash-1",
		Changed:  true,
	}
	hashMap["docs/user-stories/story2.md"] = ContentHashMap{
		FilePath: "docs/user-stories/story2.md",
		OldHash:  "old-hash-2",
		NewHash:  "old-hash-2",
		Changed:  false,
	}

	// Filter the hash map
	filteredMap := FilterChangedContent(hashMap)

	// Verify the filtered map contains only changed content
	assert.Equal(t, 1, len(filteredMap))
	assert.Contains(t, filteredMap, "docs/user-stories/story1.md")
	assert.NotContains(t, filteredMap, "docs/user-stories/story2.md")
}

func TestUpdateAllChangeRequestReferences(t *testing.T) {
	// Setup
	mockFS := io.NewMockFileSystem()
	
	// Create a mock directory structure
	mockFS.AddFile("docs/changes-request/test1.md", []byte(`
---
title: Test Change Request 1
description: This is a test change request
---

User stories:
- title: Test User Story
  file: docs/user-stories/test.md
  content-hash: oldhash123
`))
	
	mockFS.AddFile("docs/changes-request/test2.md", []byte(`
---
title: Test Change Request 2
description: This is another test change request
---

User stories:
- title: Another Test User Story
  file: docs/user-stories/another_test.md
  content-hash: anotherhash789
`))
	
	// Set up hash map
	hashMap := ContentChangeMap{
		"docs/user-stories/test.md": ContentHashMap{
			FilePath: "docs/user-stories/test.md",
			OldHash: "oldhash123",
			NewHash: "newhash456",
			Changed: true,
		},
	}
	
	// Call the function
	updatedFiles, unchangedFiles, referencesUpdated, mismatches, err := UpdateAllChangeRequestReferences("", hashMap, mockFS)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, len(updatedFiles))
	assert.Equal(t, 1, len(unchangedFiles))
	assert.Equal(t, 1, referencesUpdated)
	assert.Equal(t, 0, len(mismatches))
}

func TestUpdateAllChangeRequestReferences_NoChanges(t *testing.T) {
	// Setup
	mockFS := io.NewMockFileSystem()
	
	// Create a mock directory structure
	mockFS.AddFile("docs/changes-request/test1.md", []byte(`
---
title: Test Change Request 1
description: This is a test change request
---

User stories:
- title: Test User Story
  file: docs/user-stories/test.md
  content-hash: oldhash123
`))
	
	// Empty hash map (no changes)
	hashMap := ContentChangeMap{}
	
	// Call the function
	updatedFiles, unchangedFiles, referencesUpdated, mismatches, err := UpdateAllChangeRequestReferences("", hashMap, mockFS)
	
	// Assertions
	assert.NoError(t, err)
	assert.Nil(t, updatedFiles)
	assert.Nil(t, unchangedFiles)
	assert.Equal(t, 0, referencesUpdated)
	assert.Nil(t, mismatches)
}

func TestValidateChangedReferences(t *testing.T) {
	// Setup test data
	references := []Reference{
		{
			Title:       "Story 1",
			FilePath:    "docs/user-stories/story1.md",
			ContentHash: "old-hash-1",
		},
		{
			Title:       "Story 2",
			FilePath:    "docs/user-stories/story2.md", 
			ContentHash: "old-hash-2",
		},
		{
			Title:       "Story 3",
			FilePath:    "docs/user-stories/story3.md",
			ContentHash: "different-hash-3", // Mismatched hash
		},
	}
	
	hashMap := ContentChangeMap{
		"docs/user-stories/story1.md": {
			FilePath: "docs/user-stories/story1.md",
			OldHash:  "old-hash-1",
			NewHash:  "new-hash-1",
			Changed:  true,
		},
		"docs/user-stories/story2.md": {
			FilePath: "docs/user-stories/story2.md",
			OldHash:  "old-hash-2",
			NewHash:  "new-hash-2",
			Changed:  false, // Unchanged content
		},
		"docs/user-stories/story3.md": {
			FilePath: "docs/user-stories/story3.md",
			OldHash:  "old-hash-3", // Different from ContentHash in reference
			NewHash:  "new-hash-3",
			Changed:  true,
		},
	}
	
	// Call function
	changedRefs, mismatchedRefs := ValidateChangedReferences(references, hashMap)
	
	// Assertions
	assert.Equal(t, 2, len(changedRefs))
	assert.Equal(t, 1, len(mismatchedRefs))
	
	// Check the first reference (matches old hash)
	assert.Equal(t, "docs/user-stories/story1.md", changedRefs[0].FilePath)
	
	// Check the mismatched reference
	assert.Equal(t, "docs/user-stories/story3.md", mismatchedRefs[0].FilePath)
	assert.Equal(t, "different-hash-3", mismatchedRefs[0].ReferenceHash)
	assert.Equal(t, "old-hash-3", mismatchedRefs[0].OldHash)
}

func TestUpdateChangeRequestReferences_FilePathCorruption(t *testing.T) {
	// Setup
	mockFS := io.NewMockFileSystem()
	fileContent := `---
name: Test Change Request
created-at: 2025-03-24T20:03:55+01:00
user-stories:
  - title: Live Search Filtering
    file: docs/user-stories/create-change-request-tui/02-live-search-filtering.md
    content-hash: oldhash123
  - title: Keyboard Navigation and Selection
    file: docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md
    content-hash: oldhash456
---
`
	
	mockFS.AddFile("test_change_request.md", []byte(fileContent))
	
	// Set up hash map with very long hash values
	hashMap := ContentChangeMap{
		"docs/user-stories/create-change-request-tui/02-live-search-filtering.md": ContentHashMap{
			FilePath: "docs/user-stories/create-change-request-tui/02-live-search-filtering.md",
			OldHash: "oldhash123",
			NewHash: "448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2",
			Changed: true,
		},
		"docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md": ContentHashMap{
			FilePath: "docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md",
			OldHash: "oldhash456",
			NewHash: "feeb2080784b92262b59d45aed619d0b7980b7d3905532d52b779a88de31203d",
			Changed: true,
		},
	}
	
	// Call the function
	updated, count, mismatches, err := UpdateChangeRequestReferences("test_change_request.md", hashMap, mockFS)
	
	// Assertions
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, 2, count)
	assert.Equal(t, 0, len(mismatches))
	
	// Verify the content was updated correctly
	updatedContent, err := mockFS.ReadFile("test_change_request.md")
	assert.NoError(t, err)
	
	// The file path should remain intact, only content-hash should be updated
	assert.Contains(t, string(updatedContent), "file: docs/user-stories/create-change-request-tui/02-live-search-filtering.md")
	assert.Contains(t, string(updatedContent), "content-hash: 448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2")
	assert.Contains(t, string(updatedContent), "file: docs/user-stories/create-change-request-tui/06-keyboard-navigation-and-selection.md")
	assert.Contains(t, string(updatedContent), "content-hash: feeb2080784b92262b59d45aed619d0b7980b7d3905532d52b779a88de31203d")
	
	// The file path should NOT contain the hash
	assert.NotContains(t, string(updatedContent), "file: docs/user-stories/create-change-request-tui/02-live-search448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2")
	
	// Check for file path corruption patterns we've observed
	corruptionPatterns := []string{
		"live-search448981a2d2918b6bb7bfbc6015ef86e9dff5e1c0a944aa53d652ae3371ce40f2",
		"title: Visual Cues fore2a8e07ed7b41ff9b342035f3b3928092f0a93c3ac3b683697a651aab5136ec8",
		"7a5cdd266ccf8b42780da330b173e9a10741c5224c7560a9b367ab91c0e85889a0ced978861e1514aa",
	}
	
	for _, pattern := range corruptionPatterns {
		assert.NotContains(t, string(updatedContent), pattern, "Found corruption pattern: %s", pattern)
	}
} 