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

// MockTerminalIO implements the io.UserOutput interface for testing
type MockTerminalIO struct {
	Messages      []string
	SuccessOutput []string
	ErrorOutput   []string
}

func (m *MockTerminalIO) Print(message string) {
	m.Messages = append(m.Messages, message)
}

func (m *MockTerminalIO) PrintSuccess(message string) {
	m.SuccessOutput = append(m.SuccessOutput, message)
}

func (m *MockTerminalIO) PrintError(message string) {
	m.ErrorOutput = append(m.ErrorOutput, message)
}

func (m *MockTerminalIO) PrintTable(headers []string, rows [][]string) {}
func (m *MockTerminalIO) PrintWarning(message string) {}
func (m *MockTerminalIO) PrintProgress(message string) {}
func (m *MockTerminalIO) PrintStep(stepNumber int, totalSteps int, description string) {}
func (m *MockTerminalIO) IsDebugEnabled() bool { return false }

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
	
	// Now we can test processFormResult directly with our mock
	mockTerminal := &MockTerminalIO{}
	userStory, shouldContinue := processFormResult(mockResult, mockTerminal)
	
	// Verify the return values
	assert.True(t, shouldContinue)
	assert.Equal(t, "Test Story", userStory.Title)
	assert.Equal(t, "As a tester,\nI want to verify interfaces,\nso that components work correctly.", userStory.Description)
	assert.Equal(t, []string{"Interface works", "No runtime errors"}, userStory.Criteria)
	assert.Empty(t, mockTerminal.Messages, "No messages should be printed when confirmation is true")
	
	// Test case when confirmation is false
	mockCancelResult := &MockFormResult{
		Story: models.UserStory{
			Title: "Cancelled Story",
		},
		ConfirmationStatus: false,
	}
	
	mockTerminal = &MockTerminalIO{}
	_, shouldContinue = processFormResult(mockCancelResult, mockTerminal)
	
	// Verify the return values for cancelled submission
	assert.False(t, shouldContinue)
	assert.Len(t, mockTerminal.Messages, 1, "Should have one message when cancelled")
	assert.Equal(t, "User story empty, creation cancelled", mockTerminal.Messages[0])
} 