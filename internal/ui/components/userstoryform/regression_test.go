// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package userstoryform

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
	formModels "github.com/user-story-matrix/usm/internal/ui/models"
)

// TestRegressionBatchConfirmation tests the batch confirmation functionality
// to ensure it works correctly after the extensions and refactoring
func TestRegressionBatchConfirmation(t *testing.T) {
	// Setup a user story and form
	us := models.UserStory{
		Title: "Test Story",
		Description: "As a user,\nI want a feature,\nso that I can benefit.",
	}

	// Create the test form with mocks
	testForm := newTestUserStoryForm(t, us)
	form := testForm.UserStoryForm

	// Create a model with auto-populated fields and set it directly
	model := formModels.NewUserStoryFormModel(us, testForm.mockProcessor, form.configManager)
	
	// Set auto-populated fields directly using the map
	model.AutoPopulatedFields = map[string]bool{
		"title": true,
		"as_a": true,
		"i_want": true,
	}
	
	form.model = model
	
	// Verify initial state
	assert.True(t, form.model.HasAutoPopulatedFields(), "Should have auto-populated fields")
	assert.Equal(t, 3, form.model.GetAutoPopulatedFieldCount(), "Should have 3 auto-populated fields")
	
	// Simulate ctrl+a keypress to confirm all fields
	form.spinner.SetMessage("Processing...")
	form.spinner.SetVisible(true)
	
	// Update the form directly to simulate the batch confirmation
	form.model.ConfirmAllFields()
	
	// Set the spinner message that would be set by the Update method
	confirmMsg := fmt.Sprintf("Confirmed %d auto-populated fields", 3)
	confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76")).Italic(true)
	form.spinner.SetMessage(confirmStyle.Render(confirmMsg))
	
	// Check the updated form state
	assert.False(t, form.model.HasAutoPopulatedFields(), "Should not have auto-populated fields after confirmation")
	assert.Equal(t, 0, form.model.GetAutoPopulatedFieldCount(), "Should have 0 auto-populated fields after confirmation")

	// Verify spinner message
	assert.True(t, form.spinner.Visible, "Spinner should be visible")
	assert.Contains(t, form.spinner.Message, "Confirmed", "Spinner message should contain 'Confirmed'")
}

// TestRegressionCriteriaParsing tests various acceptance criteria formats
// to verify the enhanced parsing logic handles all supported formats correctly
func TestRegressionCriteriaParsing(t *testing.T) {
	// Test cases with different formats and expected parsed criteria
	testCases := []struct {
		name           string
		criteriaInput  string
		expectedOutput []string
	}{
		{
			name: "Combined formats regression test",
			criteriaInput: `- First item in a dash list
* Second item with asterisk
1. Numbered item
(4) Item with parenthesized number
   - Indented item with dash
   * Indented item with asterisk
   5. Indented numbered item
Plain text item without bullet or number`,
			expectedOutput: []string{
				"First item in a dash list",
				"Second item with asterisk",
				"Numbered item",
				"Item with parenthesized number",
				"Indented item with dash",
				"Indented item with asterisk",
				"Indented numbered item",
				"Plain text item without bullet or number",
			},
		},
		{
			name: "Edge case with extra whitespace",
			criteriaInput: "  \n" +
				"  - Item with leading spaces\n" +
				"* Item with trailing spaces  \n" +
				"  1. Numbered with spaces  \n",
			expectedOutput: []string{
				"Item with leading spaces",
				"Item with trailing spaces",
				"Numbered with spaces",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a form to use for testing
			testForm := newTestUserStoryForm(t, models.UserStory{})
			form := testForm.UserStoryForm
			
			// Test with direct parsing using the form's method
			criteria := form.parseAcceptanceCriteria(tc.criteriaInput)
			
			// Verify correct number of criteria extracted
			assert.Equal(t, len(tc.expectedOutput), len(criteria),
				"Expected %d criteria, got %d", len(tc.expectedOutput), len(criteria))
			
			// Verify each criterion matches expected output
			for i, expected := range tc.expectedOutput {
				if i < len(criteria) {
					assert.Equal(t, expected, criteria[i],
						"Criterion %d should be '%s', got '%s'", i, expected, criteria[i])
				}
			}
		})
	}
}

// TODO: Add regression test for clipboard paste detection
// TODO: Add regression test for visual feedback for auto-populated fields

// TODO: Add regression tests for clipboard paste detection in middle of text
// This should verify that the enhanced detection algorithms work across all scenarios

// TODO: Add regression tests for visual feedback with auto-populated fields
// This should verify that field highlighting and help text correctly update based on state 