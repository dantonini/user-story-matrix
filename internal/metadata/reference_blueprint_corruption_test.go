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

// TestBlueprintMultipleStoriesCorruption tests the scenario where updating a blueprint file
// with multiple user stories might corrupt the metadata structure.
// This test reproduces the issue seen in the custom-workflow-phase2 blueprint file.
func TestBlueprintMultipleStoriesCorruption(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Create user story files with metadata
	storyContent1 := `---
file_path: docs/user-stories/feature/story1.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-13T08:45:16+02:00
_content_hash: hash1_old_value
---

# Story 1
This is the first user story.
`
	storyPath1 := "docs/user-stories/feature/story1.md"
	fs.AddFile(storyPath1, []byte(storyContent1))

	storyContent2 := `---
file_path: docs/user-stories/feature/story2.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-13T08:45:16+02:00
_content_hash: hash2_old_value
---

# Story 2
This is the second user story.
`
	storyPath2 := "docs/user-stories/feature/story2.md"
	fs.AddFile(storyPath2, []byte(storyContent2))

	storyContent3 := `---
file_path: docs/user-stories/feature/story3.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-13T08:45:16+02:00
_content_hash: hash3_old_value
---

# Story 3
This is the third user story.
`
	storyPath3 := "docs/user-stories/feature/story3.md"
	fs.AddFile(storyPath3, []byte(storyContent3))

	// Create a blueprint file that references the user stories
	// Note: Initially we have an empty content-hash for the second story
	// This reproduces the real-world scenario that led to corruption
	blueprintContent := `---
name: test-blueprint
created-at: 2025-04-01T09:00:00+02:00
user-stories:
  - title: Story 1
    file: docs/user-stories/feature/story1.md
    content-hash: hash1_old_value
  - title: Story 2
    file: docs/user-stories/feature/story2.md
    content-hash: 
  - title: Story 3
    file: docs/user-stories/feature/story3.md
    content-hash: hash3_old_value
---
# Test Blueprint

This blueprint references multiple user stories.
`
	blueprintPath := "docs/changes-request/test-blueprint.md"
	fs.AddFile(blueprintPath, []byte(blueprintContent))

	// Update the user stories with new content to change their hash values
	// This will trigger the update of references in the blueprint
	updatedStoryContent1 := `---
file_path: docs/user-stories/feature/story1.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-14T10:15:30+02:00
_content_hash: hash1_new_value
---

# Story 1 (Updated)
This is the first user story with updated content.
`
	fs.AddFile(storyPath1, []byte(updatedStoryContent1))

	updatedStoryContent2 := `---
file_path: docs/user-stories/feature/story2.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-14T10:15:30+02:00
_content_hash: hash2_new_value
---

# Story 2 (Updated)
This is the second user story with updated content.
`
	fs.AddFile(storyPath2, []byte(updatedStoryContent2))

	// Create the hash map that would be generated during the update process
	hashMap := ContentChangeMap{
		storyPath1: ContentHashMap{
			FilePath: storyPath1,
			OldHash:  "hash1_old_value",
			NewHash:  "hash1_new_value",
			Changed:  true,
		},
		storyPath2: ContentHashMap{
			FilePath: storyPath2,
			OldHash:  "hash2_old_value", // Note: In reality, this might be empty
			NewHash:  "hash2_new_value",
			Changed:  true,
		},
	}

	// Update the references in the blueprint file
	updated, refCount, mismatches, err := UpdateChangeRequestReferences(blueprintPath, hashMap, fs)
	
	// Verify the update succeeded
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.GreaterOrEqual(t, refCount, 1)
	// We might have mismatches since we're simulating content changes
	t.Logf("Mismatches: %v", mismatches)

	// Read the updated blueprint file
	updatedContent, err := fs.ReadFile(blueprintPath)
	assert.NoError(t, err)
	
	// Print the updated content for debugging
	t.Logf("Updated blueprint content:\n%s", string(updatedContent))
	
	// IMPORTANT: Verify that the blueprint structure wasn't corrupted
	// The key issue is to check if all three user stories are still properly formatted
	
	// Check that we still have 3 title entries
	titleCount := strings.Count(string(updatedContent), "title: Story")
	assert.Equal(t, 3, titleCount, "Should have 3 title entries")
	
	// Check that each story entry is properly formatted and contained
	assert.Contains(t, string(updatedContent), "title: Story 1")
	assert.Contains(t, string(updatedContent), "file: docs/user-stories/feature/story1.md")
	assert.Contains(t, string(updatedContent), "content-hash: hash1_new_value")
	
	assert.Contains(t, string(updatedContent), "title: Story 2")
	assert.Contains(t, string(updatedContent), "file: docs/user-stories/feature/story2.md")
	assert.Contains(t, string(updatedContent), "content-hash: hash2_new_value")
	
	assert.Contains(t, string(updatedContent), "title: Story 3")
	assert.Contains(t, string(updatedContent), "file: docs/user-stories/feature/story3.md")
	assert.Contains(t, string(updatedContent), "content-hash: hash3_old_value")
	
	// The key test: each item should have exactly one title, file, and content-hash entry
	lines := strings.Split(string(updatedContent), "\n")
	storyItems := 0
	inItem := false
	itemProperties := make(map[string]int)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "- title:") {
			// New item started
			if inItem {
				// Check previous item had all properties
				assert.Equal(t, 1, itemProperties["title"], "Each item should have exactly one title")
				assert.Equal(t, 1, itemProperties["file"], "Each item should have exactly one file")
				assert.Equal(t, 1, itemProperties["content-hash"], "Each item should have exactly one content-hash")
			}
			
			inItem = true
			storyItems++
			itemProperties = make(map[string]int)
			itemProperties["title"]++
		} else if inItem && strings.HasPrefix(line, "file:") {
			itemProperties["file"]++
		} else if inItem && strings.HasPrefix(line, "content-hash:") {
			itemProperties["content-hash"]++
		} else if inItem && (line == "---" || strings.HasPrefix(line, "- title:")) {
			// Item ended
			inItem = false
			// Check properties for the last item
			assert.Equal(t, 1, itemProperties["title"], "Each item should have exactly one title")
			assert.Equal(t, 1, itemProperties["file"], "Each item should have exactly one file")
			assert.Equal(t, 1, itemProperties["content-hash"], "Each item should have exactly one content-hash")
			
			if strings.HasPrefix(line, "- title:") {
				// Start a new item
				inItem = true
				storyItems++
				itemProperties = make(map[string]int)
				itemProperties["title"]++
			}
		}
	}
	
	// Verify we found the expected number of story items
	assert.Equal(t, 3, storyItems, "Should have 3 story items in the blueprint")
} 