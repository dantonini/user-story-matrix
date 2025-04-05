// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestIntegration_UpdateChangeRequestReferences tests the UpdateChangeRequestReferences function
// using a real filesystem instead of a mock to avoid the issues that caused the original test to be skipped.
func TestIntegration_UpdateChangeRequestReferences(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create user story files
		story1Path := filepath.Join(usDir, "story1.md")
		story1Content := "# Story 1\n\nThis is a test story."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		story2Path := filepath.Join(usDir, "story2.md")
		story2Content := "# Story 2\n\nThis is another test story."
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// Create change request file with references to the user stories
		// Note: The format must exactly match what the regex expects
		crPath := filepath.Join(crDir, "cr1.blueprint.md")
		crContent := `# Change Request 1

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

- title: Story 2
  file: docs/user-stories/story2.md
  content-hash: old-hash-2

## Implementation

This change request requires implementing both Story 1 and Story 2.
`
		require.NoError(t, fs.WriteFile(crPath, []byte(crContent), 0644))
		
		// Create a hash map with changes
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
			NewHash:  "old-hash-2", // No change
			Changed:  false,
		}
		
		// Run UpdateChangeRequestReferences
		updated, refCount, err := UpdateChangeRequestReferences(crPath, hashMap, fs)
		require.NoError(t, err)
		
		// Verify results
		assert.True(t, updated, "Change request should be marked as updated")
		assert.Equal(t, 1, refCount, "One reference should be updated")
		
		// Read the updated file and verify contents
		content, err := fs.ReadFile(crPath)
		require.NoError(t, err)
		contentStr := string(content)
		
		// Verify hash was updated for the first story
		assert.Contains(t, contentStr, "content-hash: new-hash-1", "Content hash for story1 should be updated")
		assert.NotContains(t, contentStr, "content-hash: old-hash-1", "Old hash for story1 should be replaced")
		
		// Verify hash was not updated for the second story
		assert.Contains(t, contentStr, "content-hash: old-hash-2", "Content hash for story2 should remain unchanged")
	})
}

// TestIntegration_UpdateChangeRequestReferences_NoChanges tests the case where
// no references need to be updated.
func TestIntegration_UpdateChangeRequestReferences_NoChanges(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create change request file with references
		crPath := filepath.Join(crDir, "cr1.blueprint.md")
		crContent := `# Change Request 1

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

## Implementation

This change request implements Story 1.
`
		require.NoError(t, fs.WriteFile(crPath, []byte(crContent), 0644))
		
		// Create a hash map with no changes
		hashMap := make(ContentChangeMap)
		hashMap["docs/user-stories/story1.md"] = ContentHashMap{
			FilePath: "docs/user-stories/story1.md",
			OldHash:  "old-hash-1",
			NewHash:  "old-hash-1",
			Changed:  false,
		}
		
		// Run UpdateChangeRequestReferences
		updated, refCount, err := UpdateChangeRequestReferences(crPath, hashMap, fs)
		require.NoError(t, err)
		
		// Verify results
		assert.False(t, updated, "Change request should not be marked as updated")
		assert.Equal(t, 0, refCount, "No references should be updated")
		
		// Read the file and verify content hasn't changed
		content, err := fs.ReadFile(crPath)
		require.NoError(t, err)
		assert.Equal(t, crContent, string(content), "File content should remain unchanged")
	})
}

