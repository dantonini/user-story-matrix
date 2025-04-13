// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestAddMetadataToReferencedStory tests the scenario where a user story
// initially has no metadata but is already referenced in a blueprint file.
// This verifies that the update process correctly adds metadata to the story
// and updates the reference in the blueprint file.
func TestAddMetadataToReferencedStory(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Create a user story file WITHOUT metadata
	// This simulates a newly created user story file that hasn't been processed yet
	userStoryContent := `# New User Story
This is a new user story that doesn't have metadata yet.
It should be updated with proper metadata during the update process.
`
	userStoryPath := "docs/user-stories/new-feature/new-story.md"
	fs.AddFile(userStoryPath, []byte(userStoryContent))

	// Create a blueprint file that already references the user story,
	// but obviously doesn't have a content hash since the story has no metadata yet
	blueprintContent := `---
name: test-feature-blueprint
created-at: 2025-04-01T09:00:00+02:00
user-stories:
  - title: New User Story
    file: docs/user-stories/new-feature/new-story.md
    content-hash: 
---
# Blueprint for New Feature

This blueprint references a new user story that doesn't have metadata yet.
`
	blueprintPath := "docs/changes-request/test-feature.blueprint.md"
	fs.AddFile(blueprintPath, []byte(blueprintContent))

	// Process the user story first to add metadata
	updated, hashMap, err := UpdateFileMetadata(userStoryPath, "", fs)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.NotEmpty(t, hashMap.NewHash)

	// Verify the user story now has metadata
	updatedStoryContent, err := fs.ReadFile(userStoryPath)
	assert.NoError(t, err)
	assert.Contains(t, string(updatedStoryContent), "file_path:")
	assert.Contains(t, string(updatedStoryContent), "created_at:")
	assert.Contains(t, string(updatedStoryContent), "last_updated:")
	assert.Contains(t, string(updatedStoryContent), "_content_hash:")

	// Find the hash from the updated user story file
	var extractedHash string
	for _, line := range strings.Split(string(updatedStoryContent), "\n") {
		if strings.HasPrefix(line, "_content_hash:") {
			parts := strings.SplitN(line, ":", 2)
			extractedHash = strings.TrimSpace(parts[1])
			break
		}
	}
	assert.NotEmpty(t, extractedHash)

	// Update the blueprint file to include the new hash
	contentChangeMap := ContentChangeMap{
		userStoryPath: hashMap,
	}
	
	// When updating references for a file that initially had no metadata (empty hash),
	// there will likely be mismatched references since the blueprint might have an empty
	// or incorrect hash. This is expected behavior, not an error.
	updated, refCount, _, err := UpdateChangeRequestReferences(blueprintPath, contentChangeMap, fs)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.GreaterOrEqual(t, refCount, 1) // At least one reference should be updated

	// Verify the blueprint now contains the correct hash
	updatedBlueprintContent, err := fs.ReadFile(blueprintPath)
	assert.NoError(t, err)
	
	// Print the updated content for debugging
	t.Logf("Updated blueprint content:\n%s", string(updatedBlueprintContent))
	
	// Verify the hash is correctly added with proper spacing
	assert.Contains(t, string(updatedBlueprintContent), "content-hash: "+extractedHash)
	
	// Verify there are no double spaces after the colon
	assert.NotContains(t, string(updatedBlueprintContent), "content-hash:  ")
} 