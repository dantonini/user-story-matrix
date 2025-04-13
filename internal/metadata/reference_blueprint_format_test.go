// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

// TestBlueprintContentHashFormatting tests the formatting of content-hash fields
// in blueprint files, specifically checking for issues with extra spaces.
func TestBlueprintContentHashFormatting(t *testing.T) {
	// Create a mock file system
	fs := io.NewMockFileSystem()

	// Extract content hash from the user story content to use in our test
	hashValue := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	// Create a user story file with known content and hash
	userStoryContent := `---
file_path: docs/user-stories/custom-workflow/test-story.md
created_at: 2025-04-08T09:49:58+02:00
last_updated: 2025-04-13T08:45:16+02:00
_content_hash: ` + hashValue + `
---

# Test User Story
This is a test user story for testing blueprint content-hash formatting.
`
	fs.AddFile("docs/user-stories/custom-workflow/test-story.md", []byte(userStoryContent))

	// Create a blueprint file that references the user story with empty content-hash
	// This simulates the issue where the empty content-hash field gets replaced with incorrectly formatted content
	blueprintContent := `---
name: test-blueprint
created-at: 2025-04-08T20:28:02+02:00
user-stories:
  - title: Test User Story
    file: docs/user-stories/custom-workflow/test-story.md
    content-hash: 
---
# Blueprint

This is a test blueprint for testing content-hash formatting issues.
`
	fs.AddFile("docs/changes-request/test-blueprint.md", []byte(blueprintContent))

	// Verify the original content has an empty content-hash value
	originalContent, err := fs.ReadFile("docs/changes-request/test-blueprint.md")
	assert.NoError(t, err)
	assert.Contains(t, string(originalContent), "content-hash: ")

	// Create the hash map that would be generated during the update process
	hashMap := ContentChangeMap{
		"docs/user-stories/custom-workflow/test-story.md": ContentHashMap{
			FilePath: "docs/user-stories/custom-workflow/test-story.md",
			OldHash:  "",  // Empty old hash indicates a new hash being added
			NewHash:  hashValue,
			Changed:  true,
		},
	}

	// Update the references in the blueprint file
	updated, refCount, _, err := UpdateChangeRequestReferences("docs/changes-request/test-blueprint.md", hashMap, fs)
	
	// Verify the update succeeded without errors
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.GreaterOrEqual(t, refCount, 1) // At least one reference should be updated

	// Read the updated blueprint file
	updatedContent, err := fs.ReadFile("docs/changes-request/test-blueprint.md")
	assert.NoError(t, err)

	// Print the updated content for debugging
	t.Logf("Updated content:\n%s", string(updatedContent))

	// Verify the hash value is correctly inserted without extra spaces
	// The key test here is to verify there is exactly one space between the colon and the hash
	contentHashLine := regexp.MustCompile(`(?m)^\s+content-hash:.*$`).FindString(string(updatedContent))
	assert.NotEmpty(t, contentHashLine, "Could not find content-hash line")
	
	// Extract the part after the colon
	parts := strings.SplitN(contentHashLine, ":", 2)
	assert.Len(t, parts, 2, "Expected to split line at colon")
	
	afterColon := parts[1]
	// The first character should be a space
	assert.True(t, strings.HasPrefix(afterColon, " "), "Expected a space after the colon")
	
	// The second character should NOT be a space (should not have double spaces)
	assert.False(t, strings.HasPrefix(afterColon, "  "), "There should not be multiple spaces after the colon")
	
	// The value after the space should be the hash
	trimmedValue := strings.TrimSpace(afterColon)
	assert.Equal(t, hashValue, trimmedValue, 
		"The hash should be correctly inserted after exactly one space")
} 