// TestIntegration_UpdateAllChangeRequestReferences tests the UpdateAllChangeRequestReferences
// function using a real filesystem.
func TestIntegration_UpdateAllChangeRequestReferences(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create user story files
		story1Path := filepath.Join(usDir, "story1.md")
		story1Content := "# Story 1\n\nThis is a test story."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		story2Path := filepath.Join(usDir, "story2.md")
		story2Content := "# Story 2\n\nThis is another test story."
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// Create first change request file
		cr1Path := filepath.Join(crDir, "cr1.blueprint.md")
		cr1Content := `# Change Request 1

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

- title: Story 2
  file: docs/user-stories/story2.md
  content-hash: old-hash-2

## Implementation

This change request requires implementing both Story 1 and Story 2.
`
		require.NoError(t, fs.WriteFile(cr1Path, []byte(cr1Content), 0644))
		
		// Create second change request file
		cr2Path := filepath.Join(crDir, "cr2.blueprint.md")
		cr2Content := `# Change Request 2

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

## Implementation

This change request is focused on Story 1 only.
`
		require.NoError(t, fs.WriteFile(cr2Path, []byte(cr2Content), 0644))
		
		// Create a non-blueprint file which should be skipped
		nonBlueprintPath := filepath.Join(crDir, "not-a-blueprint.md")
		nonBlueprintContent := `# Not a Blueprint

This is not a blueprint file and should be ignored.

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1
`
		require.NoError(t, fs.WriteFile(nonBlueprintPath, []byte(nonBlueprintContent), 0644))
		
		// Create a hash map with changes
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
			NewHash:  "new-hash-2",
			Changed:  true,
		}
		
		// Run UpdateAllChangeRequestReferences
		updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
		require.NoError(t, err)
		
		// Get relative paths for verification
		relCR1, _ := filepath.Rel(tempDir, cr1Path)
		relCR2, _ := filepath.Rel(tempDir, cr2Path)
		
		// Verify results
		assert.Equal(t, 2, len(updatedFiles), "Two files should be updated")
		assert.Contains(t, updatedFiles, relCR1, "CR1 should be in the updated list")
		assert.Contains(t, updatedFiles, relCR2, "CR2 should be in the updated list")
		assert.Equal(t, 0, len(unchangedFiles), "No files should be unchanged")
		assert.Equal(t, 3, refCount, "Three references should be updated")
		
		// Verify content of first change request
		content1New, err := fs.ReadFile(cr1Path)
		require.NoError(t, err)
		contentStr1 := string(content1New)
		
		assert.Contains(t, contentStr1, "content-hash: new-hash-1", "Hash for story1 should be updated in CR1")
		assert.NotContains(t, contentStr1, "content-hash: old-hash-1", "Old hash for story1 should be replaced in CR1")
		assert.Contains(t, contentStr1, "content-hash: new-hash-2", "Hash for story2 should be updated in CR1")
		assert.NotContains(t, contentStr1, "content-hash: old-hash-2", "Old hash for story2 should be replaced in CR1")
		
		// Verify content of second change request
		content2New, err := fs.ReadFile(cr2Path)
		require.NoError(t, err)
		contentStr2 := string(content2New)
		
		assert.Contains(t, contentStr2, "content-hash: new-hash-1", "Hash for story1 should be updated in CR2")
		assert.NotContains(t, contentStr2, "content-hash: old-hash-1", "Old hash for story1 should be replaced in CR2")
		
		// Verify non-blueprint file was not modified
		nonBlueprintContent2, err := fs.ReadFile(nonBlueprintPath)
		require.NoError(t, err)
		assert.Equal(t, nonBlueprintContent, string(nonBlueprintContent2), "Non-blueprint file should not be modified")
	})
}

// TestIntegration_UpdateAllChangeRequestReferences_NoChanges tests the case when 
// no changes are needed in any change request.
func TestIntegration_UpdateAllChangeRequestReferences_NoChanges(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		
		// Create change request file
		crPath := filepath.Join(crDir, "cr1.blueprint.md")
		crContent := `# Change Request 1

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

## Implementation

This change request implements Story 1.
`
		require.NoError(t, fs.WriteFile(crPath, []byte(crContent), 0644))
		
		// Create a hash map with no changes
		hashMap := make(ContentChangeMap)
		hashMap["docs/user-stories/story1.md"] = ContentHashMap{
			FilePath: "docs/user-stories/story1.md",
			OldHash:  "old-hash-1",
			NewHash:  "old-hash-1",
			Changed:  false,
		}
		
		// Run UpdateAllChangeRequestReferences
		updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
		require.NoError(t, err)
		
		// Verify results
		assert.Equal(t, 0, len(updatedFiles), "No files should be updated")
		assert.Equal(t, 0, len(unchangedFiles), "No files should be in unchanged list when there are no content changes")
		assert.Equal(t, 0, refCount, "No references should be updated")
		
		// Verify file content hasn't changed
		content, err := fs.ReadFile(crPath)
		require.NoError(t, err)
		assert.Equal(t, crContent, string(content), "File content should remain unchanged")
	})
}

// TestIntegration_UpdateAllChangeRequestReferences_EmptyHashMap tests the case with
// an empty hash map.
func TestIntegration_UpdateAllChangeRequestReferences_EmptyHashMap(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		
		// Create change request file
		crPath := filepath.Join(crDir, "cr1.blueprint.md")
		crContent := `# Change Request 1

## User Stories

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

## Implementation

This change request implements Story 1.
`
		require.NoError(t, fs.WriteFile(crPath, []byte(crContent), 0644))
		
		// Create an empty hash map
		hashMap := make(ContentChangeMap)
		
		// Run UpdateAllChangeRequestReferences
		updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
		require.NoError(t, err)
		
		// Verify results
		assert.Nil(t, updatedFiles, "Updated files should be nil with empty hash map")
		assert.Nil(t, unchangedFiles, "Unchanged files should be nil with empty hash map")
		assert.Equal(t, 0, refCount, "No references should be updated")
		
		// Verify file content hasn't changed
		content, err := fs.ReadFile(crPath)
		require.NoError(t, err)
		assert.Equal(t, crContent, string(content), "File content should remain unchanged")
	})
}

