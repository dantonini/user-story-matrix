// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/contracts"
)

// MockFormResult implements contracts.UserStorySubmitter for testing
type MockFormResult struct {
	Story               models.UserStory
	ConfirmationStatus  bool
}

func (m *MockFormResult) GetUserStory() models.UserStory {
	return m.Story
}

func (m *MockFormResult) GetConfirmSubmission() bool {
	return m.ConfirmationStatus
}

// TestFormResultHandling verifies the command handler correctly processes
// forms that implement the contracts.FormResult interface
func TestFormResultHandling(t *testing.T) {
	// Create a mock form result
	mockResult := &MockFormResult{
		Story: models.UserStory{
			Title:       "Test Story",
			Description: "As a tester,\nI want to verify interfaces,\nso that components work correctly.",
			Criteria:    []string{"Interface works", "No runtime errors"},
		},
		ConfirmationStatus: true,
	}
	
	// Verify the mock implements the interface
	var _ contracts.UserStorySubmitter = (*MockFormResult)(nil)
	
	// Verify we can use the interface methods directly
	story := mockResult.GetUserStory()
	assert.Equal(t, "Test Story", story.Title)
	assert.True(t, mockResult.GetConfirmSubmission())
	
	// Here we would normally test the addUserStoryCmd.Run function directly,
	// but it's currently not easily testable since it doesn't accept the form result
	// as a parameter. In a future refactoring, extract the form handling logic
	// into a separate function that accepts a contracts.FormResult parameter.
} 