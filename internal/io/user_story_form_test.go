// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package io

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestUserStoryFormEmptyFields(t *testing.T) {
	// Create a new user story with some metadata
	us := models.UserStory{
		FilePath:    "test.md",
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
	form := NewUserStoryForm(us)
	
	// Simulate tabbing through all fields without entering any values
	form.nextField() // Title -> Description
	form.nextField() // Description -> As a
	form.nextField() // As a -> I want
	form.nextField() // I want -> So that
	form.nextField() // So that -> AC1
	form.nextField() // AC1 -> AC2
	form.nextField() // AC2 -> AC3
	form.nextField() // AC3 -> AC4
	form.nextField() // AC4 -> AC5
	form.nextField() // AC5 -> Auto-submit
	
	// Get the user story
	savedUS := form.GetUserStory()
	
	// Verify that no content was created
	assert.Equal(t, "", savedUS.Title)
	assert.Equal(t, "", form.descInput.Value())
	assert.Equal(t, "", form.asInput.Value())
	assert.Equal(t, "", form.wantInput.Value())
	assert.Equal(t, "", form.soThatInput.Value())
	
	// Verify that no acceptance criteria were created
	for _, ac := range form.acInputs {
		assert.Equal(t, "", ac.Value())
	}
	
	// Verify that the content only contains metadata and empty sections
	expectedContent := "---\n" +
		"file_path: test.md\n" +
		"created_at: " + us.CreatedAt.Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"last_updated: " + us.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"_content_hash: d41d8cd98f00b204e9800998ecf8427e\n" +
		"---\n\n" +
		"# \n" +
		"As a \n" +
		"I want \n" +
		"so that \n\n" +
		"## Acceptance criteria\n"
	
	assert.Equal(t, expectedContent, savedUS.Content)
}

func TestUserStoryFormNoCreationWhenEmpty(t *testing.T) {
	// Create a new empty user story with some metadata
	us := models.UserStory{
		FilePath:    "test.md",
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
	form := NewUserStoryForm(us)
	
	// Verify initial state
	assert.False(t, form.hasContent(), "Form should report no content when empty")
	assert.False(t, form.ConfirmSubmission, "Form should not be confirmed when empty")
	
	// Try to get the user story without any interaction
	savedUS := form.GetUserStory()
	
	// Verify that no content was created
	assert.Equal(t, "", savedUS.Title)
	assert.Equal(t, "", form.descInput.Value())
	assert.Equal(t, "", form.asInput.Value())
	assert.Equal(t, "", form.wantInput.Value())
	assert.Equal(t, "", form.soThatInput.Value())
	
	// Verify that no acceptance criteria were created
	for _, ac := range form.acInputs {
		assert.Equal(t, "", ac.Value())
	}
	
	// Verify that the content only contains metadata and empty sections
	expectedContent := "---\n" +
		"file_path: test.md\n" +
		"created_at: " + us.CreatedAt.Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"last_updated: " + us.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + "\n" +
		"_content_hash: d41d8cd98f00b204e9800998ecf8427e\n" +
		"---\n\n" +
		"# \n" +
		"As a \n" +
		"I want \n" +
		"so that \n\n" +
		"## Acceptance criteria\n"
	
	assert.Equal(t, expectedContent, savedUS.Content)
	
	// Verify that pressing Ctrl+C doesn't set the cancel flag when empty
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	model, _ := form.Update(msg)
	updatedForm := model.(*UserStoryForm)
	assert.False(t, updatedForm.cancel, "Cancel flag should not be set when form is empty")
}

func TestUserStoryFormMetadata(t *testing.T) {
	// Create a new user story with metadata
	us := models.UserStory{
		FilePath:    "test.md",
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
	form := NewUserStoryForm(us)
	
	// Set some content
	form.titleInput.SetValue("Test Title")
	form.descInput.SetValue("Test Description")
	form.asInput.SetValue("user")
	form.wantInput.SetValue("to test")
	form.soThatInput.SetValue("it works")
	form.acInputs[0].SetValue("First criteria")
	
	// Set file path
	form.SetFilePath("docs/user-stories/test.md")
	
	// Get the user story
	savedUS := form.GetUserStory()
	
	// Verify metadata
	lines := strings.Split(savedUS.Content, "\n")
	
	// Check metadata section
	assert.Equal(t, "---", lines[0])
	assert.Equal(t, "file_path: docs/user-stories/test.md", lines[1])
	assert.Contains(t, lines[2], "created_at: ")
	assert.Contains(t, lines[3], "last_updated: ")
	assert.Contains(t, lines[4], "_content_hash: ")
	assert.Equal(t, "---", lines[5])
	
	// Verify content hash is correct
	contentHash := strings.TrimPrefix(lines[4], "_content_hash: ")
	contentWithoutMetadata := "# Test Title\n" +
		"Test Description\n\n" +
		"As a user\n" +
		"I want to test\n" +
		"so that it works\n\n" +
		"## Acceptance criteria\n" +
		"- First criteria\n"
	expectedHash := models.GenerateContentHash(contentWithoutMetadata)
	assert.Equal(t, expectedHash, contentHash)
} 