// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/changerequest"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestFindIncompleteChangeRequests_NoChangeRequests(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Setup the mock to simulate an empty directory
	mockFS.AddDirectory("docs/changes-request")
	
	// Call the function being tested
	result, err := changerequest.FindIncomplete(mockFS)
	
	// Verify the results
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFindIncompleteChangeRequests_DirectoryNotFound(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	
	// Directory not found is simulated by not adding it to the filesystem
	
	// Call the function being tested
	result, err := changerequest.FindIncomplete(mockFS)
	
	// Verify the results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "change requests directory not found")
	assert.Empty(t, result)
}

func TestFindIncompleteChangeRequests_ReadDirError(t *testing.T) {
	// Skip this test as we can't easily simulate a ReadDir error with our current mock
	// In a real implementation, we would enhance the mock to support this case
	t.Skip("Skipping ReadDir error test as it's not easily simulated with current mock")
}

func TestFindIncompleteChangeRequests_WithIncompleteChangeRequests(t *testing.T) {
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
	mockFS.AddDirectory("docs/changes-request")
	
	// Add blueprint files
	mockFS.WriteFile("docs/changes-request/test-cr.blueprint.md", []byte(fileContent), 0644)
	mockFS.WriteFile("docs/changes-request/complete-cr.blueprint.md", []byte(fileContent), 0644)
	
	// Add implementation file for the complete CR only
	mockFS.WriteFile("docs/changes-request/complete-cr.implementation.md", []byte("implementation content"), 0644)
	
	// Call the function being tested
	result, err := changerequest.FindIncomplete(mockFS)
	
	// Verify the results
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Test Change Request", result[0].Name)
}

func TestFormatChangeRequestDescription(t *testing.T) {
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
	description := changerequest.FormatDescription(cr)
	
	// Verify the result
	expected := "Test Change Request (Created: 2023-01-01 12:00:00, Stories: 2)"
	assert.Equal(t, expected, description)
}

func TestDisplayRecapMessage(t *testing.T) {
	// Create a mock terminal
	mockTerminal := io.NewMockIO()
	
	// Create a test change request
	cr := models.ChangeRequest{
		Name:     "Test Change Request",
		FilePath: "docs/changes-request/2023-01-01-test-change-request.blueprint.md",
	}
	
	// Call the function being tested
	displayRecapMessage(mockTerminal, cr)
	
	// Verify the output was captured
	assert.Len(t, mockTerminal.Messages, 1)
	assert.Contains(t, mockTerminal.Messages[0], "Recap what you did in a file in docs/changes-request/2023-01-01-test-change-request.implementation.md")
}

func TestDisplayCongratulationMessage(t *testing.T) {
	// Create a mock terminal
	mockTerminal := io.NewMockIO()
	
	// Call the function being tested
	displayCongratulationMessage(mockTerminal)
	
	// Verify the output was captured
	assert.Len(t, mockTerminal.SuccessMessages, 1)
	assert.Contains(t, mockTerminal.SuccessMessages[0], "Congratulations")
} 