// TestIntegration_UpdateAllChangeRequestReferences_SharedStory tests the scenario where
// a single user story is referenced by multiple change requests to ensure all references
// are properly updated.
func TestIntegration_UpdateAllChangeRequestReferences_SharedStory(t *testing.T) {
	withTempDir(t, func(tempDir string, fs io.FileSystem) {
		// Create directory structure
		crDir := filepath.Join(tempDir, "docs", "changes-request")
		usDir := filepath.Join(tempDir, "docs", "user-stories")
		require.NoError(t, fs.MkdirAll(crDir, 0755))
		require.NoError(t, fs.MkdirAll(usDir, 0755))
		
		// Create shared user story file
		sharedStoryPath := filepath.Join(usDir, "shared-story.md")
		sharedStoryContent := "# Shared Story\n\nThis is a story referenced by multiple change requests."
		require.NoError(t, fs.WriteFile(sharedStoryPath, []byte(sharedStoryContent), 0644))
		
		// Create unique user story files
		story1Path := filepath.Join(usDir, "story1.md")
		story1Content := "# Story 1\n\nThis is a unique story."
		require.NoError(t, fs.WriteFile(story1Path, []byte(story1Content), 0644))
		
		story2Path := filepath.Join(usDir, "story2.md")
		story2Content := "# Story 2\n\nThis is another unique story."
		require.NoError(t, fs.WriteFile(story2Path, []byte(story2Content), 0644))
		
		// Create first change request file - references shared story and story1
		cr1Path := filepath.Join(crDir, "cr1.blueprint.md")
		cr1Content := `# Change Request 1

## User Stories

- title: Shared Story
  file: docs/user-stories/shared-story.md
  content-hash: old-shared-hash

- title: Story 1
  file: docs/user-stories/story1.md
  content-hash: old-hash-1

## Implementation

This change request requires implementing both stories.
`
		require.NoError(t, fs.WriteFile(cr1Path, []byte(cr1Content), 0644))
		
		// Create second change request file - references shared story and story2
		cr2Path := filepath.Join(crDir, "cr2.blueprint.md")
		cr2Content := `# Change Request 2

## User Stories

- title: Shared Story
  file: docs/user-stories/shared-story.md
  content-hash: old-shared-hash

- title: Story 2
  file: docs/user-stories/story2.md
  content-hash: old-hash-2

## Implementation

This change request also uses the shared story.
`
		require.NoError(t, fs.WriteFile(cr2Path, []byte(cr2Content), 0644))
		
		// Create a third change request file - only references the shared story
		cr3Path := filepath.Join(crDir, "cr3.blueprint.md")
		cr3Content := `# Change Request 3

## User Stories

- title: Shared Story
  file: docs/user-stories/shared-story.md
  content-hash: old-shared-hash

## Implementation

This change request only uses the shared story.
`
		require.NoError(t, fs.WriteFile(cr3Path, []byte(cr3Content), 0644))
		
		// Create a hash map with changes - only update the shared story
		hashMap := make(ContentChangeMap)
		hashMap["docs/user-stories/shared-story.md"] = ContentHashMap{
			FilePath: "docs/user-stories/shared-story.md",
			OldHash:  "old-shared-hash",
			NewHash:  "new-shared-hash",
			Changed:  true,
		}
		
		// Run UpdateAllChangeRequestReferences
		updatedFiles, unchangedFiles, refCount, err := UpdateAllChangeRequestReferences(tempDir, hashMap, fs)
		require.NoError(t, err)
		
		// Get relative paths for verification
		relCR1, _ := filepath.Rel(tempDir, cr1Path)
		relCR2, _ := filepath.Rel(tempDir, cr2Path)
		relCR3, _ := filepath.Rel(tempDir, cr3Path)
		
		// Verify results
		assert.Equal(t, 3, len(updatedFiles), "Three files should be updated")
		assert.Contains(t, updatedFiles, relCR1, "CR1 should be in the updated list")
		assert.Contains(t, updatedFiles, relCR2, "CR2 should be in the updated list")
		assert.Contains(t, updatedFiles, relCR3, "CR3 should be in the updated list")
		assert.Equal(t, 0, len(unchangedFiles), "No files should be unchanged")
		assert.Equal(t, 3, refCount, "Three references to the shared story should be updated")
		
		// Verify content of first change request
		content1, err := fs.ReadFile(cr1Path)
		require.NoError(t, err)
		contentStr1 := string(content1)
		
		assert.Contains(t, contentStr1, "content-hash: new-shared-hash", "Shared story hash should be updated in CR1")
		assert.NotContains(t, contentStr1, "content-hash: old-shared-hash", "Old shared story hash should be replaced in CR1")
		assert.Contains(t, contentStr1, "content-hash: old-hash-1", "Story1 hash should remain unchanged in CR1")
		
		// Verify content of second change request
		content2, err := fs.ReadFile(cr2Path)
		require.NoError(t, err)
		contentStr2 := string(content2)
		
		assert.Contains(t, contentStr2, "content-hash: new-shared-hash", "Shared story hash should be updated in CR2")
		assert.NotContains(t, contentStr2, "content-hash: old-shared-hash", "Old shared story hash should be replaced in CR2")
		assert.Contains(t, contentStr2, "content-hash: old-hash-2", "Story2 hash should remain unchanged in CR2")
		
		// Verify content of third change request
		content3, err := fs.ReadFile(cr3Path)
		require.NoError(t, err)
		contentStr3 := string(content3)
		
		assert.Contains(t, contentStr3, "content-hash: new-shared-hash", "Shared story hash should be updated in CR3")
		assert.NotContains(t, contentStr3, "content-hash: old-shared-hash", "Old shared story hash should be replaced in CR3")
	})
} 