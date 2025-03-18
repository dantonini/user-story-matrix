package changerequest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestFindIncomplete_NoChangeRequests(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Setup the mock to simulate an empty directory
	mockFS.Dirs["docs/changes-request"] = true
	
	// Call the function being tested
	result, err := FindIncomplete(mockFS)
	
	// Verify the results
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFindIncomplete_DirectoryNotFound(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Directory not found is simulated by not adding it to the Dirs map
	
	// Call the function being tested
	result, err := FindIncomplete(mockFS)
	
	// Verify the results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "change requests directory not found")
	assert.Empty(t, result)
}

func TestFindIncomplete_WithIncompleteChangeRequests(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Sample file content for a change request
	fileContent := `---
name: Test Change Request
created-at: 2023-01-01T00:00:00Z
user-stories:
  - title: Test User Story
    file: docs/user-stories/test.md
    content-hash: abc123
---

# Blueprint
Test content
`
	
	// Setup the mock
	mockFS.Dirs["docs/changes-request"] = true
	
	// Add blueprint files
	mockFS.Files["docs/changes-request/test-cr.blueprint.md"] = []byte(fileContent)
	mockFS.Files["docs/changes-request/complete-cr.blueprint.md"] = []byte(fileContent)
	
	// Add implementation file for the complete CR only
	mockFS.Files["docs/changes-request/complete-cr.implementation.md"] = []byte("implementation content")
	
	// Call the function being tested
	result, err := FindIncomplete(mockFS)
	
	// Verify the results
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Test Change Request", result[0].Name)
}

func TestFormatDescription(t *testing.T) {
	// Create a test change request
	cr := models.ChangeRequest{
		Name:      "Test Change Request",
		CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		UserStories: []models.UserStoryReference{
			{Title: "Story 1"},
			{Title: "Story 2"},
		},
	}
	
	// Call the function being tested
	description := FormatDescription(cr)
	
	// Verify the result
	expected := "Test Change Request (Created: 2023-01-01 12:00:00, Stories: 2)"
	assert.Equal(t, expected, description)
} 