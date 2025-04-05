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
	// This test has been implemented as an integration test in reference_integration_test.go
	// using a real filesystem instead of a mock to avoid issues with complex file operations.
	// See TestIntegration_UpdateChangeRequestReferences
	t.Skip("Test implemented as integration test in reference_integration_test.go")
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
	updated, refCount, err := UpdateChangeRequestReferences("docs/changes-request/cr2.blueprint.md", hashMap, fs)
	assert.NoError(t, err)
	assert.False(t, updated)
	assert.Equal(t, 0, refCount)

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
	// This test has been implemented as an integration test in reference_integration_test.go
	// using a real filesystem instead of a mock to avoid issues with complex file operations.
	// See TestIntegration_UpdateAllChangeRequestReferences
	t.Skip("Test implemented as integration test in reference_integration_test.go")
}

func TestUpdateAllChangeRequestReferences_NoChanges(t *testing.T) {
	fs := setupReferenceTestFiles()

	// Create a hash map with no changes
	hashMap := make(ContentChangeMap)
	hashMap["docs/user-stories/story1.md"] = ContentHashMap{
		FilePath: "docs/user-stories/story1.md",
		OldHash:  "old-hash-1",
		NewHash:  "old-hash-1",
		Changed:  false,
	}
	hashMap["docs/user-stories/story2.md"] = ContentHashMap{
		FilePath: "docs/user-stories/story2.md",
		OldHash:  "old-hash-2",
		NewHash:  "old-hash-2",
		Changed:  false,
	}

	// Test updating all change request references
	updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences("", hashMap, fs)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(updatedFiles))
	// The unchanged files list should actually be empty because when all content is unchanged, 
	// we don't actually check any change request files - the function exits early
	assert.Equal(t, 0, len(unchangedFiles))
	assert.Equal(t, 0, refCount)
}

func TestUpdateAllChangeRequestReferences_EmptyHashMap(t *testing.T) {
	fs := setupReferenceTestFiles()

	// Create an empty hash map
	hashMap := make(ContentChangeMap)

	// Test updating all change request references
	updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences("", hashMap, fs)
	assert.NoError(t, err)
	assert.Nil(t, updatedFiles)
	assert.Nil(t, unchangedFiles)
	assert.Equal(t, 0, refCount)
} 