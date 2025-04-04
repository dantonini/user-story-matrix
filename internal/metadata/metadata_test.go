// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
)

func TestExtractMetadata(t *testing.T) {
	content := `---
file_path: docs/user-stories/example/sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: abcdef1234567890
---

# Sample User Story

This is a sample user story.
`

	metadata, err := ExtractMetadata(content)
	assert.NoError(t, err)

	// Verify extracted fields
	assert.Equal(t, "docs/user-stories/example/sample.md", metadata.FilePath)
	assert.Equal(t, "abcdef1234567890", metadata.ContentHash)

	// Verify parsed timestamps
	expectedCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	expectedLastUpdated, _ := time.Parse(time.RFC3339, "2023-01-02T12:00:00Z")

	assert.Equal(t, expectedCreatedAt, metadata.CreatedAt)
	assert.Equal(t, expectedLastUpdated, metadata.LastUpdated)

	// Verify raw metadata
	assert.Equal(t, "docs/user-stories/example/sample.md", metadata.RawMetadata["file_path"])
	assert.Equal(t, "2023-01-01T12:00:00Z", metadata.RawMetadata["created_at"])
	assert.Equal(t, "2023-01-02T12:00:00Z", metadata.RawMetadata["last_updated"])
	assert.Equal(t, "abcdef1234567890", metadata.RawMetadata["_content_hash"])
}

func TestGetContentWithoutMetadata(t *testing.T) {
	content := `---
file_path: docs/user-stories/example/sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: abcdef1234567890
---

# Sample User Story

This is a sample user story.
`

	expected := `# Sample User Story

This is a sample user story.
`

	result := GetContentWithoutMetadata(content)
	assert.Equal(t, expected, result)
}

func TestCalculateContentHash(t *testing.T) {
	content := "# Sample User Story\n\nThis is a sample user story.\n"
	
	hash := CalculateContentHash(content)
	
	// The expected hash is the SHA-256 hash of the content
	expectedHash := "c24a2f89c682fea773be9292bada1e861b2f139fb38e35ada3f78f1b87e7c6f1"
	
	assert.Equal(t, expectedHash, hash)
}

func setupMockFileSystem() *io.MockFileSystem {
	fs := io.NewMockFileSystem()
	
	// Set up user stories directory
	userStoriesDir := "docs/user-stories"
	fs.AddDirectory(userStoriesDir)
	
	// Add a few user story files
	fs.AddFile("docs/user-stories/sample.md", []byte(`---
file_path: docs/user-stories/sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: abcdef1234567890
---

# Sample User Story

This is a sample user story.
`))

	fs.AddFile("docs/user-stories/another.md", []byte(`---
file_path: docs/user-stories/another.md
created_at: 2023-01-03T12:00:00Z
last_updated: 2023-01-04T12:00:00Z
_content_hash: 0987654321fedcba
---

# Another User Story

This is another user story.
`))

	// Set up change requests directory
	changeRequestsDir := "docs/changes-request"
	fs.AddDirectory(changeRequestsDir)
	
	// Add a change request file
	fs.AddFile("docs/changes-request/sample.blueprint.md", []byte(`---
name: Sample Change Request
created-at: 2023-01-05T12:00:00Z
user-stories:
  - title: Sample User Story
    file: docs/user-stories/sample.md
    content-hash: abcdef1234567890
  - title: Another User Story
    file: docs/user-stories/another.md
    content-hash: 0987654321fedcba
---

# Blueprint

This is a sample change request.
`))

	return fs
}

// Additional tests will be added for:
// - UpdateFileMetadata
// - FindMarkdownFiles
// - UpdateAllUserStoryMetadata
// - FindChangeRequestFiles
// - UpdateChangeRequestReferences
// - FilterChangedContent
// - UpdateAllChangeRequestReferences 