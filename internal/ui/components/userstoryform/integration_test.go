// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/contracts"
)

// MockCommandHandler simulates the command handler that would process the form result
type MockCommandHandler struct {
	ReceivedUserStory  models.UserStory
	ConfirmationStatus bool
	Called             bool
}

// ProcessFormResult simulates the command handler's processing of the form result
func (m *MockCommandHandler) ProcessFormResult(result contracts.UserStorySubmitter) error {
	m.Called = true
	m.ReceivedUserStory = result.GetUserStory()
	m.ConfirmationStatus = result.GetConfirmSubmission()
	return nil
}

// TestUserStoryFormEndToEndFlow verifies the complete flow from filling the form to submission
// and handling by a command handler
func TestUserStoryFormEndToEndFlow(t *testing.T) {
	// Create a sample user story
	us := models.UserStory{
		Title: "Initial Title",
	}

	// Create the test form
	testForm := newTestUserStoryForm(t, us)
	form := testForm.UserStoryForm

	// Fill in form fields
	form.inputs[FieldIndex[TitleField]].SetValue("My Test Story")
	form.inputs[FieldIndex[AsAField]].SetValue("developer")
	form.inputs[FieldIndex[IWantField]].SetValue("a robust testing framework")
	form.inputs[FieldIndex[SoThatField]].SetValue("I can ensure code quality")

	// Fill in acceptance criteria
	form.criteriasInputs[0].SetValue("Tests should run quickly")
	form.criteriasInputs[1].SetValue("Tests should be reliable")

	// Simulate pressing Enter on last criteria field which triggers submission
	form.inCriteriaSection = true
	form.focusedCriteria = len(form.criteriasInputs) - 1
	cmd := form.handleEnterKey()

	// Verify cmd is tea.Quit to indicate the form is done
	assert.NotNil(t, cmd, "Command should not be nil")

	// Verify submission state
	assert.True(t, form.submitted, "Form should be marked as submitted")
	assert.True(t, form.GetConfirmSubmission(), "GetConfirmSubmission should return true")

	// Create a mock command handler
	mockHandler := &MockCommandHandler{}

	// Process the form result with the mock handler
	err := mockHandler.ProcessFormResult(form)
	assert.NoError(t, err)

	// Verify the command handler received and processed the form correctly
	assert.True(t, mockHandler.Called, "Command handler should have been called")
	assert.True(t, mockHandler.ConfirmationStatus, "Command handler should have received confirmation")
	assert.Equal(t, "My Test Story", mockHandler.ReceivedUserStory.Title, "Command handler should have received the correct title")

	// Verify the complete story
	story := mockHandler.ReceivedUserStory
	assert.Equal(t, "My Test Story", story.Title)
	assert.Contains(t, story.Description, "As a developer,")
	assert.Contains(t, story.Description, "I want a robust testing framework,")
	assert.Contains(t, story.Description, "so that I can ensure code quality.")
	assert.Len(t, story.Criteria, 2)
	assert.Contains(t, story.Criteria, "Tests should run quickly")
	assert.Contains(t, story.Criteria, "Tests should be reliable")
}

// TestEscapeCancellation verifies that pressing Escape cancels the form
// and the result can be correctly interpreted by a command handler
func TestEscapeCancellation(t *testing.T) {
	// Create the test form
	us := models.UserStory{}
	testForm := newTestUserStoryForm(t, us)
	form := testForm.UserStoryForm

	// Fill some data
	form.inputs[FieldIndex[TitleField]].SetValue("Story that will be canceled")

	// Simulate pressing Escape
	cmd := form.handleEscapeKey()

	// Verify cmd is tea.Quit to indicate the form is done
	assert.NotNil(t, cmd, "Command should not be nil")

	// Verify submission state - should NOT be submitted
	assert.False(t, form.submitted, "Form should not be marked as submitted after ESC")
	assert.False(t, form.GetConfirmSubmission(), "GetConfirmSubmission should return false after ESC")

	// Create a mock command handler
	mockHandler := &MockCommandHandler{}

	// Process the form result with the mock handler
	err := mockHandler.ProcessFormResult(form)
	assert.NoError(t, err)

	// Verify the command handler correctly identified the cancellation
	assert.True(t, mockHandler.Called, "Command handler should have been called")
	assert.False(t, mockHandler.ConfirmationStatus, "Command handler should have received non-confirmation")
} 